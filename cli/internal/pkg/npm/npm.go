package npm

import (
	"io"
	"os/exec"
)

// Runner runs an npm subcommand with the given args. Returned errors propagate
// to the caller; stdout/stderr are streamed to the provided writers.
type Runner func(args []string, stdout, stderr io.Writer) error

// Package represents an npm package installed globally.
type Package struct {
	displayName string
	packageName string
	run         Runner
}

// NewPackage returns an npm Package bound to a runner.
func NewPackage(displayName, packageName string, run Runner) Package {
	return Package{
		displayName: displayName,
		packageName: packageName,
		run:         run,
	}
}

// Name reports the display name.
func (p Package) Name() string {
	return p.displayName
}

// Install runs `npm install -g <packageName>`.
func (p Package) Install(stdout, stderr io.Writer) error {
	return p.run([]string{"install", "-g", p.packageName}, stdout, stderr)
}

// DefaultRunner returns the production Runner: a direct execution of npm.
func DefaultRunner() Runner {
	return func(args []string, stdout, stderr io.Writer) error {
		cmd := exec.Command("npm", args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
}
