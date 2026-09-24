package forms

import (
	"os"
	"slices"

	"github.com/charmbracelet/huh"

	"tars/internal/config"
)

// ShowPiInitForm asks which packages to install and, when a local ollama is
// present, which of its models to expose and which one is the default. Other
// providers are pi's own business. When TARS_NO_FORM=1 it installs every
// package, exposes every ollama model, and pins no default.
func ShowPiInitForm(ollamaModels, packages []string, previous config.PiConfig) (config.PiConfig, error) {
	cfg := config.PiConfig{Packages: append([]string{}, packages...), OllamaModels: append([]string{}, ollamaModels...)}
	if os.Getenv("TARS_NO_FORM") != "" {
		return cfg, nil
	}

	pkgOptions := make([]huh.Option[string], len(packages))
	for i, p := range packages {
		pkgOptions[i] = huh.NewOption(p, p).Selected(true)
	}
	fields := []huh.Field{
		huh.NewMultiSelect[string]().
			Title("pi packages to install").
			Description("Subagents, MCP adapter, web search, bigpowers skills, permission system, plannotator plan review, live todo overlay, ponytail YAGNI, ask-user-question forms, clipboard image paste. Space toggles, Enter confirms.").
			Options(pkgOptions...).
			Value(&cfg.Packages),
	}
	if len(ollamaModels) == 0 {
		fields = append(fields, huh.NewNote().
			Title("No local ollama found").
			Description("Install ollama to expose local models through tars. Remote providers are configured in pi itself (/login or ~/.pi/agent/models.json)."))
	} else {
		modelOptions := make([]huh.Option[string], len(ollamaModels))
		for i, m := range ollamaModels {
			keep := len(previous.OllamaModels) == 0 || slices.Contains(previous.OllamaModels, m)
			modelOptions[i] = huh.NewOption(m, m).Selected(keep)
		}
		fields = append(fields, huh.NewMultiSelect[string]().
			Title("ollama models to expose to pi").
			Options(modelOptions...).
			Value(&cfg.OllamaModels))
	}
	if err := run(huh.NewForm(huh.NewGroup(fields...))); err != nil {
		return cfg, err
	}
	if len(cfg.OllamaModels) == 0 {
		return cfg, nil
	}

	choices := make([]huh.Option[string], 0, len(cfg.OllamaModels)+1)
	for _, m := range cfg.OllamaModels {
		choices = append(choices, huh.NewOption(m, m))
	}
	choices = append(choices, huh.NewOption("keep pi's current default", ""))
	cfg.DefaultModel = previous.DefaultModel
	if !slices.Contains(cfg.OllamaModels, cfg.DefaultModel) {
		cfg.DefaultModel = cfg.OllamaModels[0]
	}
	err := run(huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Default model").Options(choices...).Value(&cfg.DefaultModel),
	)))
	return cfg, err
}
