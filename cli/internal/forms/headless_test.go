package forms_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/config"
	"tars/internal/forms"
	"tars/internal/profiles"
)

var _ = Describe("headless defaults with TARS_NO_FORM", func() {
	BeforeEach(func() { GinkgoT().Setenv("TARS_NO_FORM", "1") })

	It("welcome is a no-op", func() {
		Expect(forms.ShowWelcome()).To(Succeed())
	})

	It("claude turns everything on and selects every rule", func() {
		cfg, err := forms.ShowClaudeInitForm([]string{"10-tdd", "20-prs"})

		Expect(err).NotTo(HaveOccurred())
		Expect(*cfg.Hook).To(BeTrue())
		Expect(*cfg.Settings).To(BeTrue())
		Expect(cfg.Rules).To(Equal([]string{"10-tdd", "20-prs"}))
	})

	It("pi installs every package, exposes every ollama model and pins no default", func() {
		cfg, err := forms.ShowPiInitForm([]string{"llama3", "qwen"}, []string{"npm:a", "npm:b"}, config.PiConfig{DefaultModel: "qwen"})

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).To(Equal(config.PiConfig{Packages: []string{"npm:a", "npm:b"}, OllamaModels: []string{"llama3", "qwen"}}))
	})

	It("project targets every agent with placeholder answers", func() {
		a, err := forms.ShowProjectForm("myproj")

		Expect(err).NotTo(HaveOccurred())
		Expect(a.Agents).To(Equal([]string{"claude", "pi", "gemini"}))
		Expect(a.Name).To(Equal("myproj"))
		Expect(a.Description).NotTo(BeEmpty())
		Expect(a.TestCommand).To(Equal("make test"))
	})

	It("wizard picker runs every offered wizard", func() {
		Expect(forms.ShowWizardPicker([]string{"claude", "pi"})).To(Equal([]string{"claude", "pi"}))
	})

	It("session picker takes the first option and rejects an empty list", func() {
		Expect(forms.ShowSessionPicker([]string{"work", "(all)"})).To(Equal("work"))
		_, err := forms.ShowSessionPicker(nil)
		Expect(err).To(HaveOccurred())
	})

	It("profile form returns the defaults unchanged", func() {
		p := profiles.Profile{Name: "acme", Alias: "ac", Email: "me@acme.io"}
		Expect(forms.ShowProfileForm(p)).To(Equal(p))
	})
})
