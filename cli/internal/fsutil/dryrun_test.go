package fsutil_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/fsutil"
)

var _ = Describe("Copier in dry-run mode", func() {
	var (
		tmp        string
		src        string
		dst        string
		backupRoot string
		log        *bytes.Buffer
		copier     fsutil.Copier
	)

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		src = filepath.Join(tmp, "src.txt")
		dst = filepath.Join(tmp, "out", "dst.txt")
		backupRoot = filepath.Join(tmp, "backups")
		log = &bytes.Buffer{}
		copier = fsutil.Copier{DryRun: true, Log: log}
		Expect(os.WriteFile(src, []byte("NEW"), 0o644)).To(Succeed())
	})

	It("reports a create without writing the destination", func() {
		Expect(copier.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

		_, err := os.Stat(dst)
		Expect(os.IsNotExist(err)).To(BeTrue(), "dry-run must not create the file")
		Expect(log.String()).To(ContainSubstring("would create"))
		Expect(log.String()).To(ContainSubstring(dst))
	})

	It("reports an overwrite and the backup version without writing anything", func() {
		Expect(os.MkdirAll(filepath.Dir(dst), 0o755)).To(Succeed())
		Expect(os.WriteFile(dst, []byte("OLD"), 0o644)).To(Succeed())

		Expect(copier.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

		b, err := os.ReadFile(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("OLD"), "dry-run must not overwrite")
		_, err = os.Stat(backupRoot)
		Expect(os.IsNotExist(err)).To(BeTrue(), "dry-run must not create backups")
		Expect(log.String()).To(ContainSubstring("would overwrite"))
		Expect(log.String()).To(ContainSubstring("v1"))
	})

	It("reports identical content as unchanged", func() {
		Expect(os.MkdirAll(filepath.Dir(dst), 0o755)).To(Succeed())
		Expect(os.WriteFile(dst, []byte("NEW"), 0o644)).To(Succeed())

		Expect(copier.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

		Expect(log.String()).To(ContainSubstring("unchanged"))
		Expect(log.String()).NotTo(ContainSubstring("would overwrite"))
	})

	It("reports a removal without removing the tree", func() {
		dir := filepath.Join(tmp, "tree")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "f.txt"), []byte("X"), 0o644)).To(Succeed())

		Expect(copier.RemoveAll(dir)).To(Succeed())

		_, err := os.Stat(dir)
		Expect(err).NotTo(HaveOccurred(), "dry-run must not remove the tree")
		Expect(log.String()).To(ContainSubstring("would remove"))
	})

	It("reports a backup without creating one", func() {
		v, err := copier.Backup(src, "demo", backupRoot)

		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(ContainSubstring("v1"))
		_, statErr := os.Stat(backupRoot)
		Expect(os.IsNotExist(statErr)).To(BeTrue(), "dry-run must not create backups")
	})

	It("still performs real writes when DryRun is false", func() {
		real := fsutil.Copier{DryRun: false, Log: log}

		Expect(real.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

		b, err := os.ReadFile(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("NEW"))
	})
})
