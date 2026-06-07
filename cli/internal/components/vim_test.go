package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

var _ = Describe("Vim.Pull", func() {
	var (
		tmp      string
		repoRoot string
		home     string
		opts     components.Options
	)

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		repoRoot = filepath.Join(tmp, "repo")
		home = filepath.Join(tmp, "home")
		Expect(os.MkdirAll(filepath.Join(repoRoot, "vim", "colors"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "vim", "vimrc"), []byte("VIMRC"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "vim", "colors", "sublimemonokai.vim"), []byte("COLORS"), 0o644)).To(Succeed())

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("copies vimrc and the colors file to the right locations", func() {
		Expect(components.NewVim(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".vimrc"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("VIMRC"))

		b, err = os.ReadFile(filepath.Join(home, ".vim", "colors", "sublimemonokai.vim"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("COLORS"))
	})

	It("backs up an existing local vimrc before overwriting", func() {
		Expect(os.MkdirAll(home, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".vimrc"), []byte("OLD"), 0o644)).To(Succeed())

		Expect(components.NewVim(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(opts.BackupRoot, "vim", "v1", ".vimrc"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("OLD"))
	})

})
