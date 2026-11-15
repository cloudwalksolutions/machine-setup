package tui

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"tars/internal/components"
	"tars/internal/forms"
	"tars/internal/report"
)

// Run drives an init flow on the terminal: headless (TARS_NO_FORM) prints plain
// text and answers with defaults; otherwise the flow runs under one tea.Program.
// This is the only piece that needs a real terminal; Start is proven with teatest.
func Run(headless bool, stdout, stderr io.Writer, flow Flow) error {
	if headless {
		return flow(report.Text{Stdout: stdout, Stderr: stderr}, forms.Headless{})
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	model := New()
	program := tea.NewProgram(model, tea.WithOutput(stdout), tea.WithContext(ctx))
	Start(ctx, program.Send, model, flow)
	final, err := program.Run()
	if err != nil {
		return err
	}
	return final.(Model).Err()
}

// Reviewer runs the pull review on the terminal; it is cmd's Reviewer.
type Reviewer struct{}

// Review shows the plans and returns the components the user accepted, or
// huh.ErrUserAborted when they quit. Like Run, it needs a real terminal;
// Review itself is proven with teatest.
func (Reviewer) Review(plans []components.Plan) ([]string, error) {
	final, err := tea.NewProgram(NewReview(plans)).Run()
	if err != nil {
		return nil, err
	}
	review := final.(Review)
	if review.Aborted() {
		return nil, huh.ErrUserAborted
	}
	return review.Accepted(), nil
}

// RunSessions runs the sessions manager on the terminal and returns what the user chose to open.
// Like Run, it is the only piece that needs a real terminal; Sessions is proven with teatest.
func RunSessions(m Manager) (string, error) {
	final, err := tea.NewProgram(NewSessions(m)).Run()
	if err != nil {
		return "", err
	}
	return final.(Sessions).Chosen(), nil
}
