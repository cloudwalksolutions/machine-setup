package cmd

import (
	"os"
	"path/filepath"
)

// BackupRoot returns where versioned backups are written: the
// MACHINE_SETUP_BACKUP_ROOT env var when set, else ~/.local/state/tars/backups.
// Living under $HOME keeps a shared (or read-only) repo clone safe and gives
// every user of a machine their own backup history.
func BackupRoot(home string) string {
	if env := os.Getenv("MACHINE_SETUP_BACKUP_ROOT"); env != "" {
		return env
	}
	return filepath.Join(home, ".local", "state", "tars", "backups")
}
