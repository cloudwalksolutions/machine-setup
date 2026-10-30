package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/config"
	"tars/internal/paths"
	"tars/internal/report"
	"tars/internal/tui"
)

type AgentsAsker interface {
	AskAgents(rules []string) (config.AgentsConfig, error)
}

type AgentsInit struct {
	Asker  AgentsAsker
	Config ConfigStore
	Rules  []string
	Pull   func(config.AgentsConfig) error
	Report report.Reporter
}

func (a *AgentsInit) Run() error {
	cfg, err := a.Config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	answer, err := a.Asker.AskAgents(a.Rules)
	if err != nil && err.Error() == "user aborted" {
		a.Report.Note("Aborted; nothing changed.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("agents form: %w", err)
	}
	cfg.Agents = answer
	if err := a.Config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	a.Report.Note("Choices saved to " + a.Config.Path())
	return a.Pull(cfg.Agents)
}

func ruleNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("listing agent rules: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, strings.TrimSuffix(e.Name(), ".md"))
	}
	return names, nil
}

func NewAgentsInit(r report.Reporter, ask AgentsAsker) (*AgentsInit, error) {
	stdout, stderr := r.Output()
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	rules, err := ruleNames(paths.For(opts.RepoRoot, opts.Home).Agents.RulesRepo)
	if err != nil {
		return nil, err
	}
	return &AgentsInit{
		Asker:  ask,
		Config: NewFileConfigStore(configPath()),
		Rules:  rules,
		Pull: func(a config.AgentsConfig) error {
			opts.Agents = a
			return SequentialPuller{
				Components: []components.Component{components.NewClaude(opts), components.NewPi(opts)},
				Report:     r,
			}.PullAll()
		},
		Report: r,
	}, nil
}

var agentsInitCmd = &cobra.Command{
	Use:     "agents",
	Aliases: []string{"a"},
	Short:   "Choose which shared rules every coding agent renders (alias: a)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runInit(cmd, func(r report.Reporter, ask tui.Asker) error {
			a, err := NewAgentsInit(r, ask)
			if err != nil {
				return err
			}
			return a.Run()
		})
	},
}

func init() {
	agentsInitCmd.SilenceUsage = true
	initCmd.AddCommand(agentsInitCmd)
}
