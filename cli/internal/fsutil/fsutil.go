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
		// Idempotent: if dst already matches src, only repair a drifted mode —
		// no backup, no copy.
		if same, err := SameContent(src, dst); err != nil {
			return err
		} else if same {
			return sameMode(src, dst)
		}
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

// SameTree reports whether src and dst are directories containing identical
// trees: the same relative entries, the same file bytes, the same exec bits.
// A missing or non-directory dst reports false without error.
func SameTree(src, dst string) (bool, error) {
	di, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !di.IsDir() {
		return false, nil
	}

	srcEntries, err := treeEntries(src)
	if err != nil {
		return false, err
	}
	dstEntries, err := treeEntries(dst)
	if err != nil {
		return false, err
	}
	if len(srcEntries) != len(dstEntries) {
		return false, nil
	}
	for rel, isDir := range srcEntries {
		dstIsDir, ok := dstEntries[rel]
		if !ok || isDir != dstIsDir {
			return false, nil
		}
		if isDir {
			continue
		}
		same, err := SameContent(filepath.Join(src, rel), filepath.Join(dst, rel))
		if err != nil {
			return false, err
		}
		if !same {
			return false, nil
		}
		if execBit(filepath.Join(src, rel)) != execBit(filepath.Join(dst, rel)) {
			return false, nil
		}
	}
	return true, nil
}

// treeEntries maps each relative path under root to whether it is a directory.
func treeEntries(root string) (map[string]bool, error) {
	entries := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		entries[rel] = d.IsDir()
		return nil
	})
	return entries, err
}

func execBit(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().Perm()&0o111 != 0
}

// sameMode chmods dst to src's permissions when they differ.
func sameMode(src, dst string) error {
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	di, err := os.Stat(dst)
	if err != nil {
		return err
	}
	if si.Mode().Perm() == di.Mode().Perm() {
		return nil
	}
	return os.Chmod(dst, si.Mode().Perm())
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
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	// The open perm is umask-filtered and ignored for pre-existing files;
	// chmod makes the copy's mode match the source unconditionally.
	return os.Chmod(dst, info.Mode().Perm())
}
