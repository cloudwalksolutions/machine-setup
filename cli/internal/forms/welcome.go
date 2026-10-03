package forms

import "charm.land/huh/v2"

// WelcomeForm is the opening screen of `tars init`.
func WelcomeForm() *huh.Form {
	return huh.NewForm(
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
	)
}
