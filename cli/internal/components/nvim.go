package components

import (
	"os"
	"path/filepath"

	"github.com/cloudwalk/machine-setup/internal/fsutil"
	"github.com/cloudwalk/machine-setup/internal/paths"
)

// Nvim ports scripts/components/nvim.sh.
type Nvim struct {
	opts Options
	p    paths.NvimPaths
}

// NewNvim returns an Nvim component bound to opts.
func NewNvim(opts Options) *Nvim {
	return &Nvim{opts: opts, p: paths.For(opts.RepoRoot, opts.Home).Nvim}
}

// Name returns "nvim".
func (n *Nvim) Name() string { return "nvim" }

// Pull replaces ~/.config/nvim with the repo's nvim/ tree, then copies the
// monokai theme into the packer plugin path.
func (n *Nvim) Pull() error {
	// Backup the existing local config (no-op if absent), then wipe so the
	// new tree is a clean replace rather than a merge.
	if _, err := fsutil.Backup(n.p.Local, n.Name(), n.opts.BackupRoot); err != nil {
		return err
	}
	if err := os.RemoveAll(n.p.Local); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(n.p.Local), 0o755); err != nil {
		return err
	}
	if err := fsutil.SafeCopy(n.p.Repo, n.p.Local, n.Name(), n.opts.BackupRoot); err != nil {
		return err
	}
	// Monokai theme.
	if err := os.MkdirAll(n.p.MonokaiLocal, 0o755); err != nil {
		return err
	}
	monokaiDst := filepath.Join(n.p.MonokaiLocal, "monokai.lua")
	return fsutil.SafeCopy(n.p.MonokaiRepo, monokaiDst, n.Name(), n.opts.BackupRoot)
}
