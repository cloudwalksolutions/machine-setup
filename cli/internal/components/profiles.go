package components

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"tars/internal/paths"
	"tars/internal/profiles"
)

// Profiles renders each account identity into git and shell config files.
type Profiles struct {
	opts       Options
	p          paths.ProfilesPaths
	ConfigPath string
}

// NewProfiles returns a Profiles component reading the default profiles config.
func NewProfiles(opts Options) *Profiles {
	return &Profiles{
		opts:       opts,
		p:          paths.For(opts.RepoRoot, opts.Home).Profiles,
		ConfigPath: profiles.DefaultPath(),
	}
}

// Name returns "profiles".
func (c *Profiles) Name() string { return "profiles" }

// Pull renders the configured profiles; a machine with no config is left alone.
func (c *Profiles) Pull() error {
	f, err := profiles.Load(c.ConfigPath, c.opts.Home)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, p := range f.Profiles {
		gitconfig := profiles.RenderGitconfig(p, c.opts.Home)
		dst := profiles.GitconfigPath(c.p.Dir, p.Alias)
		if err := c.opts.copier().SafeWrite([]byte(gitconfig), 0o644, dst, c.Name(), c.opts.BackupRoot); err != nil {
			return err
		}
		if err := c.seedEnv(p); err != nil {
			return err
		}
		c.warnIfKeyMissing(p)
	}
	return c.spliceGitconfig(f)
}

func (c *Profiles) warnIfKeyMissing(p profiles.Profile) {
	key := p.KeyPath(c.opts.Home)
	if _, err := os.Stat(key); os.IsNotExist(err) {
		fmt.Fprintf(c.opts.Stderr, "  profiles: %s has no key at %s\n", p.Alias, key)
	}
}

func (c *Profiles) seedEnv(p profiles.Profile) error {
	dst := filepath.Join(c.p.Dir, p.Alias+".env")
	if _, err := os.Stat(dst); err == nil {
		return nil
	}
	return c.opts.copier().SafeWrite([]byte(profiles.RenderEnv(p)), 0o600, dst, c.Name(), c.opts.BackupRoot)
}

func (c *Profiles) spliceGitconfig(f profiles.File) error {
	active, err := profiles.Active(c.p.Dir)
	if err != nil {
		return err
	}
	existing, err := os.ReadFile(c.p.Gitconfig)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	merged, err := profiles.Splice(existing, profiles.RenderBlock(f, active, c.p.Dir))
	if err != nil {
		return err
	}
	return c.opts.copier().SafeWrite(merged, 0o644, c.p.Gitconfig, c.Name(), c.opts.BackupRoot)
}
