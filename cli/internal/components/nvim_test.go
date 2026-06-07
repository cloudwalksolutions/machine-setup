package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

var _ = Describe("Nvim.Pull", func() {
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
		Expect(os.MkdirAll(filepath.Join(repoRoot, "nvim", "lua", "core"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "nvim", "init.lua"), []byte("INIT"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "nvim", "lua", "core", "options.lua"), []byte("OPTS"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "monokai.lua"), []byte("MONOKAI"), 0o644)).To(Succeed())

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("copies the nvim tree into ~/.config/nvim and monokai.lua to its packer path", func() {
		Expect(components.NewNvim(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".config", "nvim", "init.lua"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("INIT"))

		b, err = os.ReadFile(filepath.Join(home, ".config", "nvim", "lua", "core", "options.lua"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("OPTS"))

		b, err = os.ReadFile(filepath.Join(home, ".local", "share", "nvim", "site", "pack", "packer", "start", "monokai.nvim", "lua", "monokai.lua"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("MONOKAI"))
	})

	It("replaces (not merges) the local nvim dir so stale files are removed", func() {
		// Seed a stale plugin/file that's NOT in the repo tree.
		stale := filepath.Join(home, ".config", "nvim", "stale", "leftover.lua")
		Expect(os.MkdirAll(filepath.Dir(stale), 0o755)).To(Succeed())
		Expect(os.WriteFile(stale, []byte("STALE"), 0o644)).To(Succeed())

		Expect(components.NewNvim(opts).Pull()).To(Succeed())

		_, err := os.Stat(stale)
		Expect(os.IsNotExist(err)).To(BeTrue(), "stale leftover.lua should have been removed")

		// And the backed-up copy should still hold the stale file at v1.
		b, err := os.ReadFile(filepath.Join(opts.BackupRoot, "nvim", "v1", "nvim", "stale", "leftover.lua"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("STALE"))
	})
})
