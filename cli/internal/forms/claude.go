package forms

import (
	"charm.land/huh/v2"

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

// ClaudeForm builds the `tars init claude` questions: the edit-blocking hook, the
// settings fragment, and which global rules to render. Everything starts on.
func ClaudeForm(rules []string) (*huh.Form, func() config.ClaudeConfig) {
	hook, settings := true, true
	selected := append([]string{}, rules...)
	options := make([]huh.Option[string], len(rules))
	for i, r := range rules {
		options[i] = huh.NewOption(r, r).Selected(true)
	}
	f := huh.NewForm(
		huh.NewGroup(
			confirm("Install the edit-blocking hook?",
				"Denies sed -i / heredoc / interpreter writes so every change is a reviewable Edit or Write.",
				&hook),
			confirm("Merge settings into ~/.claude/settings.json?",
				"Model, theme, plugins and the hook entry. Machine-local keys such as autoMode and permissions are kept.",
				&settings),
			huh.NewMultiSelect[string]().
				Title("Global rules to render into ~/.claude/CLAUDE.md").
				Options(options...).
				Value(&selected),
		),
	)
	return f, func() config.ClaudeConfig {
		return config.ClaudeConfig{Hook: &hook, Settings: &settings, Rules: selected}
	}
}

// projectAgents are the tools `tars init project` can wire to AGENTS.md.
var projectAgents = []string{"claude", "pi", "gemini"}

// ProjectForm builds the `tars init project` questions: which agents the project
// targets and the values the AGENTS.md template needs, starting from placeholders.
func ProjectForm(defaultName string) (*huh.Form, func() ProjectAnswers) {
	a := ProjectAnswers{
		Agents:      append([]string{}, projectAgents...),
		Name:        defaultName,
		Description: "TODO: one paragraph on what this project is.",
		TestCommand: "make test",
	}
	options := make([]huh.Option[string], len(projectAgents))
	for i, agent := range projectAgents {
		options[i] = huh.NewOption(agent, agent).Selected(true)
	}
	f := huh.NewForm(
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
	)
	return f, func() ProjectAnswers { return a }
}
