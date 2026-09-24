package pkg

import (
	"io"
	"os"
	"os/exec"
)

// InstallStatus represents the current installation/update status of a package.
type InstallStatus int

const (
	StatusNotInstalled InstallStatus = iota
	StatusUpdateAvailable
	StatusUpToDate
)

// String returns a user-friendly representation of the install status.
func (s InstallStatus) String() string {
	switch s {
	case StatusNotInstalled:
		return "not installed"
	case StatusUpdateAvailable:
		return "update available"
	case StatusUpToDate:
		return "up to date"
	default:
		return "unknown"
	}
}

// ToolInfo is what the install picker shows for one installable.
type ToolInfo struct {
	Name        string
	Description string
	Installed   bool
	Version     string
}

// Installable is the polymorphic surface for "something the CLI knows how to
// install on a machine". Each kind (brew formula, brew cask, apt package,
// AppImage download) is its own type that satisfies this interface — there
// are no type switches anywhere downstream.
type Installable interface {
	Name() string
	Description() string
	Install(stdout, stderr io.Writer) error
	Status() (InstallStatus, string, error)
}

// PostInstall wraps an Installable with commands that run after a successful install.
type PostInstall struct {
	Installable
	steps [][]string
	run   func(cmd []string, stdout, stderr io.Writer) error
}

// WithPostInstall binds follow-up commands (e.g. `rustup default stable`) to an installable.
func WithPostInstall(inst Installable, steps [][]string, run func(cmd []string, stdout, stderr io.Writer) error) PostInstall {
	return PostInstall{Installable: inst, steps: steps, run: run}
}

// Install runs the wrapped install, then each step in order; a failing install skips the steps.
func (p PostInstall) Install(stdout, stderr io.Writer) error {
	if err := p.Installable.Install(stdout, stderr); err != nil {
		return err
	}
	for _, step := range p.steps {
		if err := p.run(step, stdout, stderr); err != nil {
			return err
		}
	}
	return nil
}

// ScriptInstaller is a generic, reusable Installable for tools installed
// via a custom shell script or command (e.g. curl | bash) and verified by
// checking if a binary/file exists on disk.
type ScriptInstaller struct {
	name        string
	description string
	checkPath   string
	installCmd  []string
	run         func(cmd []string, stdout, stderr io.Writer) error
}

// NewScriptInstaller creates a generic script-based installer.
func NewScriptInstaller(name, description, checkPath string, installCmd []string, run func(cmd []string, stdout, stderr io.Writer) error) ScriptInstaller {
	if run == nil {
		run = func(cmd []string, stdout, stderr io.Writer) error {
			c := exec.Command(cmd[0], cmd[1:]...)
			c.Stdout = stdout
			c.Stderr = stderr
			return c.Run()
		}
	}
	return ScriptInstaller{
		name:        name,
		description: description,
		checkPath:   checkPath,
		installCmd:  installCmd,
		run:         run,
	}
}

// Name returns the display name.
func (s ScriptInstaller) Name() string { return s.name }

// Description returns the picker blurb.
func (s ScriptInstaller) Description() string { return s.description }

// Install runs the custom installation command if checkPath does not exist.
func (s ScriptInstaller) Install(stdout, stderr io.Writer) error {
	if _, err := os.Stat(s.checkPath); err == nil {
		return nil
	}
	return s.run(s.installCmd, stdout, stderr)
}

// Status reports the version `<checkPath> --version` prints when checkPath exists.
func (s ScriptInstaller) Status() (InstallStatus, string, error) {
	if _, err := os.Stat(s.checkPath); err != nil {
		return StatusNotInstalled, "", nil
	}
	probe := PathProbe{LookPath: func(string) (string, error) { return s.checkPath, nil }, Run: s.run}
	return probe.Status(s.checkPath)
}
