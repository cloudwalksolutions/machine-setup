package components

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"tars/internal/fsutil"
)

// shareableKeys are the settings.json keys safe to sync between machines.
var shareableKeys = []string{"model", "theme", "enabledPlugins"}

// settings is a schemaless settings.json; unknown keys pass through untouched.
type settings map[string]any

// hookGroup is one entry of settings.hooks.<event>; identified by its commands.
type hookGroup struct {
	Matcher string           `json:"matcher,omitempty"`
	Hooks   []map[string]any `json:"hooks"`
}

func (g hookGroup) key() string {
	var cmds []string
	for _, h := range g.Hooks {
		if cmd, ok := h["command"].(string); ok {
			cmds = append(cmds, cmd)
		}
	}
	return strings.Join(cmds, "\n")
}

// loadSettings parses path; a missing file yields an empty settings plus the
// os.IsNotExist error so callers choose between "start empty" and "skip".
func loadSettings(path string) (settings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return settings{}, err
	}
	var s settings
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return s, nil
}

func (s settings) events() []string {
	hooks, _ := s["hooks"].(map[string]any)
	events := make([]string, 0, len(hooks))
	for event := range hooks {
		events = append(events, event)
	}
	slices.Sort(events)
	return events
}

func (s settings) hooks(event string) []hookGroup {
	var groups []hookGroup
	if hooks, ok := s["hooks"].(map[string]any); ok {
		convert(hooks[event], &groups)
	}
	return groups
}

func (s settings) setHooks(event string, groups []hookGroup) {
	hooks, _ := s["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
		s["hooks"] = hooks
	}
	var raw any
	convert(groups, &raw)
	hooks[event] = raw
}

// save writes s as indented JSON with a backup first; identical bytes are a no-op.
func (s settings) save(path, component, backupRoot string) error {
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, out) {
		return nil
	}
	if _, err := fsutil.Backup(path, component, backupRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// upsert replaces each base group sharing a key with incoming, appending new ones.
func upsert(base, incoming []hookGroup) []hookGroup {
	for _, in := range incoming {
		i := slices.IndexFunc(base, func(g hookGroup) bool { return g.key() == in.key() })
		if i < 0 {
			base = append(base, in)
		} else {
			base[i] = in
		}
	}
	return base
}

// refresh updates base groups from the matching local ones, ignoring local-only groups.
func refresh(base, local []hookGroup) []hookGroup {
	for i, g := range base {
		if j := slices.IndexFunc(local, func(l hookGroup) bool { return l.key() == g.key() }); j >= 0 {
			base[i] = local[j]
		}
	}
	return base
}

// convert round-trips v through JSON into out.
func convert(v, out any) {
	b, _ := json.Marshal(v)
	_ = json.Unmarshal(b, out)
}
