// Package cmd defines the cobra commands and the orchestration of the setup
// flow. Production wiring lives in NewProductionSetup (the composition root);
// tests construct Setup directly with test doubles. The package holds no
// mutable global state — every dependency Setup needs is injected.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"tars/internal/components"
	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/pkg"
	"tars/internal/pkg/apt"
	"tars/internal/pkg/brew"
	"tars/internal/pkg/rvm"
	"tars/internal/shell"
)

// ── Interfaces (collaborators of Setup) ──────────────────────────────────

// Welcomer shows the welcome screen. Implementations may be the real TUI form
// or a no-op for tests.
type Welcomer interface {
	Show() error
}

// ToolPicker presents the multi-select install form and returns the chosen
// names (a subset of the offered names).
type ToolPicker interface {
	Pick(offered []string) ([]string, error)
}

// ConfigStore loads, saves, and reports the path of the persistent config.
type ConfigStore interface {
	Load() (*config.Config, error)
	Save(*config.Config) error
	Path() string
}

// Registry exposes both the curated install list (for the installer) and the
// list's names (for the picker form).
type Registry interface {
	Installables() []pkg.Installable
	Names() []string
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

// ── Setup ────────────────────────────────────────────────────────────────

// Setup orchestrates the `tars setup` flow. All collaborators are
// injected via interfaces, so tests can substitute spies without mutating
// package state.
type Setup struct {
	Welcome   Welcomer
	Picker    ToolPicker
	Config    ConfigStore
	Registry  Registry
	Installer PackageInstaller
	OhMyZsh   Installer
	P10k      Installer
	Pull      Puller

	Stdout io.Writer
	Stderr io.Writer
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
	s.Installer.InstallAll(available, selected)
	s.runShellInstaller("oh-my-zsh", s.OhMyZsh)
	s.runShellInstaller("powerlevel10k", s.P10k)
	s.runPull()

	fmt.Fprintln(s.Stdout, "\nSetup complete.")
	s.printNextSteps()
	return nil
}

// greet shows the welcome screen; user-aborted is non-fatal.
func (s *Setup) greet() error {
	err := s.Welcome.Show()
	if err == nil || err.Error() == "user aborted" {
		return nil
	}
	return fmt.Errorf("welcome: %w", err)
}

// pickTools offers the registry's names to the picker; user-aborted is non-fatal.
func (s *Setup) pickTools() ([]string, error) {
	selected, err := s.Picker.Pick(s.Registry.Names())
	if err != nil && err.Error() != "user aborted" {
		return nil, fmt.Errorf("tool picker: %w", err)
	}
	return selected, nil
}

func (s *Setup) announceConfig(arch string) {
	fmt.Fprintf(s.Stdout, "Config written to %s\n", s.Config.Path())
	fmt.Fprintf(s.Stdout, "Detected architecture: %s\n", arch)
}

func (s *Setup) runShellInstaller(name string, i Installer) {
	fmt.Fprintf(s.Stdout, "\nInstalling %s...\n", name)
	if err := i.Install(); err != nil {
		fmt.Fprintf(s.Stderr, "  %s: %v\n", name, err)
	}
}

// runPull applies the configs; failures are non-fatal here — the user can
// re-run `tars pull` after fixing the cause.
func (s *Setup) runPull() {
	fmt.Fprintln(s.Stdout, "\nPulling configuration files...")
	if s.Pull.PullAll() != nil {
		fmt.Fprintln(s.Stderr, "  some components failed; re-run `tars pull` after fixing the cause")
	}
}

func (s *Setup) printNextSteps() {
	fmt.Fprintln(s.Stdout, "\nNext steps:")
	fmt.Fprintln(s.Stdout, "  • Open a new terminal — Powerlevel10k launches its configuration wizard")
	fmt.Fprintln(s.Stdout, "  • Run `rustup install stable && rustup default stable` to bootstrap the Rust toolchain")
	fmt.Fprintln(s.Stdout, "  • Run `ghcup tui` to pick GHC / Cabal / HLS versions")
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

// FormsWelcomer wraps forms.ShowWelcome.
type FormsWelcomer struct{}

func (FormsWelcomer) Show() error { return forms.ShowWelcome() }

// FormsPicker wraps forms.ShowInstallForm.
type FormsPicker struct{}

func (FormsPicker) Pick(offered []string) ([]string, error) {
	return forms.ShowInstallForm(offered)
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
	Stdout io.Writer
	Stderr io.Writer
}

func (p IterativeInstaller) InstallAll(available []pkg.Installable, selected []string) {
	picked := stringSet(selected)
	for _, inst := range available {
		if !picked[inst.Name()] {
			continue
		}
		fmt.Fprintf(p.Stdout, "Installing %s...\n", inst.Name())
		if err := inst.Install(p.Stdout, p.Stderr); err != nil {
			fmt.Fprintf(p.Stderr, "  %s: %v\n", inst.Name(), err)
		}
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
	Stdout     io.Writer
	Stderr     io.Writer
}

// PullAll pulls every component, continuing past failures and returning them joined.
func (p SequentialPuller) PullAll() error {
	return runComponents(p.Components, components.Component.Name, components.Component.Pull, p.Stdout, p.Stderr)
}

// runComponents drives one action across a component list, printing progress,
// tolerating per-component failures, and returning them aggregated — the shared
// engine behind SequentialPuller and SequentialPusher (they change together).
func runComponents[T any](items []T, name func(T) string, act func(T) error, stdout, stderr io.Writer) error {
	var failed []error
	for _, c := range items {
		fmt.Fprintf(stdout, "  → %s\n", name(c))
		if err := act(c); err != nil {
			fmt.Fprintf(stderr, "  %s: %v\n", name(c), err)
			failed = append(failed, fmt.Errorf("%s: %w", name(c), err))
		}
	}
	return errors.Join(failed...)
}

// ── Composition root ─────────────────────────────────────────────────────

// NewSetup wires Setup with its collaborators — this is the only place in
// the cli that assembles the dependency graph. The cobra RunE calls it; tests
// either call it too or construct Setup directly with their own collaborators.
func NewSetup(stdout, stderr io.Writer, cfgPath string) (*Setup, error) {
	compOpts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	home := compOpts.Home
	p10kDir := filepath.Join(home, ".oh-my-zsh", "custom", "themes", "powerlevel10k")

	return &Setup{
		Welcome: FormsWelcomer{},
		Picker:  FormsPicker{},
		Config:  NewFileConfigStore(cfgPath),
		Registry: pkg.NewRegistryFactory(
			brew.DefaultRunner(),
			apt.DefaultKit(),
			rvm.NewInstaller(filepath.Join(home, ".rvm"), rvm.DefaultRunner()),
		).For(runtime.GOOS),
		Installer: IterativeInstaller{Stdout: stdout, Stderr: stderr},
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
			Stdout:     stdout,
			Stderr:     stderr,
		},
		Stdout: stdout,
		Stderr: stderr,
	}, nil
}

// ── Cobra command ────────────────────────────────────────────────────────

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initialize this machine with the tars defaults",
	Long: `Full machine bootstrap: pick dev tools to install, install them (brew on
macOS, apt/tarball on Linux), install oh-my-zsh and Powerlevel10k, then apply
all dotfile configs — overwriting ~/.zshrc, ~/.config/nvim, ~/.byobu, and
~/.vimrc, each archived first under ~/.local/state/tars/backups/<component>/vN.
Also seeds ~/.zshrc_secret from a template when absent and saves the tool
selection to the tars config file.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		s, err := NewSetup(cmd.OutOrStdout(), cmd.ErrOrStderr(), configPath())
		if err != nil {
			return err
		}
		return s.Run()
	},
}
