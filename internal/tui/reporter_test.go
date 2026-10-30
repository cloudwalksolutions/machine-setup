package tui

import (
	"errors"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/report"
)

var _ = Describe("reporter", func() {
	var (
		sent      []tea.Msg
		printed   []string
		cancelled atomic.Bool
		r         report.Reporter
	)

	BeforeEach(func() {
		sent, printed = nil, nil
		cancelled.Store(false)
		r = newReporter(
			func(msg tea.Msg) { sent = append(sent, msg) },
			func(line string) { printed = append(printed, ansi.Strip(line)) },
			&cancelled,
		)
	})

	It("prints finished work above the pane, in order with the state messages", func() {
		boom := errors.New("boom")
		r.StepStarted("Installing packages", 3)
		r.ItemStarted("jq")
		r.ItemDone("jq", nil)
		r.ItemStarted("go")
		r.ItemDone("go", boom)
		r.Note("Config written")
		r.StepDone("Installing packages", nil)
		r.StepDone("Installing oh-my-zsh", errors.New("git missing"))

		Expect(printed).To(Equal([]string{
			"Installing packages",
			"  ✓ jq",
			"  ✗ go: boom",
			"Config written",
			"  ✗ Installing oh-my-zsh: git missing",
		}))
		Expect(sent).To(Equal([]tea.Msg{
			StepStartedMsg{Title: "Installing packages", Total: 3},
			ItemStartedMsg{Name: "jq"},
			ItemDoneMsg{Name: "jq"},
			ItemStartedMsg{Name: "go"},
			ItemDoneMsg{Name: "go", Err: boom},
			StepDoneMsg{Title: "Installing packages"},
			StepDoneMsg{Title: "Installing oh-my-zsh", Err: errors.New("git missing")},
		}))
	})

	It("streams child output line by line from both writers", func() {
		stdout, stderr := r.Output()
		_, _ = stdout.Write([]byte("fetching\ndone\n"))
		_, _ = stderr.Write([]byte("warning: slow\n"))

		Expect(sent).To(Equal([]tea.Msg{
			OutputLineMsg{Line: "fetching"},
			OutputLineMsg{Line: "done"},
			OutputLineMsg{Line: "warning: slow"},
		}))
	})

	It("relays the cancel flag", func() {
		Expect(r.Cancelled()).To(BeFalse())
		cancelled.Store(true)
		Expect(r.Cancelled()).To(BeTrue())
	})
})
