package fsutil_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/fsutil"
)

var _ = Describe("Backup", func() {
	var (
		tmp        string
		backupRoot string
	)

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		backupRoot = filepath.Join(tmp, "backups")
	})

	It("creates v1 under the component dir on the first backup", func() {
		src := filepath.Join(tmp, "src.txt")
		Expect(os.WriteFile(src, []byte("hello"), 0o644)).To(Succeed())

		dst, err := fsutil.Backup(src, "zsh", backupRoot)
		Expect(err).NotTo(HaveOccurred())
		Expect(dst).To(Equal(filepath.Join(backupRoot, "zsh", "v1")))
		Expect(filepath.Join(dst, "src.txt")).To(BeARegularFile())
	})

	It("increments to vN+1 when prior versions exist", func() {
		src := filepath.Join(tmp, "src.txt")
		Expect(os.WriteFile(src, []byte("x"), 0o644)).To(Succeed())

		_, err := fsutil.Backup(src, "zsh", backupRoot)
		Expect(err).NotTo(HaveOccurred())
		dst, err := fsutil.Backup(src, "zsh", backupRoot)
		Expect(err).NotTo(HaveOccurred())
		Expect(dst).To(Equal(filepath.Join(backupRoot, "zsh", "v2")))
	})

	It("is a no-op (no error, empty dst) when src does not exist", func() {
		dst, err := fsutil.Backup(filepath.Join(tmp, "nope"), "zsh", backupRoot)
		Expect(err).NotTo(HaveOccurred())
		Expect(dst).To(BeEmpty())
	})

	It("backs up a directory tree recursively", func() {
		srcDir := filepath.Join(tmp, "config")
		Expect(os.MkdirAll(filepath.Join(srcDir, "sub"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("A"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("B"), 0o644)).To(Succeed())

		dst, err := fsutil.Backup(srcDir, "nvim", backupRoot)
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Join(dst, "config", "a.txt")).To(BeARegularFile())
		Expect(filepath.Join(dst, "config", "sub", "b.txt")).To(BeARegularFile())
	})
})

var _ = Describe("SameTree", func() {
	var tmp string

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
	})

	writeTree := func(root string, files map[string]string) {
		for rel, content := range files {
			path := filepath.Join(root, rel)
			Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
			Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())
		}
	}

	It("reports true for identical trees", func() {
		writeTree(filepath.Join(tmp, "a"), map[string]string{"x.txt": "X", "sub/y.txt": "Y"})
		writeTree(filepath.Join(tmp, "b"), map[string]string{"x.txt": "X", "sub/y.txt": "Y"})

		Expect(fsutil.SameTree(filepath.Join(tmp, "a"), filepath.Join(tmp, "b"))).To(BeTrue())
	})

	It("reports false when dst has an extra file", func() {
		writeTree(filepath.Join(tmp, "a"), map[string]string{"x.txt": "X"})
		writeTree(filepath.Join(tmp, "b"), map[string]string{"x.txt": "X", "extra.txt": "E"})

		Expect(fsutil.SameTree(filepath.Join(tmp, "a"), filepath.Join(tmp, "b"))).To(BeFalse())
	})

	It("reports false when a file's content differs", func() {
		writeTree(filepath.Join(tmp, "a"), map[string]string{"x.txt": "X"})
		writeTree(filepath.Join(tmp, "b"), map[string]string{"x.txt": "CHANGED"})

		Expect(fsutil.SameTree(filepath.Join(tmp, "a"), filepath.Join(tmp, "b"))).To(BeFalse())
	})

	It("reports false when an exec bit differs", func() {
		writeTree(filepath.Join(tmp, "a"), map[string]string{"x.sh": "#!/bin/sh"})
		writeTree(filepath.Join(tmp, "b"), map[string]string{"x.sh": "#!/bin/sh"})
		Expect(os.Chmod(filepath.Join(tmp, "a", "x.sh"), 0o755)).To(Succeed())

		Expect(fsutil.SameTree(filepath.Join(tmp, "a"), filepath.Join(tmp, "b"))).To(BeFalse())
	})

	It("reports false (no error) when dst is missing", func() {
		writeTree(filepath.Join(tmp, "a"), map[string]string{"x.txt": "X"})

		same, err := fsutil.SameTree(filepath.Join(tmp, "a"), filepath.Join(tmp, "missing"))
		Expect(err).NotTo(HaveOccurred())
		Expect(same).To(BeFalse())
	})
})

var _ = Describe("SafeCopy", func() {
	var (
		tmp        string
		backupRoot string
	)

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		backupRoot = filepath.Join(tmp, "backups")
	})

	It("copies src to dst, creating parent directories", func() {
		src := filepath.Join(tmp, "src.txt")
		Expect(os.WriteFile(src, []byte("payload"), 0o644)).To(Succeed())
		dst := filepath.Join(tmp, "nested", "deep", "dst.txt")

		Expect(fsutil.SafeCopy(src, dst, "zsh", backupRoot)).To(Succeed())
		b, err := os.ReadFile(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("payload"))
	})

	It("backs up the existing dst before overwriting", func() {
		src := filepath.Join(tmp, "src.txt")
		dst := filepath.Join(tmp, "dst.txt")
		Expect(os.WriteFile(src, []byte("new"), 0o644)).To(Succeed())
		Expect(os.WriteFile(dst, []byte("old"), 0o644)).To(Succeed())

		Expect(fsutil.SafeCopy(src, dst, "zsh", backupRoot)).To(Succeed())

		newContent, err := os.ReadFile(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(newContent)).To(Equal("new"))

		backedUp, err := os.ReadFile(filepath.Join(backupRoot, "zsh", "v1", "dst.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(backedUp)).To(Equal("old"))
	})

	It("returns an error when src does not exist", func() {
		err := fsutil.SafeCopy(
			filepath.Join(tmp, "missing"),
			filepath.Join(tmp, "dst"),
			"zsh", backupRoot,
		)
		Expect(err).To(HaveOccurred())
	})

	It("preserves the source's exec bit on copy", func() {
		src := filepath.Join(tmp, "script.sh")
		Expect(os.WriteFile(src, []byte("#!/bin/sh\n"), 0o755)).To(Succeed())
		dst := filepath.Join(tmp, "out", "script.sh")

		Expect(fsutil.SafeCopy(src, dst, "byobu", backupRoot)).To(Succeed())

		info, err := os.Stat(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o755)))
	})

	It("repairs a mode mismatch on identical content without backing up", func() {
		src := filepath.Join(tmp, "script.sh")
		dst := filepath.Join(tmp, "dst.sh")
		Expect(os.WriteFile(src, []byte("#!/bin/sh\n"), 0o755)).To(Succeed())
		Expect(os.WriteFile(dst, []byte("#!/bin/sh\n"), 0o644)).To(Succeed())

		Expect(fsutil.SafeCopy(src, dst, "byobu", backupRoot)).To(Succeed())

		info, err := os.Stat(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o755)))
		_, err = os.Stat(filepath.Join(backupRoot, "byobu"))
		Expect(os.IsNotExist(err)).To(BeTrue(), "mode repair must not create a backup")
	})

	It("is a no-op (no backup) when dst already equals src (idempotent)", func() {
		src := filepath.Join(tmp, "src.txt")
		dst := filepath.Join(tmp, "dst.txt")
		Expect(os.WriteFile(src, []byte("same"), 0o644)).To(Succeed())
		Expect(os.WriteFile(dst, []byte("same"), 0o644)).To(Succeed())

		Expect(fsutil.SafeCopy(src, dst, "zsh", backupRoot)).To(Succeed())

		_, err := os.Stat(filepath.Join(backupRoot, "zsh"))
		Expect(os.IsNotExist(err)).To(BeTrue(), "no backup should be created for identical content")
	})
})
