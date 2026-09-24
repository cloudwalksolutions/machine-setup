package components

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"

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

// pullRules renders the selected rule files into CLAUDE.md via a temp file so
// SafeCopy provides backup-before-overwrite and the identical-content skip.
func (c *Claude) pullRules() error {
	entries, err := os.ReadDir(c.p.RulesRepo)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString(claudeMDHeader)
	for _, e := range entries {
		if !c.ruleSelected(strings.TrimSuffix(e.Name(), ".md")) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(c.p.RulesRepo, e.Name()))
		if err != nil {
			return err
		}
		buf.Write(b)
		buf.WriteString("\n")
	}
	tmp, err := os.CreateTemp("", "tars-claude-*.md")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return fsutil.SafeCopy(tmp.Name(), c.p.ClaudeMDLocal, c.Name(), c.opts.BackupRoot)
}

// ruleSelected reports whether a rule is in the configured set; no config means all.
func (c *Claude) ruleSelected(name string) bool {
	return len(c.cfg.Rules) == 0 || slices.Contains(c.cfg.Rules, name)
}

// enabled treats an unset toggle as on.
func enabled(b *bool) bool { return b == nil || *b }
