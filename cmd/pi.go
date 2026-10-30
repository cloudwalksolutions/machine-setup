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

// PiAsker presents the `tars init pi` questions over what the machine reports.
type PiAsker interface {
	AskPi(in forms.PiInputs) (config.PiConfig, error)
}

// PiOps are the machine-touching steps of init, in the order Run calls them.
type PiOps interface {
	Pull(config.PiConfig) error
	InstallPackages([]string) error
}

// PiInit orchestrates `tars init pi`: ask, persist the choices, then apply them.
type PiInit struct {
	Asker  PiAsker
	Inputs func() forms.PiInputs
	Config ConfigStore
	Ops    PiOps
	Report report.Reporter
}

// Run asks for the pi setup, saves it, and applies it to the machine.
func (p *PiInit) Run() error {
	cfg, err := p.Config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	answer, err := p.Asker.AskPi(p.Inputs())
	if err != nil && err.Error() == "user aborted" {
		p.Report.Note("Aborted; nothing changed.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("pi form: %w", err)
	}
	cfg.Pi = answer
	if err := p.Config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	p.Report.Note("Choices saved to " + p.Config.Path())
	if err := p.Ops.Pull(answer); err != nil {
		return err
	}
	p.Report.StepStarted("Installing pi packages", len(answer.Packages))
	if err := p.Ops.InstallPackages(answer.Packages); err != nil {
		return err
	}
	p.Report.Note("\nNext steps:")
	p.Report.Note("  • Run `pi` and try /tdd, /plan, /pr; the `tars` agent is available to the subagent tool")
	p.Report.Note("  • Other model providers: `/login` inside pi or ~/.pi/agent/models.json")
	return nil
}

// ── Production collaborators ─────────────────────────────────────────────

// componentPiOps runs the init steps against a Pi component built with the saved choices.
type componentPiOps struct {
	opts   components.Options
	report report.Reporter
}

func (o componentPiOps) with(cfg config.PiConfig) *components.Pi {
	opts := o.opts
	opts.Pi = cfg
	return components.NewPi(opts)
}

func (o componentPiOps) Pull(cfg config.PiConfig) error {
	return SequentialPuller{Components: []components.Component{o.with(cfg)}, Report: o.report}.PullAll()
}
func (o componentPiOps) InstallPackages(p []string) error {
	return o.with(o.opts.Pi).InstallPackages(p)
}

// NewPiInit wires the production `tars init pi`.
func NewPiInit(r report.Reporter, ask PiAsker) (*PiInit, error) {
	stdout, stderr := r.Output()
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	pi := components.NewPi(opts)
	return &PiInit{
		Asker: ask,
		Inputs: func() forms.PiInputs {
			return forms.PiInputs{OllamaModels: pi.OllamaModels(), Packages: pi.Packages(), Previous: opts.Pi}
		},
		Config: NewFileConfigStore(configPath()),
		Ops:    componentPiOps{opts: opts, report: r},
		Report: r,
	}, nil
}

// ── Cobra commands ───────────────────────────────────────────────────────

var piInitCmd = &cobra.Command{
	Use:   "pi",
	Short: "Choose packages and local ollama models for the pi coding agent and apply them to ~/.pi/agent",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runInit(cmd, func(r report.Reporter, ask tui.Asker) error {
			p, err := NewPiInit(r, ask)
			if err != nil {
				return err
			}
			return p.Run()
		})
	},
}

func init() {
	piInitCmd.SilenceUsage = true
	initCmd.AddCommand(piInitCmd)
}
