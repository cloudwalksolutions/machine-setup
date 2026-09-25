package forms_test

import (
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/pkg"
)

var _ = Describe("Headless", func() {
	var h forms.Headless

	It("answers every prompt with the defaults", func() {
		Expect(h.Welcome()).To(Succeed())
		Expect(h.PickTools([]pkg.ToolInfo{{Name: "neovim", Installed: true}, {Name: "fzf"}})).To(Equal([]string{"fzf"}))
		Expect(h.PickWizards([]string{"claude", "pi"})).To(Equal([]string{"claude", "pi"}))

		claude, err := h.AskClaude([]string{"10-tdd"})
		Expect(err).NotTo(HaveOccurred())
		Expect(*claude.Hook && *claude.Settings).To(BeTrue())
		Expect(claude.Rules).To(Equal([]string{"10-tdd"}))

		pi, err := h.AskPi(forms.PiInputs{OllamaModels: []string{"llama3"}, Packages: []string{"npm:a"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(pi).To(Equal(config.PiConfig{Packages: []string{"npm:a"}, OllamaModels: []string{"llama3"}}))

		project, err := h.AskProject("app")
		Expect(err).NotTo(HaveOccurred())
		Expect(project.Agents).To(Equal([]string{"claude", "pi", "gemini"}))
		Expect(project.Name).To(Equal("app"))
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
		_, collect := forms.ClaudeForm([]string{"10-tdd", "20-prs"})
		cfg := collect()
		Expect(*cfg.Hook && *cfg.Settings).To(BeTrue())
		Expect(cfg.Rules).To(Equal([]string{"10-tdd", "20-prs"}))
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
