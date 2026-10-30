// Package sessions drives byobu: lists, creates, attaches, kills and renames sessions.
package sessions

import (
	"path/filepath"
	"strings"
)

// WindowName derives a tmux-safe window name from a dir's basename.
func WindowName(dir string) string {
	name := filepath.Base(dir)
	return strings.NewReplacer(".", "_", ":", "_").Replace(name)
}
