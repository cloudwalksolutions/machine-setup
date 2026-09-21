package components

import (
	"os"
	"path/filepath"

	"github.com/cloudwalk/machine-setup/internal/paths"
)

// Nvim pulls/pushes the ~/.config/nvim tree (clean replace) + monokai theme.
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
	if _, err := n.opts.copier().Backup(n.p.Local, n.Name(), n.opts.BackupRoot); err != nil {
		return err
	}
	if err := n.opts.copier().RemoveAll(n.p.Local); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(n.p.Local), 0o755); err != nil {
		return err
	}
	if err := n.opts.copier().SafeCopy(n.p.Repo, n.p.Local, n.Name(), n.opts.BackupRoot); err != nil {
		return err
	}
	// Monokai theme.
	if err := os.MkdirAll(n.p.MonokaiLocal, 0o755); err != nil {
		return err
	}
	monokaiDst := filepath.Join(n.p.MonokaiLocal, "monokai.lua")
	return n.opts.copier().SafeCopy(n.p.MonokaiRepo, monokaiDst, n.Name(), n.opts.BackupRoot)
}

// Push replaces the repo's nvim/ tree with ~/.config/nvim, archiving the old
// repo tree under "nvim-repo" (clean replace, mirroring Pull).
func (n *Nvim) Push() error {
	comp := n.Name() + "-repo"
	if _, err := n.opts.copier().Backup(n.p.Repo, comp, n.opts.BackupRoot); err != nil {
		return err
	}
	if err := n.opts.copier().RemoveAll(n.p.Repo); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(n.p.Repo), 0o755); err != nil {
		return err
	}
	return n.opts.copier().SafeCopy(n.p.Local, n.p.Repo, comp, n.opts.BackupRoot)
}
