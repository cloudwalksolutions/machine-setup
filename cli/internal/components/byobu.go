package components

import (
	"os"
	"path/filepath"

	"github.com/cloudwalk/machine-setup/internal/fsutil"
	"github.com/cloudwalk/machine-setup/internal/paths"
)

// Byobu ports scripts/components/byobu.sh.
type Byobu struct {
	opts Options
	p    paths.ByobuPaths
}

// NewByobu returns a Byobu component bound to opts.
func NewByobu(opts Options) *Byobu {
	return &Byobu{opts: opts, p: paths.For(opts.RepoRoot, opts.Home).Byobu}
}

// Name returns "byobu".
func (b *Byobu) Name() string { return "byobu" }

// Pull copies all byobu config files (excluding bin/) into ~/.byobu/.
func (b *Byobu) Pull() error {
	copies := []struct{ src, dst string }{
		{b.p.TmuxConfRepo, b.p.TmuxConfLocal},
		{b.p.KeybindingsRepo, b.p.KeybindingsLocal},
		{b.p.DatetimeRepo, b.p.DatetimeLocal},
		{b.p.StatusrcRepo, b.p.StatusrcLocal},
	}
	for _, c := range copies {
		if err := fsutil.SafeCopy(c.src, c.dst, b.Name(), b.opts.BackupRoot); err != nil {
			return err
		}
	}
	return b.pullBin()
}

// pullBin copies each file inside <repo>/byobu/bin into ~/.byobu/bin, flat.
// Mirrors `cp -r byobu/bin/* ~/.byobu/bin/` in scripts/components/byobu.sh:29.
func (b *Byobu) pullBin() error {
	entries, err := os.ReadDir(b.p.BinRepo)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(b.p.BinLocal, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		src := filepath.Join(b.p.BinRepo, e.Name())
		dst := filepath.Join(b.p.BinLocal, e.Name())
		if err := fsutil.SafeCopy(src, dst, b.Name(), b.opts.BackupRoot); err != nil {
			return err
		}
	}
	return nil
}
