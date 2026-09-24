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
	Ask() (forms.PiAnswers, error)
}

// PiOps are the machine-touching steps of init, in the order Run calls them.
type PiOps interface {
	StoreKeys(map[string]string) error
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
	cfg.Pi = answer.Config
	if err := p.Config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Fprintf(p.Stdout, "Choices saved to %s\n", p.Config.Path())
	if err := p.Ops.StoreKeys(answer.Secrets); err != nil {
		return err
	}
	if err := p.Ops.Pull(answer.Config); err != nil {
		return err
	}
	if err := p.Ops.InstallPackages(answer.Config.Packages); err != nil {
		return err
	}
	p.printNextSteps(answer.Config)
	return nil
}

func (p *PiInit) printNextSteps(cfg config.PiConfig) {
	fmt.Fprintln(p.Stdout, "\nNext steps:")
	for _, pr := range cfg.Providers {
		switch pr.Kind {
		case "openai":
			fmt.Fprintf(p.Stdout, "  • Ensure %s is exported (~/.zshrc_secret), then open a new shell\n", pr.KeyEnv)
		case "llamacpp":
			fmt.Fprintf(p.Stdout, "  • Start llama-server at %s; pick a model with /models inside pi\n", pr.BaseURL)
		}
	}
	fmt.Fprintln(p.Stdout, "  • Run `pi` and try /tdd, /plan, /pr; the `tars` agent is available to the subagent tool")
}

// ── Production collaborators ─────────────────────────────────────────────

// FormsPiAsker wraps forms.ShowPiInitForm, detecting the machine's state first.
type FormsPiAsker struct{ pi *components.Pi }

func (a FormsPiAsker) Ask() (forms.PiAnswers, error) {
	detected := forms.PiDetected{Providers: a.pi.DetectProviders(), OllamaModels: a.pi.OllamaModels()}
	return forms.ShowPiInitForm(detected, a.pi.Packages())
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
func (o componentPiOps) StoreKeys(keys map[string]string) error {
	return o.with(o.opts.Pi).StoreKeys(keys)
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
		Asker:  FormsPiAsker{pi: components.NewPi(opts)},
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

  init  choose packages and model providers (ollama, llama.cpp, any OpenAI-compatible
        endpoint), save the choices, and apply them to ~/.pi/agent

'tars pull' keeps the agent, prompts, extension, permissions and AGENTS.md in sync
afterwards; providers are rendered from the saved choices, never from the repo.`,
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
