// Package components implements the per-tool config operations (pull to the
// machine, push back to the repo) natively in Go.
package components

import (
	"io"

	"tars/internal/fsutil"
)

// Component is the unit the orchestrator iterates over during setup.
type Component interface {
	Name() string
	Pull() error
}

// Pushable is a component that can copy local config back into the repo.
// Fonts is intentionally not Pushable (system fonts are install-only).
type Pushable interface {
	Name() string
	Push() error
}

// Options is the per-run configuration every component needs.
type Options struct {
	RepoRoot   string    // root of the machine-setup repo
	Home       string    // user's HOME (destination root)
	BackupRoot string    // <repoRoot>/backups in normal use
	DryRun     bool      // report intended writes instead of performing them
	Stdout     io.Writer // progress output
	Stderr     io.Writer // error/warning output
}

// copier is the writer every component routes its writes through.
func (o Options) copier() fsutil.Copier {
	return fsutil.Copier{DryRun: o.DryRun, Log: o.Stdout}
}

// AllPullable returns the pullable components in canonical order.
func AllPullable(opts Options) []Component {
	return []Component{
		NewVim(opts),
		NewZsh(opts),
		NewByobu(opts),
		NewNvim(opts),
		NewFonts(opts),
		NewTerminal(opts),
		NewProfiles(opts),
	}
}

// AllPushable returns the components that can push local config back to the repo,
// in canonical order (fonts excluded — install-only).
func AllPushable(opts Options) []Pushable {
	return []Pushable{
		NewVim(opts),
		NewZsh(opts),
		NewByobu(opts),
		NewNvim(opts),
		NewTerminal(opts),
	}
}
