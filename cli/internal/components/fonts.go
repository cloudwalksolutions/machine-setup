package components

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/cloudwalk/machine-setup/internal/paths"
)

// Fonts ports scripts/components/fonts.sh.
//
// On darwin, the system font dir (/Library/Fonts) requires sudo, so the
// default CopyFn shells out to `sudo cp`. On other OSes, a plain copy is used.
// Tests inject CopyFn (and optionally LocalOverride) to avoid sudo.
type Fonts struct {
	opts          Options
	p             paths.FontsPaths
	CopyFn        func(src, dst string) error
	LocalOverride string // when non-empty, overrides p.Local (test seam)
}

// NewFonts returns a Fonts component with platform defaults for the current OS.
func NewFonts(opts Options) *Fonts {
	return NewFontsForOS(opts, runtime.GOOS)
}

// NewFontsForOS is the OS-explicit form, useful for tests.
func NewFontsForOS(opts Options, goos string) *Fonts {
	p := paths.ForOS(opts.RepoRoot, opts.Home, goos).Fonts
	f := &Fonts{opts: opts, p: p}
	f.CopyFn = defaultFontCopy(goos)
	return f
}

// Name returns "fonts".
func (f *Fonts) Name() string { return "fonts" }

// Pull copies every file in <repo>/fonts/ to the OS-appropriate font directory.
func (f *Fonts) Pull() error {
	dst := f.p.Local
	if f.LocalOverride != "" {
		dst = f.LocalOverride
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(f.p.Repo)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		src := filepath.Join(f.p.Repo, e.Name())
		dstFile := filepath.Join(dst, e.Name())
		if err := f.CopyFn(src, dstFile); err != nil {
			return fmt.Errorf("install font %s: %w", e.Name(), err)
		}
	}
	return nil
}

// defaultFontCopy returns a per-OS copy function. darwin uses `sudo cp` because
// /Library/Fonts is system-owned; other OSes use a plain in-process copy.
func defaultFontCopy(goos string) func(src, dst string) error {
	if goos == "darwin" {
		return func(src, dst string) error {
			cmd := exec.Command("sudo", "cp", src, dst)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
	}
	return plainCopy
}

func plainCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
