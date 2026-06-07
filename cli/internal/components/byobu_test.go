package components_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

var _ = Describe("Byobu.Pull", func() {
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
		Expect(os.MkdirAll(filepath.Join(repoRoot, "byobu", "bin"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "byobu", ".tmux.conf"), []byte("TMUX"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "byobu", "keybindings.tmux"), []byte("KEYS"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "byobu", "datetime.tmux"), []byte("TIME"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "byobu", "statusrc"), []byte("STATUS"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repoRoot, "byobu", "bin", "1_git"), []byte("GIT_SCRIPT"), 0o755)).To(Succeed())

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: filepath.Join(tmp, "backups"),
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("copies tmux.conf and the other byobu config files to ~/.byobu/", func() {
		Expect(components.NewByobu(opts).Pull()).To(Succeed())

		mustEqual := func(path, want string) {
			b, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred(), path)
			Expect(string(b)).To(Equal(want), path)
		}
		mustEqual(filepath.Join(home, ".byobu", ".tmux.conf"), "TMUX")
		mustEqual(filepath.Join(home, ".byobu", "keybindings.tmux"), "KEYS")
		mustEqual(filepath.Join(home, ".byobu", "datetime.tmux"), "TIME")
		mustEqual(filepath.Join(home, ".byobu", "statusrc"), "STATUS")
	})

	It("copies repo byobu/bin/* into ~/.byobu/bin/ (flat, not nested)", func() {
		Expect(components.NewByobu(opts).Pull()).To(Succeed())

		b, err := os.ReadFile(filepath.Join(home, ".byobu", "bin", "1_git"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(b)).To(Equal("GIT_SCRIPT"))

		// And NOT nested as ~/.byobu/bin/bin/1_git.
		_, err = os.Stat(filepath.Join(home, ".byobu", "bin", "bin"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
})
