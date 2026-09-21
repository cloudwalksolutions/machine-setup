package forms

import (
	"errors"
	"os"

	"github.com/charmbracelet/huh"
)

// ShowSessionPicker displays a single-select over the given session options.
// When MACHINE_SETUP_NO_FORM=1 it returns the first option (tests/CI).
func ShowSessionPicker(options []string) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no sessions to pick from")
	}
	if os.Getenv("MACHINE_SETUP_NO_FORM") != "" {
		return options[0], nil
	}

	selected := options[0]
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Open a byobu session").
				Options(opts...).
				Value(&selected),
		),
	).Run()

	return selected, err
}
