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

var _ = Describe("ForOS Pi", func() {
	const (
		repoRoot = "/repo"
		home     = "/home/u"
	)

	It("maps the repo pi/ dir onto ~/.pi/agent", func() {
		p := paths.ForOS(repoRoot, home, "linux")
		agent := filepath.Join(home, ".pi", "agent")
		Expect(p.Pi.SettingsRepo).To(Equal(filepath.Join(repoRoot, "pi", "settings.json")))
		Expect(p.Pi.SettingsLocal).To(Equal(filepath.Join(agent, "settings.json")))
		Expect(p.Pi.ModelsLocal).To(Equal(filepath.Join(agent, "models.json")))
		Expect(p.Pi.AgentRepo).To(Equal(filepath.Join(repoRoot, "pi", "agents", "tars.md")))
		Expect(p.Pi.AgentLocal).To(Equal(filepath.Join(agent, "agents", "tars.md")))
		Expect(p.Pi.PromptsRepo).To(Equal(filepath.Join(repoRoot, "pi", "prompts")))
		Expect(p.Pi.PromptsLocal).To(Equal(filepath.Join(agent, "prompts")))
		Expect(p.Pi.ExtensionRepo).To(Equal(filepath.Join(repoRoot, "pi", "extensions", "block-unreviewable-edits.ts")))
		Expect(p.Pi.ExtensionLocal).To(Equal(filepath.Join(agent, "extensions", "block-unreviewable-edits.ts")))
		Expect(p.Pi.PermissionsRepo).To(Equal(filepath.Join(repoRoot, "pi", "permissions.json")))
		Expect(p.Pi.PermissionsLocal).To(Equal(filepath.Join(agent, "extensions", "pi-permission-system", "config.json")))
		Expect(p.Pi.AgentsMDLocal).To(Equal(filepath.Join(agent, "AGENTS.md")))
		Expect(p.Pi.KeybindingsRepo).To(Equal(filepath.Join(repoRoot, "pi", "keybindings.json")))
		Expect(p.Pi.KeybindingsLocal).To(Equal(filepath.Join(agent, "keybindings.json")))
	})
})

var _ = Describe("ForOS Claude", func() {
	const (
		repoRoot = "/repo"
		home     = "/home/u"
	)

	It("maps the repo claude/ dir onto ~/.claude", func() {
		p := paths.ForOS(repoRoot, home, "linux")
		Expect(p.Claude.SettingsRepo).To(Equal(filepath.Join(repoRoot, "claude", "settings.json")))
		Expect(p.Claude.SettingsLocal).To(Equal(filepath.Join(home, ".claude", "settings.json")))
		Expect(p.Claude.HookRepo).To(Equal(filepath.Join(repoRoot, "claude", "hooks", "block-unreviewable-edits.sh")))
		Expect(p.Claude.HookLocal).To(Equal(filepath.Join(home, ".claude", "hooks", "block-unreviewable-edits.sh")))
		Expect(p.Claude.RulesRepo).To(Equal(filepath.Join(repoRoot, "claude", "rules")))
		Expect(p.Claude.ClaudeMDLocal).To(Equal(filepath.Join(home, ".claude", "CLAUDE.md")))
		Expect(p.Claude.ProjectTemplateRepo).To(Equal(filepath.Join(repoRoot, "claude", "templates", "AGENTS.project.md")))
	})
})
