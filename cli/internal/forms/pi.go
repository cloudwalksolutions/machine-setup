package forms

import (
	"errors"
	"os"
	"strings"

	"github.com/charmbracelet/huh"

	"tars/internal/config"
)

// PiDetected is what init found on the machine before asking anything.
type PiDetected struct {
	Providers    []config.PiProvider // parsed from an existing ~/.pi/agent/models.json
	OllamaModels []string            // from `ollama list`; empty when ollama is absent
}

// PiAnswers separates what is persisted (Config) from what only goes to ~/.zshrc_secret.
type PiAnswers struct {
	Config  config.PiConfig
	Secrets map[string]string // env var name → API key, for variables not exported yet
}

// ShowPiInitForm asks which packages to install and which model providers to
// configure. When TARS_NO_FORM=1 it installs every package, keeps the detected
// providers, pins no default, and collects no secrets.
func ShowPiInitForm(d PiDetected, packages []string) (PiAnswers, error) {
	answers := PiAnswers{
		Config:  config.PiConfig{Packages: append([]string{}, packages...), Providers: d.Providers},
		Secrets: map[string]string{},
	}
	if os.Getenv("TARS_NO_FORM") != "" {
		return answers, nil
	}

	pkgOptions := make([]huh.Option[string], len(packages))
	for i, p := range packages {
		pkgOptions[i] = huh.NewOption(p, p).Selected(true)
	}
	var kinds []string
	if len(d.OllamaModels) > 0 {
		kinds = append(kinds, "ollama")
	}
	if detected(d.Providers, "openai").Name != "" {
		kinds = append(kinds, "openai")
	}
	err := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("pi packages to install").
			Description("Subagents, MCP adapter, web search, bigpowers skills, permission system, plannotator plan review, live todo overlay, pi-lens diagnostics, ponytail YAGNI. Space toggles, Enter confirms.").
			Options(pkgOptions...).
			Value(&answers.Config.Packages),
		huh.NewMultiSelect[string]().
			Title("Model providers to configure").
			Description("ollama and llama.cpp are local; openai is any OpenAI-compatible endpoint with a key from ~/.zshrc_secret.").
			Options(
				huh.NewOption("ollama (local, http://localhost:11434)", "ollama").Selected(contains(kinds, "ollama")),
				huh.NewOption("llama.cpp server via pi-llama-cpp", "llamacpp"),
				huh.NewOption("OpenAI-compatible endpoint", "openai").Selected(contains(kinds, "openai")),
			).
			Value(&kinds),
	)).Run()
	if err != nil {
		return answers, err
	}

	answers.Config.Providers, err = askProviders(kinds, d, answers.Secrets)
	if err != nil {
		return answers, err
	}
	answers.Config.DefaultProvider, answers.Config.DefaultModel, err = askDefault(answers.Config.Providers)
	return answers, err
}

// askProviders collects one PiProvider per chosen kind, pre-filled from detection.
// An openai provider whose env var is not exported also asks for the key itself.
func askProviders(kinds []string, d PiDetected, secrets map[string]string) ([]config.PiProvider, error) {
	var out []config.PiProvider
	for _, kind := range kinds {
		p := detected(d.Providers, kind)
		p.Kind = kind
		switch kind {
		case "ollama":
			p.Name = "ollama"
			models := append([]string{}, p.Models...)
			opts := make([]huh.Option[string], len(d.OllamaModels))
			for i, m := range d.OllamaModels {
				opts[i] = huh.NewOption(m, m).Selected(len(models) == 0 || contains(models, m))
			}
			if err := huh.NewForm(huh.NewGroup(
				huh.NewMultiSelect[string]().Title("ollama models to expose").Options(opts...).Value(&models),
			)).Run(); err != nil {
				return nil, err
			}
			p.Models = models
		case "llamacpp":
			p.Name = "llamacpp"
			if p.BaseURL == "" {
				p.BaseURL = "http://localhost:8080"
			}
			if err := huh.NewForm(huh.NewGroup(
				huh.NewInput().Title("llama.cpp server URL").Value(&p.BaseURL),
			)).Run(); err != nil {
				return nil, err
			}
		case "openai":
			models := strings.Join(p.Models, ",")
			if p.Name == "" {
				p.Name = "remote"
			}
			if err := huh.NewForm(huh.NewGroup(
				huh.NewInput().Title("Provider name").Value(&p.Name),
				huh.NewInput().Title("Base URL (…/v1)").Value(&p.BaseURL),
				huh.NewInput().Title("Env var holding the API key").Description("Exported from ~/.zshrc_secret.").Value(&p.KeyEnv),
				huh.NewInput().Title("Model ids, comma-separated").Value(&models),
			)).Run(); err != nil {
				return nil, err
			}
			p.Models = splitList(models)
			if p.KeyEnv == "" {
				p.KeyEnv = config.KeyEnvFor(p.Name)
			}
			if os.Getenv(p.KeyEnv) == "" {
				var key string
				if err := huh.NewForm(huh.NewGroup(
					huh.NewInput().Title("API key for " + p.Name).
						Description(p.KeyEnv + " is not exported; the key is stored in ~/.zshrc_secret, never in the repo. Leave empty to add it yourself later.").
						EchoMode(huh.EchoModePassword).Value(&key),
				)).Run(); err != nil {
					return nil, err
				}
				if key != "" {
					secrets[p.KeyEnv] = key
				}
			}
		}
		out = append(out, p)
	}
	return out, nil
}

// askDefault picks the default provider/model among the configured ones.
func askDefault(providers []config.PiProvider) (string, string, error) {
	var choices []huh.Option[string]
	for _, p := range providers {
		for _, m := range p.Models {
			choices = append(choices, huh.NewOption(p.Name+"/"+m, p.Name+"/"+m))
		}
	}
	if len(choices) == 0 {
		return "", "", nil
	}
	pick := choices[0].Value
	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Default model").Options(choices...).Value(&pick),
	)).Run(); err != nil {
		return "", "", err
	}
	provider, model, ok := strings.Cut(pick, "/")
	if !ok {
		return "", "", errors.New("invalid default selection")
	}
	return provider, model, nil
}

func detected(providers []config.PiProvider, kind string) config.PiProvider {
	for _, p := range providers {
		if p.Kind == kind {
			return p
		}
	}
	return config.PiProvider{}
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}
