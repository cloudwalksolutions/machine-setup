package forms

import (
	"os"

	"github.com/charmbracelet/huh"

	"tars/internal/profiles"
)

// ShowProfileForm asks for a new profile's identity, starting from defaults.
// When TARS_NO_FORM=1 it returns defaults unchanged (tests/CI).
func ShowProfileForm(defaults profiles.Profile) (profiles.Profile, error) {
	if os.Getenv("TARS_NO_FORM") != "" {
		return defaults, nil
	}

	p := defaults
	err := run(huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name (also the folder under your projects dir)").Value(&p.Name),
			huh.NewInput().Title("Alias (short handle; key is ~/.ssh/id_rsa.<alias>)").Value(&p.Alias),
			huh.NewInput().Title("Email").Value(&p.Email),
			huh.NewInput().Title("GitHub user").Value(&p.GitHub),
			huh.NewInput().Title("Full name (blank to use the shared one)").Value(&p.FullName),
		),
	))

	return p, err
}
