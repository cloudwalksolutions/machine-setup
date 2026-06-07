// Package components ports scripts/components/*.sh to native Go so the CLI
// can pull dotfile configs without shelling out to bash.
package components

import "io"

// Component is the unit the orchestrator iterates over during setup.
type Component interface {
	Name() string
	Pull() error
}

// Options is the per-run configuration every component needs.
type Options struct {
	RepoRoot   string    // root of the machine-setup repo
	Home       string    // user's HOME (destination root)
	BackupRoot string    // <repoRoot>/backups in normal use
	Stdout     io.Writer // progress output
	Stderr     io.Writer // error/warning output
}

// AllPullable returns the components in the order scripts/pull.sh iterates them.
func AllPullable(opts Options) []Component {
	return []Component{
		NewVim(opts),
		NewZsh(opts),
		NewByobu(opts),
		NewNvim(opts),
		NewFonts(opts),
	}
}
