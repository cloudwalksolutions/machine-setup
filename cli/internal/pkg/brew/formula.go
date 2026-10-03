package brew

import (
	"bytes"
	"io"
	"strings"

	"tars/internal/pkg"
)

// Runner runs a brew subcommand with the given args. Returned errors propagate
// to the caller; stdout/stderr are streamed to the provided writers.
type Runner func(args []string, stdout, stderr io.Writer) error

// Formula is a brew package installed via `brew install <name>`.
type Formula struct {
	name        string
	description string
	run         Runner
	// Binary is the executable to look for when brew does not list the formula; defaults to name.
	Binary string
	Probe  pkg.PathProbe
}

// NewFormula returns a Formula bound to a runner.
func NewFormula(name, description string, run Runner) Formula {
	return Formula{name: name, description: description, run: run}
}

// Name returns the formula's brew name.
func (f Formula) Name() string { return f.name }

// Description returns the picker blurb.
func (f Formula) Description() string { return f.description }

// Install runs `brew install <name>`.
func (f Formula) Install(stdout, stderr io.Writer) error {
	return f.run([]string{"install", f.name}, stdout, stderr)
}

// Status reports the brew-listed version, else whatever the binary on PATH reports.
func (f Formula) Status() (pkg.InstallStatus, string, error) {
	if version, ok := listedVersion(f.run, "list", "--versions", f.name); ok {
		return pkg.StatusUpToDate, version, nil
	}
	return f.Probe.Status(binaryOr(f.Binary, f.name))
}

// listedVersion parses `<name> <version>` from brew's stdout; a non-zero exit means brew does not manage it.
func listedVersion(run Runner, args ...string) (string, bool) {
	var out bytes.Buffer
	if err := run(args, &out, io.Discard); err != nil {
		return "", false
	}
	fields := strings.Fields(out.String())
	if len(fields) < 2 {
		return "", false
	}
	return fields[len(fields)-1], true
}

func binaryOr(binary, name string) string {
	if binary != "" {
		return binary
	}
	return name
}
