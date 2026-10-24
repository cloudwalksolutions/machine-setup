package cmd

import (
	"bytes"
	"fmt"
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/forms"
	"tars/internal/report"
)

var pullDryRun, pullDiff, pullYes bool

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Apply repo configs to this machine (dotfiles, fonts, terminals) with versioned backups",
	Long: `Copy the dotfiles, fonts, and terminal settings into your machine and render
your git profiles, backing up anything that already exists under
~/.local/state/tars/backups/<component>/vN (override with TARS_BACKUP_ROOT).
~/.zshrc_secret and ~/.zprofile_local are seeded once and never overwritten.

Unlike 'setup', pull does NOT install packages, oh-my-zsh, or anything from the
network — it only lays down configuration, so it's safe to run repeatedly.
On a terminal it shows each component's diff and applies it on Enter; -y/--yes,
TARS_NO_FORM or no terminal applies everything without asking. --diff prints
every diff and --dry-run lists the writes, both without changing anything.
Exits non-zero when any component fails, listing each failure.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		stdout, stderr := cmd.OutOrStdout(), cmd.ErrOrStderr()
		opts, err := buildOptions(stdout, stderr)
		if err != nil {
			return err
		}
		opts.DryRun = pullDryRun || pullDiff
		opts.Diff = pullDiff
		if opts.DryRun {
			fmt.Fprintln(stdout, "Dry run — no files will be written.")
		} else {
			fmt.Fprintln(stdout, "Applying configuration files...")
		}
		r := report.Text{Stdout: stdout, Stderr: stderr}
		var puller interface{ PullAll() error } = SequentialPuller{Components: components.AllPullable(opts), Report: r}
		if PullConfirms(pullYes, opts.DryRun, os.Getenv("TARS_NO_FORM") != "", term.IsTerminal(os.Stdin.Fd())) {
			puller = NewConfirmingPuller(opts, r)
		}
		if err := puller.PullAll(); err != nil {
			return fmt.Errorf("pull completed with failures: %w", err)
		}
		if opts.DryRun {
			fmt.Fprintln(stdout, "\nDry run complete — nothing was written.")
		} else {
			fmt.Fprintln(stdout, "\nPull complete.")
		}
		return nil
	},
	SilenceUsage: true,
}

// NewConfirmingPuller pairs each component's diff preview with its real pull.
func NewConfirmingPuller(opts components.Options, r report.Reporter) ConfirmingPuller {
	previews := &bytes.Buffer{}
	preview := opts
	preview.DryRun, preview.Diff, preview.Stdout = true, true, previews
	apply := components.AllPullable(opts)
	steps := make([]PullStep, len(apply))
	for i, c := range components.AllPullable(preview) {
		steps[i] = PullStep{Preview: c, Apply: apply[i]}
	}
	return ConfirmingPuller{Steps: steps, Previews: previews, Confirm: forms.PullConfirm{Out: opts.Stdout}, Report: r}
}

// PullConfirms reports whether pull should show each component's diff and ask
// before applying it: only a writing pull with a terminal and no --yes.
func PullConfirms(yes, dryRun, headless, tty bool) bool {
	return !yes && !dryRun && !headless && tty
}

// PullStep pairs a component's diff-mode dry run with the real one.
type PullStep struct {
	Preview components.Component
	Apply   components.Component
}

func (s PullStep) name() string { return s.Apply.Name() }

// Confirmer shows a component's pending diff and asks whether to apply it.
type Confirmer interface {
	ConfirmApply(name, diff string) (bool, error)
}

// ConfirmingPuller previews each component into Previews and applies it only
// when the user confirms its diff.
type ConfirmingPuller struct {
	Steps    []PullStep
	Previews *bytes.Buffer
	Confirm  Confirmer
	Report   report.Reporter
}

// PullAll previews, confirms and applies each step, returning failures joined.
func (p ConfirmingPuller) PullAll() error {
	return runComponents(p.Steps, PullStep.name, p.pull, p.Report)
}

func (p ConfirmingPuller) pull(s PullStep) error {
	p.Previews.Reset()
	if err := s.Preview.Pull(); err != nil {
		return err
	}
	if p.Previews.Len() > 0 {
		ok, err := p.Confirm.ConfirmApply(s.name(), p.Previews.String())
		if err != nil {
			return err
		}
		if !ok {
			p.Report.Note("    skipped " + s.name())
			return nil
		}
	}
	return s.Apply.Pull()
}
