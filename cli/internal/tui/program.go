package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/pkg"
	"tars/internal/report"
)

// Asker is every question an init flow can ask; Prompts answers through the
// program, forms.Headless with defaults.
type Asker interface {
	Welcome() error
	PickTools(tools []pkg.ToolInfo) ([]string, error)
	PickWizards(offered []string) ([]string, error)
	AskClaude(rules []string) (config.ClaudeConfig, error)
	AskPi(in forms.PiInputs) (config.PiConfig, error)
	AskProject(defaultName string) (forms.ProjectAnswers, error)
}

// Flow is an init orchestration: it reports progress and asks questions, and
// its error is the command's result.
type Flow func(report.Reporter, Asker) error

// Start runs the flow in its own goroutine against the program that renders
// model, reached only through send: finished lines are printed above the pane,
// state and prompts become messages, and the flow's result ends the program.
func Start(ctx context.Context, send func(tea.Msg), model Model, flow Flow) {
	println := func(line string) { send(tea.Println(line)()) }
	go func() {
		err := flow(newReporter(send, println, model.cancelled), Prompts{Send: send, Ctx: ctx})
		send(RunDoneMsg{Err: err})
	}()
}
