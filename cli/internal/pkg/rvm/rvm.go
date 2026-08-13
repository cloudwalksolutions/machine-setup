// Package rvm provides the Ruby Version Manager installer. RVM is not in
// brew; its canonical install is the official curl-pipe script at get.rvm.io.
// This package wraps that install so it composes alongside brew/apt entries
// in the dev-tool registry via the pkg.Installable interface.
package rvm

import (
	"io"
	"os"
	"os/exec"
)

// installScript is the official RVM bootstrap command. Piped through bash with
// the `stable` channel — same as the canonical instructions on rvm.io.
const installScript = `\curl -sSL https://get.rvm.io | bash -s stable`

// Installer installs RVM by running its official curl-pipe bootstrap.
type Installer struct {
	// Dir is the path that signals "already installed" (typically ~/.rvm).
	Dir string
	// Runner is the side-effect; tests replace it.
	Runner func(stdout, stderr io.Writer) error
}

// NewInstaller binds an Installer to a target Dir and a Runner.
func NewInstaller(dir string, run func(stdout, stderr io.Writer) error) Installer {
	return Installer{Dir: dir, Runner: run}
}

// Name reports "rvm" for registry/log display.
func (Installer) Name() string { return "rvm" }

// Install runs the bootstrap if Dir does not exist; otherwise no-ops.
func (i Installer) Install(stdout, stderr io.Writer) error {
	if _, err := os.Stat(i.Dir); err == nil {
		return nil
	}
	return i.Runner(stdout, stderr)
}

// DefaultRunner returns the production Runner: a bash pipe of the official
// RVM install script with the `stable` channel.
func DefaultRunner() func(stdout, stderr io.Writer) error {
	return func(stdout, stderr io.Writer) error {
		cmd := exec.Command("bash", "-c", installScript)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
}
