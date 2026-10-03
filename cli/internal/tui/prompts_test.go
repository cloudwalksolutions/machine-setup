package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/forms"
	"tars/internal/pkg"
)

var _ = Describe("Prompts", func() {
	// answering stands in for the program: it resolves every prompt with the given outcome.
	answering := func(outcome error) (Prompts, *[]*huh.Form) {
		var shown []*huh.Form
		send := func(msg tea.Msg) {
			p := msg.(PromptMsg)
			shown = append(shown, p.Form)
			p.Done <- outcome
		}
		return Prompts{Send: send, Ctx: context.Background()}, &shown
	}

	It("shows each form through the program and returns the collected answers", func() {
		p, shown := answering(nil)

		tools, err := p.PickTools([]pkg.ToolInfo{{Name: "neovim", Installed: true}, {Name: "fzf"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(tools).To(Equal([]string{"fzf"}))

		claude, err := p.AskClaude([]string{"10-tdd"})
		Expect(err).NotTo(HaveOccurred())
		Expect(claude.Rules).To(Equal([]string{"10-tdd"}))

		Expect(p.Welcome()).To(Succeed())
		Expect(*shown).To(HaveLen(3))
	})

	It("asks pi's default model only when models are exposed", func() {
		p, shown := answering(nil)

		cfg, err := p.AskPi(forms.PiInputs{OllamaModels: []string{"llama3"}, Packages: []string{"npm:a"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.DefaultModel).To(Equal("llama3"))
		Expect(*shown).To(HaveLen(2))

		_, err = p.AskPi(forms.PiInputs{Packages: []string{"npm:a"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(*shown).To(HaveLen(3))
	})

	It("returns the abort error the orchestrators already recognise", func() {
		p, _ := answering(huh.ErrUserAborted)

		_, err := p.PickWizards([]string{"claude"})
		Expect(err).To(MatchError("user aborted"))
	})

	It("gives up when the program is gone", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		p := Prompts{Send: func(tea.Msg) {}, Ctx: ctx}

		_, err := p.AskProject("app")
		Expect(err).To(MatchError(huh.ErrUserAborted))
	})
})
