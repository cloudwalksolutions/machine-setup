package profiles

import (
	"os"
	"path/filepath"
	"strings"
)

const activeFile = "active"

// Active returns the alias recorded by SetActive, or "" when none is set.
func Active(dir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dir, activeFile))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// SetActive records alias as the machine-wide default profile.
func SetActive(dir, alias string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, activeFile), []byte(alias+"\n"), 0o644)
}
