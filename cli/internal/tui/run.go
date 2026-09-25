package tui

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"

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
