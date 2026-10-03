package tui_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/report"
	"tars/internal/tui"
)

var _ = Describe("Start under a real program", func() {
	It("reports, streams output, runs a form on the user's keys, and returns the flow's result", func() {
		model := tui.New()
		tm := teatest.NewTestModel(GinkgoTB(), model, teatest.WithInitialTermSize(100, 40))
		var picked []string
		flow := func(r report.Reporter, ask tui.Asker) error {
			r.StepStarted("Installing packages", 1)
			r.ItemStarted("jq")
			stdout, _ := r.Output()
			_, _ = io.WriteString(stdout, "fetching jq\n")
			r.ItemDone("jq", nil)
			var err error
			if picked, err = ask.PickWizards([]string{"claude", "pi"}); err != nil {
				return err
			}
			r.Note("Setup complete.")
			return errors.New("flow result")
		}

		tui.Start(context.Background(), tm.Send, model, flow)

		// WaitFor consumes the output stream, so keep a copy of everything it reads.
		var screen bytes.Buffer
		teatest.WaitFor(GinkgoTB(), io.TeeReader(tm.Output(), &screen), func(b []byte) bool {
			return strings.Contains(ansi.Strip(string(b)), "Agent setups to run now")
		}, teatest.WithDuration(5*time.Second))
		tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

		final := tm.FinalModel(GinkgoTB(), teatest.WithFinalTimeout(5*time.Second)).(tui.Model)
		Expect(final.Err()).To(MatchError("flow result"))
		Expect(picked).To(Equal([]string{"claude", "pi"}))

		_, err := io.Copy(&screen, tm.FinalOutput(GinkgoTB()))
		Expect(err).NotTo(HaveOccurred())
		text := ansi.Strip(screen.String())
		Expect(text).To(ContainSubstring("Installing packages"))
		Expect(text).To(ContainSubstring("✓ jq"))
		Expect(text).To(ContainSubstring("Setup complete."))
	})
})
