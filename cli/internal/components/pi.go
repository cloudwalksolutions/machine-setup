package components

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"tars/internal/config"
	"tars/internal/paths"
)

// Pi provisions the pi coding agent: the baseline agent, prompts, the edit
// guard extension, permissions, a settings fragment, and AGENTS.md rendered
// from the shared rules. Providers are rendered from config, never from the repo.
type Pi struct {
	opts   Options
	cfg    config.PiConfig
	p      paths.PiPaths
	rules  string
	secret string

	// Run executes an external command (pi, ollama); the seam tests replace.
	Run func(name string, args ...string) (string, error)
}

// NewPi builds the component from opts, including the `tars pi init` choices.
func NewPi(opts Options) *Pi {
	all := paths.For(opts.RepoRoot, opts.Home)
	return &Pi{opts: opts, cfg: opts.Pi, p: all.Pi, rules: all.Claude.RulesRepo, secret: all.Zsh.SecretLocal, Run: defaultRun}
}

func defaultRun(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

// Packages lists the packages the repo fragment installs.
func (c *Pi) Packages() []string {
	fragment, err := loadSettings(c.p.SettingsRepo)
	if err != nil {
		return nil
	}
	return fragment.strings("packages")
}

// DetectProviders reads the existing models.json into PiProvider entries so the
// init form can pre-fill them; a localhost:11434 base URL means ollama.
func (c *Pi) DetectProviders() []config.PiProvider {
	models, err := loadSettings(c.p.ModelsLocal)
	if err != nil {
		return nil
	}
	providers, _ := models["providers"].(map[string]any)
	var out []config.PiProvider
	for name, raw := range providers {
		entry, _ := raw.(map[string]any)
		baseURL, _ := entry["baseUrl"].(string)
		p := config.PiProvider{Name: name, Kind: "openai", BaseURL: baseURL, KeyEnv: config.KeyEnvFor(name)}
		if strings.Contains(baseURL, "localhost:11434") {
			p.Kind, p.KeyEnv = "ollama", ""
		}
		list, _ := entry["models"].([]any)
		for _, m := range list {
			if mm, ok := m.(map[string]any); ok {
				if id, ok := mm["id"].(string); ok {
					p.Models = append(p.Models, id)
				}
			}
		}
		out = append(out, p)
	}
	return out
}

// OllamaModels returns the model names a local ollama reports; empty when ollama is absent.
func (c *Pi) OllamaModels() []string {
	out, err := c.Run("ollama", "list")
	if err != nil {
		return nil
	}
	var names []string
	for i, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.Fields(line)
		if i == 0 || len(fields) == 0 {
			continue
		}
		names = append(names, fields[0])
	}
	return names
}

// InstallPackages runs `pi install` for each selected package `pi list` does not show.
func (c *Pi) InstallPackages(selected []string) error {
	listed, err := c.Run("pi", "list")
	if err != nil {
		return fmt.Errorf("pi list: %w", err)
	}
	for _, src := range selected {
		if strings.Contains(listed, src) {
			continue
		}
		fmt.Fprintf(c.opts.Stdout, "  installing %s\n", src)
		if out, err := c.Run("pi", "install", src); err != nil {
			return fmt.Errorf("pi install %s: %w\n%s", src, err, out)
		}
	}
	return nil
}

// StoreKeys appends API keys to ~/.zshrc_secret as exports so models.json can
// reference them as $VAR; variables already present are left untouched.
func (c *Pi) StoreKeys(keys map[string]string) error {
	names := make([]string, 0, len(keys))
	for name := range keys {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := c.appendSecret(name, keys[name]); err != nil {
			return err
		}
	}
	return nil
}

// appendSecret adds an export line to ~/.zshrc_secret unless the variable is already there.
func (c *Pi) appendSecret(name, value string) error {
	existing, err := os.ReadFile(c.secret)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(existing), "export "+name+"=") {
		return nil
	}
	content := append(existing, []byte(fmt.Sprintf("export %s=\"%s\"\n", name, value))...)
	return c.opts.copier().SafeWrite(content, 0o600, c.secret, c.Name(), c.opts.BackupRoot)
}

func (c *Pi) Name() string { return "pi" }

// files lists the repo↔local pairs copied verbatim in both directions.
func (c *Pi) files() [][2]string {
	return [][2]string{
		{c.p.AgentRepo, c.p.AgentLocal},
		{c.p.ExtensionRepo, c.p.ExtensionLocal},
		{c.p.PermissionsRepo, c.p.PermissionsLocal},
	}
}

// Pull applies the repo's pi config to ~/.pi/agent.
func (c *Pi) Pull() error {
	for _, f := range c.files() {
		if err := c.opts.copier().SafeCopy(f[0], f[1], c.Name(), c.opts.BackupRoot); err != nil {
			return err
		}
	}
	if err := c.copyDir(c.p.PromptsRepo, c.p.PromptsLocal, c.Name()); err != nil {
		return err
	}
	if err := c.pullSettings(); err != nil {
		return err
	}
	return c.pullRules()
}

// Push copies the local agent, prompts, extension and permissions back to the
// repo, archiving the repo copies under "pi-repo". Providers and models never move.
func (c *Pi) Push() error {
	comp := c.Name() + "-repo"
	for _, f := range c.files() {
		if err := c.opts.copier().SafeCopy(f[1], f[0], comp, c.opts.BackupRoot); err != nil {
			return err
		}
	}
	return c.copyDir(c.p.PromptsLocal, c.p.PromptsRepo, comp)
}

// copyDir copies every file in src flat into dst under component.
func (c *Pi) copyDir(src, dst, component string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := c.opts.copier().SafeCopy(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name()), component, c.opts.BackupRoot); err != nil {
			return err
		}
	}
	return nil
}

// piShareableKeys are the settings.json keys the repo fragment owns.
var piShareableKeys = []string{"defaultThinkingLevel"}

// pullSettings merges the fragment into ~/.pi/agent/settings.json: shareable
// keys overwrite, packages are unioned fragment-first, everything else is kept.
func (c *Pi) pullSettings() error {
	fragment, err := loadSettings(c.p.SettingsRepo)
	if err != nil {
		return err
	}
	local, err := loadSettings(c.p.SettingsLocal)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, k := range piShareableKeys {
		if v, ok := fragment[k]; ok {
			local[k] = v
		}
	}
	local["packages"] = unionStrings(fragment.strings("packages"), local.strings("packages"))
	if c.cfg.DefaultProvider != "" {
		local["defaultProvider"] = c.cfg.DefaultProvider
	}
	if c.cfg.DefaultModel != "" {
		local["defaultModel"] = c.cfg.DefaultModel
	}
	for _, p := range c.cfg.Providers {
		if p.Kind == "llamacpp" {
			local["llamaServerUrl"] = p.BaseURL
		}
	}
	if err := local.save(c.p.SettingsLocal, c.Name(), c.opts.BackupRoot); err != nil {
		return err
	}
	return c.pullModels()
}

// pullModels upserts the configured providers into models.json by name; local-only
// providers stay. Keys are env references, never literals.
func (c *Pi) pullModels() error {
	if len(c.cfg.Providers) == 0 {
		return nil
	}
	models, err := loadSettings(c.p.ModelsLocal)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	providers, _ := models["providers"].(map[string]any)
	if providers == nil {
		providers = map[string]any{}
	}
	changed := false
	for _, p := range c.cfg.Providers {
		if entry := providerEntry(p); entry != nil {
			providers[p.Name] = entry
			changed = true
		}
	}
	if !changed {
		return nil
	}
	models["providers"] = providers
	return models.save(c.p.ModelsLocal, c.Name(), c.opts.BackupRoot)
}

// providerEntry renders one models.json provider; llama.cpp has none (pi-llama-cpp registers it).
func providerEntry(p config.PiProvider) map[string]any {
	ids := make([]any, 0, len(p.Models))
	for _, id := range p.Models {
		ids = append(ids, map[string]any{"id": id})
	}
	switch p.Kind {
	case "ollama":
		return map[string]any{"baseUrl": "http://localhost:11434/v1", "api": "openai-completions", "apiKey": "ollama", "models": ids}
	case "openai":
		return map[string]any{"baseUrl": p.BaseURL, "api": "openai-completions", "apiKey": "$" + p.KeyEnv, "models": ids}
	}
	return nil
}

func unionStrings(first, second []string) []any {
	out := make([]any, 0, len(first)+len(second))
	seen := map[string]bool{}
	for _, s := range append(append([]string{}, first...), second...) {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

const agentsMDHeader = "# Working rules\n\nManaged by `tars pi init`; edit the rule files under claude/rules in the machine-setup repo, not this file.\n\n"

// pullRules renders the same rule selection Claude uses into ~/.pi/agent/AGENTS.md.
func (c *Pi) pullRules() error {
	content, err := renderRules(c.rules, agentsMDHeader, c.opts.Claude.Rules)
	if err != nil {
		return err
	}
	return c.opts.copier().SafeWrite(content, 0o644, c.p.AgentsMDLocal, c.Name(), c.opts.BackupRoot)
}
