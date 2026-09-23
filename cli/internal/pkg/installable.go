package pkg

import "io"

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
