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

var _ = Describe("ForOS Zsh.ProfileLocalOverride", func() {
	const (
		repoRoot = "/repo"
		home     = "/home/u"
	)

	It("maps the repo template to ~/.zprofile_local on darwin", func() {
		p := paths.ForOS(repoRoot, home, "darwin")
		Expect(p.Zsh.ProfileLocalTemplate).To(Equal(filepath.Join(repoRoot, "zsh", "zprofile_local.template")))
		Expect(p.Zsh.ProfileLocalOverride).To(Equal(filepath.Join(home, ".zprofile_local")))
	})

	It("maps to the same override name on linux", func() {
		p := paths.ForOS(repoRoot, home, "linux")
		Expect(p.Zsh.ProfileLocalTemplate).To(Equal(filepath.Join(repoRoot, "zsh", "zprofile_local.template")))
		Expect(p.Zsh.ProfileLocalOverride).To(Equal(filepath.Join(home, ".zprofile_local")))
	})
})

var _ = Describe("ForOS Terminal", func() {
	const (
		repoRoot = "/repo"
		home     = "/home/u"
	)

	It("points the repo artifact at terminal/font", func() {
		p := paths.ForOS(repoRoot, home, "darwin")
		Expect(p.Terminal.FontRepo).To(Equal(filepath.Join(repoRoot, "terminal", "font")))
	})
})
