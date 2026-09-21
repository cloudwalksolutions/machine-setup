package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/fsutil"
	"tars/internal/paths"
)

// ClaudeAsker presents the `tars claude init` form over the available rule names.
type ClaudeAsker interface {
	Ask(rules []string) (config.ClaudeConfig, error)
}

// ClaudeInit orchestrates `tars claude init`: ask, persist the choices, pull.
type ClaudeInit struct {
	Asker  ClaudeAsker
	Config ConfigStore
	Rules  []string
	Pull   func(config.ClaudeConfig) error

	Stdout io.Writer
}

// Run asks the user which pieces to provision, saves the answers, and applies them.
func (c *ClaudeInit) Run() error {
	cfg, err := c.Config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	answer, err := c.Asker.Ask(c.Rules)
	if err != nil && err.Error() == "user aborted" {
		fmt.Fprintln(c.Stdout, "Aborted; nothing changed.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("claude form: %w", err)
	}
	cfg.Claude = answer
	if err := c.Config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Fprintf(c.Stdout, "Choices saved to %s\n", c.Config.Path())
	return c.Pull(cfg.Claude)
}

// ProjectAsker presents the `tars claude project` form.
type ProjectAsker interface {
	Ask(defaultName string) (forms.ProjectAnswers, error)
}

// ClaudeProject scaffolds a project's CLAUDE.md from the repo template.
type ClaudeProject struct {
	Asker      ProjectAsker
	Template   string
	BackupRoot string
	Force      bool

	Stdout io.Writer
}

// Run renders the template into <dir>/CLAUDE.md; an existing file needs --force.
func (p *ClaudeProject) Run(dir string) error {
	target := filepath.Join(dir, "CLAUDE.md")
	if _, err := os.Stat(target); err == nil && !p.Force {
		return fmt.Errorf("%s exists; re-run with --force to back it up and overwrite", target)
	}
	answers, err := p.Asker.Ask(filepath.Base(dir))
	if err != nil {
		return fmt.Errorf("project form: %w", err)
	}
	tpl, err := template.ParseFiles(p.Template)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, answers); err != nil {
		return err
	}
	if _, err := fsutil.Backup(target, "claude-project", p.BackupRoot); err != nil {
		return err
	}
	if err := os.WriteFile(target, buf.Bytes(), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(p.Stdout, "Wrote %s\n", target)
	return nil
}

// ── Production collaborators ─────────────────────────────────────────────

// FormsClaudeAsker wraps forms.ShowClaudeInitForm.
type FormsClaudeAsker struct{}

func (FormsClaudeAsker) Ask(rules []string) (config.ClaudeConfig, error) {
	return forms.ShowClaudeInitForm(rules)
}

// FormsProjectAsker wraps forms.ShowClaudeProjectForm.
type FormsProjectAsker struct{}

func (FormsProjectAsker) Ask(defaultName string) (forms.ProjectAnswers, error) {
	return forms.ShowClaudeProjectForm(defaultName)
}

// configPath honors --config, else the default (env-overridable) location.
func configPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return config.DefaultConfigPath()
}

// LoadClaudeConfig reads the saved `tars claude init` choices; a missing
// config yields the all-on zero value so pull/setup work before init ran.
func LoadClaudeConfig() (config.ClaudeConfig, error) {
	cfg, err := config.Init(configPath())
	if err != nil {
		return config.ClaudeConfig{}, fmt.Errorf("loading config: %w", err)
	}
	return cfg.Claude, nil
}

// ruleNames lists the rule files in dir without their .md suffix.
func ruleNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("listing claude rules: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, strings.TrimSuffix(e.Name(), ".md"))
	}
	return names, nil
}

// NewClaudeInit wires the production `tars claude init`.
func NewClaudeInit(stdout, stderr io.Writer) (*ClaudeInit, error) {
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	rules, err := ruleNames(paths.For(opts.RepoRoot, opts.Home).Claude.RulesRepo)
	if err != nil {
		return nil, err
	}
	return &ClaudeInit{
		Asker:  FormsClaudeAsker{},
		Config: NewFileConfigStore(configPath()),
		Rules:  rules,
		Pull: func(c config.ClaudeConfig) error {
			opts.Claude = c
			return SequentialPuller{
				Components: []components.Component{components.NewClaude(opts)},
				Stdout:     stdout,
				Stderr:     stderr,
			}.PullAll()
		},
		Stdout: stdout,
	}, nil
}

// NewClaudeProject wires the production `tars claude project`.
func NewClaudeProject(stdout, stderr io.Writer, force bool) (*ClaudeProject, error) {
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	return &ClaudeProject{
		Asker:      FormsProjectAsker{},
		Template:   paths.For(opts.RepoRoot, opts.Home).Claude.ProjectTemplateRepo,
		BackupRoot: opts.BackupRoot,
		Force:      force,
		Stdout:     stdout,
	}, nil
}

// ── Cobra commands ───────────────────────────────────────────────────────

var claudeCmd = &cobra.Command{
	Use:     "claude",
	Aliases: []string{"c", "cc"},
	Short:   "Provision Claude Code: hook, settings, global rules, project CLAUDE.md (aliases: c, cc)",
	Long: `Manage Claude Code configuration the same way tars manages dotfiles.

  init     pick which pieces to provision (edit-blocking hook, settings fragment,
           global rules), save the choices, and apply them to ~/.claude
  project  scaffold a CLAUDE.md for a project from the repo template

'tars pull' and 'tars push' keep the pieces in sync afterwards.`,
}

var claudeInitCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Choose and apply the Claude Code pieces for this machine (alias: i)",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := NewClaudeInit(cmd.OutOrStdout(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		return c.Run()
	},
}

var claudeProjectForce bool

var claudeProjectCmd = &cobra.Command{
	Use:     "project [dir]",
	Aliases: []string{"p"},
	Short:   "Scaffold <dir>/CLAUDE.md from the repo template, default cwd (alias: p)",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) == 1 {
			dir = args[0]
		}
		dir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		p, err := NewClaudeProject(cmd.OutOrStdout(), cmd.ErrOrStderr(), claudeProjectForce)
		if err != nil {
			return err
		}
		return p.Run(dir)
	},
}

func init() {
	claudeProjectCmd.Flags().BoolVar(&claudeProjectForce, "force", false, "back up and overwrite an existing CLAUDE.md")
	// Runtime errors (existing CLAUDE.md, unwritable HOME) shouldn't dump usage.
	for _, c := range []*cobra.Command{claudeCmd, claudeInitCmd, claudeProjectCmd} {
		c.SilenceUsage = true
	}
	claudeCmd.AddCommand(claudeInitCmd, claudeProjectCmd)
}
