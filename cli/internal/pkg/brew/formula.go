package brew

import "io"

// Runner runs a brew subcommand with the given args. Returned errors propagate
// to the caller; stdout/stderr are streamed to the provided writers.
type Runner func(args []string, stdout, stderr io.Writer) error

// Formula is a brew package installed via `brew install <name>`.
type Formula struct {
	name string
	run  Runner
}

// NewFormula returns a Formula bound to a runner.
func NewFormula(name string, run Runner) Formula {
	return Formula{name: name, run: run}
}

// Name returns the formula's brew name.
func (f Formula) Name() string { return f.name }

// Install runs `brew install <name>`.
func (f Formula) Install(stdout, stderr io.Writer) error {
	return f.run([]string{"install", f.name}, stdout, stderr)
}
