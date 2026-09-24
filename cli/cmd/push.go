package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"tars/internal/components"
	"tars/internal/repo"
)

// SequentialPusher iterates the configured pushable components, printing
// progress and capturing per-component failures (mirrors SequentialPuller).
type SequentialPusher struct {
	Components []components.Pushable
	Stdout     io.Writer
	Stderr     io.Writer
}

// PushAll pushes every component, reporting failures inline without aborting,
// and returns an aggregate error naming the components that failed.
func (p SequentialPusher) PushAll() error {
	return runComponents(p.Components, components.Pushable.Name, components.Pushable.Push, p.Stdout, p.Stderr)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Copy local config changes back into the repo (with versioned backups)",
	Long: `Push local dotfile and terminal configuration into the repo, archiving the
previous repo copies under ~/.local/state/tars/backups/<component>-repo/vN
(override the location with TARS_BACKUP_ROOT).

Requires a real clone of the repo — the configs embedded in the binary are
read-only. Exits non-zero when any component fails, listing each failure.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		stdout, stderr := cmd.OutOrStdout(), cmd.ErrOrStderr()
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("locating home dir: %w", err)
		}
		// Push writes into the repo working tree, so it needs a real clone —
		// the embedded-assets fallback is read-only by design.
		root, err := repo.Find()
		if err != nil {
			return fmt.Errorf("push requires a clone of the repo (set TARS_REPO or run from inside one): %w", err)
		}
		opts := components.Options{
			RepoRoot:   root,
			Home:       home,
			BackupRoot: BackupRoot(home),
			Stdout:     stdout,
			Stderr:     stderr,
		}
		fmt.Fprintln(stdout, "Pushing configuration files...")
		if err := (SequentialPusher{
			Components: components.AllPushable(opts),
			Stdout:     stdout,
			Stderr:     stderr,
		}).PushAll(); err != nil {
			return fmt.Errorf("push completed with failures: %w", err)
		}
		fmt.Fprintln(stdout, "\nPush complete.")
		return nil
	},
	SilenceUsage: true,
}
