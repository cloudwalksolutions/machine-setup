package paths_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/paths"
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

var _ = Describe("ForOS Profiles", func() {
	It("renders under ~/.config/tars/profiles and manages ~/.gitconfig on every OS", func() {
		for _, goos := range []string{"darwin", "linux"} {
			p := paths.ForOS("/repo", "/home/u", goos)
			Expect(p.Profiles.Dir).To(Equal("/home/u/.config/tars/profiles"))
			Expect(p.Profiles.Gitconfig).To(Equal("/home/u/.gitconfig"))
		}
	})
})

var _ = Describe("ForOS Terminal", func() {
	const (
		repoRoot = "/repo"
		home     = "/home/u"
	)

	It("points the repo artifacts at terminal/font and the Terminal.app profile", func() {
		p := paths.ForOS(repoRoot, home, "darwin")
		Expect(p.Terminal.FontRepo).To(Equal(filepath.Join(repoRoot, "terminal", "font")))
		Expect(p.Terminal.ProfileRepo).To(Equal(filepath.Join(repoRoot, "terminal", "tars.terminal")))
	})
})
