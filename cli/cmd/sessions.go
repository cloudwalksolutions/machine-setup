package cmd

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"tars/internal/forms"
	"tars/internal/profiles"
	"tars/internal/projects"
	"tars/internal/sessions"
	"tars/internal/tui"
)

// LiveSessions drives the running byobu sessions.
type LiveSessions interface {
	List() ([]sessions.Live, error)
	Attach(name string) error
	Open(name string, windows []sessions.Window) error
	Kill(name string) error
	Rename(old, name string) error
}

// ProjectCatalog finds the projects and their session templates.
type ProjectCatalog interface {
	Projects() ([]projects.Project, error)
	Templates() (projects.File, error)
	Path() string
	Seed() error
	SaveTemplates(projects.File) error
}

// SessionPicker asks the user to choose one of the offered options.
type SessionPicker interface {
	Pick(options []string) (string, error)
}

// Sessions orchestrates the `tars sessions` verbs over injected collaborators.
type Sessions struct {
	Live     LiveSessions
	Catalog  ProjectCatalog
	Picker   SessionPicker
	EditFn   func(path string) error // real: $EDITOR on the projects file
	Getwd    func() (string, error)
	Manage   func(tui.Manager) (string, error) // real: the interactive manager; returns what to open
	Headless bool                              // TARS_NO_FORM: no terminal to manage in
	Stdout   io.Writer
}

const pickNew = " (new)"

// PickAndOpen offers the running sessions and the projects without one, then opens the choice.
func (s *Sessions) PickAndOpen() error {
	live, err := s.Live.List()
	if err != nil {
		return err
	}
	found, err := s.Catalog.Projects()
	if err != nil {
		return err
	}
	running := map[string]bool{}
	var options []string
	for _, l := range live {
		running[l.Name] = true
		options = append(options, l.Name)
	}
	for _, p := range found {
		if !running[p.Name] {
			options = append(options, p.Name+pickNew)
		}
	}
	choice, err := s.Picker.Pick(options)
	if err != nil && err.Error() == "user aborted" { // huh's Ctrl+C, non-fatal (as in setup)
		return nil
	}
	if err != nil {
		return err
	}
	return s.Open(strings.TrimSuffix(choice, pickNew))
}

// Interactive runs the sessions manager, then opens what the user picked in it.
func (s *Sessions) Interactive() error {
	if s.Headless {
		return s.List()
	}
	chosen, err := s.Manage(sessionsManager{s})
	if err != nil || chosen == "" {
		return err
	}
	return s.Open(chosen)
}

// sessionsManager is the tui.Manager over the live sessions and the project catalog.
type sessionsManager struct{ s *Sessions }

func (m sessionsManager) Rows() ([]tui.Row, error) {
	live, err := m.s.Live.List()
	if err != nil {
		return nil, err
	}
	found, err := m.s.Catalog.Projects()
	if err != nil {
		return nil, err
	}
	dirs := map[string]string{}
	for _, p := range found {
		dirs[p.Name] = p.Dir
	}
	var rows []tui.Row
	for _, l := range live {
		rows = append(rows, tui.Row{Name: l.Name, Live: true, Windows: l.Windows, Attached: l.Attached, Dir: dirs[l.Name]})
		delete(dirs, l.Name)
	}
	for _, p := range found {
		if dir, idle := dirs[p.Name]; idle {
			rows = append(rows, tui.Row{Name: p.Name, Dir: dir})
		}
	}
	return rows, nil
}

func (m sessionsManager) Kill(name string) error { return m.s.Live.Kill(name) }

func (m sessionsManager) Rename(old, name string) error { return m.s.Live.Rename(old, name) }

func (m sessionsManager) Template(project string) (projects.Template, error) {
	found, err := m.s.Catalog.Projects()
	if err != nil {
		return projects.Template{}, err
	}
	templates, err := m.s.Catalog.Templates()
	if err != nil {
		return projects.Template{}, err
	}
	for _, p := range found {
		if p.Name == project {
			return templates.Template(p), nil
		}
	}
	return projects.Template{}, fmt.Errorf("no project %q", project)
}

func (m sessionsManager) SaveTemplate(project string, t projects.Template) error {
	templates, err := m.s.Catalog.Templates()
	if err != nil {
		return err
	}
	saved := maps.Clone(templates.Projects)
	if saved == nil {
		saved = map[string]projects.Template{}
	}
	saved[project] = t
	return m.s.Catalog.SaveTemplates(projects.File{Projects: saved})
}

// List prints the running sessions, marking the attached one with '*' and naming their project.
func (s *Sessions) List() error {
	live, err := s.Live.List()
	if err != nil {
		return err
	}
	found, err := s.Catalog.Projects()
	if err != nil {
		return err
	}
	dirs := map[string]string{}
	for _, p := range found {
		dirs[p.Name] = p.Dir
	}
	w := tabwriter.NewWriter(s.Stdout, 0, 0, 2, ' ', 0)
	for _, l := range live {
		marker := " "
		if l.Attached {
			marker = "*"
		}
		fmt.Fprintf(w, "%s %s\t%d windows", marker, l.Name, l.Windows)
		if dir, ok := dirs[l.Name]; ok {
			fmt.Fprintf(w, "\tproject %s", dir)
		}
		fmt.Fprintln(w)
	}
	return w.Flush()
}

// Projects prints every project, marking running ones with '●' and naming its template.
func (s *Sessions) Projects() error {
	live, err := s.Live.List()
	if err != nil {
		return err
	}
	found, err := s.Catalog.Projects()
	if err != nil {
		return err
	}
	templates, err := s.Catalog.Templates()
	if err != nil {
		return err
	}
	running := map[string]bool{}
	for _, l := range live {
		running[l.Name] = true
	}
	w := tabwriter.NewWriter(s.Stdout, 0, 0, 2, ' ', 0)
	for _, p := range found {
		marker, template := " ", "default"
		if running[p.Name] {
			marker = "●"
		}
		if _, ok := templates.Projects[p.Name]; ok {
			template = "template"
		}
		fmt.Fprintf(w, "%s %s\t%s\t%s\n", marker, p.Name, p.Dir, template)
	}
	return w.Flush()
}

// Edit seeds the projects file if missing and opens it in the editor.
func (s *Sessions) Edit() error {
	if err := s.Catalog.Seed(); err != nil {
		return err
	}
	return s.EditFn(s.Catalog.Path())
}

// Kill ends a running session.
func (s *Sessions) Kill(name string) error { return s.Live.Kill(name) }

// Rename gives a running session a new name.
func (s *Sessions) Rename(old, name string) error { return s.Live.Rename(old, name) }

// Open attaches to the running session called target, or starts the project of that name;
// an empty target means the project containing the current dir.
func (s *Sessions) Open(target string) error {
	if target == "" {
		here, err := s.projectHere()
		if err != nil {
			return err
		}
		target = here
	}
	live, err := s.Live.List()
	if err != nil {
		return err
	}
	var running []string
	for _, l := range live {
		if l.Name == target {
			return s.Live.Attach(target)
		}
		running = append(running, l.Name)
	}
	found, err := s.Catalog.Projects()
	if err != nil {
		return err
	}
	templates, err := s.Catalog.Templates()
	if err != nil {
		return err
	}
	var names []string
	for _, p := range found {
		if p.Name == target {
			return s.Live.Open(p.Name, templates.Windows(p))
		}
		names = append(names, p.Name)
	}
	return fmt.Errorf("no session or project %q (running: %s; projects: %s)",
		target, strings.Join(running, ", "), strings.Join(names, ", "))
}

func (s *Sessions) projectHere() (string, error) {
	cwd, err := s.Getwd()
	if err != nil {
		return "", err
	}
	found, err := s.Catalog.Projects()
	if err != nil {
		return "", err
	}
	for _, p := range found {
		if rel, err := filepath.Rel(p.Dir, cwd); err == nil && !strings.HasPrefix(rel, "..") {
			return p.Name, nil
		}
	}
	return "", fmt.Errorf("%s is not inside a project", cwd)
}

// FileProjectCatalog finds projects under the profiles' projects_dir and templates in the projects file.
type FileProjectCatalog struct {
	ProfilesPath string
	File         string
	Home         string
}

// Projects discovers the git repos under projects_dir.
func (c FileProjectCatalog) Projects() ([]projects.Project, error) {
	f, err := profiles.Load(c.ProfilesPath, c.Home)
	if errors.Is(err, os.ErrNotExist) {
		return projects.Discover(profiles.DefaultProjectsDir(c.Home))
	}
	if err != nil {
		return nil, err
	}
	return projects.Discover(f.ProjectsDir)
}

// Templates loads the projects file.
func (c FileProjectCatalog) Templates() (projects.File, error) { return projects.Load(c.File) }

// Path is the projects file location.
func (c FileProjectCatalog) Path() string { return c.File }

// Seed writes the example projects file if none exists.
func (c FileProjectCatalog) Seed() error { return projects.Seed(c.File) }

// SaveTemplates writes the projects file.
func (c FileProjectCatalog) SaveTemplates(f projects.File) error { return projects.Save(c.File, f) }

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
func NewSessions(stdout io.Writer) *Sessions {
	home, _ := os.UserHomeDir()
	return &Sessions{
		Live: sessions.Launcher{
			Run:       sessions.DefaultRunner(),
			Output:    sessions.DefaultOutput(),
			LookupEnv: os.LookupEnv,
		},
		Catalog:  FileProjectCatalog{ProfilesPath: profiles.DefaultPath(), File: projects.DefaultPath(), Home: home},
		Picker:   FormsSessionPicker{},
		EditFn:   defaultEditor,
		Getwd:    os.Getwd,
		Manage:   tui.RunSessions,
		Headless: os.Getenv("TARS_NO_FORM") != "",
		Stdout:   stdout,
	}
}

// sessionsRunE builds the production Sessions and runs fn on it.
func sessionsRunE(fn func(s *Sessions, args []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fn(NewSessions(cmd.OutOrStdout()), args)
	}
}

var sessionsCmd = &cobra.Command{
	Use:     "sessions",
	Aliases: []string{"s", "by"},
	Short:   "Manage live byobu sessions and start them from project templates (aliases: s, by)",
	Long: `Every git repo under projects_dir (profiles.yaml, default ~/Desktop/projects)
is a project; opening one starts a byobu session named after it, laid out by its
template in ~/.config/tars/projects.yaml (default: one window at the repo root).
Opening is idempotent: a running session is attached, never duplicated.
Run bare for a picker over running sessions and projects, or with -i for a
manager that opens, kills and renames sessions and edits project templates.`,
	RunE: sessionsRunE(func(s *Sessions, _ []string) error {
		if sessionsInteractive {
			return s.Interactive()
		}
		return s.PickAndOpen()
	}),
}

var sessionsInteractive bool

var sessionsOpenCmd = &cobra.Command{
	Use:     "open [session|project]",
	Aliases: []string{"o"},
	Short:   "Attach to a running session, or start one for a project; defaults to the current dir's project (alias: o)",
	Args:    cobra.MaximumNArgs(1),
	RunE: sessionsRunE(func(s *Sessions, args []string) error {
		return s.Open(strings.Join(args, ""))
	}),
}

var sessionsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "ls"},
	Short:   "List the running byobu sessions and their projects (aliases: l, ls)",
	Args:    cobra.NoArgs,
	RunE:    sessionsRunE(func(s *Sessions, _ []string) error { return s.List() }),
}

var sessionsProjectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"p"},
	Short:   "List the projects under projects_dir, which are running, and their templates (alias: p)",
	Args:    cobra.NoArgs,
	RunE:    sessionsRunE(func(s *Sessions, _ []string) error { return s.Projects() }),
}

var sessionsKillCmd = &cobra.Command{
	Use:     "kill <session>",
	Aliases: []string{"k"},
	Short:   "Kill a running byobu session (alias: k)",
	Args:    cobra.ExactArgs(1),
	RunE:    sessionsRunE(func(s *Sessions, args []string) error { return s.Kill(args[0]) }),
}

var sessionsRenameCmd = &cobra.Command{
	Use:     "rename <session> <new-name>",
	Aliases: []string{"r"},
	Short:   "Rename a running byobu session (alias: r)",
	Args:    cobra.ExactArgs(2),
	RunE:    sessionsRunE(func(s *Sessions, args []string) error { return s.Rename(args[0], args[1]) }),
}

var sessionsEditCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Short:   "Edit the project session templates in $EDITOR, seeding an example if missing (alias: e)",
	Args:    cobra.NoArgs,
	RunE:    sessionsRunE(func(s *Sessions, _ []string) error { return s.Edit() }),
}

func init() {
	// Runtime errors (unknown session, byobu failures) shouldn't dump usage.
	sessionsCmd.SilenceUsage = true
	sessionsCmd.Flags().BoolVarP(&sessionsInteractive, "interactive", "i", false,
		"manage sessions and project templates in a full-screen list")
	for _, c := range []*cobra.Command{
		sessionsOpenCmd, sessionsListCmd, sessionsProjectsCmd, sessionsKillCmd, sessionsRenameCmd, sessionsEditCmd,
	} {
		c.SilenceUsage = true
		sessionsCmd.AddCommand(c)
	}
}
