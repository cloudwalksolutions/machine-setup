package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/paths"
)

// ClaudeAsker presents the `tars init claude` form over the available rule names.
type ClaudeAsker interface {
	Ask(rules []string) (config.ClaudeConfig, error)
}

// ClaudeInit orchestrates `tars init claude`: ask, persist the choices, pull.
type ClaudeInit struct {
	Asker  ClaudeAsker
	Config ConfigStore
	Rules  []string
	Pull   func(config.ClaudeConfig) error

	Stdout io.Writer
}

// Run asks the user which pieces to provision, saves the answers, and applies them.
func (c *ClaudeInit) Run() error {
	cfg, err := c.Config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	answer, err := c.Asker.Ask(c.Rules)
	if err != nil && err.Error() == "user aborted" {
		fmt.Fprintln(c.Stdout, "Aborted; nothing changed.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("claude form: %w", err)
	}
	cfg.Claude = answer
	if err := c.Config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Fprintf(c.Stdout, "Choices saved to %s\n", c.Config.Path())
	return c.Pull(cfg.Claude)
}

// ProjectAsker presents the `tars claude project` form.
type ProjectAsker interface {
	Ask(defaultName string) (forms.ProjectAnswers, error)
}

// ── Production collaborators ─────────────────────────────────────────────

// FormsClaudeAsker wraps forms.ShowClaudeInitForm.
type FormsClaudeAsker struct{}

func (FormsClaudeAsker) Ask(rules []string) (config.ClaudeConfig, error) {
	return forms.ShowClaudeInitForm(rules)
}

// configPath honors --config, else the default (env-overridable) location.
func configPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return config.DefaultConfigPath()
}

// LoadClaudeConfig reads the saved `tars init claude` choices; a missing
// config yields the all-on zero value so pull/setup work before init ran.
func LoadClaudeConfig() (config.ClaudeConfig, error) {
	cfg, err := config.Init(configPath())
	if err != nil {
		return config.ClaudeConfig{}, fmt.Errorf("loading config: %w", err)
	}
	return cfg.Claude, nil
}

// ruleNames lists the rule files in dir without their .md suffix.
func ruleNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("listing claude rules: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, strings.TrimSuffix(e.Name(), ".md"))
	}
	return names, nil
}

// NewClaudeInit wires the production `tars init claude`.
func NewClaudeInit(stdout, stderr io.Writer) (*ClaudeInit, error) {
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	rules, err := ruleNames(paths.For(opts.RepoRoot, opts.Home).Claude.RulesRepo)
	if err != nil {
		return nil, err
	}
	return &ClaudeInit{
		Asker:  FormsClaudeAsker{},
		Config: NewFileConfigStore(configPath()),
		Rules:  rules,
		Pull: func(c config.ClaudeConfig) error {
			opts.Claude = c
			return SequentialPuller{
				Components: []components.Component{components.NewClaude(opts)},
				Stdout:     stdout,
				Stderr:     stderr,
			}.PullAll()
		},
		Stdout: stdout,
	}, nil
}

// ── Cobra commands ───────────────────────────────────────────────────────

var claudeInitCmd = &cobra.Command{
	Use:     "claude",
	Aliases: []string{"c"},
	Short:   "Choose and apply the Claude Code pieces for this machine (alias: c)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := NewClaudeInit(cmd.OutOrStdout(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		return c.Run()
	},
}

func init() {
	claudeInitCmd.SilenceUsage = true
	initCmd.AddCommand(claudeInitCmd)
}
