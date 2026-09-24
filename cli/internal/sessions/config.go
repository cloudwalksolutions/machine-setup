// Package sessions manages the declarative byobu session config and launcher.
package sessions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Session is one named byobu session; each dir becomes a window.
type Session struct {
	Name string   `yaml:"name"`
	Dirs []string `yaml:"dirs"`
}

// File is the parsed sessions config.
type File struct {
	Sessions []Session `yaml:"sessions"`
}

// DefaultPath is ~/.config/tars/sessions.yaml unless TARS_SESSIONS_PATH overrides it.
func DefaultPath() string {
	if envPath := os.Getenv("TARS_SESSIONS_PATH"); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "tars/sessions.yaml"
	}
	return filepath.Join(home, ".config", "tars", "sessions.yaml")
}

// Load reads and parses the sessions config at path.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return File{}, err
	}
	for _, s := range f.Sessions {
		// tmux rejects '.'/':' in session names (target separators).
		if strings.ContainsAny(s.Name, ".:") {
			return File{}, fmt.Errorf("invalid session name %q: '.' and ':' are not allowed", s.Name)
		}
		if len(s.Dirs) == 0 {
			return File{}, fmt.Errorf("session %q has no dirs", s.Name)
		}
		for i, dir := range s.Dirs {
			s.Dirs[i] = expandHome(dir)
		}
	}
	return f, nil
}

const seedContent = `# tars sessions config — each session opens one byobu window per dir.
# Session names may not contain '.' or ':'.
sessions:
  - name: cloudwalk
    dirs:
      - ~/Desktop/projects/cloudwalk/api
      - ~/Desktop/projects/cloudwalk/infra
`

// Seed writes a commented example config if none exists. Idempotent.
func Seed(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(seedContent), 0o644)
}

// WindowName derives a tmux-safe window name from a dir's basename.
func WindowName(dir string) string {
	name := filepath.Base(dir)
	return strings.NewReplacer(".", "_", ":", "_").Replace(name)
}

// expandHome resolves a leading ~ (tmux's -c does not expand it).
func expandHome(dir string) string {
	if dir != "~" && !strings.HasPrefix(dir, "~/") {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return dir
	}
	return filepath.Join(home, strings.TrimPrefix(dir, "~"))
}
