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
})
