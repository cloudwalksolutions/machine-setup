package fsutil_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/fsutil"
)

var _ = Describe("Copier.SafeWrite", func() {
	var (
		tmp        string
		dst        string
		backupRoot string
	)

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		dst = filepath.Join(tmp, "home", ".gitconfig")
		backupRoot = filepath.Join(tmp, "backups")
	})

	It("writes the content to a new destination with the given mode, creating parents", func() {
		Expect(fsutil.Copier{}.SafeWrite([]byte("[user]\n"), 0o600, dst, "profiles", backupRoot)).To(Succeed())

		b, err := os.ReadFile(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("[user]\n"))
		info, err := os.Stat(dst)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o600)))
	})

	It("backs up a differing destination once and leaves an identical one alone", func() {
		Expect(os.MkdirAll(filepath.Dir(dst), 0o755)).To(Succeed())
		Expect(os.WriteFile(dst, []byte("OLD"), 0o644)).To(Succeed())

		Expect(fsutil.Copier{}.SafeWrite([]byte("NEW"), 0o644, dst, "profiles", backupRoot)).To(Succeed())
		Expect(fsutil.Copier{}.SafeWrite([]byte("NEW"), 0o644, dst, "profiles", backupRoot)).To(Succeed())

		b, err := os.ReadFile(filepath.Join(backupRoot, "profiles", "v1", ".gitconfig"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("OLD"))
		_, err = os.Stat(filepath.Join(backupRoot, "profiles", "v2"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("in dry-run reports the intended write and touches nothing", func() {
		log := &bytes.Buffer{}

		Expect(fsutil.Copier{DryRun: true, Log: log}.SafeWrite([]byte("NEW"), 0o644, dst, "profiles", backupRoot)).To(Succeed())

		_, err := os.Stat(dst)
		Expect(os.IsNotExist(err)).To(BeTrue())
		Expect(log.String()).To(ContainSubstring("would create  " + dst))
	})
})
