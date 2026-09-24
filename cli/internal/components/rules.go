package components

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// renderRules concatenates header plus the selected rule files from dir in
// name order; an empty selection means every rule.
func renderRules(dir, header string, selected []string) ([]byte, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteString(header)
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		if len(selected) > 0 && !slices.Contains(selected, name) {
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
