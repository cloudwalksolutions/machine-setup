package cmd

import (
	"fmt"
	"os"

	"github.com/cloudwalk/machine-setup/internal/components"
	"github.com/spf13/cobra"
)

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
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("locating home dir: %w", err)
		}
		root, err := ResolveRepo(home)
		if err != nil {
			return fmt.Errorf("locating repo root: %w", err)
		}
		opts := components.Options{
			RepoRoot:   root,
			Home:       home,
			BackupRoot: BackupRoot(home),
			Stdout:     stdout,
			Stderr:     stderr,
		}
		fmt.Fprintln(stdout, "Applying configuration files...")
		if err := (SequentialPuller{
			Components: components.AllPullable(opts),
			Stdout:     stdout,
			Stderr:     stderr,
		}).PullAll(); err != nil {
			return fmt.Errorf("pull completed with failures: %w", err)
		}
		fmt.Fprintln(stdout, "\nPull complete.")
		return nil
	},
	SilenceUsage: true,
}
