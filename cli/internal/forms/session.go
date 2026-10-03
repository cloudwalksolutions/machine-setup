package forms

import (
	"errors"

	"charm.land/huh/v2"
)

// ShowSessionPicker displays a single-select over the given session options.
// When TARS_NO_FORM=1 it returns the first option (tests/CI).
func ShowSessionPicker(options []string) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no sessions to pick from")
	}
	if headless() {
		return options[0], nil
	}

	selected := options[0]
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}

	err := run(huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Open a byobu session").
				Options(opts...).
				Value(&selected),
		),
	))

	return selected, err
}
