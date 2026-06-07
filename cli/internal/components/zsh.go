package components

import (
	"os"

	"github.com/cloudwalk/machine-setup/internal/fsutil"
	"github.com/cloudwalk/machine-setup/internal/paths"
)

// Zsh ports scripts/components/zsh.sh.
type Zsh struct {
	opts Options
	p    paths.ZshPaths
}

// NewZsh returns a Zsh component bound to opts.
func NewZsh(opts Options) *Zsh {
	return &Zsh{opts: opts, p: paths.For(opts.RepoRoot, opts.Home).Zsh}
}

// Name returns "zsh".
func (z *Zsh) Name() string { return "zsh" }

// Pull copies zshrc, aliases, and profile into HOME.
func (z *Zsh) Pull() error {
	copies := []struct{ src, dst string }{
		{z.p.ZshrcRepo, z.p.ZshrcLocal},
		{z.p.AliasesRepo, z.p.AliasesLocal},
		{z.p.ProfileRepo, z.p.ProfileLocal},
	}
	for _, c := range copies {
		if err := fsutil.SafeCopy(c.src, c.dst, z.Name(), z.opts.BackupRoot); err != nil {
			return err
		}
	}
	if _, err := os.Stat(z.p.FuncsRepo); err == nil {
		if err := fsutil.SafeCopy(z.p.FuncsRepo, z.p.FuncsLocal, z.Name(), z.opts.BackupRoot); err != nil {
			return err
		}
	}
	return z.seedSecret()
}

// seedSecret copies the template to ~/.zshrc_secret iff the local file does
// not yet exist. Existing local secrets are left untouched (they hold real keys).
func (z *Zsh) seedSecret() error {
	if _, err := os.Stat(z.p.SecretLocal); err == nil {
		return nil
	}
	if _, err := os.Stat(z.p.SecretTemplate); err != nil {
		return nil
	}
	return fsutil.SafeCopy(z.p.SecretTemplate, z.p.SecretLocal, z.Name(), z.opts.BackupRoot)
}
