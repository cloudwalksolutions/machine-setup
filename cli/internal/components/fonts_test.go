package components_test

import (
	"bytes"
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
})
