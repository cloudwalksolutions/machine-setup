package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/aymanbagabas/go-udiff"

	"tars/internal/components"
	"tars/internal/fsutil"
)

var (
	moveKeys   = key.NewBinding(key.WithKeys("up", "down", "j", "k"), key.WithHelp("↑/↓", "move"))
	toggleKey  = key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle"))
	acceptAll  = key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "accept all"))
	skipAll    = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "skip all"))
	hunkKeys   = key.NewBinding(key.WithKeys("n", "N"), key.WithHelp("n/N", "hunk"))
	fileKeys   = key.NewBinding(key.WithKeys("]", "["), key.WithHelp("]/[", "file"))
	paneKey    = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "pane"))
	applyKey   = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "apply"))
	abortKey   = key.NewBinding(key.WithKeys("q", "esc"), key.WithHelp("q", "quit"))
	reviewKeys = []key.Binding{moveKeys, toggleKey, acceptAll, skipAll, hunkKeys, fileKeys, paneKey, applyKey, abortKey}
)

// reviewRow is one line of the left pane: a component (change -1) or one of its changes.
type reviewRow struct{ plan, change int }

// Review is the interactive pull review: components and their changes on the
// left, the selected change's colored diff on the right.
type Review struct {
	plans    []components.Plan
	accepted []bool
	rows     []reviewRow
	cursor   int
	size     *tea.WindowSizeMsg
	dark     bool
	diff     viewport.Model
	hunks    []int
	hunk     int
	diffHas  bool // the right pane has the keyboard
	help     help.Model
	done     bool
	aborted  bool
}

// NewReview starts with every component accepted and the first change selected.
func NewReview(plans []components.Plan) Review {
	r := Review{plans: plans, accepted: make([]bool, len(plans)), dark: true, help: help.New(), diff: viewport.New()}
	r.diff.SoftWrap = false
	for i, p := range plans {
		r.accepted[i] = true
		r.rows = append(r.rows, reviewRow{i, -1})
		for j := range p.Changes {
			r.rows = append(r.rows, reviewRow{i, j})
		}
	}
	return r
}

// Accepted is the components the user chose to apply; Aborted tells whether they quit instead.
func (r Review) Accepted() []string {
	var names []string
	for i, p := range r.plans {
		if r.accepted[i] {
			names = append(names, p.Component)
		}
	}
	return names
}

// Aborted is true when the review ended without applying.
func (r Review) Aborted() bool { return r.aborted }

// Init asks the terminal whether its background is dark.
func (r Review) Init() tea.Cmd { return tea.RequestBackgroundColor }

// Update sizes the panes, picks the palette and handles keys.
func (r Review) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.size = &msg
		r.layout()
	case tea.BackgroundColorMsg:
		r.dark = msg.IsDark()
		r.showSelected()
	case tea.MouseWheelMsg:
		r.diff, _ = r.diff.Update(msg)
	case selectRowMsg:
		r.cursor = int(msg)
		r.showSelected()
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || key.Matches(msg, abortKey) {
			r.aborted = true
			return r, tea.Quit
		}
		if key.Matches(msg, applyKey) {
			r.done = true
			return r, tea.Quit
		}
		if key.Matches(msg, paneKey) {
			r.diffHas = !r.diffHas
		}
		if key.Matches(msg, moveKeys) {
			step := 1
			if k := msg.String(); k == "up" || k == "k" {
				step = -1
			}
			switch {
			case r.diffHas && step > 0:
				r.diff.ScrollDown(1)
			case r.diffHas:
				r.diff.ScrollUp(1)
			default:
				r.move(step)
			}
		}
		if key.Matches(msg, toggleKey) {
			plan := r.rows[r.cursor].plan
			r.accepted[plan] = !r.accepted[plan]
		}
		if key.Matches(msg, acceptAll, skipAll) {
			for i := range r.accepted {
				r.accepted[i] = key.Matches(msg, acceptAll)
			}
		}
		if key.Matches(msg, hunkKeys) {
			r.jumpHunk(msg.String() == "n")
		}
		if key.Matches(msg, fileKeys) {
			r.jumpFile(msg.String() == "]")
		}
	}
	return r, nil
}

// jumpHunk scrolls the diff so the next or previous hunk header is the top row.
func (r *Review) jumpHunk(forward bool) {
	if len(r.hunks) == 0 {
		return
	}
	step := 1
	if !forward {
		step = -1
	}
	r.hunk = min(max(r.hunk+step, 0), len(r.hunks)-1)
	r.diff.SetYOffset(r.hunks[r.hunk])
}

// jumpFile selects the change after or before the one on show, skipping component rows.
func (r *Review) jumpFile(forward bool) {
	step, from := 1, r.cursor
	if !forward {
		step = -1
	}
	if r.rows[from].change < 0 && from+1 < len(r.rows) && r.rows[from+1].change >= 0 {
		from++
	}
	for i := from + step; i >= 0 && i < len(r.rows); i += step {
		if r.rows[i].change >= 0 {
			r.cursor = i
			r.showSelected()
			return
		}
	}
}

func (r *Review) move(step int) {
	r.cursor = min(max(r.cursor+step, 0), len(r.rows)-1)
	r.showSelected()
}

func (r *Review) layout() {
	leftW, rightW := r.paneWidths()
	r.help.SetWidth(r.size.Width)
	r.diff.SetWidth(rightW)
	r.diff.SetHeight(max(r.bodyHeight()-1, 1))
	_ = leftW
	r.showSelected()
}

func (r Review) paneWidths() (left, right int) {
	left = r.size.Width / 3
	return left, r.size.Width - left - 1
}

// bodyHeight is the height left for the panes under the title and above the footer.
func (r Review) bodyHeight() int { return r.size.Height - 3 }

func (r Review) selected() (components.Plan, *fsutil.Change) {
	row := r.rows[r.cursor]
	plan := r.plans[row.plan]
	if row.change >= 0 {
		return plan, &plan.Changes[row.change]
	}
	if len(plan.Changes) > 0 {
		return plan, &plan.Changes[0]
	}
	return plan, nil
}

func (r *Review) showSelected() {
	if r.size == nil {
		return
	}
	_, ch := r.selected()
	var rows []string
	r.hunks, r.hunk = nil, 0
	switch {
	case ch == nil:
	case len(ch.Diff.Hunks) > 0:
		rows, r.hunks = RenderDiff(ch.Diff, r.diff.Width(), r.dark)
	default:
		rows = []string{dimTxt.Render(describeChange(*ch))}
	}
	r.diff.SetContentLines(rows)
	r.diff.GotoTop()
}

func describeChange(ch fsutil.Change) string {
	switch {
	case ch.Kind == fsutil.Create:
		return "new file"
	case ch.Kind == fsutil.Remove:
		return "will be removed"
	case ch.Kind == fsutil.Set:
		return ch.Detail
	case ch.Binary:
		return "binary file, no diff"
	}
	return "no diff"
}

// View draws the title, the two panes and the footer; before the first size it draws nothing.
func (r Review) View() tea.View {
	if r.size == nil {
		return tea.NewView("")
	}
	v := tea.NewView(r.screen())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.OnMouse = r.onMouse
	return v
}

// selectRowMsg selects the left-pane row clicked on.
type selectRowMsg int

// onMouse turns a click on a left-pane row into its selection.
func (r Review) onMouse(msg tea.MouseMsg) tea.Cmd {
	click, ok := msg.(tea.MouseClickMsg)
	if !ok || click.Button != tea.MouseLeft {
		return nil
	}
	leftW, _ := r.paneWidths()
	row := click.Y - 1
	if click.X >= leftW || row < 0 || row >= len(r.rows) {
		return nil
	}
	return func() tea.Msg { return selectRowMsg(row) }
}

func (r Review) screen() string {
	if r.size.Width < 60 || r.size.Height < 10 {
		return failedTxt.Render("terminal too small for the review; resize or run tars pull -y")
	}
	leftW, rightW := r.paneWidths()
	body := r.bodyHeight()
	left := lipgloss.NewStyle().Width(leftW).Height(body).MaxHeight(body).Render(r.leftPane(leftW))
	right := lipgloss.NewStyle().Width(rightW).Height(body).MaxHeight(body).Render(r.rightPane(rightW))
	line := lipgloss.NewStyle().MaxWidth(r.size.Width)
	plan := fmt.Sprintf("%s · %d of %d accepted", components.PlanLine(r.plans), len(r.Accepted()), len(r.plans))
	return strings.Join([]string{
		titleTxt.Render("Review pull"),
		lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right),
		line.Render(dimTxt.Render(plan)),
		line.Render(r.help.ShortHelpView(reviewKeys)),
	}, "\n")
}

func (r Review) leftPane(width int) string {
	lines := make([]string, 0, len(r.rows))
	for i, row := range r.rows {
		text := r.rowText(row)
		if i == r.cursor {
			text = lipgloss.NewStyle().Bold(true).Background(dim).Width(width).Render(text)
		}
		lines = append(lines, text)
	}
	return strings.Join(lines, "\n")
}

func (r Review) rowText(row reviewRow) string {
	plan := r.plans[row.plan]
	if row.change < 0 {
		mark := lipgloss.NewStyle().Foreground(accent).Render("✓")
		if !r.accepted[row.plan] {
			mark = failedTxt.Render("✗")
		}
		return fmt.Sprintf("%s %s  %s", mark, titleTxt.Render(plan.Component), dimTxt.Render(plan.Summary()))
	}
	ch := plan.Changes[row.change]
	glyph, detail := changeGlyph(ch), changeDetail(ch)
	if detail != "" {
		detail = "  " + dimTxt.Render(detail)
	}
	return fmt.Sprintf("  %s %s%s", glyph, ch.Path, detail)
}

func changeGlyph(ch fsutil.Change) string {
	switch ch.Kind {
	case fsutil.Create:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")).Render("+")
	case fsutil.Remove:
		return failedTxt.Render("-")
	case fsutil.Set:
		return lipgloss.NewStyle().Foreground(accent).Render("⚙")
	}
	return lipgloss.NewStyle().Foreground(accent).Render("~")
}

func changeDetail(ch fsutil.Change) string {
	if ch.Binary {
		return "binary"
	}
	if ch.Kind == fsutil.Set {
		return ch.Detail
	}
	if adds, dels := countLines(ch.Diff); adds+dels > 0 {
		return fmt.Sprintf("+%d −%d", adds, dels)
	}
	return ""
}

func countLines(d udiff.UnifiedDiff) (adds, dels int) {
	for _, h := range d.Hunks {
		for _, l := range h.Lines {
			switch l.Kind {
			case udiff.Insert:
				adds++
			case udiff.Delete:
				dels++
			}
		}
	}
	return adds, dels
}

func (r Review) rightPane(width int) string {
	_, ch := r.selected()
	header := ""
	if ch != nil {
		header = fmt.Sprintf("%s · %s", ch.Path, [...]string{"create", "modify", "remove", "set"}[ch.Kind])
		if detail := changeDetail(*ch); detail != "" {
			header += " · " + detail
		}
		if len(r.hunks) > 0 {
			header += fmt.Sprintf(" · %d hunks", len(r.hunks))
		}
	}
	return titleTxt.Width(width).Render(header) + "\n" + r.diff.View()
}
