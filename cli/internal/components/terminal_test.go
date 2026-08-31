package components_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

var _ = Describe("Terminal", func() {
	var (
		tmp      string
		repoRoot string
		home     string
		opts     components.Options
	)

	// newTerm builds a darwin Terminal with no-op seams; specs override what they assert.
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
		Expect(os.WriteFile(filepath.Join(repoRoot, "terminal", "CloudWalk.terminal"), []byte("PROFILE"), 0o644)).To(Succeed())

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

	It("Pull is a no-op for iTerm2 when the font already matches (idempotent)", func() {
		setCalls := 0
		t := newTerm()
		t.CurrentFontFn = func() (string, error) { return "Hack Nerd Font 13", nil }
		t.SetFontFn = func(string) error { setCalls++; return nil }

		Expect(t.Pull()).To(Succeed())
		Expect(setCalls).To(Equal(0))
		_, err := os.Stat(filepath.Join(opts.BackupRoot, "terminal"))
		Expect(os.IsNotExist(err)).To(BeTrue(), "no backup when font already matches")
	})

	It("Pull configures Terminal.app via the CloudWalk profile", func() {
		var gotPath, gotName string
		t := newTerm()
		t.ApplyFn = func(p, n string) error { gotPath = p; gotName = n; return nil }

		Expect(t.Pull()).To(Succeed())
		Expect(gotPath).To(Equal(filepath.Join(repoRoot, "terminal", "CloudWalk.terminal")))
		Expect(gotName).To(Equal("CloudWalk"))
	})

	It("Pull skips the Terminal.app import when CloudWalk is already the default (no window popup)", func() {
		applyCalls := 0
		t := newTerm()
		t.DefaultProfileFn = func() (string, error) { return "CloudWalk", nil }
		t.ApplyFn = func(string, string) error { applyCalls++; return nil }

		Expect(t.Pull()).To(Succeed())
		Expect(applyCalls).To(Equal(0))
	})

	It("Push writes the live iTerm2 font into the repo, archiving the prior value", func() {
		var exportName, exportDst string
		t := newTerm()
		t.CurrentFontFn = func() (string, error) { return "Foo 20", nil }
		t.ExportFn = func(n, d string) error { exportName = n; exportDst = d; return nil }

		Expect(t.Push()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(repoRoot, "terminal", "font"))
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.TrimSpace(string(b))).To(Equal("Foo 20"))

		ab, err := os.ReadFile(filepath.Join(opts.BackupRoot, "terminal-repo", "v1", "font"))
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.TrimSpace(string(ab))).To(Equal("Hack Nerd Font 13"))

		Expect(exportName).To(Equal("CloudWalk"))
		Expect(exportDst).To(Equal(filepath.Join(repoRoot, "terminal", "CloudWalk.terminal")))
	})

	It("is a no-op on non-darwin", func() {
		t := components.NewTerminalForOS(opts, "linux")
		called := false
		t.CurrentFontFn = func() (string, error) { called = true; return "", nil }
		t.SetFontFn = func(string) error { called = true; return nil }
		t.ApplyFn = func(string, string) error { called = true; return nil }
		t.ExportFn = func(string, string) error { called = true; return nil }

		Expect(t.Pull()).To(Succeed())
		Expect(t.Push()).To(Succeed())
		Expect(called).To(BeFalse())
	})
})
