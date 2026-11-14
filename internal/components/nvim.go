package components

import (
	"path/filepath"

	"tars/internal/fsutil"
	"tars/internal/paths"
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
	// Idempotent: an unchanged local tree needs no backup, wipe, or copy.
	same, err := fsutil.SameTree(n.p.Repo, n.p.Local, n.p.Exclude...)
	if err != nil {
		return err
	}
	if !same {
		if err := n.opts.copier().Replace(n.p.Repo, n.p.Local, n.p.Exclude, n.Name(), n.opts.BackupRoot); err != nil {
			return err
		}
	}
	monokaiDst := filepath.Join(n.p.MonokaiLocal, "monokai.lua")
	return n.opts.copier().SafeCopy(n.p.MonokaiRepo, monokaiDst, n.Name(), n.opts.BackupRoot)
}

// Push replaces the repo's nvim/ tree with ~/.config/nvim, archiving the old
// repo tree under "nvim-repo" (clean replace, mirroring Pull).
func (n *Nvim) Push() error {
	return n.opts.copier().Replace(n.p.Local, n.p.Repo, n.p.Exclude, n.Name()+"-repo", n.opts.BackupRoot)
}
