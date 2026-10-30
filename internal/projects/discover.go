// Package projects finds the git repos under projects_dir and the session template for each.
package projects

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"tars/internal/sessions"
)

// Project is a git repo under projects_dir; Name doubles as its byobu session name.
type Project struct {
	Name string
	Dir  string
}

// Discover lists the git repos at root/<repo> and root/<group>/<repo>.
func Discover(root string) ([]Project, error) {
	found, err := walk(root, 2)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	count := map[string]int{}
	for _, p := range found {
		count[p.Name]++
	}
	for i, p := range found {
		if count[p.Name] > 1 {
			p.Name = filepath.Base(filepath.Dir(p.Dir)) + "-" + p.Name
		}
		found[i].Name = sessions.WindowName(p.Name)
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Name < found[j].Name })
	return found, nil
}

func walk(dir string, depth int) ([]Project, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var found []Project
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		child := filepath.Join(dir, e.Name())
		if isRepo(child) {
			found = append(found, Project{Name: e.Name(), Dir: child})
			continue
		}
		if depth > 1 {
			nested, err := walk(child, depth-1)
			if err != nil {
				return nil, err
			}
			found = append(found, nested...)
		}
	}
	return found, nil
}

func isRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
