package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

var _ = Describe("Component Push (local → repo, archiving the repo copy)", func() {
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
		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	write := func(path, content string) {
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())
	}
	read := func(path string) string {
		b, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred(), path)
		return string(b)
	}

	It("vim pushes local vimrc to the repo and archives the old repo copy", func() {
		write(filepath.Join(repoRoot, "vim", "vimrc"), "OLD")
		write(filepath.Join(repoRoot, "vim", "colors", "sublimemonokai.vim"), "OLDC")
		write(filepath.Join(home, ".vimrc"), "NEW")
		write(filepath.Join(home, ".vim", "colors", "sublimemonokai.vim"), "NEWC")

		Expect(components.NewVim(opts).Push()).To(Succeed())

		Expect(read(filepath.Join(repoRoot, "vim", "vimrc"))).To(Equal("NEW"))
		Expect(read(filepath.Join(opts.BackupRoot, "vim-repo", "v1", "vimrc"))).To(Equal("OLD"))
	})

	It("zsh pushes local zshrc to the repo and archives the old repo copy", func() {
		write(filepath.Join(repoRoot, "zsh", "zshrc"), "OLD")
		write(filepath.Join(repoRoot, "zsh", "zshrc_aliases"), "OLDA")
		write(filepath.Join(home, ".zshrc"), "NEW")
		write(filepath.Join(home, ".zshrc_aliases"), "NEWA")

		Expect(components.NewZsh(opts).Push()).To(Succeed())

		Expect(read(filepath.Join(repoRoot, "zsh", "zshrc"))).To(Equal("NEW"))
		Expect(read(filepath.Join(opts.BackupRoot, "zsh-repo", "v1", "zshrc"))).To(Equal("OLD"))
	})

	It("byobu pushes local keybindings to the repo and archives the old repo copy", func() {
		write(filepath.Join(repoRoot, "byobu", "keybindings.tmux"), "OLD")
		write(filepath.Join(home, ".byobu", "keybindings.tmux"), "NEW")

		Expect(components.NewByobu(opts).Push()).To(Succeed())

		Expect(read(filepath.Join(repoRoot, "byobu", "keybindings.tmux"))).To(Equal("NEW"))
		Expect(read(filepath.Join(opts.BackupRoot, "byobu-repo", "v1", "keybindings.tmux"))).To(Equal("OLD"))
	})

	It("nvim pushes the local config tree to the repo and archives the old repo copy", func() {
		write(filepath.Join(repoRoot, "nvim", "init.lua"), "OLD")
		write(filepath.Join(home, ".config", "nvim", "init.lua"), "NEW")

		Expect(components.NewNvim(opts).Push()).To(Succeed())

		Expect(read(filepath.Join(repoRoot, "nvim", "init.lua"))).To(Equal("NEW"))
		// old repo tree archived under nvim-repo/v1/nvim/
		Expect(read(filepath.Join(opts.BackupRoot, "nvim-repo", "v1", "nvim", "init.lua"))).To(Equal("OLD"))
	})

	It("AllPushable lists vim, zsh, byobu, nvim, terminal (not fonts)", func() {
		var names []string
		for _, p := range components.AllPushable(opts) {
			names = append(names, p.Name())
		}
		Expect(names).To(Equal([]string{"vim", "zsh", "byobu", "nvim", "terminal"}))
	})
})
