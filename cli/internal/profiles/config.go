// Package profiles manages per-account identities: git name/email, GitHub user,
// SSH key, and the project directory each one applies to.
package profiles

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile is one account identity; its dir and ssh key follow from name and alias.
type Profile struct {
	Name     string `yaml:"name"`
	Alias    string `yaml:"alias"`
	Email    string `yaml:"email"`
	GitHub   string `yaml:"github"`
	FullName string `yaml:"full_name"`
}

// File is the parsed profiles config.
type File struct {
	ProjectsDir string    `yaml:"projects_dir"`
	FullName    string    `yaml:"full_name"`
	Profiles    []Profile `yaml:"profiles"`
}

var identifierRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// Load reads and parses the profiles config at path, resolving conventions against home.
func Load(path, home string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return File{}, err
	}
	if f.ProjectsDir == "" {
		f.ProjectsDir = filepath.Join(home, "Desktop", "projects")
	}
	f.ProjectsDir = expandHome(f.ProjectsDir, home)
	names, aliases := map[string]bool{}, map[string]bool{}
	for i := range f.Profiles {
		p := &f.Profiles[i]
		if p.FullName == "" {
			p.FullName = f.FullName
		}
		if err := p.Validate(); err != nil {
			return File{}, err
		}
		if names[p.Name] || aliases[p.Alias] {
			return File{}, fmt.Errorf("profile %q (%s) repeats an existing name or alias", p.Name, p.Alias)
		}
		names[p.Name], aliases[p.Alias] = true, true
	}
	sort.Slice(f.Profiles, func(i, j int) bool { return f.Profiles[i].Alias < f.Profiles[j].Alias })
	return f, nil
}

// Validate reports the first missing or malformed field of a resolved profile.
func (p Profile) Validate() error {
	for _, id := range []string{p.Name, p.Alias} {
		if !identifierRe.MatchString(id) {
			return fmt.Errorf("invalid profile identifier %q: use lowercase letters, digits, '-' or '_'", id)
		}
	}
	if p.Email == "" {
		return fmt.Errorf("profile %q has no email", p.Alias)
	}
	if p.GitHub == "" {
		return fmt.Errorf("profile %q has no github user", p.Alias)
	}
	if p.FullName == "" {
		return fmt.Errorf("profile %q has no full_name (set it per profile or at the top level)", p.Alias)
	}
	return nil
}

// Dir is where the profile's repositories live: <projectsDir>/<name>.
func (p Profile) Dir(projectsDir string) string {
	return filepath.Join(projectsDir, p.Name)
}

const seedContent = `# tars profiles — one identity per account.
# Each profile owns <projects_dir>/<name> and signs with ~/.ssh/id_rsa.<alias>.
# projects_dir: ~/Desktop/projects
full_name: Your Name
profiles:
  - name: cloudwalk
    alias: cws
    email: you@cloudwalk.example
    github: your-github-user
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

// Append adds p as a new list item at the end of the config, seeding it first when
// missing, so the user's own comments and ordering survive.
func Append(path string, p Profile) error {
	if err := Seed(path); err != nil {
		return err
	}
	existing, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var b strings.Builder
	b.Write(existing)
	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "  - name: %s\n    alias: %s\n    email: %s\n    github: %s\n", p.Name, p.Alias, p.Email, p.GitHub)
	if p.FullName != "" {
		fmt.Fprintf(&b, "    full_name: %s\n", p.FullName)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// Find returns the profile whose alias or name matches id.
func (f File) Find(id string) (Profile, bool) {
	for _, p := range f.Profiles {
		if p.Alias == id || p.Name == id {
			return p, true
		}
	}
	return Profile{}, false
}

// KeyPath is the profile's private key: ~/.ssh/id_rsa.<alias>.
func (p Profile) KeyPath(home string) string {
	return filepath.Join(home, ".ssh", "id_rsa."+p.Alias)
}

func expandHome(dir, home string) string {
	if dir != "~" && !strings.HasPrefix(dir, "~/") {
		return dir
	}
	return filepath.Join(home, strings.TrimPrefix(dir, "~"))
}

// DefaultPath is ~/.config/tars/profiles.yaml unless TARS_PROFILES_PATH overrides it.
func DefaultPath() string {
	if envPath := os.Getenv("TARS_PROFILES_PATH"); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "tars/profiles.yaml"
	}
	return filepath.Join(home, ".config", "tars", "profiles.yaml")
}
