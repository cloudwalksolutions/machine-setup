// Package cmd defines the cobra commands and the orchestration of the setup
// flow. Production wiring lives in NewProductionSetup (the composition root);
// tests construct Setup directly with test doubles. The package holds no
// mutable global state — every dependency Setup needs is injected.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/config"
	"tars/internal/pkg"
	"tars/internal/pkg/apt"
	"tars/internal/pkg/brew"
	"tars/internal/pkg/npm"
	"tars/internal/pkg/registry"
	"tars/internal/pkg/rvm"
	"tars/internal/report"
	"tars/internal/shell"
	"tars/internal/tui"
)

// ── Interfaces (collaborators of Setup) ──────────────────────────────────

// Welcomer shows the welcome screen. Implementations may be the real TUI form
// or a no-op for tests.
type Welcomer interface {
	Welcome() error
}

// WizardPicker presents the agent wizards and returns the chosen subset.
type WizardPicker interface {
	PickWizards(offered []string) ([]string, error)
}

// InstallPicker presents the install form over the catalog and returns the chosen names.
type InstallPicker interface {
	PickTools(offered []pkg.ToolInfo) ([]string, error)
}

// ConfigStore loads, saves, and reports the path of the persistent config.
type ConfigStore interface {
	Load() (*config.Config, error)
	Save(*config.Config) error
	Path() string
}

// Registry exposes both the curated install list (for the installer) and the
// catalog (for the picker form).
type Registry interface {
	Installables() []pkg.Installable
	Catalog() []pkg.ToolInfo
}

// PackageInstaller installs the named subset of available installables.
// Failures are reported inline; the loop continues.
type PackageInstaller interface {
	InstallAll(available []pkg.Installable, selected []string)
}

// Installer is the single-op contract for things like oh-my-zsh and powerlevel10k.
type Installer interface {
	Install() error
}

// Puller pulls every dotfile component, reporting failures inline and
// returning an aggregate error naming the components that failed.
type Puller interface {
	PullAll() error
}

// NamedInit is one agent wizard (`tars init claude`, `tars init pi`) the bootstrap can run.
type NamedInit struct {
	Name string
	Run  func() error
}

// ── Setup ────────────────────────────────────────────────────────────────

// Setup orchestrates the `tars init` flow. All collaborators are
// injected via interfaces, so tests can substitute spies without mutating
// package state.
type Setup struct {
	Welcome   Welcomer
	Picker    InstallPicker
	Config    ConfigStore
	Registry  Registry
	Installer PackageInstaller
	OhMyZsh   Installer
	P10k      Installer
	Pull      Puller
	Wizards   WizardPicker
	Inits     []NamedInit
	Report    report.Reporter
}

// Run drives the orchestration. Each step is a single method call on an
// injected collaborator; failures are non-fatal where the user can still
// recover from a partial run, fatal where they cannot.
func (s *Setup) Run() error {
	if err := s.greet(); err != nil {
		return err
	}

	cfg, err := s.Config.Load()
	if err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	available := s.Registry.Installables()
	selected, err := s.pickTools()
	if err != nil {
		return err
	}

	cfg.Packages = packagesFromNames(selected)
	if err := s.Config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	s.announceConfig(cfg.Architecture)
	s.Report.StepStarted("Installing packages", len(selected))
	s.Installer.InstallAll(available, selected)
	s.runShellInstaller("Installing oh-my-zsh", s.OhMyZsh)
	s.runShellInstaller("Installing powerlevel10k", s.P10k)
	s.runPull()
	s.runWizards()

	s.Report.Note("\nSetup complete.")
	s.printNextSteps()
	return nil
}

// runWizards offers the agent wizards and runs the picked ones; each failure is
// reported inline and the next wizard still runs.
func (s *Setup) runWizards() {
	names := make([]string, len(s.Inits))
	for i, in := range s.Inits {
		names[i] = in.Name
	}
	picked, err := s.Wizards.PickWizards(names)
	if err != nil {
		return
	}
	for _, in := range s.Inits {
		if !slices.Contains(picked, in.Name) {
			continue
		}
		title := "Initializing " + in.Name
		s.Report.StepStarted(title, 0)
		s.Report.StepDone(title, in.Run())
	}
}

// greet shows the welcome screen; user-aborted is non-fatal.
func (s *Setup) greet() error {
	err := s.Welcome.Welcome()
	if err == nil || err.Error() == "user aborted" {
		return nil
	}
	return fmt.Errorf("welcome: %w", err)
}

// pickTools probes the machine, then offers the catalog to the picker; user-aborted is non-fatal.
func (s *Setup) pickTools() ([]string, error) {
	const probing = "Checking installed tools"
	s.Report.StepStarted(probing, 0)
	catalog := s.Registry.Catalog()
	s.Report.StepDone(probing, nil)
	selected, err := s.Picker.PickTools(catalog)
	if err != nil && err.Error() != "user aborted" {
		return nil, fmt.Errorf("tool picker: %w", err)
	}
	return selected, nil
}

func (s *Setup) announceConfig(arch string) {
	s.Report.Note("Config written to " + s.Config.Path())
	s.Report.Note("Detected architecture: " + arch)
}

func (s *Setup) runShellInstaller(title string, i Installer) {
	s.Report.StepStarted(title, 0)
	s.Report.StepDone(title, i.Install())
}

// runPull applies the configs; failures are non-fatal here — the user can
// re-run `tars pull` after fixing the cause.
func (s *Setup) runPull() {
	const pulling = "Pulling configuration files"
	s.Report.StepStarted(pulling, 0)
	if s.Pull.PullAll() != nil {
		s.Report.StepDone(pulling, errors.New("some components failed; re-run `tars pull` after fixing the cause"))
	}
}

func (s *Setup) printNextSteps() {
	s.Report.Note("\nNext steps:")
	s.Report.Note("  • Open a new terminal — Powerlevel10k launches its configuration wizard")
}

// packagesFromNames builds the persistable config slice from selected names.
func packagesFromNames(names []string) []config.Package {
	out := make([]config.Package, len(names))
	for i, n := range names {
		out[i] = config.Package{Name: n}
	}
	return out
}

// ── Production collaborators ─────────────────────────────────────────────

// runInit runs an init flow inside the TUI, or headless when TARS_NO_FORM is set.
func runInit(cmd *cobra.Command, flow func(report.Reporter, tui.Asker) error) error {
	return tui.Run(os.Getenv("TARS_NO_FORM") != "", cmd.OutOrStdout(), cmd.ErrOrStderr(), flow)
}

// FileConfigStore reads/writes the YAML config at a fixed path.
type FileConfigStore struct{ path string }

func NewFileConfigStore(path string) FileConfigStore    { return FileConfigStore{path: path} }
func (s FileConfigStore) Load() (*config.Config, error) { return config.Init(s.path) }
func (s FileConfigStore) Save(cfg *config.Config) error { return config.Save(s.path, cfg) }
func (s FileConfigStore) Path() string                  { return s.path }

// IterativeInstaller is the production PackageInstaller. It filters the
// available list to selected names, then calls Install on each.
type IterativeInstaller struct {
	Report report.Reporter
}

func (p IterativeInstaller) InstallAll(available []pkg.Installable, selected []string) {
	picked := stringSet(selected)
	stdout, stderr := p.Report.Output()
	for _, inst := range available {
		if !picked[inst.Name()] {
			continue
		}
		p.Report.ItemStarted(inst.Name())
		p.Report.ItemDone(inst.Name(), inst.Install(stdout, stderr))
	}
}

func stringSet(s []string) map[string]bool {
	set := make(map[string]bool, len(s))
	for _, v := range s {
		set[v] = true
	}
	return set
}

// SequentialPuller is the production Puller. It iterates the configured
// component list, printing progress and capturing per-component failures.
type SequentialPuller struct {
	Components []components.Component
	Report     report.Reporter
}

// PullAll pulls every component, continuing past failures and returning them joined.
func (p SequentialPuller) PullAll() error {
	return runComponents(p.Components, components.Component.Name, components.Component.Pull, p.Report)
}

// runComponents drives one action across a component list, reporting each item,
// tolerating per-component failures, and returning them aggregated — the shared
// engine behind SequentialPuller and SequentialPusher (they change together).
func runComponents[T any](items []T, name func(T) string, act func(T) error, r report.Reporter) error {
	var failed []error
	for _, c := range items {
		r.ItemStarted(name(c))
		err := act(c)
		r.ItemDone(name(c), err)
		if err != nil {
			failed = append(failed, fmt.Errorf("%s: %w", name(c), err))
		}
	}
	return errors.Join(failed...)
}

// ── Composition root ─────────────────────────────────────────────────────

// NewSetup wires Setup with its collaborators — this is the only place in
// the cli that assembles the dependency graph. The cobra RunE calls it; tests
// either call it too or construct Setup directly with their own collaborators.
func NewSetup(r report.Reporter, ask tui.Asker, cfgPath string) (*Setup, error) {
	stdout, stderr := r.Output()
	compOpts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	home := compOpts.Home
	p10kDir := filepath.Join(home, ".oh-my-zsh", "custom", "themes", "powerlevel10k")

	return &Setup{
		Welcome: ask,
		Picker:  ask,
		Config:  NewFileConfigStore(cfgPath),
		Registry: registry.NewRegistryFactory(
			brew.DefaultRunner(),
			apt.DefaultKit(),
			pkg.PathProbe{},
			rvm.NewInstaller(filepath.Join(home, ".rvm"), rvm.DefaultRunner()),
			npm.NewPackage("gemini-cli", "@google/gemini-cli", "Google's Gemini CLI agent", npm.DefaultRunner()),
			pkg.NewScriptInstaller(
				"claude-code",
				"Anthropic's Claude Code agent",
				filepath.Join(home, ".local", "bin", "claude"),
				[]string{"bash", "-c", "curl -fsSL https://claude.ai/install.sh | bash"},
				nil,
			),
			npm.NewPackage("pi", "@earendil-works/pi-coding-agent", "pi coding agent", npm.DefaultRunner()),
		).For(runtime.GOOS),
		Installer: IterativeInstaller{Report: r},
		OhMyZsh: shell.OhMyZshInstaller{
			Dir:    filepath.Join(home, ".oh-my-zsh"),
			Runner: shell.DefaultRunner(),
			Stdout: stdout,
			Stderr: stderr,
		},
		P10k: shell.Powerlevel10kInstaller{
			Dir:    p10kDir,
			Runner: shell.DefaultP10kRunner(p10kDir),
			Stdout: stdout,
			Stderr: stderr,
		},
		Pull: SequentialPuller{
			Components: components.AllPullable(compOpts),
			Report:     r,
		},
		Wizards: ask,
		Inits: []NamedInit{
			{Name: "claude", Run: func() error {
				c, err := NewClaudeInit(r, ask)
				if err != nil {
					return err
				}
				return c.Run()
			}},
			{Name: "pi", Run: func() error {
				p, err := NewPiInit(r, ask)
				if err != nil {
					return err
				}
				return p.Run()
			}},
		},
		Report: r,
	}, nil
}

// ── Cobra command ────────────────────────────────────────────────────────

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Initialize this machine with the tars defaults (alias: i)",
	Args:    cobra.NoArgs,
	Long: `Full machine bootstrap: pick dev tools to install, install them (brew on
macOS, apt/tarball on Linux), install oh-my-zsh and Powerlevel10k, then apply
all dotfile configs — overwriting ~/.zshrc, ~/.config/nvim, ~/.byobu, and
~/.vimrc, each archived first under ~/.local/state/tars/backups/<component>/vN.
Also installs fonts, points iTerm2/Terminal.app at them (macOS), renders any
configured profiles, seeds ~/.zshrc_secret and ~/.zprofile_local from templates
when absent, and saves the tool selection to ~/.config/tars/config.yaml.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runInit(cmd, func(r report.Reporter, ask tui.Asker) error {
			s, err := NewSetup(r, ask, configPath())
			if err != nil {
				return err
			}
			return s.Run()
		})
	},
}
