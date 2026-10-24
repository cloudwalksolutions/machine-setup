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
		p := SequentialPuller{Components: components.AllPullable(opts), Report: report.Text{Stdout: stdout, Stderr: stderr}}
		if !pullYes && !opts.DryRun && os.Getenv("TARS_NO_FORM") == "" && term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd()) {
			diffs, preview := &bytes.Buffer{}, opts
			preview.DryRun, preview.Diff, preview.Stdout = true, true, diffs
			p.Previews, p.Diffs, p.Confirm = components.AllPullable(preview), diffs, forms.PullConfirm{Out: stdout}
		}
		if err := p.PullAll(); err != nil && err.Error() == "user aborted" {
			fmt.Fprintln(stdout, "\nPull cancelled; components after this one were left as they were.")
			return nil
		} else if err != nil {
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

// Confirmer shows a component's pending diff and asks whether to apply it.
type Confirmer interface {
	ConfirmApply(name, diff string) (bool, error)
}
