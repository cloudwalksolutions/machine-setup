package forms

import (
	"os"

	"charm.land/huh/v2"
)

// ShowWizardPicker asks which agent wizards to run at the end of `tars init`,
// all pre-checked. When TARS_NO_FORM=1 it returns every offered name.
func ShowWizardPicker(offered []string) ([]string, error) {
	selected := append([]string{}, offered...)
	if os.Getenv("TARS_NO_FORM") != "" {
		return selected, nil
	}
	options := make([]huh.Option[string], len(offered))
	for i, name := range offered {
		options[i] = huh.NewOption(name, name).Selected(true)
	}
	err := run(huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Agent setups to run now").
			Description("Each one asks its own questions; skip any and run `tars init <name>` later.").
			Options(options...).
			Value(&selected),
	)))
	return selected, err
}
