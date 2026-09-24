package components

import (
	"os"

	"tars/internal/config"
	"tars/internal/fsutil"
	"tars/internal/paths"
)

// Claude provisions Claude Code: the edit-blocking hook, the shareable
// settings fragment, and a global CLAUDE.md rendered from the repo's rule files.
type Claude struct {
	opts Options
	cfg  config.ClaudeConfig
	p    paths.ClaudePaths
}

// NewClaude builds the component from opts, including the `tars claude init` choices.
func NewClaude(opts Options) *Claude {
	return &Claude{opts: opts, cfg: opts.Claude, p: paths.For(opts.RepoRoot, opts.Home).Claude}
}

func (c *Claude) Name() string { return "claude" }

// Pull applies the repo's Claude config to ~/.claude.
func (c *Claude) Pull() error {
	if enabled(c.cfg.Hook) {
		if err := fsutil.SafeCopy(c.p.HookRepo, c.p.HookLocal, c.Name(), c.opts.BackupRoot); err != nil {
			return err
		}
	}
	if enabled(c.cfg.Settings) {
		if err := c.pullSettings(); err != nil {
			return err
		}
	}
	return c.pullRules()
}

// Push copies the local hook and the shareable settings keys back to the repo,
// archiving the repo copies under "claude-repo".
func (c *Claude) Push() error {
	comp := c.Name() + "-repo"
	if err := fsutil.SafeCopy(c.p.HookLocal, c.p.HookRepo, comp, c.opts.BackupRoot); err != nil {
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
	return fragment.save(c.p.SettingsRepo, comp, c.opts.BackupRoot)
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
	for _, event := range fragment.events() {
		local.setHooks(event, upsert(local.hooks(event), fragment.hooks(event)))
	}
	return local.save(c.p.SettingsLocal, c.Name(), c.opts.BackupRoot)
}

const claudeMDHeader = "# Global Claude Code rules\n\nManaged by `tars claude init`; edit the rule files in the machine-setup repo, not this file.\n\n"

// pullRules renders the selected rule files into CLAUDE.md.
func (c *Claude) pullRules() error {
	content, err := renderRules(c.p.RulesRepo, claudeMDHeader, c.cfg.Rules)
	if err != nil {
		return err
	}
	return c.opts.copier().SafeWrite(content, 0o644, c.p.ClaudeMDLocal, c.Name(), c.opts.BackupRoot)
}

// enabled treats an unset toggle as on.
func enabled(b *bool) bool { return b == nil || *b }
