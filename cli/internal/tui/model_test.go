package tui_test

import (
	"errors"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/forms"
	"tars/internal/tui"
)

type harness struct{ model tui.Model }

func newHarness() *harness { return &harness{model: tui.New()} }

// send feeds messages to the model, running returned commands and feeding their
// messages back, the way a program would.
func (h *harness) send(msgs ...tea.Msg) {
	for _, msg := range msgs {
		m, cmd := h.model.Update(msg)
		h.model = m.(tui.Model)
		h.run(cmd)
	}
}

func (h *harness) run(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	switch msg := cmd().(type) {
	case nil:
	case tea.BatchMsg:
		for _, c := range msg {
			h.run(c)
		}
	default:
		h.send(msg)
	}
}

func (h *harness) view() string { return ansi.Strip(h.model.View().Content) }

var _ = Describe("Model", func() {
	var h *harness

	BeforeEach(func() { h = newHarness() })

	It("shows the step, its progress and the running item", func() {
		h.send(tui.StepStartedMsg{Title: "Installing packages", Total: 3}, tui.ItemStartedMsg{Name: "jq"})
		Expect(h.view()).To(ContainSubstring("Installing packages"))
		Expect(h.view()).To(ContainSubstring("0/3"))
		Expect(h.view()).To(ContainSubstring("jq"))

		h.send(tui.ItemDoneMsg{Name: "jq"})
		Expect(h.view()).To(ContainSubstring("1/3"))
		Expect(h.view()).NotTo(ContainSubstring("jq"))
	})

	It("shows the last five output lines under the running item and drops them when it finishes", func() {
		h.send(tui.StepStartedMsg{Title: "Installing packages", Total: 1}, tui.ItemStartedMsg{Name: "neovim"})
		for _, line := range []string{"one", "two", "three", "four", "five", "six"} {
			h.send(tui.OutputLineMsg{Line: line})
		}

		Expect(h.view()).NotTo(ContainSubstring("one"))
		Expect(h.view()).To(ContainSubstring("two"))
		Expect(h.view()).To(ContainSubstring("six"))

		h.send(tui.ItemDoneMsg{Name: "neovim"})
		Expect(h.view()).NotTo(ContainSubstring("six"))
	})

	It("counts failed items and steps without quitting", func() {
		h.send(tui.StepStartedMsg{Title: "Installing packages", Total: 2}, tui.ItemStartedMsg{Name: "go"}, tui.ItemDoneMsg{Name: "go", Err: errors.New("mirror down")})
		h.send(tui.StepStartedMsg{Title: "Installing oh-my-zsh"}, tui.StepDoneMsg{Title: "Installing oh-my-zsh", Err: errors.New("git missing")})

		Expect(h.view()).To(ContainSubstring("2 failed"))
	})

	Describe("prompts", func() {
		var done chan error

		BeforeEach(func() {
			done = make(chan error, 1)
			h.send(tui.PromptMsg{Form: forms.WelcomeForm(), Done: done})
		})

		It("shows the form and reports completion", func() {
			Expect(h.view()).To(ContainSubstring("Welcome to tars"))

			h.send(tea.KeyPressMsg{Code: tea.KeyEnter})

			Expect(done).To(Receive(BeNil()))
			Expect(h.view()).NotTo(ContainSubstring("Welcome to tars"))
		})

		It("reports an abort when the form is cancelled", func() {
			h.send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

			Expect(done).To(Receive(MatchError(huh.ErrUserAborted)))
		})
	})

	It("turns ctrl+c during work into a cooperative cancel, then quits when the run ends", func() {
		h.send(tui.StepStartedMsg{Title: "Installing packages", Total: 2}, tui.ItemStartedMsg{Name: "jq"})
		Expect(h.model.Cancelled()).To(BeFalse())

		h.send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

		Expect(h.model.Cancelled()).To(BeTrue())
		Expect(h.view()).To(ContainSubstring("stopping"))

		_, cmd := h.model.Update(tui.RunDoneMsg{Err: errors.New("interrupted")})
		Expect(cmd).NotTo(BeNil())
		Expect(cmd()).To(Equal(tea.Quit()))
	})
})
