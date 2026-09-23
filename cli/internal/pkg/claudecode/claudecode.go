package claudecode

import (
	"io"
	"os"
	"os/exec"
	"tars/internal/pkg"
)

// installScript is the official native installer script for Claude Code.
const installScript = `curl -fsSL https://claude.ai/install.sh | bash`

// Installer installs Claude Code by running its official bootstrap.
type Installer struct {
	// Path is the binary location that signals "already installed" (typically ~/.local/bin/claude).
	Path string
	// Runner is the side-effect; tests replace it.
	Runner func(stdout, stderr io.Writer) error
}

// NewInstaller binds an Installer to a target Path and a Runner.
func NewInstaller(path string, run func(stdout, stderr io.Writer) error) Installer {
	return Installer{Path: path, Runner: run}
}

// Name reports "claude-code" for registry/log display.
func (Installer) Name() string { return "claude-code" }

// Install runs the bootstrap if Path does not exist; otherwise no-ops.
func (i Installer) Install(stdout, stderr io.Writer) error {
	if _, err := os.Stat(i.Path); err == nil {
		return nil
	}
	return i.Runner(stdout, stderr)
}

// Status returns the current installation status of Claude Code.
func (i Installer) Status() (pkg.InstallStatus, string, error) {
	if _, err := os.Stat(i.Path); err == nil {
		return pkg.StatusUpToDate, "installed", nil
	}
	return pkg.StatusNotInstalled, "", nil
}

// DefaultRunner returns the production Runner: a direct bash pipe of the official install script.
func DefaultRunner() func(stdout, stderr io.Writer) error {
	return func(stdout, stderr io.Writer) error {
		cmd := exec.Command("bash", "-c", installScript)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
}
