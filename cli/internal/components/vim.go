package components

import (
	"github.com/cloudwalk/machine-setup/internal/fsutil"
	"github.com/cloudwalk/machine-setup/internal/paths"
)

// Vim ports scripts/components/vim.sh.
type Vim struct {
	opts Options
	p    paths.VimPaths
}

// NewVim returns a Vim component bound to opts.
func NewVim(opts Options) *Vim {
	return &Vim{opts: opts, p: paths.For(opts.RepoRoot, opts.Home).Vim}
}

// Name returns the component's name used for backup directory grouping.
func (v *Vim) Name() string { return "vim" }

// Pull copies vimrc and the sublimemonokai color scheme into HOME.
func (v *Vim) Pull() error {
	if err := fsutil.SafeCopy(v.p.VimrcRepo, v.p.VimrcLocal, v.Name(), v.opts.BackupRoot); err != nil {
		return err
	}
	return fsutil.SafeCopy(v.p.ColorsRepo, v.p.ColorsLocal, v.Name(), v.opts.BackupRoot)
}
