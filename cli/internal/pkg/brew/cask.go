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

// Status reports the installed version via `brew list --versions --cask <name>`.
func (c Cask) Status() (pkg.InstallStatus, string, error) {
	return listedVersion(c.run, "list", "--versions", "--cask", c.name)
}
