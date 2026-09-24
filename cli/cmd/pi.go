package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/config"
	"tars/internal/forms"
)

// PiAsker presents the `tars pi init` form.
type PiAsker interface {
	Ask() (config.PiConfig, error)
}

// PiOps are the machine-touching steps of init, in the order Run calls them.
type PiOps interface {
	Pull(config.PiConfig) error
	InstallPackages([]string) error
}

// PiInit orchestrates `tars pi init`: ask, persist the choices, then apply them.
type PiInit struct {
	Asker  PiAsker
	Config ConfigStore
	Ops    PiOps

	Stdout io.Writer
}

// Run asks for the pi setup, saves it, and applies it to the machine.
func (p *PiInit) Run() error {
	cfg, err := p.Config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	answer, err := p.Asker.Ask()
	if err != nil && err.Error() == "user aborted" {
		fmt.Fprintln(p.Stdout, "Aborted; nothing changed.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("pi form: %w", err)
	}
	cfg.Pi = answer
	if err := p.Config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Fprintf(p.Stdout, "Choices saved to %s\n", p.Config.Path())
	if err := p.Ops.Pull(answer); err != nil {
		return err
	}
	if err := p.Ops.InstallPackages(answer.Packages); err != nil {
		return err
	}
	fmt.Fprintln(p.Stdout, "\nNext steps:")
	fmt.Fprintln(p.Stdout, "  • Run `pi` and try /tdd, /plan, /pr; the `tars` agent is available to the subagent tool")
	fmt.Fprintln(p.Stdout, "  • Other model providers: `/login` inside pi or ~/.pi/agent/models.json")
	return nil
}

// ── Production collaborators ─────────────────────────────────────────────

// FormsPiAsker wraps forms.ShowPiInitForm with what the machine reports.
type FormsPiAsker struct {
	pi       *components.Pi
	previous config.PiConfig
}

func (a FormsPiAsker) Ask() (config.PiConfig, error) {
	return forms.ShowPiInitForm(a.pi.OllamaModels(), a.pi.Packages(), a.previous)
}

// componentPiOps runs the init steps against a Pi component built with the saved choices.
type componentPiOps struct {
	opts   components.Options
	stdout io.Writer
	stderr io.Writer
}

func (o componentPiOps) with(cfg config.PiConfig) *components.Pi {
	opts := o.opts
	opts.Pi = cfg
	return components.NewPi(opts)
}

func (o componentPiOps) Pull(cfg config.PiConfig) error {
	return SequentialPuller{Components: []components.Component{o.with(cfg)}, Stdout: o.stdout, Stderr: o.stderr}.PullAll()
}
func (o componentPiOps) InstallPackages(p []string) error {
	return o.with(o.opts.Pi).InstallPackages(p)
}

// NewPiInit wires the production `tars pi init`.
func NewPiInit(stdout, stderr io.Writer) (*PiInit, error) {
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	return &PiInit{
		Asker:  FormsPiAsker{pi: components.NewPi(opts), previous: opts.Pi},
		Config: NewFileConfigStore(configPath()),
		Ops:    componentPiOps{opts: opts, stdout: stdout, stderr: stderr},
		Stdout: stdout,
	}, nil
}

// ── Cobra commands ───────────────────────────────────────────────────────

var piCmd = &cobra.Command{
	Use:   "pi",
	Short: "Provision the pi coding agent: baseline agent, prompts, edit guard, packages, providers",
	Long: `Manage the pi coding agent the same way tars manages dotfiles.

  init  choose packages and which local ollama models to expose, save the choices,
        and apply them to ~/.pi/agent

'tars pull' keeps the agent, prompts, extension, permissions and AGENTS.md in sync
afterwards. Other model providers are configured in pi itself.`,
}

var piInitCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Choose and apply the pi setup for this machine (alias: i)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		p, err := NewPiInit(cmd.OutOrStdout(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		return p.Run()
	},
}

func init() {
	for _, c := range []*cobra.Command{piCmd, piInitCmd} {
		c.SilenceUsage = true
	}
	piCmd.AddCommand(piInitCmd)
}
