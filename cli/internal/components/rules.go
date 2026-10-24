package components

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

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
