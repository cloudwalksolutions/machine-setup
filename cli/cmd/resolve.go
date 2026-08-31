package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cloudwalk/machine-setup/internal/assets"
	"github.com/cloudwalk/machine-setup/internal/repo"
)

// binaryVersion is the release tag baked in via SetVersion; "dev" otherwise.
var binaryVersion = "dev"

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
