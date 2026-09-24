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

// Installable is the polymorphic surface for "something the CLI knows how to
// install on a machine". Each kind (brew formula, brew cask, apt package,
// AppImage download) is its own type that satisfies this interface — there
// are no type switches anywhere downstream.
type Installable interface {
	Name() string
	Install(stdout, stderr io.Writer) error
	Status() (InstallStatus, string, error)
}

// ScriptInstaller is a generic, reusable Installable for tools installed
// via a custom shell script or command (e.g. curl | bash) and verified by
// checking if a binary/file exists on disk.
type ScriptInstaller struct {
	name       string
	checkPath  string
	installCmd []string
	run        func(cmd []string, stdout, stderr io.Writer) error
}

// NewScriptInstaller creates a generic script-based installer.
func NewScriptInstaller(name, checkPath string, installCmd []string, run func(cmd []string, stdout, stderr io.Writer) error) ScriptInstaller {
	if run == nil {
		run = func(cmd []string, stdout, stderr io.Writer) error {
			c := exec.Command(cmd[0], cmd[1:]...)
			c.Stdout = stdout
			c.Stderr = stderr
			return c.Run()
		}
	}
	return ScriptInstaller{
		name:       name,
		checkPath:  checkPath,
		installCmd: installCmd,
		run:        run,
	}
}

// Name returns the display name.
func (s ScriptInstaller) Name() string { return s.name }

// Install runs the custom installation command if checkPath does not exist.
func (s ScriptInstaller) Install(stdout, stderr io.Writer) error {
	if _, err := os.Stat(s.checkPath); err == nil {
		return nil
	}
	return s.run(s.installCmd, stdout, stderr)
}

// Status returns UpToDate if checkPath exists, otherwise NotInstalled.
func (s ScriptInstaller) Status() (InstallStatus, string, error) {
	if _, err := os.Stat(s.checkPath); err == nil {
		return StatusUpToDate, "installed", nil
	}
	return StatusNotInstalled, "", nil
}
