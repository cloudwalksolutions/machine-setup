package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudwalk/machine-setup/internal/components"
	"github.com/cloudwalk/machine-setup/internal/repo"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Apply repo configs to this machine (dotfiles, fonts, terminals) with versioned backups",
	Long: `Copy the repo's dotfiles, fonts, and terminal settings into your machine,
backing up anything that already exists under backups/<component>/vN.

Unlike 'setup', pull does NOT install packages, oh-my-zsh, or anything from the
network — it only lays down configuration, so it's safe to run repeatedly.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		stdout, stderr := cmd.OutOrStdout(), cmd.ErrOrStderr()
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("locating home dir: %w", err)
		}
		root, err := repo.Find()
		if err != nil {
			return fmt.Errorf("locating repo root: %w", err)
		}
		opts := components.Options{
			RepoRoot:   root,
			Home:       home,
			BackupRoot: filepath.Join(root, "backups"),
			Stdout:     stdout,
			Stderr:     stderr,
		}
		fmt.Fprintln(stdout, "Applying configuration files...")
		SequentialPuller{
			Components: components.AllPullable(opts),
			Stdout:     stdout,
			Stderr:     stderr,
		}.PullAll()
		fmt.Fprintln(stdout, "\nPull complete.")
		return nil
	},
}
