package forms

import (
	"os"

	"github.com/charmbracelet/huh"
)

// ShowWelcome displays a full-screen welcome Note using huh.
// Set MACHINE_SETUP_NO_FORM=1 to skip the TUI (used in tests/CI).
func ShowWelcome() error {
	if os.Getenv("MACHINE_SETUP_NO_FORM") != "" {
		return nil
	}
	return huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("CloudWalk Machine Setup").
				Description(
					"Welcome to the CloudWalk machine setup!\n\n"+
						"This tool will initialize your development environment\n"+
						"config at ~/.config/.machine-setup/config.yaml.\n\n"+
						"Press *Enter* to continue or *Ctrl+C* to abort.",
				).
				Next(true).
				NextLabel("Continue"),
		),
	).Run()
}
