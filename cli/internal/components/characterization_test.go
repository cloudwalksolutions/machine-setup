package components_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
)

func relPaths(root string) []string {
	var out []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		Expect(err).NotTo(HaveOccurred())
	}
	sort.Strings(out)
	return out
}

func backupVersions(backupRoot, component string) int {
	entries, err := os.ReadDir(filepath.Join(backupRoot, component))
	if os.IsNotExist(err) {
		return 0
	}
	Expect(err).NotTo(HaveOccurred())
	n := 0
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "v") {
			n++
		}
	}
	return n
}

func loginProfile() string {
	if runtime.GOOS == "darwin" {
		return ".zprofile"
	}
	return ".profile"
}

var _ = Describe("pull", func() {
	var (
		tmp        string
		repoRoot   string
		home       string
		backupRoot string
		opts       components.Options
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
		backupRoot = filepath.Join(tmp, "backups")

		write("vim/vimrc", "VIMRC")
		write("vim/colors/sublimemonokai.vim", "COLORS")

		write("zsh/zshrc", "ZSHRC")
		write("zsh/zshrc_aliases", "ALIASES")
		write("zsh/profile", "PROFILE")
		write("zsh/zshrc_funcs", "FUNCS")
		write("zsh/zshrc_secret.template", "TEMPLATE")

		write("byobu/.tmux.conf", "TMUXCONF")
		write("byobu/keybindings.tmux", "KEYS")
		write("byobu/datetime.tmux", "DATETIME")
		write("byobu/statusrc", "STATUSRC")
		write("byobu/color.tmux", "COLOR")
		write("byobu/bin/1_git", "GITSCRIPT")

		write("nvim/init.lua", "INIT")
		write("nvim/lua/core/settings.lua", "SETTINGS")
		write("monokai.lua", "MONOKAI")

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: backupRoot,
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	Describe("Vim", func() {
		It("writes exactly the vimrc and the colorscheme", func() {
			Expect(components.NewVim(opts).Pull()).To(Succeed())
			Expect(relPaths(home)).To(Equal([]string{
				".vim/colors/sublimemonokai.vim",
				".vimrc",
			}))
		})

		It("creates no backup when pulled twice with no change", func() {
			Expect(components.NewVim(opts).Pull()).To(Succeed())
			Expect(components.NewVim(opts).Pull()).To(Succeed())
			Expect(backupVersions(backupRoot, "vim")).To(Equal(0))
		})

		It("archives the previous file before overwriting a changed one", func() {
			Expect(components.NewVim(opts).Pull()).To(Succeed())
			write("vim/vimrc", "VIMRC_V2")

			Expect(components.NewVim(opts).Pull()).To(Succeed())

			Expect(backupVersions(backupRoot, "vim")).To(Equal(1))
			b, err := os.ReadFile(filepath.Join(backupRoot, "vim", "v1", ".vimrc"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal("VIMRC"))
		})
	})

	Describe("Zsh", func() {
		It("writes exactly the five managed zsh files", func() {
			Expect(components.NewZsh(opts).Pull()).To(Succeed())
			Expect(relPaths(home)).To(ConsistOf(
				".zshrc",
				".zshrc_aliases",
				".zshrc_funcs",
				".zshrc_secret",
				loginProfile(),
			))
		})

		It("creates no backup when pulled twice with no change", func() {
			Expect(components.NewZsh(opts).Pull()).To(Succeed())
			Expect(components.NewZsh(opts).Pull()).To(Succeed())
			Expect(backupVersions(backupRoot, "zsh")).To(Equal(0))
		})

		It("creates no ~/.zprofile_local when the repo ships no template", func() {
			Expect(components.NewZsh(opts).Pull()).To(Succeed())
			_, err := os.Stat(filepath.Join(home, ".zprofile_local"))
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Describe("Byobu", func() {
		It("writes exactly the five config files plus bin/", func() {
			Expect(components.NewByobu(opts).Pull()).To(Succeed())
			Expect(relPaths(home)).To(Equal([]string{
				".byobu/.tmux.conf",
				".byobu/bin/1_git",
				".byobu/color.tmux",
				".byobu/datetime.tmux",
				".byobu/keybindings.tmux",
				".byobu/statusrc",
			}))
		})

		It("creates no backup when pulled twice with no change", func() {
			Expect(components.NewByobu(opts).Pull()).To(Succeed())
			Expect(components.NewByobu(opts).Pull()).To(Succeed())
			Expect(backupVersions(backupRoot, "byobu")).To(Equal(0))
		})

	})

	Describe("Nvim", func() {
		It("writes the nvim tree and the packer-path monokai copy", func() {
			Expect(components.NewNvim(opts).Pull()).To(Succeed())
			Expect(relPaths(home)).To(Equal([]string{
				".config/nvim/init.lua",
				".config/nvim/lua/core/settings.lua",
				".local/share/nvim/site/pack/packer/start/monokai.nvim/lua/monokai.lua",
			}))
		})

		It("copies repo-side .undo and .netrwhist scratch state into the live config", func() {
			write("nvim/.netrwhist", "HISTORY")
			write("nvim/.undo/%Users%someone%.zshrc_secret", "UNDOBLOB")

			Expect(components.NewNvim(opts).Pull()).To(Succeed())

			Expect(relPaths(filepath.Join(home, ".config", "nvim"))).To(ContainElements(
				".netrwhist",
				".undo/%Users%someone%.zshrc_secret",
			))
		})
	})

	Describe("Fonts", func() {
		var localDir string

		newFonts := func() *components.Fonts {
			f := components.NewFontsForOS(opts, "linux")
			f.LocalOverride = localDir
			return f
		}

		BeforeEach(func() {
			localDir = filepath.Join(tmp, "installed-fonts")
			Expect(os.MkdirAll(localDir, 0o755)).To(Succeed())
			write("fonts/Hack Regular.ttf", "FONT_A")
			write("fonts/Hack Bold.ttf", "FONT_B")
		})

		It("installs every font file from the repo", func() {
			Expect(newFonts().Pull()).To(Succeed())
			Expect(relPaths(localDir)).To(Equal([]string{
				"Hack Bold.ttf",
				"Hack Regular.ttf",
			}))
		})

		It("archives an existing font before replacing it", func() {
			Expect(os.WriteFile(filepath.Join(localDir, "Hack Regular.ttf"), []byte("OLD_FONT"), 0o644)).To(Succeed())

			Expect(newFonts().Pull()).To(Succeed())

			Expect(backupVersions(backupRoot, "fonts")).To(Equal(1))
			b, err := os.ReadFile(filepath.Join(localDir, "Hack Regular.ttf"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal("FONT_A"))
		})
	})

	Describe("Terminal", func() {
		var fontSet []string

		newTerminal := func(current string) *components.Terminal {
			t := components.NewTerminalForOS(opts, "darwin")
			t.CurrentFontFn = func() (string, error) { return current, nil }
			t.SetFontFn = func(font string) error {
				fontSet = append(fontSet, font)
				return nil
			}
			t.DefaultProfileFn = func() (string, error) { return "Basic", nil }
			t.ApplyFn = func(string, string) error { return nil }
			return t
		}

		BeforeEach(func() {
			fontSet = nil
			write("terminal/font", "Hack Nerd Font 13\n")
		})

		It("sets the font when the current value differs", func() {
			Expect(newTerminal("Monaco 12").Pull()).To(Succeed())
			Expect(fontSet).To(Equal([]string{"Hack Nerd Font 13"}))
		})

		It("leaves the font alone when it already matches", func() {
			Expect(newTerminal("Hack Nerd Font 13").Pull()).To(Succeed())
			Expect(fontSet).To(BeEmpty())
		})
	})
})
