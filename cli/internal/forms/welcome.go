package forms

import (
	"os"

	"github.com/charmbracelet/huh"
)

// ShowWelcome displays a full-screen welcome Note using huh.
// Set TARS_NO_FORM=1 to skip the TUI (used in tests/CI).
func ShowWelcome() error {
	if os.Getenv("TARS_NO_FORM") != "" {
		return nil
	}
	return run(huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("tars — dev-machine setup").
				Description(
					"Welcome to tars!\n\n" +
						"This tool will initialize your development environment\n" +
						"config at ~/.config/tars/config.yaml.\n\n" +
						"Press *Enter* to continue or *Ctrl+C* to abort.",
				).
				Next(true).
				NextLabel("Continue"),
		),
	))
}
