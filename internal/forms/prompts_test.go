package forms_test

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/pkg"
	"tars/internal/projects"
	"tars/internal/sessions"
)

var _ = Describe("Headless", func() {
	var h forms.Headless

	It("answers every prompt with the defaults", func() {
		Expect(h.Welcome()).To(Succeed())
		Expect(h.PickTools([]pkg.ToolInfo{{Name: "neovim", Installed: true}, {Name: "fzf"}})).To(Equal([]string{"fzf"}))
		Expect(h.PickWizards([]string{"claude", "pi"})).To(Equal([]string{"claude", "pi"}))

		agents, err := h.AskAgents([]string{"10-tdd"})
		Expect(err).NotTo(HaveOccurred())
		Expect(agents.SkipRules).To(BeEmpty())

		claude, err := h.AskClaude()
		Expect(err).NotTo(HaveOccurred())
		Expect(*claude.Hook && *claude.Settings).To(BeTrue())

		pi, err := h.AskPi(forms.PiInputs{OllamaModels: []string{"llama3"}, Packages: []string{"npm:a"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(pi).To(Equal(config.PiConfig{Packages: []string{"npm:a"}, OllamaModels: []string{"llama3"}}))

		project, err := h.AskProject("app")
		Expect(err).NotTo(HaveOccurred())
		Expect(project.Agents).To(Equal([]string{"claude", "pi", "gemini"}))
		Expect(project.Name).To(Equal("app"))
	})
})

var _ = Describe("session form builders", func() {
	It("TemplateForm starts from the project's current template", func() {
		t := projects.Template{Windows: []sessions.Window{{Name: "code", Command: "nvim"}}}

		_, collect := forms.TemplateForm("api", t)

		Expect(collect()).To(Equal(t))
	})

	It("RenameForm starts from the session's current name", func() {
		_, collect := forms.RenameForm("scratch")

		Expect(collect()).To(Equal("scratch"))
	})

	It("KillForm keeps the session unless confirmed", func() {
		_, collect := forms.KillForm("scratch")

		Expect(collect()).To(BeFalse())
	})
})

var _ = Describe("form builders", func() {
	It("InstallForm starts with only the missing tools selected", func() {
		_, collect := forms.InstallForm([]pkg.ToolInfo{{Name: "neovim", Installed: true}, {Name: "fzf"}, {Name: "go"}})
		Expect(collect()).To(Equal([]string{"fzf", "go"}))
	})

	It("InstallForm shows every tool of a category without scrolling", func() {
		names := []string{"neovim", "byobu", "gh", "lazygit", "jq", "bat", "eza", "k9s", "fzf", "ripgrep"}
		tools := make([]pkg.ToolInfo, len(names))
		for i, n := range names {
			tools[i] = pkg.ToolInfo{Name: n}
		}
		form, _ := forms.InstallForm(tools)
		form = forms.Styled(form)
		form.Init()

		for _, n := range names {
			Expect(ansi.Strip(form.View())).To(ContainSubstring(n))
		}
	})

	It("InstallForm lists tree-sitter-cli with the editors", func() {
		form, _ := forms.InstallForm([]pkg.ToolInfo{{Name: "neovim"}, {Name: "tree-sitter-cli"}})
		form.Init()

		Expect(ansi.Strip(form.View())).To(And(ContainSubstring("Terminal Utilities & Editors"), ContainSubstring("tree-sitter-cli")))
	})

	It("WizardForm starts with every wizard selected", func() {
		_, collect := forms.WizardForm([]string{"claude", "pi"})
		Expect(collect()).To(Equal([]string{"claude", "pi"}))
	})

	It("ClaudeForm starts with everything on", func() {
		_, collect := forms.ClaudeForm()
		cfg := collect()
		Expect(*cfg.Hook && *cfg.Settings).To(BeTrue())
	})

	It("AgentsForm starts with every rule selected", func() {
		_, collect := forms.AgentsForm([]string{"10-tdd", "20-prs"})
		Expect(collect().SkipRules).To(BeEmpty())
	})

	It("AgentsForm records the unticked rules as skipped", func() {
		form, collect := forms.AgentsForm([]string{"10-tdd", "20-prs"})
		press := func(msg tea.Msg) {
			_, cmd := form.Update(msg)
			for cmd != nil {
				next := cmd()
				if next == nil {
					return
				}
				_, cmd = form.Update(next)
			}
		}
		form.Init()
		press(tea.KeyPressMsg{Code: 'x', Text: "x"})

		Expect(collect().SkipRules).To(Equal([]string{"10-tdd"}))
	})

	It("PiForm starts with every package and the previously exposed models", func() {
		_, collect := forms.PiForm(forms.PiInputs{
			OllamaModels: []string{"llama3", "qwen"},
			Packages:     []string{"npm:a"},
			Previous:     config.PiConfig{OllamaModels: []string{"qwen"}},
		})
		Expect(collect()).To(Equal(config.PiConfig{Packages: []string{"npm:a"}, OllamaModels: []string{"qwen"}}))
	})

	It("PiDefaultForm starts on the previous default when it is still exposed, else the first model", func() {
		_, collect := forms.PiDefaultForm([]string{"llama3", "qwen"}, "qwen")
		Expect(collect()).To(Equal("qwen"))
		_, collect = forms.PiDefaultForm([]string{"llama3", "qwen"}, "gone")
		Expect(collect()).To(Equal("llama3"))
	})

	It("ProjectForm starts with every agent and the placeholders", func() {
		_, collect := forms.ProjectForm("app")
		a := collect()
		Expect(a.Agents).To(Equal([]string{"claude", "pi", "gemini"}))
		Expect(a.Name).To(Equal("app"))
		Expect(a.TestCommand).To(Equal("make test"))
	})

	It("WelcomeForm builds", func() {
		Expect(forms.WelcomeForm()).NotTo(BeNil())
	})
})
