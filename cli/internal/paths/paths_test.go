package paths_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudwalk/machine-setup/internal/paths"
)

var _ = Describe("ForOS Zsh.ProfileLocal", func() {
	const (
		repoRoot = "/repo"
		home     = "/home/u"
	)

	It("targets ~/.zprofile on darwin (the file zsh login shells read)", func() {
		p := paths.ForOS(repoRoot, home, "darwin")
		Expect(p.Zsh.ProfileLocal).To(Equal(filepath.Join(home, ".zprofile")))
	})

	It("targets ~/.profile on linux", func() {
		p := paths.ForOS(repoRoot, home, "linux")
		Expect(p.Zsh.ProfileLocal).To(Equal(filepath.Join(home, ".profile")))
	})
})

var _ = Describe("ForOS Terminal", func() {
	const (
		repoRoot = "/repo"
		home     = "/home/u"
	)

	It("points the repo artifacts at terminal/font and terminal/CloudWalk.terminal", func() {
		p := paths.ForOS(repoRoot, home, "darwin")
		Expect(p.Terminal.FontRepo).To(Equal(filepath.Join(repoRoot, "terminal", "font")))
		Expect(p.Terminal.ProfileRepo).To(Equal(filepath.Join(repoRoot, "terminal", "CloudWalk.terminal")))
	})
})
