package fsutil_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/fsutil"
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

	Context("with Diff on", func() {
		BeforeEach(func() {
			copier.Diff = true
			Expect(os.MkdirAll(filepath.Dir(dst), 0o755)).To(Succeed())
		})

		It("shows a unified diff under an overwrite", func() {
			Expect(os.WriteFile(src, []byte("keep\nnew line\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(dst, []byte("keep\nold line\n"), 0o644)).To(Succeed())

			Expect(copier.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

			Expect(log.String()).To(ContainSubstring("would overwrite"))
			Expect(log.String()).To(ContainSubstring("-old line\n"))
			Expect(log.String()).To(ContainSubstring("+new line\n"))
		})

		It("shows a create as one line, never the whole new file", func() {
			Expect(os.WriteFile(src, []byte("first\n"), 0o644)).To(Succeed())

			Expect(copier.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

			Expect(log.String()).To(Equal("    would create  " + dst + "\n"))
		})

		It("reports a binary change without dumping its bytes", func() {
			Expect(os.WriteFile(src, []byte("font\x00new"), 0o644)).To(Succeed())
			Expect(os.WriteFile(dst, []byte("font\x00old"), 0o644)).To(Succeed())

			Expect(copier.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

			Expect(log.String()).To(ContainSubstring("would overwrite"))
			Expect(log.String()).NotTo(ContainSubstring("font"))
		})

		It("stays silent on unchanged files so an empty log means nothing changes", func() {
			Expect(os.WriteFile(dst, []byte("NEW"), 0o644)).To(Succeed())

			Expect(copier.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

			Expect(log.String()).To(BeEmpty())
		})

		It("diffs a tree file by file, showing only what changes", func() {
			repo, local := filepath.Join(tmp, "repo"), filepath.Join(tmp, "local")
			for _, d := range []string{repo, local} {
				Expect(os.MkdirAll(filepath.Join(d, "lua"), 0o755)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(d, "init.lua"), []byte("same\n"), 0o644)).To(Succeed())
			}
			Expect(os.WriteFile(filepath.Join(repo, "lua", "keys.lua"), []byte("new map\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(local, "lua", "keys.lua"), []byte("old map\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(local, "stale.lua"), []byte("gone\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(repo, "fresh.lua"), []byte("brand new\n"), 0o644)).To(Succeed())

			Expect(copier.SafeCopy(repo, local, "demo", backupRoot)).To(Succeed())

			Expect(log.String()).To(ContainSubstring("-old map\n+new map\n"))
			Expect(log.String()).To(ContainSubstring("would remove  " + filepath.Join(local, "stale.lua") + "\n"))
			Expect(log.String()).To(ContainSubstring("would create  " + filepath.Join(local, "fresh.lua") + "\n"))
			Expect(log.String()).NotTo(ContainSubstring("gone"))
			Expect(log.String()).NotTo(ContainSubstring("brand new"))
			Expect(log.String()).NotTo(ContainSubstring("init.lua"))
		})
	})

	It("still performs real writes when DryRun is false", func() {
		real := fsutil.Copier{DryRun: false, Log: log}

		Expect(real.SafeCopy(src, dst, "demo", backupRoot)).To(Succeed())

		b, err := os.ReadFile(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("NEW"))
	})
})
