package tui

import (
	"context"
	"io"

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

// Run drives an init flow: headless (TARS_NO_FORM) prints plain text and answers
// with defaults; otherwise the flow runs in a goroutine while the program renders
// its progress and forms. The flow's error is the result.
func Run(headless bool, stdout, stderr io.Writer, flow func(report.Reporter, Asker) error) error {
	if headless {
		return flow(report.Text{Stdout: stdout, Stderr: stderr}, forms.Headless{})
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	model := New()
	program := tea.NewProgram(model, tea.WithOutput(stdout), tea.WithContext(ctx))
	go func() {
		println := func(line string) { program.Println(line) }
		err := flow(newReporter(program.Send, println, model.cancelled), Prompts{Send: program.Send, Ctx: ctx})
		program.Send(RunDoneMsg{Err: err})
	}()
	final, err := program.Run()
	if err != nil {
		return err
	}
	return final.(Model).Err()
}
