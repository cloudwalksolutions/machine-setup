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

// Status reports the installed version via `brew list --versions <name>`.
func (f Formula) Status() (pkg.InstallStatus, string, error) {
	return listedVersion(f.run, "list", "--versions", f.name)
}

// listedVersion parses `<name> <version>` from brew's stdout; a non-zero exit means not installed.
func listedVersion(run Runner, args ...string) (pkg.InstallStatus, string, error) {
	var out bytes.Buffer
	if err := run(args, &out, io.Discard); err != nil {
		return pkg.StatusNotInstalled, "", nil
	}
	fields := strings.Fields(out.String())
	if len(fields) < 2 {
		return pkg.StatusNotInstalled, "", nil
	}
	return pkg.StatusUpToDate, fields[len(fields)-1], nil
}
