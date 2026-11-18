package components

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"tars/internal/config"
	"tars/internal/paths"
)

// Claude provisions Claude Code: the edit-blocking hook, the shareable
// settings fragment, and the repo's rule files as user-level rules.
type Claude struct {
	opts Options
	cfg  config.ClaudeConfig
	p    paths.ClaudePaths
}

// NewClaude builds the component from opts, including the `tars init claude` choices.
func NewClaude(opts Options) *Claude {
	return &Claude{opts: opts, cfg: opts.Claude, p: paths.For(opts.RepoRoot, opts.Home).Claude}
}

func (c *Claude) Name() string { return "claude" }

// Pull applies the repo's Claude config to ~/.claude.
func (c *Claude) Pull() error {
	if enabled(c.cfg.Hook) {
		if err := c.copyHooks(c.p.HooksRepo, c.p.HooksLocal, c.Name()); err != nil {
			return err
		}
		if err := pullEditGuard(c.opts, c.Name(), c.p.EditGuardLocal); err != nil {
			return err
		}
	}
	if enabled(c.cfg.Settings) {
		if err := c.pullSettings(); err != nil {
			return err
		}
	}
	if err := c.pullSkills(); err != nil {
		return err
	}
	return c.pullRules()
}

// pullSkills copies each skill folder the repo ships into ~/.claude/skills,
// next to the user's own skills.
func (c *Claude) pullSkills() error {
	return filepath.WalkDir(c.p.SkillsRepo, func(src string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(c.p.SkillsRepo, src)
		if err != nil {
			return err
		}
		return c.opts.copier().SafeCopy(src, filepath.Join(c.p.SkillsLocal, rel), c.Name(), c.opts.BackupRoot)
	})
}

// Push copies the local hook and the shareable settings keys back to the repo,
// archiving the repo copies under "claude-repo".
func (c *Claude) Push() error {
	comp := c.Name() + "-repo"
	if err := c.copyHooks(c.p.HooksLocal, c.p.HooksRepo, comp); err != nil {
		return err
	}
	local, err := loadSettings(c.p.SettingsLocal)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	fragment, err := loadSettings(c.p.SettingsRepo)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, k := range shareableKeys {
		if v, ok := local[k]; ok {
			fragment[k] = v
		}
	}
	for _, event := range fragment.events() {
		fragment.setHooks(event, refresh(fragment.hooks(event), local.hooks(event)))
	}
	return fragment.save(c.opts.copier(), c.p.SettingsRepo, comp, c.opts.BackupRoot)
}

// copyHooks copies the hooks the repo ships, never machine-only hooks.
func (c *Claude) copyHooks(src, dst, component string) error {
	shipped, err := os.ReadDir(c.p.HooksRepo)
	if err != nil {
		return err
	}
	for _, e := range shipped {
		if err := c.opts.copier().SafeCopy(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name()), component, c.opts.BackupRoot); err != nil {
			return err
		}
	}
	return nil
}

// pullSettings merges the fragment into ~/.claude/settings.json: shareable keys
// are overwritten, hook groups upserted by command, everything else kept.
func (c *Claude) pullSettings() error {
	fragment, err := loadSettings(c.p.SettingsRepo)
	if err != nil {
		return err
	}
	local, err := loadSettings(c.p.SettingsLocal)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, k := range shareableKeys {
		if v, ok := fragment[k]; ok {
			local[k] = v
		}
	}
	if enabled(c.cfg.Hook) {
		for _, event := range fragment.events() {
			local.setHooks(event, upsert(local.hooks(event), fragment.hooks(event)))
		}
	}
	return local.save(c.opts.copier(), c.p.SettingsLocal, c.Name(), c.opts.BackupRoot)
}

const claudeMDHeader = "# Global Claude Code rules\n\nManaged by `tars init agents`; edit the rule files under agents/rules in the machine-setup repo, not this file.\n\n"

func (c *Claude) pullRules() error {
	dir := paths.For(c.opts.RepoRoot, c.opts.Home).Agents.RulesRepo
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	wanted := map[string]bool{}
	for _, e := range entries {
		if slices.Contains(c.opts.Agents.SkipRules, strings.TrimSuffix(e.Name(), ".md")) {
			continue
		}
		wanted[e.Name()] = true
		if err := c.opts.copier().SafeCopy(filepath.Join(dir, e.Name()), filepath.Join(c.p.RulesLocal, e.Name()), c.Name(), c.opts.BackupRoot); err != nil {
			return err
		}
	}
	if err := c.removeStaleRules(wanted); err != nil {
		return err
	}
	return c.removeRenderedClaudeMD()
}

func (c *Claude) removeRenderedClaudeMD() error {
	content, err := os.ReadFile(c.p.ClaudeMDLocal)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(content, []byte(claudeMDHeader)) {
		return nil
	}
	if _, err := c.opts.copier().Backup(c.p.ClaudeMDLocal, c.Name(), c.opts.BackupRoot); err != nil {
		return err
	}
	return c.opts.copier().RemoveAll(c.p.ClaudeMDLocal)
}

func (c *Claude) removeStaleRules(wanted map[string]bool) error {
	copied, err := os.ReadDir(c.p.RulesLocal)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range copied {
		if wanted[e.Name()] {
			continue
		}
		stale := filepath.Join(c.p.RulesLocal, e.Name())
		if _, err := c.opts.copier().Backup(stale, c.Name(), c.opts.BackupRoot); err != nil {
			return err
		}
		if err := c.opts.copier().RemoveAll(stale); err != nil {
			return err
		}
	}
	return nil
}

// enabled treats an unset toggle as on.
func enabled(b *bool) bool { return b == nil || *b }
