package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"tars/internal/fsutil"
	"tars/internal/paths"
	"tars/internal/report"
	"tars/internal/tui"
)

// ProjectInit scaffolds a project's agent harness: one AGENTS.md from the repo
// template plus a pointer file per selected agent that imports it.
type ProjectInit struct {
	Asker      ProjectAsker
	Template   string
	BackupRoot string
	Force      bool
	Report     report.Reporter
}

// pointers maps an agent to the file it reads and the import line that file carries.
var pointers = map[string][2]string{
	"claude": {"CLAUDE.md", "@AGENTS.md\n"},
	"gemini": {"GEMINI.md", "@./AGENTS.md\n"},
}

// Run renders AGENTS.md into dir and writes the selected pointer files; an
// existing AGENTS.md needs --force.
func (p *ProjectInit) Run(dir string) error {
	target := filepath.Join(dir, "AGENTS.md")
	if _, err := os.Stat(target); err == nil && !p.Force {
		return fmt.Errorf("%s exists; re-run with --force to back it up and overwrite", target)
	}
	answers, err := p.Asker.AskProject(filepath.Base(dir))
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
	if _, err := fsutil.Backup(target, "project", p.BackupRoot); err != nil {
		return err
	}
	if err := os.WriteFile(target, buf.Bytes(), 0o644); err != nil {
		return err
	}
	for _, agent := range answers.Agents {
		ptr, ok := pointers[agent]
		if !ok {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, ptr[0]), []byte(ptr[1]), 0o644); err != nil {
			return err
		}
	}
	p.Report.Note("Wrote " + target)
	return nil
}

// NewProjectInit wires the production `tars init project`.
func NewProjectInit(r report.Reporter, ask ProjectAsker, force bool) (*ProjectInit, error) {
	stdout, stderr := r.Output()
	opts, err := buildOptions(stdout, stderr)
	if err != nil {
		return nil, err
	}
	return &ProjectInit{
		Asker:      ask,
		Template:   paths.For(opts.RepoRoot, opts.Home).Claude.ProjectTemplateRepo,
		BackupRoot: opts.BackupRoot,
		Force:      force,
		Report:     r,
	}, nil
}

var initProjectForce bool

var initProjectCmd = &cobra.Command{
	Use:     "project [dir]",
	Aliases: []string{"p"},
	Short:   "Scaffold <dir>/AGENTS.md plus CLAUDE.md/GEMINI.md pointers for the agents you pick, default cwd (alias: p)",
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
		return runInit(cmd, func(r report.Reporter, ask tui.Asker) error {
			p, err := NewProjectInit(r, ask, initProjectForce)
			if err != nil {
				return err
			}
			return p.Run(dir)
		})
	},
}

func init() {
	initProjectCmd.Flags().BoolVar(&initProjectForce, "force", false, "back up and overwrite an existing AGENTS.md")
	initProjectCmd.SilenceUsage = true
	initCmd.AddCommand(initProjectCmd)
}
