package components

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"tars/internal/paths"
)

func pullRules(opts Options, component, header, dest string) error {
	dir := paths.For(opts.RepoRoot, opts.Home).Agents.RulesRepo
	content, err := renderRules(dir, header, opts.Agents.SkipRules)
	if err != nil {
		return err
	}
	return opts.copier().SafeWrite(content, 0o644, dest, component, opts.BackupRoot)
}

func pullEditGuard(opts Options, component, dest string) error {
	src := paths.For(opts.RepoRoot, opts.Home).Agents.EditGuardRepo
	return opts.copier().SafeCopy(src, dest, component, opts.BackupRoot)
}

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
