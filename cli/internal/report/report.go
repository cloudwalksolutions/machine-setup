// Package report is how orchestrators tell the user what is happening: a
// Reporter receives steps, items and child-process output, and either prints
// plain text (Text) or renders a TUI.
package report

import (
	"fmt"
	"io"
)

// Reporter receives progress from an orchestrator.
type Reporter interface {
	StepStarted(title string, total int)
	ItemStarted(name string)
	ItemDone(name string, err error)
	Note(line string)
	StepDone(title string, err error)
	Output() (stdout, stderr io.Writer)
	Cancelled() bool
}

// Text prints progress as plain lines, the headless and CI path.
type Text struct {
	Stdout io.Writer
	Stderr io.Writer
}

// StepStarted announces a step on its own paragraph.
func (t Text) StepStarted(title string, _ int) {
	fmt.Fprintf(t.Stdout, "\n%s...\n", title)
}

// ItemStarted lists an item under the current step.
func (t Text) ItemStarted(name string) {
	fmt.Fprintf(t.Stdout, "  → %s\n", name)
}

// ItemDone prints a failed item to stderr; successes print nothing.
func (t Text) ItemDone(name string, err error) {
	if err != nil {
		fmt.Fprintf(t.Stderr, "  %s: %v\n", name, err)
	}
}

// Note prints one informational line.
func (t Text) Note(line string) {
	fmt.Fprintln(t.Stdout, line)
}

// StepDone prints a failed step to stderr; successes print nothing.
func (t Text) StepDone(title string, err error) {
	if err != nil {
		fmt.Fprintf(t.Stderr, "  %s: %v\n", title, err)
	}
}

// Output streams child-process output straight through.
func (t Text) Output() (io.Writer, io.Writer) {
	return t.Stdout, t.Stderr
}

// Cancelled is always false: plain text has no interrupt handling of its own.
func (t Text) Cancelled() bool { return false }
