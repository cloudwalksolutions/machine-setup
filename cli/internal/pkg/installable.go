package pkg

import "io"

// Installable is the polymorphic surface for "something the CLI knows how to
// install on a machine". Each kind (brew formula, brew cask, apt package,
// AppImage download) is its own type that satisfies this interface — there
// are no type switches anywhere downstream.
type Installable interface {
	Name() string
	Install(stdout, stderr io.Writer) error
}
