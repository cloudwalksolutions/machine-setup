package components

import (
	"os"

	"github.com/cloudwalk/machine-setup/internal/fsutil"
	"github.com/cloudwalk/machine-setup/internal/paths"
)

// Zsh pulls/pushes the zsh dotfiles and seeds the secret template.
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

// Push copies local zsh files back to the repo, archiving the repo copies under
// "zsh-repo". Funcs and profile are pushed only when they exist locally.
func (z *Zsh) Push() error {
	comp := z.Name() + "-repo"
	if err := fsutil.SafeCopy(z.p.ZshrcLocal, z.p.ZshrcRepo, comp, z.opts.BackupRoot); err != nil {
		return err
	}
	if err := fsutil.SafeCopy(z.p.AliasesLocal, z.p.AliasesRepo, comp, z.opts.BackupRoot); err != nil {
		return err
	}
	for _, c := range []struct{ local, repo string }{
		{z.p.FuncsLocal, z.p.FuncsRepo},
		{z.p.ProfileLocal, z.p.ProfileRepo},
	} {
		if _, err := os.Stat(c.local); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err := fsutil.SafeCopy(c.local, c.repo, comp, z.opts.BackupRoot); err != nil {
			return err
		}
	}
	return nil
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
