package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"tars/internal/components"
	"tars/internal/forms"
	"tars/internal/paths"
	"tars/internal/profiles"
)

// ProfileStore loads and grows the profiles config file.
type ProfileStore interface {
	Load() (profiles.File, error)
	Path() string
	Seed() error
	Append(profiles.Profile) error
}

// Profiles orchestrates the `tars profiles` verbs over injected collaborators.
type Profiles struct {
	Home      string
	Store     ProfileStore
	Picker    SessionPicker
	Prompt    func(defaults profiles.Profile) (profiles.Profile, error)
	EditFn    func(path string) error
	Apply     func() error
	SetActive func(alias string) error
	ActiveFn  func() (string, error)
	Run       func(name string, args ...string) error
	LookupEnv func(key string) (string, bool)
	Stdout    io.Writer
	Stderr    io.Writer
}

// PickAndRun offers the aliases and activates the chosen profile.
func (p *Profiles) PickAndRun() error {
	f, err := p.load()
	if err != nil {
		return err
	}
	choice, err := p.Picker.Pick(aliases(f))
	if err != nil {
		if err.Error() == "user aborted" {
			return nil
		}
		return err
	}
	return p.Use(choice)
}

func aliases(f profiles.File) []string {
	out := make([]string, len(f.Profiles))
	for i, prof := range f.Profiles {
		out[i] = prof.Alias
	}
	return out
}

// List prints every profile with its dir and email, marking the active one with '*'.
func (p *Profiles) List() error {
	f, err := p.load()
	if err != nil {
		return err
	}
	active, err := p.ActiveFn()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(p.Stdout, 0, 0, 2, ' ', 0)
	for _, prof := range f.Profiles {
		marker := " "
		if prof.Alias == active {
			marker = "*"
		}
		fmt.Fprintf(w, "%s %s\t%s\t%s\t%s\n", marker, prof.Alias, prof.Name, prof.Dir(f.ProjectsDir), prof.Email)
	}
	return w.Flush()
}

// Show prints one profile's identity and derived paths; with no id, the active one.
func (p *Profiles) Show(id string) error {
	f, err := p.load()
	if err != nil {
		return err
	}
	if id == "" {
		if id, err = p.ActiveFn(); err != nil {
			return err
		}
		if id == "" {
			return errors.New("no active profile — run `tars profiles use <alias>` or `tars profiles show <alias>`")
		}
	}
	prof, found := f.Find(id)
	if !found {
		return fmt.Errorf("unknown profile %q (configured: %s)", id, strings.Join(profileLabels(f), ", "))
	}
	fmt.Fprintf(p.Stdout, "%s (%s)\n", prof.Name, prof.Alias)
	fmt.Fprintf(p.Stdout, "  dir     %s\n", prof.Dir(f.ProjectsDir))
	fmt.Fprintf(p.Stdout, "  key     %s\n", prof.KeyPath(p.Home))
	fmt.Fprintf(p.Stdout, "  email   %s\n", prof.Email)
	fmt.Fprintf(p.Stdout, "  github  %s\n", prof.GitHub)
	return nil
}

// Edit seeds the config if missing, opens it in the editor, and re-renders.
func (p *Profiles) Edit() error {
	if err := p.Store.Seed(); err != nil {
		return err
	}
	if err := p.EditFn(p.Store.Path()); err != nil {
		return err
	}
	return p.Apply()
}

// Add prompts for a new profile, appends it to the config, and re-renders.
func (p *Profiles) Add() error {
	prof, err := p.Prompt(profiles.Profile{})
	if err != nil {
		if err.Error() == "user aborted" {
			return nil
		}
		return err
	}
	if err := p.Store.Append(prof); err != nil {
		return err
	}
	return p.Apply()
}

// Use makes the profile the machine-wide default and points gh at its account.
func (p *Profiles) Use(id string) error {
	f, err := p.load()
	if err != nil {
		return err
	}
	prof, found := f.Find(id)
	if !found {
		return fmt.Errorf("unknown profile %q (configured: %s)", id, strings.Join(profileLabels(f), ", "))
	}
	if err := p.SetActive(prof.Alias); err != nil {
		return err
	}
	if err := p.Apply(); err != nil {
		return err
	}
	if _, exported := p.LookupEnv("GITHUB_TOKEN"); exported {
		fmt.Fprintf(p.Stderr, "  GITHUB_TOKEN is exported and overrides gh's account; move it into %s.env\n", prof.Alias)
	}
	if err := p.Run("gh", "auth", "switch", "--user", prof.GitHub); err != nil {
		fmt.Fprintf(p.Stderr, "  gh account not switched (%v); git identity is active regardless\n", err)
	}
	return nil
}

// FileProfileStore is the production ProfileStore over the yaml config file.
type FileProfileStore struct {
	path string
	home string
}

// Load reads the profiles config from disk.
func (s FileProfileStore) Load() (profiles.File, error) { return profiles.Load(s.path, s.home) }

// Path returns the config file location.
func (s FileProfileStore) Path() string { return s.path }

// Seed writes the example config if none exists.
func (s FileProfileStore) Seed() error { return profiles.Seed(s.path) }

// Append adds a profile to the config file.
func (s FileProfileStore) Append(p profiles.Profile) error { return profiles.Append(s.path, p) }

// NewProfiles wires the production collaborators for the profiles verbs.
func NewProfiles(stdout, stderr io.Writer) (*Profiles, error) {
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	dir := paths.For(opts.RepoRoot, opts.Home).Profiles.Dir
	return &Profiles{
		Home:      opts.Home,
		Store:     FileProfileStore{path: profiles.DefaultPath(), home: opts.Home},
		Picker:    FormsSessionPicker{},
		Prompt:    forms.ShowProfileForm,
		EditFn:    defaultEditor,
		Apply:     components.NewProfiles(opts).Pull,
		SetActive: func(alias string) error { return profiles.SetActive(dir, alias) },
		ActiveFn:  func() (string, error) { return profiles.Active(dir) },
		Run:       runAttached,
		LookupEnv: os.LookupEnv,
		Stdout:    stdout,
		Stderr:    stderr,
	}, nil
}

// runAttached runs a command with the terminal attached (gh may prompt).
func runAttached(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

// profilesRunE builds the production Profiles and runs fn on it.
func profilesRunE(fn func(p *Profiles, args []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		p, err := NewProfiles(cmd.OutOrStdout(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		return fn(p, args)
	}
}

var profilesCmd = &cobra.Command{
	Use:     "profiles",
	Aliases: []string{"p"},
	Short:   "Switch git/GitHub/SSH identity per project dir (alias: p)",
	Long: `Manage account identities declared in ~/.config/tars/profiles.yaml.
Each profile (name + alias + email + github) owns <projects_dir>/<name> and signs
with ~/.ssh/id_rsa.<alias>. tars renders a per-profile gitconfig, a managed
includeIf block in ~/.gitconfig, and a private <alias>.env your shell sources
while that profile is active. Run bare for an interactive picker.`,
	RunE: profilesRunE(func(p *Profiles, _ []string) error { return p.PickAndRun() }),
}

var profilesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List profiles, marking the active one (aliases: l, ls)",
	Args:    cobra.NoArgs,
	RunE:    profilesRunE(func(p *Profiles, _ []string) error { return p.List() }),
}

var profilesUseCmd = &cobra.Command{
	Use:   "use <alias|name>",
	Short: "Make a profile the default outside its dir and switch gh to its account",
	Args:  cobra.ExactArgs(1),
	RunE:  profilesRunE(func(p *Profiles, args []string) error { return p.Use(args[0]) }),
}

var profilesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a profile interactively",
	Args:  cobra.NoArgs,
	RunE:  profilesRunE(func(p *Profiles, _ []string) error { return p.Add() }),
}

var profilesEditCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Edit profiles.yaml in $EDITOR (seeding an example), then re-render (alias: e)",
	Args:    cobra.NoArgs,
	RunE:    profilesRunE(func(p *Profiles, _ []string) error { return p.Edit() }),
}

var profilesShowCmd = &cobra.Command{
	Use:   "show [alias|name]",
	Short: "Show one profile's identity and derived paths (default: the active one)",
	Args:  cobra.MaximumNArgs(1),
	RunE: profilesRunE(func(p *Profiles, args []string) error {
		id := ""
		if len(args) == 1 {
			id = args[0]
		}
		return p.Show(id)
	}),
}

var profilesApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Re-render the git and shell files from profiles.yaml",
	Args:  cobra.NoArgs,
	RunE:  profilesRunE(func(p *Profiles, _ []string) error { return p.Apply() }),
}

func init() {
	profilesCmd.SilenceUsage = true
	for _, c := range []*cobra.Command{
		profilesListCmd, profilesUseCmd, profilesAddCmd, profilesEditCmd, profilesShowCmd, profilesApplyCmd,
	} {
		c.SilenceUsage = true
		profilesCmd.AddCommand(c)
	}
}

func (p *Profiles) load() (profiles.File, error) {
	f, err := p.Store.Load()
	if errors.Is(err, os.ErrNotExist) {
		return f, fmt.Errorf("no profiles configured — run `tars profiles edit` to create %s", p.Store.Path())
	}
	return f, err
}

func profileLabels(f profiles.File) []string {
	labels := make([]string, len(f.Profiles))
	for i, prof := range f.Profiles {
		labels[i] = fmt.Sprintf("%s (%s)", prof.Name, prof.Alias)
	}
	return labels
}
