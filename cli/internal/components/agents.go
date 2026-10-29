package components

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"tars/internal/paths"
)

// pullRules renders the shared agent rules, minus the skipped ones, under header
// into dest on behalf of component; every agent adapter calls it with its own
// header and destination.
func pullRules(opts Options, component, header, dest string) error {
	dir := paths.For(opts.RepoRoot, opts.Home).Agents.RulesRepo
	content, err := renderRules(dir, header, opts.Agents.SkipRules)
	if err != nil {
		return err
	}
	return opts.copier().SafeWrite(content, 0o644, dest, component, opts.BackupRoot)
}

// pullEditGuard copies the shared edit-guard rule table to dest, where the
// agent's guard reads it, on behalf of component.
func pullEditGuard(opts Options, component, dest string) error {
	src := paths.For(opts.RepoRoot, opts.Home).Agents.EditGuardRepo
	return opts.copier().SafeCopy(src, dest, component, opts.BackupRoot)
}

// renderRules concatenates header and every rule file in dir not in skip.
func renderRules(dir, header string, skip []string) ([]byte, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString(header)
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		if slices.Contains(skip, name) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		buf.Write(b)
		buf.WriteString("\n")
	}
	return buf.Bytes(), nil
}
