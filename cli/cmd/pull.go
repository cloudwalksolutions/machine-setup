package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"tars/internal/components"
)

var pullDryRun bool

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Apply repo configs to this machine (dotfiles, fonts, terminals) with versioned backups",
	Long: `Copy the dotfiles, fonts, and terminal settings into your machine, backing up
anything that already exists under ~/.local/state/tars/backups/<component>/vN
(override the location with MACHINE_SETUP_BACKUP_ROOT).

Unlike 'setup', pull does NOT install packages, oh-my-zsh, or anything from the
network — it only lays down configuration, so it's safe to run repeatedly.
Exits non-zero when any component fails, listing each failure.`,
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
			Stdout:     stdout,
			Stderr:     stderr,
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
