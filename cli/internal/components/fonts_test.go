package components_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

var _ = Describe("Fonts.Pull", func() {
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
		Expect(os.MkdirAll(filepath.Join(repoRoot, "fonts"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "fonts", "Hack Regular.ttf"), []byte("FONT_A"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "fonts", "Hack Bold.ttf"), []byte("FONT_B"), 0o644)).To(Succeed())

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("copies each font from repo/fonts to the configured local dir via CopyFn", func() {
		localDir := filepath.Join(tmp, "installed-fonts")
		Expect(os.MkdirAll(localDir, 0o755)).To(Succeed())

		f := components.NewFontsForOS(opts, "linux")
		// Pin the dst to a test-controlled dir; on linux that's ~/.local/share/fonts.
		// We override CopyFn to a non-sudo, no-fc-cache implementation.
		f.LocalOverride = localDir
		f.CopyFn = func(src, dst string) error {
			b, err := os.ReadFile(src)
			if err != nil {
				return err
			}
			out, err := os.Create(dst)
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, bytes.NewReader(b))
			return err
		}

		Expect(f.Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(localDir, "Hack Regular.ttf"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("FONT_A"))

		b, err = os.ReadFile(filepath.Join(localDir, "Hack Bold.ttf"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("FONT_B"))
	})

	It("backs up an existing font before replacing it", func() {
		localDir := filepath.Join(tmp, "installed-fonts")
		Expect(os.MkdirAll(localDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(localDir, "Hack Regular.ttf"), []byte("OLD_FONT"), 0o644)).To(Succeed())

		f := components.NewFontsForOS(opts, "linux")
		f.LocalOverride = localDir
		f.CopyFn = func(src, dst string) error { return nil }

		Expect(f.Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(opts.BackupRoot, "fonts", "v1", "Hack Regular.ttf"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("OLD_FONT"))
	})

	It("creates no backup when the font is not already installed", func() {
		localDir := filepath.Join(tmp, "installed-fonts")
		Expect(os.MkdirAll(localDir, 0o755)).To(Succeed())

		f := components.NewFontsForOS(opts, "linux")
		f.LocalOverride = localDir
		f.CopyFn = func(src, dst string) error { return nil }

		Expect(f.Pull()).To(Succeed())

		_, err := os.Stat(filepath.Join(opts.BackupRoot, "fonts"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("leaves the backup intact when the copy fails", func() {
		localDir := filepath.Join(tmp, "installed-fonts")
		Expect(os.MkdirAll(localDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(localDir, "Hack Regular.ttf"), []byte("OLD_FONT"), 0o644)).To(Succeed())

		f := components.NewFontsForOS(opts, "linux")
		f.LocalOverride = localDir
		f.CopyFn = func(src, dst string) error {
			if filepath.Base(dst) == "Hack Regular.ttf" {
				return errors.New("sudo denied")
			}
			return nil
		}

		Expect(f.Pull()).To(MatchError(ContainSubstring("sudo denied")))

		b, err := os.ReadFile(filepath.Join(opts.BackupRoot, "fonts", "v1", "Hack Regular.ttf"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("OLD_FONT"))
	})

	It("skips CopyFn for fonts already present and identical (idempotent, no sudo)", func() {
		localDir := filepath.Join(tmp, "installed-fonts")
		Expect(os.MkdirAll(localDir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(localDir, "Hack Regular.ttf"), []byte("FONT_A"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(localDir, "Hack Bold.ttf"), []byte("FONT_B"), 0o644)).To(Succeed())

		calls := 0
		f := components.NewFontsForOS(opts, "linux")
		f.LocalOverride = localDir
		f.CopyFn = func(src, dst string) error { calls++; return nil }

		Expect(f.Pull()).To(Succeed())
		Expect(calls).To(Equal(0))
	})
})
