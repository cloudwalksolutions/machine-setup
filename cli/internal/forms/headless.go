package forms

import (
	"os"

	"tars/internal/config"
	"tars/internal/pkg"
)

// headless is true when TARS_NO_FORM is set: every prompt answers with its defaults (tests/CI).
func headless() bool { return os.Getenv("TARS_NO_FORM") != "" }

// Headless answers every prompt with the defaults the interactive forms start from.
type Headless struct{}

// Welcome shows nothing.
func (Headless) Welcome() error { return nil }

// PickTools selects every tool that is not installed yet.
func (Headless) PickTools(tools []pkg.ToolInfo) ([]string, error) {
	_, collect := InstallForm(tools)
	return collect(), nil
}

// PickWizards runs every offered wizard.
func (Headless) PickWizards(offered []string) ([]string, error) {
	_, collect := WizardForm(offered)
	return collect(), nil
}

// AskClaude turns everything on and selects every rule.
func (Headless) AskClaude(rules []string) (config.ClaudeConfig, error) {
	_, collect := ClaudeForm(rules)
	return collect(), nil
}

// AskPi installs every package, exposes every ollama model and pins no default.
func (Headless) AskPi(in PiInputs) (config.PiConfig, error) {
	in.Previous = config.PiConfig{}
	_, collect := PiForm(in)
	return collect(), nil
}

// AskProject targets every agent with placeholder answers.
func (Headless) AskProject(defaultName string) (ProjectAnswers, error) {
	_, collect := ProjectForm(defaultName)
	return collect(), nil
}
