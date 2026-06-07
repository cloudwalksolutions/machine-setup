package brew

import "io"

// Cask is a brew package installed via `brew install --cask <name>`.
type Cask struct {
	name string
	run  Runner
}

// NewCask returns a Cask bound to a runner.
func NewCask(name string, run Runner) Cask {
	return Cask{name: name, run: run}
}

// Name returns the cask's brew name.
func (c Cask) Name() string { return c.name }

// Install runs `brew install --cask <name>`.
func (c Cask) Install(stdout, stderr io.Writer) error {
	return c.run([]string{"install", "--cask", c.name}, stdout, stderr)
}
