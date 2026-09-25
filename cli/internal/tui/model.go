// Package tui renders `tars init` as one Bubble Tea program: finished steps and
// items scroll into the terminal history, one live pane shows what is running,
// and forms take over the pane when the orchestrator asks a question.
package tui

import (
	"fmt"
	"strings"
	"sync/atomic"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"tars/internal/forms"
)

const tailLines = 5

var (
	accent    = lipgloss.Color("#7AA2F7")
	dim       = lipgloss.Color("240")
	red       = lipgloss.Color("#ED567A")
	titleTxt  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	dimTxt    = lipgloss.NewStyle().Foreground(dim)
	failedTxt = lipgloss.NewStyle().Foreground(red)
)

type phase int

const (
	working phase = iota
	inForm
	finished
)

// Model is the live pane: the step in progress, or the form being answered.
type Model struct {
	phase     phase
	title     string
	item      string
	done      int
	total     int
	tail      []string
	failures  int
	spinner   spinner.Model
	form      *huh.Form
	formDone  chan<- error
	cancelled *atomic.Bool
	err       error
}

// New returns an idle model.
func New() Model {
	return Model{
		spinner:   spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(accent))),
		cancelled: &atomic.Bool{},
	}
}

// Init starts the spinner.
func (m Model) Init() tea.Cmd { return m.spinner.Tick }

// Cancelled reports whether the user asked to stop; the reporter relays it to the orchestrator.
func (m Model) Cancelled() bool { return m.cancelled.Load() }

// Err is the orchestrator's result once the run has ended.
func (m Model) Err() error { return m.err }

// Update routes orchestrator messages, form input and interrupts.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.phase == inForm {
		return m.updateForm(msg)
	}
	switch msg := msg.(type) {
	case PromptMsg:
		m.phase, m.form, m.formDone = inForm, forms.Styled(msg.Form), msg.Done
		return m, m.form.Init()
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled.Store(true)
		}
	case tea.InterruptMsg:
		m.cancelled.Store(true)
	case RunDoneMsg:
		m.phase, m.err = finished, msg.Err
		return m, tea.Quit
	case StepStartedMsg:
		m.title, m.item, m.done, m.total, m.tail = msg.Title, "", 0, msg.Total, nil
	case ItemStartedMsg:
		m.item, m.tail = msg.Name, nil
	case ItemDoneMsg:
		m.done++
		m.item, m.tail = "", nil
		if msg.Err != nil {
			m.failures++
		}
	case StepDoneMsg:
		m.item, m.tail = "", nil
		if msg.Err != nil {
			m.failures++
		}
	case OutputLineMsg:
		m.tail = append(m.tail, msg.Line)
		if len(m.tail) > tailLines {
			m.tail = m.tail[len(m.tail)-tailLines:]
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// updateForm forwards input to the running form until it completes or aborts.
func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.form.Update(msg)
	m.form = next.(*huh.Form)
	switch m.form.State {
	case huh.StateCompleted:
		m.formDone <- nil
	case huh.StateAborted:
		m.formDone <- huh.ErrUserAborted
	default:
		return m, cmd
	}
	m.phase, m.form, m.formDone = working, nil, nil
	return m, cmd
}

// View renders the live pane: the running form, or the step in progress and the
// tail of its output.
func (m Model) View() tea.View {
	if m.phase == inForm {
		return tea.NewView(m.form.View())
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", m.spinner.View(), titleTxt.Render(m.title))
	if m.total > 0 {
		fmt.Fprintf(&b, " %s", dimTxt.Render(fmt.Sprintf("%d/%d", m.done, m.total)))
	}
	if m.item != "" {
		fmt.Fprintf(&b, "  %s", m.item)
	}
	if m.failures > 0 {
		fmt.Fprintf(&b, "  %s", failedTxt.Render(fmt.Sprintf("%d failed", m.failures)))
	}
	if m.cancelled.Load() {
		fmt.Fprintf(&b, "  %s", failedTxt.Render("stopping after the current item"))
	}
	for _, line := range m.tail {
		fmt.Fprintf(&b, "\n    %s", dimTxt.Render(line))
	}
	return tea.NewView(b.String())
}
