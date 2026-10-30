package components

import (
	"os"

	"tars/internal/paths"
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
		if err := z.opts.copier().SafeCopy(c.src, c.dst, z.Name(), z.opts.BackupRoot); err != nil {
			return err
		}
	}
	if _, err := os.Stat(z.p.FuncsRepo); err == nil {
		if err := z.opts.copier().SafeCopy(z.p.FuncsRepo, z.p.FuncsLocal, z.Name(), z.opts.BackupRoot); err != nil {
			return err
		}
	}
	if err := z.seed(z.p.SecretTemplate, z.p.SecretLocal); err != nil {
		return err
	}
	return z.seed(z.p.ProfileLocalTemplate, z.p.ProfileLocalOverride)
}

// Push copies local zsh files back to the repo, archiving the repo copies under
// "zsh-repo". The profile is pushed only when it exists locally; funcs stay personal.
func (z *Zsh) Push() error {
	comp := z.Name() + "-repo"
	if err := z.opts.copier().SafeCopy(z.p.ZshrcLocal, z.p.ZshrcRepo, comp, z.opts.BackupRoot); err != nil {
		return err
	}
	if err := z.opts.copier().SafeCopy(z.p.AliasesLocal, z.p.AliasesRepo, comp, z.opts.BackupRoot); err != nil {
		return err
	}
	if _, err := os.Stat(z.p.ProfileLocal); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return z.opts.copier().SafeCopy(z.p.ProfileLocal, z.p.ProfileRepo, comp, z.opts.BackupRoot)
}

// seed copies template to local only when local does not already exist.
func (z *Zsh) seed(template, local string) error {
	if _, err := os.Stat(local); err == nil {
		return nil
	}
	if _, err := os.Stat(template); err != nil {
		return nil
	}
	return z.opts.copier().SafeCopy(template, local, z.Name(), z.opts.BackupRoot)
}
