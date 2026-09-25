package forms

import (
	"slices"

	"charm.land/huh/v2"

	"tars/internal/config"
)

// PiInputs is what the machine reports before the pi questions: the local ollama
// models, the default package list, and the previously saved choices.
type PiInputs struct {
	OllamaModels []string
	Packages     []string
	Previous     config.PiConfig
}

// PiForm builds the package picker and, when a local ollama is present, the model
// picker. Every package starts checked; models start as previously exposed, or all
// of them the first time. Other providers are pi's own business.
func PiForm(in PiInputs) (*huh.Form, func() config.PiConfig) {
	cfg := config.PiConfig{Packages: append([]string{}, in.Packages...)}
	pkgOptions := make([]huh.Option[string], len(in.Packages))
	for i, p := range in.Packages {
		pkgOptions[i] = huh.NewOption(p, p).Selected(true)
	}
	fields := []huh.Field{
		huh.NewMultiSelect[string]().
			Title("pi packages to install").
			Description("Subagents, MCP adapter, web search, bigpowers skills, permission system, plannotator plan review, live todo overlay, ponytail YAGNI, ask-user-question forms, clipboard image paste, context-mode output sandboxing, background shell tasks.").
			Options(pkgOptions...).
			Value(&cfg.Packages),
	}
	if len(in.OllamaModels) == 0 {
		fields = append(fields, huh.NewNote().
			Title("No local ollama found").
			Description("Install ollama to expose local models through tars. Remote providers are configured in pi itself (/login or ~/.pi/agent/models.json)."))
	} else {
		modelOptions := make([]huh.Option[string], len(in.OllamaModels))
		for i, m := range in.OllamaModels {
			keep := len(in.Previous.OllamaModels) == 0 || slices.Contains(in.Previous.OllamaModels, m)
			modelOptions[i] = huh.NewOption(m, m).Selected(keep)
			if keep {
				cfg.OllamaModels = append(cfg.OllamaModels, m)
			}
		}
		fields = append(fields, huh.NewMultiSelect[string]().
			Title("ollama models to expose to pi").
			Options(modelOptions...).
			Value(&cfg.OllamaModels))
	}
	return huh.NewForm(huh.NewGroup(fields...)), func() config.PiConfig { return cfg }
}

// PiDefaultForm builds the default-model choice over the exposed models, starting on
// the previous default when it is still exposed, else the first model.
func PiDefaultForm(models []string, previous string) (*huh.Form, func() string) {
	choices := make([]huh.Option[string], 0, len(models)+1)
	for _, m := range models {
		choices = append(choices, huh.NewOption(m, m))
	}
	choices = append(choices, huh.NewOption("keep pi's current default", ""))
	chosen := previous
	if !slices.Contains(models, chosen) {
		chosen = models[0]
	}
	f := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Default model").Options(choices...).Value(&chosen),
	))
	return f, func() string { return chosen }
}
