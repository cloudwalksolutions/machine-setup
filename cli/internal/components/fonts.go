package components

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"tars/internal/paths"
)

// Fonts installs the bundled Nerd Font files into the per-user font directory.
type Fonts struct {
	opts          Options
	p             paths.FontsPaths
	LocalOverride string // when non-empty, overrides p.Local (test seam)
}

// NewFonts returns a Fonts component for the current OS.
func NewFonts(opts Options) *Fonts {
	return NewFontsForOS(opts, runtime.GOOS)
}

// NewFontsForOS is the OS-explicit form, useful for tests.
func NewFontsForOS(opts Options, goos string) *Fonts {
	return &Fonts{opts: opts, p: paths.ForOS(opts.RepoRoot, opts.Home, goos).Fonts}
}

// Name returns "fonts".
func (f *Fonts) Name() string { return "fonts" }

// Pull copies every file in <repo>/fonts/ into the font directory, backing up
// and skipping identical files like every other component.
func (f *Fonts) Pull() error {
	dst := f.p.Local
	if f.LocalOverride != "" {
		dst = f.LocalOverride
	}
	entries, err := os.ReadDir(f.p.Repo)
	if err != nil {
		return err
	}
	copier := f.opts.copier()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		src := filepath.Join(f.p.Repo, e.Name())
		if err := copier.SafeCopy(src, filepath.Join(dst, e.Name()), f.Name(), f.opts.BackupRoot); err != nil {
			return fmt.Errorf("install font %s: %w", e.Name(), err)
		}
	}
	return nil
}
