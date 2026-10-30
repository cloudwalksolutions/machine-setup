package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
)

var _ = Describe("Zsh.Pull seeding ~/.zprofile_local", func() {
	var (
		tmp      string
		repoRoot string
		home     string
		opts     components.Options
	)

	write := func(rel, content string) {
		p := filepath.Join(repoRoot, rel)
		Expect(os.MkdirAll(filepath.Dir(p), 0o755)).To(Succeed())
		Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())
	}

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		repoRoot = filepath.Join(tmp, "repo")
		home = filepath.Join(tmp, "home")

		write("zsh/zshrc", "ZSHRC")
		write("zsh/zshrc_aliases", "ALIASES")
		write("zsh/profile", "PROFILE")

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("creates ~/.zprofile_local from the template when it is missing", func() {
		write("zsh/zprofile_local.template", "# per-machine overrides")

		Expect(components.NewZsh(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".zprofile_local"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("# per-machine overrides"))
	})

	It("does NOT overwrite an existing ~/.zprofile_local", func() {
		write("zsh/zprofile_local.template", "# per-machine overrides")
		Expect(os.MkdirAll(home, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".zprofile_local"),
			[]byte("export PATH=\"$PATH:/my/machine/bin\""), 0o644)).To(Succeed())

		Expect(components.NewZsh(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".zprofile_local"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("export PATH=\"$PATH:/my/machine/bin\""))
	})

	It("is a no-op when the repo has no template", func() {
		Expect(components.NewZsh(opts).Pull()).To(Succeed())

		_, err := os.Stat(filepath.Join(home, ".zprofile_local"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
})
