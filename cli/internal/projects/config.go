package projects

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultPath is ~/.config/tars/projects.yaml unless TARS_PROJECTS_PATH overrides it.
func DefaultPath() string {
	if envPath := os.Getenv("TARS_PROJECTS_PATH"); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "tars/projects.yaml"
	}
	return filepath.Join(home, ".config", "tars", "projects.yaml")
}

// Load reads the templates at path; a missing file means every project uses the default.
func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return File{}, nil
	}
	if err != nil {
		return File{}, err
	}
	var f File
	err = yaml.Unmarshal(data, &f)
	return f, err
}

const seedContent = `# tars projects: what a new byobu session looks like for each project.
# Projects are the git repos under projects_dir (see profiles.yaml); one
# without an entry opens a single window at its root. Dirs are project-relative.
#
# projects:
#   api:
#     windows:
#       - {name: code, command: nvim}
#       - {name: server, command: make dev}
#       - {name: web, dir: frontend}
`

// Seed writes a commented example config if none exists.
func Seed(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(seedContent), 0o644)
}

// Save writes f to path, creating its parent dirs.
func Save(path string, f File) error {
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
