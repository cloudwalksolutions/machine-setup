package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
)

var _ = Describe("Fonts.Pull", func() {
	var (
		tmp      string
		localDir string
		opts     components.Options
		stdout   *bytes.Buffer
	)

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		repoRoot := filepath.Join(tmp, "repo")
		Expect(os.MkdirAll(filepath.Join(repoRoot, "fonts"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "fonts", "Hack Regular.ttf"), []byte("FONT_A"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "fonts", "Hack Bold.ttf"), []byte("FONT_B"), 0o644)).To(Succeed())
		localDir = filepath.Join(tmp, "installed-fonts")

		stdout = &bytes.Buffer{}
		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       filepath.Join(tmp, "home"),
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     stdout,
			Stderr:     &bytes.Buffer{},
		}
	})

	fonts := func(goos string) *components.Fonts {
		f := components.NewFontsForOS(opts, goos)
		f.LocalOverride = localDir
		return f
	}

	It("copies each font from repo/fonts into the local font dir, creating it", func() {
		Expect(fonts("darwin").Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(localDir, "Hack Regular.ttf"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("FONT_A"))
		b, err = os.ReadFile(filepath.Join(localDir, "Hack Bold.ttf"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("FONT_B"))
	})

	It("backs up an existing font before replacing it", func() {
		Expect(os.MkdirAll(localDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(localDir, "Hack Regular.ttf"), []byte("OLD_FONT"), 0o644)).To(Succeed())

		Expect(fonts("linux").Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(opts.BackupRoot, "fonts", "v1", "Hack Regular.ttf"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("OLD_FONT"))
	})

	It("is idempotent: a second pull neither copies nor backs up", func() {
		Expect(fonts("linux").Pull()).To(Succeed())
		Expect(fonts("linux").Pull()).To(Succeed())

		_, err := os.Stat(filepath.Join(opts.BackupRoot, "fonts"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("reports instead of writing in dry-run mode", func() {
		opts.DryRun = true

		Expect(fonts("linux").Pull()).To(Succeed())

		Expect(stdout.String()).To(ContainSubstring("would create"))
		Expect(stdout.String()).To(ContainSubstring("Hack Regular.ttf"))
		_, err := os.Stat(filepath.Join(localDir, "Hack Regular.ttf"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
})
