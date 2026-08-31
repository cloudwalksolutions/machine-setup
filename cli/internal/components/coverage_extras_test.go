package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

// newOpts builds a minimal repo/home fixture shared by the branch-coverage specs.
func newOpts() (components.Options, string, string) {
	tmp := GinkgoT().TempDir()
	repoRoot := filepath.Join(tmp, "repo")
	home := filepath.Join(tmp, "home")
	Expect(os.MkdirAll(repoRoot, 0o755)).To(Succeed())
	Expect(os.MkdirAll(home, 0o755)).To(Succeed())
	return components.Options{
		RepoRoot:   repoRoot,
		Home:       home,
		BackupRoot: filepath.Join(tmp, "backups"),
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	}, repoRoot, home
}

var _ = Describe("AllPullable", func() {
	It("lists the six components in pull order", func() {
		opts, _, _ := newOpts()
		var names []string
		for _, c := range components.AllPullable(opts) {
			names = append(names, c.Name())
		}
		Expect(names).To(Equal([]string{"vim", "zsh", "byobu", "nvim", "fonts", "terminal"}))
	})
})

var _ = Describe("Fonts on linux (real default copy)", func() {
	It("copies fonts to ~/.local/share/fonts without sudo, idempotently", func() {
		opts, repoRoot, home := newOpts()
		Expect(os.MkdirAll(filepath.Join(repoRoot, "fonts"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "fonts", "Hack.ttf"), []byte("FONT"), 0o644)).To(Succeed())

		f := components.NewFontsForOS(opts, "linux")
		Expect(f.Name()).To(Equal("fonts"))
		Expect(f.Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".local", "share", "fonts", "Hack.ttf"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("FONT"))

		Expect(f.Pull()).To(Succeed(), "second pull skips identical fonts")
	})

	It("errors when the repo fonts dir is missing", func() {
		opts, _, _ := newOpts()
		Expect(components.NewFontsForOS(opts, "linux").Pull()).NotTo(Succeed())
	})
})

var _ = Describe("NewFonts", func() {
	It("builds a component for the current OS", func() {
		opts, _, _ := newOpts()
		Expect(components.NewFonts(opts).Name()).To(Equal("fonts"))
	})
})

var _ = Describe("Byobu branch coverage", func() {
	It("Push copies local bin scripts back into the repo", func() {
		opts, repoRoot, home := newOpts()
		Expect(os.MkdirAll(filepath.Join(home, ".byobu", "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".byobu", "bin", "2_cpu"), []byte("CPU"), 0o755)).To(Succeed())

		Expect(components.NewByobu(opts).Push()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(repoRoot, "byobu", "bin", "2_cpu"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("CPU"))
	})

	It("Pull errors when the repo byobu dir is missing", func() {
		opts, _, _ := newOpts()
		Expect(components.NewByobu(opts).Pull()).NotTo(Succeed())
	})
})

var _ = Describe("Zsh branch coverage", func() {
	seedRepo := func(repoRoot string) {
		Expect(os.MkdirAll(filepath.Join(repoRoot, "zsh"), 0o755)).To(Succeed())
		for _, f := range []string{"zshrc", "zshrc_aliases", "profile"} {
			Expect(os.WriteFile(filepath.Join(repoRoot, "zsh", f), []byte(f), 0o644)).To(Succeed())
		}
	}

	It("Pull succeeds without a secret template (nothing seeded)", func() {
		opts, repoRoot, home := newOpts()
		seedRepo(repoRoot)

		Expect(components.NewZsh(opts).Pull()).To(Succeed())

		_, err := os.Stat(filepath.Join(home, ".zshrc_secret"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("Push includes local funcs when present and skips a missing profile", func() {
		opts, repoRoot, home := newOpts()
		seedRepo(repoRoot)
		Expect(os.WriteFile(filepath.Join(home, ".zshrc"), []byte("rc"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".zshrc_aliases"), []byte("al"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(home, ".zshrc_funcs"), []byte("fn"), 0o644)).To(Succeed())

		Expect(components.NewZsh(opts).Push()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(repoRoot, "zsh", "zshrc_funcs"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("fn"))
	})

	It("Push errors when the local zshrc is missing", func() {
		opts, repoRoot, _ := newOpts()
		seedRepo(repoRoot)
		Expect(components.NewZsh(opts).Push()).NotTo(Succeed())
	})
})

var _ = Describe("Vim branch coverage", func() {
	It("Push errors when local vim files are missing", func() {
		opts, _, _ := newOpts()
		Expect(components.NewVim(opts).Push()).NotTo(Succeed())
	})
})

var _ = Describe("Nvim branch coverage", func() {
	It("Pull propagates an unreadable repo tree", func() {
		opts, repoRoot, home := newOpts()
		locked := filepath.Join(repoRoot, "nvim")
		Expect(os.MkdirAll(locked, 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(home, ".config", "nvim"), 0o755)).To(Succeed())
		Expect(os.Chmod(locked, 0o000)).To(Succeed())
		DeferCleanup(func() { _ = os.Chmod(locked, 0o755) })

		Expect(components.NewNvim(opts).Pull()).NotTo(Succeed())
	})

	It("Push errors when the local nvim config is missing", func() {
		opts, _, _ := newOpts()
		Expect(components.NewNvim(opts).Push()).NotTo(Succeed())
	})
})

var _ = Describe("Terminal branch coverage", func() {
	newTerm := func(opts components.Options) *components.Terminal {
		t := components.NewTerminalForOS(opts, "darwin")
		t.CurrentFontFn = func() (string, error) { return "Hack Nerd Font 13", nil }
		t.SetFontFn = func(string) error { return nil }
		t.DefaultProfileFn = func() (string, error) { return "CloudWalk", nil }
		t.ApplyFn = func(string, string) error { return nil }
		t.ExportFn = func(string, string) error { return nil }
		return t
	}

	It("Pull errors when the repo font file is missing", func() {
		opts, _, _ := newOpts()
		Expect(newTerm(opts).Pull()).NotTo(Succeed())
	})

	It("Push propagates a CurrentFontFn failure", func() {
		opts, repoRoot, _ := newOpts()
		Expect(os.MkdirAll(filepath.Join(repoRoot, "terminal"), 0o755)).To(Succeed())
		t := newTerm(opts)
		t.CurrentFontFn = func() (string, error) { return "", os.ErrPermission }

		Expect(t.Push()).NotTo(Succeed())
	})
})
