package tui

import "charm.land/huh/v2"

// StepStartedMsg opens a step with an optional item count.
type StepStartedMsg struct {
	Title string
	Total int
}

// ItemStartedMsg names the item now in progress.
type ItemStartedMsg struct{ Name string }

// ItemDoneMsg closes an item, failed when Err is set.
type ItemDoneMsg struct {
	Name string
	Err  error
}

// OutputLineMsg is one line of child-process output.
type OutputLineMsg struct{ Line string }

// StepDoneMsg closes a step, failed when Err is set.
type StepDoneMsg struct {
	Title string
	Err   error
}

// PromptMsg asks the model to run a form; Done receives nil or huh.ErrUserAborted.
type PromptMsg struct {
	Form *huh.Form
	Done chan<- error
}

// RunDoneMsg ends the program with the orchestrator's result.
type RunDoneMsg struct{ Err error }
