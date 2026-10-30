package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/report"
	"tars/internal/tui"
)

// ClaudeAsker presents the `tars init claude` form.
type ClaudeAsker interface {
	AskClaude() (config.ClaudeConfig, error)
}

// ClaudeInit orchestrates `tars init claude`: ask, persist the choices, pull.
type ClaudeInit struct {
	Asker  ClaudeAsker
	Config ConfigStore
	Pull   func(config.ClaudeConfig) error
	Report report.Reporter
}

// Run asks the user which pieces to provision, saves the answers, and applies them.
func (c *ClaudeInit) Run() error {
	cfg, err := c.Config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	answer, err := c.Asker.AskClaude()
	if err != nil && err.Error() == "user aborted" {
		c.Report.Note("Aborted; nothing changed.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("claude form: %w", err)
	}
	cfg.Claude = answer
	if err := c.Config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	c.Report.Note("Choices saved to " + c.Config.Path())
	return c.Pull(cfg.Claude)
}

// ProjectAsker presents the `tars claude project` form.
type ProjectAsker interface {
	AskProject(defaultName string) (forms.ProjectAnswers, error)
}

// ── Production collaborators ─────────────────────────────────────────────

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

// NewClaudeInit wires the production `tars init claude`.
func NewClaudeInit(r report.Reporter, ask ClaudeAsker) (*ClaudeInit, error) {
	stdout, stderr := r.Output()
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	return &ClaudeInit{
		Asker:  ask,
		Config: NewFileConfigStore(configPath()),
		Pull: func(c config.ClaudeConfig) error {
			opts.Claude = c
			return SequentialPuller{
				Components: []components.Component{components.NewClaude(opts)},
				Report:     r,
			}.PullAll()
		},
		Report: r,
	}, nil
}

// ── Cobra commands ───────────────────────────────────────────────────────

var claudeInitCmd = &cobra.Command{
	Use:     "claude",
	Aliases: []string{"c"},
	Short:   "Choose and apply the Claude Code pieces for this machine (alias: c)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runInit(cmd, func(r report.Reporter, ask tui.Asker) error {
			c, err := NewClaudeInit(r, ask)
			if err != nil {
				return err
			}
			return c.Run()
		})
	},
}

func init() {
	claudeInitCmd.SilenceUsage = true
	initCmd.AddCommand(claudeInitCmd)
}
