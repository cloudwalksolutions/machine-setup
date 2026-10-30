package forms

import (
	"slices"

	"charm.land/huh/v2"

	"tars/internal/config"
)

// AgentsForm asks which shared rules to render; every rule starts selected.
func AgentsForm(rules []string) (*huh.Form, func() config.AgentsConfig) {
	selected := append([]string{}, rules...)
	options := make([]huh.Option[string], len(rules))
	for i, r := range rules {
		options[i] = huh.NewOption(r, r).Selected(true)
	}
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Global rules to render for every agent").
				Description("Rendered into ~/.claude/CLAUDE.md and ~/.pi/agent/AGENTS.md.").
				Options(options...).
				Value(&selected),
		),
	)
	return f, func() config.AgentsConfig {
		var skip []string
		for _, r := range rules {
			if !slices.Contains(selected, r) {
				skip = append(skip, r)
			}
		}
		return config.AgentsConfig{SkipRules: skip}
	}
}

// ProjectAnswers fills the per-project AGENTS.md template; Agents picks which
// tools get a pointer file (claude, gemini; pi reads AGENTS.md natively).
type ProjectAnswers struct {
	Agents      []string
	Name        string
	Description string
	TestCommand string
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
