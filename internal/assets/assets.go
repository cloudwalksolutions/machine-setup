// Package assets serves the dotfiles embedded in the binary (embed.go at the
// repo root) so a brew-installed tars can provision a machine with no clone:
// the tree is materialized on demand and then serves as the repo root.
package assets

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"tars"
)

// Dirs are the repo dotfile directories embedded in the binary.
var Dirs = []string{"nvim", "zsh", "byobu", "vim", "fonts", "terminal", "claude", "pi", "agents"}

// RootFiles are repo-root files components read (paths.go), embedded alongside Dirs.
var RootFiles = []string{"monokai.lua"}

const versionMarker = ".tars-version"

// Files returns the embedded files under one top-level dir (or a single
// root file), keyed by repo-relative slash path (e.g. "zsh/zshrc").
func Files(dir string) (map[string][]byte, error) {
	out := map[string][]byte{}
	err := fs.WalkDir(tars.Dotfiles, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := tars.Dotfiles.ReadFile(path)
		if err != nil {
			return err
		}
		out[path] = b
		return nil
	})
	return out, err
}

// Materialize extracts the embedded dotfile tree to dst and stamps it with
// version. A matching marker short-circuits; "dev" builds always re-extract
// so a locally built binary never serves a stale cache.
func Materialize(dst, version string) error {
	if version != "dev" {
		if b, err := os.ReadFile(filepath.Join(dst, versionMarker)); err == nil && string(b) == version {
			return nil
		}
	}
	for _, top := range append(append([]string{}, Dirs...), RootFiles...) {
		files, err := Files(top)
		if err != nil {
			return err
		}
		for rel, content := range files {
			path := filepath.Join(dst, filepath.FromSlash(rel))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			// go:embed drops file modes; bin/ and hooks/ scripts must stay executable.
			mode := os.FileMode(0o644)
			if strings.Contains(rel, "/bin/") || strings.Contains(rel, "/hooks/") {
				mode = 0o755
			}
			if err := os.WriteFile(path, content, mode); err != nil {
				return err
			}
			// WriteFile's perm is ignored for pre-existing files.
			if err := os.Chmod(path, mode); err != nil {
				return err
			}
		}
	}
	return os.WriteFile(filepath.Join(dst, versionMarker), []byte(version), 0o644)
}
