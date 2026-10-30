package components_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/components"
)

var _ = Describe("Terminal", func() {
	var (
		tmp      string
		repoRoot string
		home     string
		opts     components.Options
	)

	newTerm := func() *components.Terminal {
		t := components.NewTerminalForOS(opts, "darwin")
		t.CurrentFontFn = func() (string, error) { return "Monaco 12", nil }
		t.SetFontFn = func(string) error { return nil }
		t.DefaultProfileFn = func() (string, error) { return "Basic", nil }
		t.ApplyFn = func(string, string) error { return nil }
		t.ExportFn = func(string, string) error { return nil }
		return t
	}

	BeforeEach(func() {
		tmp = GinkgoT().TempDir()
		repoRoot = filepath.Join(tmp, "repo")
		home = filepath.Join(tmp, "home")
		Expect(os.MkdirAll(filepath.Join(repoRoot, "terminal"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "terminal", "font"), []byte("Hack Nerd Font 13\n"), 0o644)).To(Succeed())

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("reports the name terminal", func() {
		Expect(components.NewTerminal(opts).Name()).To(Equal("terminal"))
	})

	It("Pull sets the iTerm2 font to the repo font and archives the old value", func() {
		var setTo string
		setCalls := 0
		t := newTerm()
		t.CurrentFontFn = func() (string, error) { return "Monaco 12", nil }
		t.SetFontFn = func(f string) error { setTo = f; setCalls++; return nil }

		Expect(t.Pull()).To(Succeed())
		Expect(setTo).To(Equal("Hack Nerd Font 13"))
		Expect(setCalls).To(Equal(1))

		b, err := os.ReadFile(filepath.Join(opts.BackupRoot, "terminal", "v1", "font"))
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.TrimSpace(string(b))).To(Equal("Monaco 12"))
	})

	It("Pull touches nothing when the font already matches", func() {
		t := newTerm()
		t.CurrentFontFn = func() (string, error) { return "Hack Nerd Font 13", nil }
		t.SetFontFn = func(string) error {
			Fail("SetFontFn must not be called when the font already matches")
			return nil
		}

		Expect(t.Pull()).To(Succeed())

		_, err := os.Stat(filepath.Join(opts.BackupRoot, "terminal"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("Pull skips the Terminal.app import when the tars profile is already the default (no window popup)", func() {
		applyCalls := 0
		t := newTerm()
		t.DefaultProfileFn = func() (string, error) { return "tars", nil }
		t.ApplyFn = func(string, string) error { applyCalls++; return nil }

		Expect(t.Pull()).To(Succeed())
		Expect(applyCalls).To(Equal(0))
	})

	It("Push writes the live iTerm2 font into the repo, archiving the prior value", func() {
		t := newTerm()
		t.CurrentFontFn = func() (string, error) { return "Foo 20", nil }

		Expect(t.Push()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(repoRoot, "terminal", "font"))
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.TrimSpace(string(b))).To(Equal("Foo 20"))

		ab, err := os.ReadFile(filepath.Join(opts.BackupRoot, "terminal-repo", "v1", "font"))
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.TrimSpace(string(ab))).To(Equal("Hack Nerd Font 13"))
	})

	It("warns to restart iTerm2 when it is running during a font change", func() {
		stderr := &bytes.Buffer{}
		opts.Stderr = stderr
		t := newTerm()
		t.IsRunningFn = func() bool { return true }

		Expect(t.Pull()).To(Succeed())

		Expect(stderr.String()).To(ContainSubstring("quit and reopen"))
	})

	It("does not warn when iTerm2 is not running", func() {
		stderr := &bytes.Buffer{}
		opts.Stderr = stderr
		t := newTerm()
		t.IsRunningFn = func() bool { return false }

		Expect(t.Pull()).To(Succeed())

		Expect(stderr.String()).To(BeEmpty())
	})

	It("dry-run reports the font change and profile import without calling the seams", func() {
		stdout := &bytes.Buffer{}
		opts.Stdout = stdout
		opts.DryRun = true
		t := newTerm()
		t.SetFontFn = func(string) error { Fail("SetFontFn called in dry-run"); return nil }
		t.ApplyFn = func(string, string) error { Fail("ApplyFn called in dry-run"); return nil }

		Expect(t.Pull()).To(Succeed())

		Expect(stdout.String()).To(ContainSubstring("would set terminal font  Monaco 12 → Hack Nerd Font 13"))
		Expect(stdout.String()).To(ContainSubstring("would import Terminal.app profile  tars"))
		_, err := os.Stat(filepath.Join(opts.BackupRoot, "terminal"))
		Expect(os.IsNotExist(err)).To(BeTrue(), "no backup in dry-run")
	})

	It("is a no-op on non-darwin", func() {
		t := components.NewTerminalForOS(opts, "linux")
		called := false
		t.CurrentFontFn = func() (string, error) { called = true; return "", nil }
		t.SetFontFn = func(string) error { called = true; return nil }

		Expect(t.Pull()).To(Succeed())
		Expect(t.Push()).To(Succeed())
		Expect(called).To(BeFalse())
	})
})
