package tui

import (
	"fmt"
	"io"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"

	"tars/internal/report"
)

// reporter prints finished work above the live pane and sends the model what is
// in progress. Both go through the program's message queue, so lines and state
// arrive in the order the orchestrator produced them.
type reporter struct {
	send      func(tea.Msg)
	println   func(string)
	cancelled *atomic.Bool
}

func newReporter(send func(tea.Msg), println func(string), cancelled *atomic.Bool) *reporter {
	return &reporter{send: send, println: println, cancelled: cancelled}
}

func (r *reporter) StepStarted(title string, total int) {
	r.println(titleTxt.Render(title))
	r.send(StepStartedMsg{Title: title, Total: total})
}

func (r *reporter) ItemStarted(name string) { r.send(ItemStartedMsg{Name: name}) }

func (r *reporter) ItemDone(name string, err error) {
	if err != nil {
		r.println(failedTxt.Render(fmt.Sprintf("  ✗ %s: %v", name, err)))
	} else {
		r.println(fmt.Sprintf("  ✓ %s", name))
	}
	r.send(ItemDoneMsg{Name: name, Err: err})
}

func (r *reporter) Note(line string) { r.println(line) }

func (r *reporter) StepDone(title string, err error) {
	if err != nil {
		r.println(failedTxt.Render(fmt.Sprintf("  ✗ %s: %v", title, err)))
	}
	r.send(StepDoneMsg{Title: title, Err: err})
}

func (r *reporter) Cancelled() bool { return r.cancelled.Load() }

// Output hands out line-splitting writers so child output lands in the live pane.
func (r *reporter) Output() (io.Writer, io.Writer) {
	emit := func(line string) { r.send(OutputLineMsg{Line: line}) }
	return &report.LineWriter{Emit: emit}, &report.LineWriter{Emit: emit}
}
