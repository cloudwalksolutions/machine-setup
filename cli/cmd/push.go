package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudwalk/machine-setup/internal/components"
	"github.com/cloudwalk/machine-setup/internal/repo"
	"github.com/spf13/cobra"
)

// SequentialPusher iterates the configured pushable components, printing
// progress and capturing per-component failures (mirrors SequentialPuller).
type SequentialPusher struct {
	Components []components.Pushable
	Stdout     io.Writer
	Stderr     io.Writer
}

// PushAll pushes every component, reporting failures inline without aborting.
func (p SequentialPusher) PushAll() {
	for _, c := range p.Components {
		fmt.Fprintf(p.Stdout, "  → %s\n", c.Name())
		if err := c.Push(); err != nil {
			fmt.Fprintf(p.Stderr, "  %s: %v\n", c.Name(), err)
		}
	}
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Copy local config changes back into the repo (with versioned backups)",
	Long: `Push local dotfile and terminal configuration into the repo, archiving the
previous repo copies under backups/<component>-repo/vN.`,
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
		fmt.Fprintln(stdout, "Pushing configuration files...")
		SequentialPusher{
			Components: components.AllPushable(opts),
			Stdout:     stdout,
			Stderr:     stderr,
		}.PushAll()
		fmt.Fprintln(stdout, "\nPush complete.")
		return nil
	},
}
