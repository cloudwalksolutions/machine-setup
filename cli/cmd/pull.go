package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/report"
)

var pullDryRun bool

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Apply repo configs to this machine (dotfiles, fonts, terminals) with versioned backups",
	Long: `Copy the dotfiles, fonts, and terminal settings into your machine and render
your git profiles, backing up anything that already exists under
~/.local/state/tars/backups/<component>/vN (override with TARS_BACKUP_ROOT).
~/.zshrc_secret and ~/.zprofile_local are seeded once and never overwritten.

Unlike 'setup', pull does NOT install packages, oh-my-zsh, or anything from the
network — it only lays down configuration, so it's safe to run repeatedly.
Use --dry-run to see the writes without making them. Exits non-zero when any
component fails, listing each failure.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		stdout, stderr := cmd.OutOrStdout(), cmd.ErrOrStderr()
		opts, err := buildOptions(stdout, stderr)
		if err != nil {
			return err
		}
		opts.DryRun = pullDryRun
		if pullDryRun {
			fmt.Fprintln(stdout, "Dry run — no files will be written.")
		} else {
			fmt.Fprintln(stdout, "Applying configuration files...")
		}
		if err := (SequentialPuller{
			Components: components.AllPullable(opts),
			Report:     report.Text{Stdout: stdout, Stderr: stderr},
		}).PullAll(); err != nil {
			return fmt.Errorf("pull completed with failures: %w", err)
		}
		if pullDryRun {
			fmt.Fprintln(stdout, "\nDry run complete — nothing was written.")
		} else {
			fmt.Fprintln(stdout, "\nPull complete.")
		}
		return nil
	},
	SilenceUsage: true,
}
