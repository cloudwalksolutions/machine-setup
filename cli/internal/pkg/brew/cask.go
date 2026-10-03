package brew

import (
	"io"
	"tars/internal/pkg"
)

// Cask is a brew package installed via `brew install --cask <name>`.
type Cask struct {
	name        string
	description string
	run         Runner
	// Binary is the executable to look for when brew does not list the cask; defaults to name.
	Binary string
	Probe  pkg.PathProbe
}

// NewCask returns a Cask bound to a runner.
func NewCask(name, description string, run Runner) Cask {
	return Cask{name: name, description: description, run: run}
}

// Name returns the cask's brew name.
func (c Cask) Name() string { return c.name }

// Description returns the picker blurb.
func (c Cask) Description() string { return c.description }

// Install runs `brew install --cask <name>`.
func (c Cask) Install(stdout, stderr io.Writer) error {
	return c.run([]string{"install", "--cask", c.name}, stdout, stderr)
}

// Status reports the brew-listed version, else whatever the binary on PATH reports.
func (c Cask) Status() (pkg.InstallStatus, string, error) {
	if version, ok := listedVersion(c.run, "list", "--versions", "--cask", c.name); ok {
		return pkg.StatusUpToDate, version, nil
	}
	return c.Probe.Status(binaryOr(c.Binary, c.name))
}
