// Package fsutil ports backup_file and safe_copy from scripts/lib/common.sh.
package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var versionRe = regexp.MustCompile(`^v(\d+)$`)

// Backup copies src into <backupRoot>/<component>/v<N>/ and returns the v<N> dir.
// N is the highest existing v<digits> directory under the component dir + 1, or 1.
func Backup(src, component, backupRoot string) (string, error) {
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
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return "", err
	}
	return dst, copyPath(src, filepath.Join(dst, filepath.Base(src)))
}

// SafeCopy validates src exists, backs up dst (if present) under component,
// then copies src to dst, creating dst's parent if needed.
func SafeCopy(src, dst, component, backupRoot string) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}
	if _, err := os.Stat(dst); err == nil {
		if _, err := Backup(dst, component, backupRoot); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return copyPath(src, dst)
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
