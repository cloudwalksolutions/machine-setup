package forms

import (
	"os"

	"github.com/charmbracelet/huh"

	"tars/internal/config"
)

// ProjectAnswers fills the per-project AGENTS.md template; Agents picks which
// tools get a pointer file (claude, gemini; pi reads AGENTS.md natively).
type ProjectAnswers struct {
	Agents      []string
	Name        string
	Description string
	TestCommand string
}

// ShowClaudeInitForm asks which Claude Code pieces to provision: the
// edit-blocking hook, the settings fragment, and which global rules to render.
// When TARS_NO_FORM=1 everything is on and every rule is selected.
func ShowClaudeInitForm(rules []string) (config.ClaudeConfig, error) {
	hook, settings := true, true
	selected := make([]string, len(rules))
	copy(selected, rules)
	if os.Getenv("TARS_NO_FORM") != "" {
		return config.ClaudeConfig{Hook: &hook, Settings: &settings, Rules: selected}, nil
	}

	options := make([]huh.Option[string], len(rules))
	for i, r := range rules {
		options[i] = huh.NewOption(r, r).Selected(true)
	}
	err := run(huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Install the edit-blocking hook?").
				Description("Denies sed -i / heredoc / interpreter writes so every change is a reviewable Edit or Write.").
				Value(&hook),
			huh.NewConfirm().
				Title("Merge model, theme, plugins and the hook entry into ~/.claude/settings.json?").
				Description("Machine-local keys such as autoMode and permissions are kept as they are.").
				Value(&settings),
			huh.NewMultiSelect[string]().
				Title("Global rules to render into ~/.claude/CLAUDE.md").
				Description("All are pre-selected. Space to toggle, Enter to confirm.").
				Options(options...).
				Value(&selected),
		),
	))
	return config.ClaudeConfig{Hook: &hook, Settings: &settings, Rules: selected}, err
}

// projectAgents are the tools `tars init project` can wire to AGENTS.md.
var projectAgents = []string{"claude", "pi", "gemini"}

// ShowProjectForm asks which agents the project targets and the values the
// AGENTS.md template needs. When TARS_NO_FORM=1 it picks every agent and placeholders.
func ShowProjectForm(defaultName string) (ProjectAnswers, error) {
	a := ProjectAnswers{
		Agents:      append([]string{}, projectAgents...),
		Name:        defaultName,
		Description: "TODO: one paragraph on what this project is.",
		TestCommand: "make test",
	}
	if os.Getenv("TARS_NO_FORM") != "" {
		return a, nil
	}
	options := make([]huh.Option[string], len(projectAgents))
	for i, agent := range projectAgents {
		options[i] = huh.NewOption(agent, agent).Selected(true)
	}
	err := run(huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Agents used in this project").
				Description("AGENTS.md is written once; claude and gemini get a pointer file that imports it, pi reads it directly.").
				Options(options...).
				Value(&a.Agents),
			huh.NewInput().Title("Project name").Value(&a.Name),
			huh.NewInput().Title("One-line description").Value(&a.Description),
			huh.NewInput().Title("Canonical test command").Value(&a.TestCommand),
		),
	))
	return a, err
}
