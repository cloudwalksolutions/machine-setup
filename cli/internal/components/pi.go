package components

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tars/internal/config"
	"tars/internal/paths"
)

// Pi provisions the pi coding agent: the baseline agent, prompts, the edit
// guard extension, permissions, a settings fragment, and AGENTS.md rendered
// from the shared rules. The only provider tars manages is local ollama.
type Pi struct {
	opts  Options
	cfg   config.PiConfig
	p     paths.PiPaths
	rules string

	// Run executes an external command (pi, ollama); the seam tests replace.
	Run func(name string, args ...string) (string, error)
}

// NewPi builds the component from opts, including the `tars pi init` choices.
func NewPi(opts Options) *Pi {
	all := paths.For(opts.RepoRoot, opts.Home)
	return &Pi{opts: opts, cfg: opts.Pi, p: all.Pi, rules: all.Claude.RulesRepo, Run: defaultRun}
}

func defaultRun(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
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
	if err := c.pullKeybindings(); err != nil {
		return err
	}
	return c.pullRules()
}

// pullKeybindings merges the fragment into ~/.pi/agent/keybindings.json: fragment
// actions overwrite (an empty list unbinds), the user's other bindings stay.
func (c *Pi) pullKeybindings() error {
	fragment, err := loadSettings(c.p.KeybindingsRepo)
	if err != nil {
		return err
	}
	local, err := loadSettings(c.p.KeybindingsLocal)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	maps.Copy(local, fragment)
	return local.save(c.p.KeybindingsLocal, c.Name(), c.opts.BackupRoot)
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
	if c.cfg.DefaultModel != "" {
		local["defaultProvider"] = "ollama"
		local["defaultModel"] = c.cfg.DefaultModel
	}
	if err := local.save(c.p.SettingsLocal, c.Name(), c.opts.BackupRoot); err != nil {
		return err
	}
	return c.pullModels()
}

// pullModels upserts the ollama provider into models.json; other providers stay.
func (c *Pi) pullModels() error {
	if len(c.cfg.OllamaModels) == 0 {
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
	ids := make([]any, 0, len(c.cfg.OllamaModels))
	for _, id := range c.cfg.OllamaModels {
		ids = append(ids, map[string]any{"id": id})
	}
	providers["ollama"] = map[string]any{"baseUrl": "http://localhost:11434/v1", "api": "openai-completions", "apiKey": "ollama", "models": ids}
	models["providers"] = providers
	return models.save(c.p.ModelsLocal, c.Name(), c.opts.BackupRoot)
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

// Packages lists the packages the repo fragment installs.
func (c *Pi) Packages() []string {
	fragment, err := loadSettings(c.p.SettingsRepo)
	if err != nil {
		return nil
	}
	return fragment.strings("packages")
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
