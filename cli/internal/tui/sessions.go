package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"tars/internal/forms"
	"tars/internal/projects"
)

// Row is one entry of the sessions manager: a running session, or a project without one.
type Row struct {
	Name     string
	Live     bool
	Windows  int
	Attached bool
	Dir      string // the project's dir; empty for a session that is no project
}

// Title is the session or project name.
func (r Row) Title() string { return r.Name }

// Description summarizes the session, or says the project is not running.
func (r Row) Description() string {
	if !r.Live {
		return "not running · " + r.Dir
	}
	desc := fmt.Sprintf("%d windows", r.Windows)
	if r.Attached {
		desc += " · attached"
	}
	return desc
}

// FilterValue matches rows by name.
func (r Row) FilterValue() string { return r.Name }

// Manager is everything the sessions manager does outside the terminal.
type Manager interface {
	Rows() ([]Row, error)
	Kill(name string) error
	Rename(old, name string) error
	Template(project string) (projects.Template, error)
	SaveTemplate(project string, t projects.Template) error
}

type rowsMsg struct {
	rows []Row
	err  error
}

var (
	openKey     = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open"))
	killKey     = key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "kill"))
	renameKey   = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename"))
	templateKey = key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "edit"))
	quitKey     = key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
)

// Sessions is the interactive sessions manager: a filterable list of sessions and projects.
type Sessions struct {
	manager Manager
	list    list.Model
	size    *tea.WindowSizeMsg
	form    *formHost
	onDone  func() error // runs the form's answer once it completes
	status  string       // why the last action failed
	chosen  string
}

// NewSessions returns a manager that loads its rows on Init.
func NewSessions(m Manager) Sessions {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "byobu sessions"
	l.DisableQuitKeybindings() // its default quit key is "v"; q and ctrl+c are handled here
	l.KeyMap.CursorUp.SetHelp("↑", "up")
	l.KeyMap.CursorDown.SetHelp("↓", "down")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{openKey, killKey, renameKey, templateKey, quitKey}
	}
	return Sessions{manager: m, list: l}
}

// Chosen is the session or project to open once the program has quit; empty if none.
func (s Sessions) Chosen() string { return s.chosen }

// Init loads the rows.
func (s Sessions) Init() tea.Cmd { return s.load }

func (s Sessions) load() tea.Msg {
	rows, err := s.manager.Rows()
	return rowsMsg{rows: rows, err: err}
}

// Update routes sizes, rows and keys to the list, or to the open form.
func (s Sessions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s.form != nil {
		return s.updateForm(msg)
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.size = &msg
		s.list.SetSize(msg.Width, msg.Height-1) // the status line goes under the list
	case rowsMsg:
		if msg.err != nil {
			s.status = msg.err.Error()
			return s, nil
		}
		items := make([]list.Item, len(msg.rows))
		for i, r := range msg.rows {
			items[i] = r
		}
		return s, s.list.SetItems(items)
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return s, tea.Quit
		}
		if s.list.FilterState() == list.Filtering {
			break
		}
		if key.Matches(msg, quitKey) {
			return s, tea.Quit
		}
		if row, ok := s.list.SelectedItem().(Row); ok && key.Matches(msg, openKey, killKey, renameKey, templateKey) {
			return s.act(msg, row)
		}
	}
	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

// act runs the action bound to msg on row; actions that do not apply to the row do nothing.
func (s Sessions) act(msg tea.KeyPressMsg, row Row) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, openKey):
		s.chosen = row.Name
		return s, tea.Quit
	case key.Matches(msg, killKey) && row.Live:
		f, kill := forms.KillForm(row.Name)
		return s.ask(f, func() error {
			if !kill() {
				return nil
			}
			return s.manager.Kill(row.Name)
		})
	case key.Matches(msg, renameKey) && row.Live:
		f, name := forms.RenameForm(row.Name)
		return s.ask(f, func() error { return s.manager.Rename(row.Name, name()) })
	case key.Matches(msg, templateKey) && row.Dir != "":
		current, err := s.manager.Template(row.Name)
		if err != nil {
			s.status = err.Error()
			return s, nil
		}
		f, template := forms.TemplateForm(row.Name, current)
		return s.ask(f, func() error { return s.manager.SaveTemplate(row.Name, template()) })
	}
	return s, nil
}

// ask opens f over the list; done runs once the user completes it.
func (s Sessions) ask(f *huh.Form, done func() error) (tea.Model, tea.Cmd) {
	host, cmd := openForm(f, s.size)
	s.form, s.onDone = &host, done
	return s, cmd
}

func (s Sessions) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	host, cmd, state := s.form.update(msg)
	switch state {
	case huh.StateCompleted:
		done := s.onDone
		s.form, s.onDone, s.status = nil, nil, ""
		if err := done(); err != nil {
			s.status = err.Error()
		}
		return s, tea.Batch(cmd, s.load)
	case huh.StateAborted:
		s.form, s.onDone = nil, nil
		return s, cmd
	}
	s.form = &host
	return s, cmd
}

// View renders the open form, or the list.
func (s Sessions) View() tea.View {
	if s.form != nil {
		return tea.NewView(s.form.view())
	}
	if s.status != "" {
		return tea.NewView(s.list.View() + "\n" + failedTxt.Render(s.status))
	}
	return tea.NewView(s.list.View())
}
