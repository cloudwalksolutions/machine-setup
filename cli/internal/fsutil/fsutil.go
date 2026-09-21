// Package fsutil provides versioned backups and safe (backup-before-overwrite)
// copies shared by all components.
package fsutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var versionRe = regexp.MustCompile(`^v(\d+)$`)

// Copier performs backup-before-overwrite copies. With DryRun set it reports
// each intended action to Log and writes nothing.
type Copier struct {
	DryRun bool
	Log    io.Writer
}

func (c Copier) report(format string, args ...any) {
	if c.Log == nil {
		return
	}
	fmt.Fprintf(c.Log, format+"\n", args...)
}

// Backup copies src into <backupRoot>/<component>/v<N>/ and returns the v<N> dir.
// N is the highest existing v<digits> directory under the component dir + 1, or 1.
func (c Copier) Backup(src, component, backupRoot string) (string, error) {
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	componentDir := filepath.Join(backupRoot, component)
	version, err := nextVersion(componentDir)
	if err != nil {
		return "", err
	}
	dst := filepath.Join(componentDir, fmt.Sprintf("v%d", version))
	if c.DryRun {
		return dst, nil
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return "", err
	}
	return dst, copyPath(src, filepath.Join(dst, filepath.Base(src)))
}

// SafeCopy validates src exists, backs up dst (if present) under component,
// then copies src to dst, creating dst's parent if needed.
func (c Copier) SafeCopy(src, dst, component, backupRoot string) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}
	if _, err := os.Stat(dst); err == nil {
		// Idempotent: if dst already matches src, do nothing — no backup, no copy.
		if same, err := SameContent(src, dst); err != nil {
			return err
		} else if same {
			if c.DryRun {
				c.report("    unchanged  %s", dst)
			}
			return nil
		}
		version, err := c.Backup(dst, component, backupRoot)
		if err != nil {
			return err
		}
		if c.DryRun {
			c.report("    would overwrite  %s (backup %s)", dst, filepath.Base(version))
			return nil
		}
	} else if !os.IsNotExist(err) {
		return err
	} else if c.DryRun {
		c.report("    would create  %s", dst)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return copyPath(src, dst)
}

// RemoveAll deletes path, or reports the intent when in dry-run mode.
func (c Copier) RemoveAll(path string) error {
	if c.DryRun {
		if _, err := os.Stat(path); err == nil {
			c.report("    would remove  %s", path)
		}
		return nil
	}
	return os.RemoveAll(path)
}

// Backup is the non-dry-run form of Copier.Backup.
func Backup(src, component, backupRoot string) (string, error) {
	return Copier{}.Backup(src, component, backupRoot)
}

// SafeCopy is the non-dry-run form of Copier.SafeCopy.
func SafeCopy(src, dst, component, backupRoot string) error {
	return Copier{}.SafeCopy(src, dst, component, backupRoot)
}

// SameContent reports whether src and dst are regular files with identical bytes.
// Directories always report false so directory copies proceed normally.
func SameContent(src, dst string) (bool, error) {
	si, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	di, err := os.Stat(dst)
	if err != nil {
		return false, err
	}
	if si.IsDir() || di.IsDir() || si.Size() != di.Size() {
		return false, nil
	}
	a, err := os.ReadFile(src)
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(dst)
	if err != nil {
		return false, err
	}
	return bytes.Equal(a, b), nil
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func nextVersion(componentDir string) (int, error) {
	entries, err := os.ReadDir(componentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m := versionRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
