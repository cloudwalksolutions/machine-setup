package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

var _ = Describe("Zsh.Pull", func() {
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
		Expect(os.MkdirAll(filepath.Join(repoRoot, "zsh"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "zsh", "zshrc"), []byte("ZSHRC"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "zsh", "zshrc_aliases"), []byte("ALIASES"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "zsh", "profile"), []byte("PROFILE"), 0o644)).To(Succeed())

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("copies zshrc, aliases, and profile to HOME", func() {
		Expect(components.NewZsh(opts).Pull()).To(Succeed())

		mustEqual := func(path, want string) {
			b, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred(), path)
			Expect(string(b)).To(Equal(want), path)
		}
		mustEqual(filepath.Join(home, ".zshrc"), "ZSHRC")
		mustEqual(filepath.Join(home, ".zshrc_aliases"), "ALIASES")
		mustEqual(filepath.Join(home, ".profile"), "PROFILE")
	})

	It("copies zshrc_funcs only when present in repo", func() {
		Expect(os.WriteFile(filepath.Join(repoRoot, "zsh", "zshrc_funcs"), []byte("FUNCS"), 0o644)).To(Succeed())

		Expect(components.NewZsh(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".zshrc_funcs"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("FUNCS"))
	})

	It("does not create .zshrc_funcs when repo lacks the source", func() {
		Expect(components.NewZsh(opts).Pull()).To(Succeed())
		_, err := os.Stat(filepath.Join(home, ".zshrc_funcs"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("seeds ~/.zshrc_secret from the template when local secret is missing", func() {
		Expect(os.WriteFile(filepath.Join(repoRoot, "zsh", "zshrc_secret.template"), []byte("# template"), 0o644)).To(Succeed())

		Expect(components.NewZsh(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".zshrc_secret"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("# template"))
	})

	It("does NOT overwrite an existing ~/.zshrc_secret (contains real secrets)", func() {
		Expect(os.WriteFile(filepath.Join(repoRoot, "zsh", "zshrc_secret.template"), []byte("# template"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(home, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".zshrc_secret"), []byte("MY_KEY=abc"), 0o600)).To(Succeed())

		Expect(components.NewZsh(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".zshrc_secret"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("MY_KEY=abc"))
	})
})
