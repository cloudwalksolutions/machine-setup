package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"tars/internal/assets"
	"tars/internal/components"
	"tars/internal/repo"
)

// binaryVersion is the release tag baked in via SetVersion; "dev" otherwise.
var binaryVersion = "dev"

// buildOptions resolves home + repo root and assembles the shared component
// options — the common preamble of pull and setup (push differs: it requires
// a real clone, so it resolves the repo itself).
func buildOptions(stdout, stderr io.Writer) (components.Options, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return components.Options{}, fmt.Errorf("locating home dir: %w", err)
	}
	root, err := ResolveRepo(home)
	if err != nil {
		return components.Options{}, fmt.Errorf("locating repo root: %w", err)
	}
	return components.Options{
		RepoRoot:   root,
		Home:       home,
		BackupRoot: BackupRoot(home),
		Stdout:     stdout,
		Stderr:     stderr,
	}, nil
}

// ResolveRepo returns the repo root for config reads: a real clone when one
// is found (env override or upward walk), else the embedded dotfiles
// materialized under ~/.local/share/tars/repo — so a brew-installed tars
// works with no clone at all.
func ResolveRepo(home string) (string, error) {
	if root, err := repo.Find(); err == nil {
		return root, nil
	}
	dst := filepath.Join(home, ".local", "share", "tars", "repo")
	if err := assets.Materialize(dst, binaryVersion); err != nil {
		return "", fmt.Errorf("materializing embedded configs: %w", err)
	}
	return dst, nil
}
