// Package repo resolves the machine-setup repository root.
package repo

import (
	"errors"
	"os"
	"path/filepath"
)

// Find returns the repository root. Honors MACHINE_SETUP_REPO if set,
// otherwise walks up from the current working directory.
func Find() (string, error) {
	if env := os.Getenv("MACHINE_SETUP_REPO"); env != "" {
		return env, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return FindFrom(cwd)
}

// FindFrom walks up from start looking for a directory containing both a
// Makefile and a scripts/components subdirectory (the markers of this repo).
func FindFrom(start string) (string, error) {
	dir := start
	for {
		if hasMarkers(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("repo root not found from " + start)
		}
		dir = parent
	}
}

func hasMarkers(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err != nil {
		return false
	}
	if info, err := os.Stat(filepath.Join(dir, "scripts", "components")); err != nil || !info.IsDir() {
		return false
	}
	return true
}
