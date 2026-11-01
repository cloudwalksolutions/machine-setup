package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tars/internal/components"
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
		p := &SequentialPuller{Report: report.Text{Stdout: stdout, Stderr: stderr}}
		opts.Record = p.Record
		p.Components = components.AllPullable(opts)
		if err := p.PullAll(); err != nil && err.Error() == "user aborted" {
			fmt.Fprintln(stdout, "\nPull cancelled; nothing was written.")
			return nil
		} else if err != nil {
			return fmt.Errorf("pull completed with failures: %w", err)
		}
		if opts.DryRun {
			fmt.Fprintln(stdout, "\nDry run complete — nothing was written.")
		} else {
			fmt.Fprintln(stdout, "\n"+components.Recap(p.Applied, len(p.Components)-len(p.Applied)-len(p.Skipped)))
		}
		return nil
	},
	SilenceUsage: true,
}
