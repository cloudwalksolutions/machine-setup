package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/pkg"
)

// Prompts answers the orchestrators' questions by running each form inside the
// program: it sends the form and blocks the orchestrator goroutine until the
// user completes or aborts it, or the program goes away.
type Prompts struct {
	Send func(tea.Msg)
	Ctx  context.Context
}

func (p Prompts) ask(f *huh.Form) error {
	done := make(chan error, 1)
	p.Send(PromptMsg{Form: f, Done: done})
	select {
	case err := <-done:
		return err
	case <-p.Ctx.Done():
		return huh.ErrUserAborted
	}
}

// Welcome shows the opening screen.
func (p Prompts) Welcome() error { return p.ask(forms.WelcomeForm()) }

// PickTools runs the tool picker.
func (p Prompts) PickTools(tools []pkg.ToolInfo) ([]string, error) {
	f, collect := forms.InstallForm(tools)
	if err := p.ask(f); err != nil {
		return nil, err
	}
	return collect(), nil
}

// PickWizards runs the agent-setup picker.
func (p Prompts) PickWizards(offered []string) ([]string, error) {
	f, collect := forms.WizardForm(offered)
	if err := p.ask(f); err != nil {
		return nil, err
	}
	return collect(), nil
}

// AskClaude runs the claude questions.
func (p Prompts) AskClaude(rules []string) (config.ClaudeConfig, error) {
	f, collect := forms.ClaudeForm(rules)
	if err := p.ask(f); err != nil {
		return config.ClaudeConfig{}, err
	}
	return collect(), nil
}

// AskPi runs the pi questions, then the default-model choice when models are exposed.
func (p Prompts) AskPi(in forms.PiInputs) (config.PiConfig, error) {
	f, collect := forms.PiForm(in)
	if err := p.ask(f); err != nil {
		return config.PiConfig{}, err
	}
	cfg := collect()
	if len(cfg.OllamaModels) == 0 {
		return cfg, nil
	}
	choice, collectDefault := forms.PiDefaultForm(cfg.OllamaModels, in.Previous.DefaultModel)
	if err := p.ask(choice); err != nil {
		return config.PiConfig{}, err
	}
	cfg.DefaultModel = collectDefault()
	return cfg, nil
}

// AskProject runs the project questions.
func (p Prompts) AskProject(defaultName string) (forms.ProjectAnswers, error) {
	f, collect := forms.ProjectForm(defaultName)
	if err := p.ask(f); err != nil {
		return forms.ProjectAnswers{}, err
	}
	return collect(), nil
}
