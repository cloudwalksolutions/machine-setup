package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudwalk/machine-setup/internal/forms"
	"github.com/cloudwalk/machine-setup/internal/sessions"
	"github.com/spf13/cobra"
)

// SessionStore loads the sessions config file.
type SessionStore interface {
	Load() (sessions.File, error)
	Path() string
	Seed() error
}

// SessionOpener drives byobu for configured sessions.
type SessionOpener interface {
	Open(sessions.Session) error
	OpenAll(sessions.File) error
	Fresh(name string) error
}

// SessionPicker asks the user to choose one of the offered options.
type SessionPicker interface {
	Pick(options []string) (string, error)
}

const (
	pickAll   = "(all)"
	pickFresh = "(fresh)"
)

// Sessions orchestrates the `tars sessions` verbs over injected collaborators.
type Sessions struct {
	Store  SessionStore
	Opener SessionOpener
	Picker SessionPicker
	EditFn func(path string) error // real: $EDITOR on the config file
	Stdout io.Writer
	Stderr io.Writer
}

// PickAndRun shows the session picker and runs the chosen action.
func (s *Sessions) PickAndRun() error {
	f, err := s.load()
	if err != nil {
		return err
	}
	choice, err := s.Picker.Pick(append(sessionNames(f), pickAll, pickFresh))
	if err != nil {
		if err.Error() == "user aborted" { // huh's Ctrl+C, non-fatal (as in setup)
			return nil
		}
		return err
	}
	switch choice {
	case pickAll:
		return s.Opener.OpenAll(f)
	case pickFresh:
		return s.Opener.Fresh("")
	default:
		return s.Open(choice)
	}
}

// Edit seeds the config file if missing and opens it in the editor.
func (s *Sessions) Edit() error {
	if err := s.Store.Seed(); err != nil {
		return err
	}
	return s.EditFn(s.Store.Path())
}

// load wraps Store.Load with guidance when no config exists yet.
func (s *Sessions) load() (sessions.File, error) {
	f, err := s.Store.Load()
	if errors.Is(err, os.ErrNotExist) {
		return f, fmt.Errorf("no sessions configured — run `tars sessions edit` to create %s", s.Store.Path())
	}
	return f, err
}

// Open opens the configured session with the given name.
func (s *Sessions) Open(name string) error {
	f, err := s.load()
	if err != nil {
		return err
	}
	for _, sess := range f.Sessions {
		if sess.Name == name {
			return s.Opener.Open(sess)
		}
	}
	return fmt.Errorf("unknown session %q (configured: %s)", name, strings.Join(sessionNames(f), ", "))
}

// All opens every configured session, attaching to the first.
func (s *Sessions) All() error {
	f, err := s.load()
	if err != nil {
		return err
	}
	return s.Opener.OpenAll(f)
}

// New opens a fresh byobu session (optionally named); reads no config.
func (s *Sessions) New(name string) error {
	return s.Opener.Fresh(name)
}

// List prints the configured sessions and their dirs.
func (s *Sessions) List() error {
	f, err := s.load()
	if err != nil {
		return err
	}
	for _, sess := range f.Sessions {
		fmt.Fprintf(s.Stdout, "%s\n", sess.Name)
		for _, dir := range sess.Dirs {
			fmt.Fprintf(s.Stdout, "  %s\n", dir)
		}
	}
	return nil
}

func sessionNames(f sessions.File) []string {
	names := make([]string, len(f.Sessions))
	for i, s := range f.Sessions {
		names[i] = s.Name
	}
	return names
}

// FileSessionStore is the production SessionStore over the yaml config file.
type FileSessionStore struct{ path string }

// Load reads the sessions config from disk.
func (s FileSessionStore) Load() (sessions.File, error) { return sessions.Load(s.path) }

// Path returns the config file location.
func (s FileSessionStore) Path() string { return s.path }

// Seed writes the example config if none exists.
func (s FileSessionStore) Seed() error { return sessions.Seed(s.path) }

// FormsSessionPicker adapts the huh single-select to SessionPicker.
type FormsSessionPicker struct{}

// Pick shows the picker and returns the chosen option.
func (FormsSessionPicker) Pick(options []string) (string, error) {
	return forms.ShowSessionPicker(options)
}

// defaultEditor opens path in $EDITOR (fallback: vim) with the terminal attached.
func defaultEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	c := exec.Command(editor, path)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

// NewSessions wires the production collaborators for the sessions verbs.
func NewSessions(stdout, stderr io.Writer) (*Sessions, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("locating home dir: %w", err)
	}
	return &Sessions{
		Store: FileSessionStore{path: sessions.DefaultPath()},
		Opener: sessions.Launcher{
			Run:       sessions.DefaultRunner(),
			LookupEnv: os.LookupEnv,
			Home:      home,
			Stdout:    stdout,
			Stderr:    stderr,
		},
		Picker: FormsSessionPicker{},
		EditFn: defaultEditor,
		Stdout: stdout,
		Stderr: stderr,
	}, nil
}

// sessionsRunE builds the production Sessions and runs fn on it.
func sessionsRunE(fn func(s *Sessions, args []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		s, err := NewSessions(cmd.OutOrStdout(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		return fn(s, args)
	}
}

var sessionsCmd = &cobra.Command{
	Use:     "sessions",
	Aliases: []string{"s", "by"},
	Short:   "Open byobu sessions from a simple config of dirs (aliases: s, by)",
	Long: `Manage a set of byobu sessions declared in a simple yaml config
(~/.config/.machine-setup/sessions.yaml): each session is a name plus a list
of dirs, one window per dir. Opening is idempotent — existing sessions are
attached, never duplicated. Run bare for an interactive picker.`,
	RunE: sessionsRunE(func(s *Sessions, _ []string) error { return s.PickAndRun() }),
}

var sessionsOpenCmd = &cobra.Command{
	Use:     "open <name>",
	Aliases: []string{"o"},
	Short:   "Open one configured session (alias: o)",
	Args:    cobra.ExactArgs(1),
	RunE:    sessionsRunE(func(s *Sessions, args []string) error { return s.Open(args[0]) }),
}

var sessionsAllCmd = &cobra.Command{
	Use:     "all",
	Aliases: []string{"a"},
	Short:   "Open every configured session and attach to the first (alias: a)",
	Args:    cobra.NoArgs,
	RunE:    sessionsRunE(func(s *Sessions, _ []string) error { return s.All() }),
}

var sessionsNewCmd = &cobra.Command{
	Use:     "new [name]",
	Aliases: []string{"n"},
	Short:   "Open a fresh byobu session, optionally named (alias: n)",
	Args:    cobra.MaximumNArgs(1),
	RunE: sessionsRunE(func(s *Sessions, args []string) error {
		name := ""
		if len(args) == 1 {
			name = args[0]
		}
		return s.New(name)
	}),
}

var sessionsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List the configured sessions and their dirs (aliases: l, ls)",
	Args:    cobra.NoArgs,
	RunE:    sessionsRunE(func(s *Sessions, _ []string) error { return s.List() }),
}

var sessionsEditCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Edit the sessions config in $EDITOR, seeding an example if missing (alias: e)",
	Args:    cobra.NoArgs,
	RunE:    sessionsRunE(func(s *Sessions, _ []string) error { return s.Edit() }),
}

func init() {
	// Runtime errors (unknown session, byobu failures) shouldn't dump usage.
	sessionsCmd.SilenceUsage = true
	for _, c := range []*cobra.Command{
		sessionsOpenCmd, sessionsAllCmd, sessionsNewCmd, sessionsListCmd, sessionsEditCmd,
	} {
		c.SilenceUsage = true
		sessionsCmd.AddCommand(c)
	}
}
