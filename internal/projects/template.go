package projects

import (
	"fmt"
	"path/filepath"
	"strings"

	"tars/internal/sessions"
)

// Template is what a new session for a project looks like; window dirs are project-relative.
type Template struct {
	Windows []sessions.Window `yaml:"windows"`
}

// File is the parsed projects config: templates keyed by project name.
type File struct {
	Projects map[string]Template `yaml:"projects"`
}

// String writes one `name | dir | command` window per line, as ParseTemplate reads them.
func (t Template) String() string {
	var b strings.Builder
	for _, w := range t.Windows {
		fmt.Fprintf(&b, "%s | %s | %s\n", w.Name, w.Dir, w.Command)
	}
	return b.String()
}

// ParseTemplate reads one `name | dir | command` window per line.
func ParseTemplate(text string) (Template, error) {
	var t Template
	for n, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var fields [3]string
		for i, f := range strings.SplitN(line, "|", 3) {
			fields[i] = strings.TrimSpace(f)
		}
		switch name := fields[0]; {
		case name == "":
			return Template{}, fmt.Errorf("line %d: window needs a name", n+1)
		case strings.ContainsAny(name, ".:"):
			return Template{}, fmt.Errorf("line %d: window %q: '.' and ':' are not allowed", n+1, name)
		}
		t.Windows = append(t.Windows, sessions.Window{Name: fields[0], Dir: fields[1], Command: fields[2]})
	}
	return t, nil
}

// Template is p's configured template, or one window at its root.
func (f File) Template(p Project) Template {
	if t := f.Projects[p.Name]; len(t.Windows) > 0 {
		return t
	}
	return Template{Windows: []sessions.Window{{Name: sessions.WindowName(p.Dir)}}}
}

// Windows resolves the windows of a new session for p.
func (f File) Windows(p Project) []sessions.Window {
	t := f.Template(p)
	windows := make([]sessions.Window, len(t.Windows))
	for i, w := range t.Windows {
		w.Dir = filepath.Join(p.Dir, w.Dir)
		windows[i] = w
	}
	return windows
}
