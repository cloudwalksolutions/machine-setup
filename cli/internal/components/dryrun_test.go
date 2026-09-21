package components_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

var _ = Describe("dry-run pull", func() {
	var (
		tmp        string
		repoRoot   string
		home       string
		backupRoot string
		log        *bytes.Buffer
	)

	write := func(rel, content string) {
		p := filepath.Join(repoRoot, rel)
		Expect(os.MkdirAll(filepath.Dir(p), 0o755)).To(Succeed())
		Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())
	}

	newOpts := func(dryRun bool) components.Options {
		return components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: backupRoot,
			DryRun:     dryRun,
			Stdout:     log,
			Stderr:     &bytes.Buffer{},
		}
	}

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		repoRoot = filepath.Join(tmp, "repo")
		home = filepath.Join(tmp, "home")
		backupRoot = filepath.Join(tmp, "backups")
		log = &bytes.Buffer{}

		write("vim/vimrc", "VIMRC")
		write("vim/colors/sublimemonokai.vim", "COLORS")
		write("zsh/zshrc", "ZSHRC")
		write("zsh/zshrc_aliases", "ALIASES")
		write("zsh/profile", "PROFILE")
		write("zsh/zshrc_secret.template", "TEMPLATE")
		write("nvim/init.lua", "INIT")
		write("monokai.lua", "MONOKAI")
	})

	It("writes nothing to HOME", func() {
		Expect(components.NewZsh(newOpts(true)).Pull()).To(Succeed())
		Expect(components.NewVim(newOpts(true)).Pull()).To(Succeed())

		_, err := os.Stat(home)
		Expect(os.IsNotExist(err)).To(BeTrue(), "dry-run must not create HOME")
	})

	It("creates no backups", func() {
		Expect(components.NewVim(newOpts(false)).Pull()).To(Succeed())
		write("vim/vimrc", "VIMRC_V2")

		Expect(components.NewVim(newOpts(true)).Pull()).To(Succeed())

		Expect(backupVersions(backupRoot, "vim")).To(Equal(0))
		b, err := os.ReadFile(filepath.Join(home, ".vimrc"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("VIMRC"), "the original must survive a dry run")
	})

	It("does not wipe an existing nvim config", func() {
		local := filepath.Join(home, ".config", "nvim")
		Expect(os.MkdirAll(local, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(local, "mine.lua"), []byte("MINE"), 0o644)).To(Succeed())

		Expect(components.NewNvim(newOpts(true)).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(local, "mine.lua"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("MINE"))
	})

	It("does not install fonts", func() {
		write("fonts/Hack Regular.ttf", "FONT_A")
		localDir := filepath.Join(tmp, "installed-fonts")
		Expect(os.MkdirAll(localDir, 0o755)).To(Succeed())

		calls := 0
		f := components.NewFontsForOS(newOpts(true), "linux")
		f.LocalOverride = localDir
		f.CopyFn = func(src, dst string) error { calls++; return nil }

		Expect(f.Pull()).To(Succeed())
		Expect(calls).To(Equal(0))
	})

	It("does not touch terminal preferences", func() {
		write("terminal/font", "Hack Nerd Font 13\n")

		setCalls := 0
		t := components.NewTerminalForOS(newOpts(true), "darwin")
		t.CurrentFontFn = func() (string, error) { return "Monaco 12", nil }
		t.SetFontFn = func(string) error { setCalls++; return nil }

		Expect(t.Pull()).To(Succeed())

		Expect(setCalls).To(Equal(0))
	})

	It("predicts exactly the paths a real pull writes", func() {
		Expect(components.NewVim(newOpts(true)).Pull()).To(Succeed())
		Expect(components.NewZsh(newOpts(true)).Pull()).To(Succeed())

		var predicted []string
		for _, line := range strings.Split(log.String(), "\n") {
			if _, path, found := strings.Cut(line, "would create  "); found {
				rel, err := filepath.Rel(home, path)
				Expect(err).NotTo(HaveOccurred())
				predicted = append(predicted, filepath.ToSlash(rel))
			}
		}
		sort.Strings(predicted)

		Expect(components.NewVim(newOpts(false)).Pull()).To(Succeed())
		Expect(components.NewZsh(newOpts(false)).Pull()).To(Succeed())

		Expect(predicted).To(Equal(relPaths(home)))
	})
})
