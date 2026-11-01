package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"tars/internal/components"
	"tars/internal/report"
	"tars/internal/tui"
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
Every write is listed as it happens and the run ends with a recap. --diff prints
a header per changed component, its unified diffs and a Plan line, and --dry-run
lists the writes; neither changes anything. Exits non-zero when any component
fails, listing each failure.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		stdout, stderr := cmd.OutOrStdout(), cmd.ErrOrStderr()
		opts, err := buildOptions(stdout, stderr)
		if err != nil {
			return err
		}
		p := &SequentialPuller{Report: report.Text{Stdout: stdout, Stderr: stderr}}
		opts.Record = p.Record
		if pullDiff {
			opts.DryRun, opts.Diff, opts.Stdout = true, true, io.Discard
			p.Previews = components.AllPullable(opts)
			plans, err := p.Plans()
			if err != nil {
				return err
			}
			components.WritePlans(stdout, plans)
			return nil
		}
		opts.DryRun = pullDryRun
		if opts.DryRun {
			fmt.Fprintln(stdout, "Dry run — no files will be written.")
		} else {
			fmt.Fprintln(stdout, "Applying configuration files...")
		}
		p.Components = components.AllPullable(opts)
		if Interactive(pullYes, opts.DryRun, os.Getenv("TARS_NO_FORM") != "", term.IsTerminal(os.Stdin.Fd()), term.IsTerminal(os.Stdout.Fd())) {
			preview := opts
			preview.DryRun, preview.Diff, preview.Stdout = true, true, io.Discard
			p.Previews, p.Review = components.AllPullable(preview), tui.Reviewer{}
		}
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

// Interactive reports whether pull should open the review: only a writing pull
// on a terminal, without -y or TARS_NO_FORM.
func Interactive(yes, dryRun, headless, ttyIn, ttyOut bool) bool {
	return !yes && !dryRun && !headless && ttyIn && ttyOut
}
