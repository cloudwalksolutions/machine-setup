package report_test

import (
	"bytes"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"tars/internal/report"
)

var _ = Describe("Text", func() {
	var (
		stdout, stderr *bytes.Buffer
		text           report.Text
	)

	BeforeEach(func() {
		stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
		text = report.Text{Stdout: stdout, Stderr: stderr}
	})

	It("StepStarted announces the step and ItemStarted lists each item under it", func() {
		text.StepStarted("Pulling configuration files", 2)
		text.ItemStarted("vim")
		text.ItemStarted("zsh")

		Expect(stdout.String()).To(Equal("\nPulling configuration files...\n  → vim\n  → zsh\n"))
	})

	It("Note prints a line, StepDone reports only failures, Output hands out the writers, nothing is ever cancelled", func() {
		text.Note("Config written to /tmp/config.yaml")
		text.StepDone("Pulling configuration files", nil)
		text.StepDone("Pulling configuration files", errors.New("zsh: disk full"))
		out, errw := text.Output()

		Expect(stdout.String()).To(Equal("Config written to /tmp/config.yaml\n"))
		Expect(stderr.String()).To(Equal("  Pulling configuration files: zsh: disk full\n"))
		Expect(out).To(BeIdenticalTo(io.Writer(stdout)))
		Expect(errw).To(BeIdenticalTo(io.Writer(stderr)))
		Expect(text.Cancelled()).To(BeFalse())
		var _ report.Reporter = text
	})

	It("ItemDone writes the failure to stderr and nothing on success", func() {
		text.ItemDone("jq", errors.New("mirror down"))
		text.ItemDone("fzf", nil)

		Expect(stderr.String()).To(Equal("  jq: mirror down\n"))
		Expect(stdout.String()).To(BeEmpty())
	})
})
