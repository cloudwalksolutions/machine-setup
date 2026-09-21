package components_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/components"
)

const goldenPath = "testdata/pull_manifest.golden"

// Regenerate with UPDATE_GOLDEN=1 after reviewing the diff.
func manifestLine(root, path string) string {
	info, err := os.Lstat(path)
	Expect(err).NotTo(HaveOccurred())
	rel, err := filepath.Rel(root, path)
	Expect(err).NotTo(HaveOccurred())

	b, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	sum := sha256.Sum256(b)

	return fmt.Sprintf("%s %04o %s",
		filepath.ToSlash(rel), info.Mode().Perm(), hex.EncodeToString(sum[:])[:12])
}

func manifest(root string) []string {
	var lines []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			lines = append(lines, manifestLine(root, p))
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		Expect(err).NotTo(HaveOccurred())
	}
	sort.Strings(lines)
	return lines
}

var _ = Describe("pull write manifest", func() {
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

	pullEveryFileWritingComponent := func() {
		Expect(components.NewVim(opts).Pull()).To(Succeed())
		Expect(components.NewZsh(opts).Pull()).To(Succeed())
		Expect(components.NewByobu(opts).Pull()).To(Succeed())
		Expect(components.NewNvim(opts).Pull()).To(Succeed())
		Expect(components.NewFontsForOS(opts, "linux").Pull()).To(Succeed())
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
		write("zsh/zprofile_local.template", "LOCAL_TEMPLATE")

		write("byobu/.tmux.conf", "TMUXCONF")
		write("byobu/keybindings.tmux", "KEYS")
		write("byobu/datetime.tmux", "DATETIME")
		write("byobu/statusrc", "STATUSRC")
		write("byobu/color.tmux", "COLOR")
		write("byobu/bin/1_git", "GITSCRIPT")

		write("nvim/init.lua", "INIT")
		write("nvim/lua/core/settings.lua", "SETTINGS")
		write("monokai.lua", "MONOKAI")

		write("fonts/Hack Regular.ttf", "FONT_A")
		write("fonts/Hack Bold.ttf", "FONT_B")

		opts = components.Options{
			RepoRoot:   repoRoot,
			Home:       home,
			BackupRoot: backupRoot,
			Stdout:     &bytes.Buffer{},
			Stderr:     &bytes.Buffer{},
		}
	})

	It("writes exactly the files, modes, and backups recorded in the golden manifest", func() {
		pullEveryFileWritingComponent()
		sections := []string{
			"# fresh pull onto an empty HOME",
			strings.Join(manifest(home), "\n"),
		}

		write("zsh/zshrc", "ZSHRC_V2")
		pullEveryFileWritingComponent()
		sections = append(sections,
			"",
			"# after re-pull with zsh/zshrc changed upstream",
			strings.Join(manifest(home), "\n"),
			"",
			"# backups created",
			strings.Join(manifest(backupRoot), "\n"),
		)
		got := strings.Join(sections, "\n") + "\n"

		if os.Getenv("UPDATE_GOLDEN") != "" {
			Expect(os.MkdirAll(filepath.Dir(goldenPath), 0o755)).To(Succeed())
			Expect(os.WriteFile(goldenPath, []byte(got), 0o644)).To(Succeed())
			Skip("golden manifest regenerated — review the diff before committing")
		}

		want, err := os.ReadFile(goldenPath)
		Expect(err).NotTo(HaveOccurred(), "golden manifest missing; regenerate with UPDATE_GOLDEN=1")
		Expect(got).To(Equal(string(want)))
	})
})
