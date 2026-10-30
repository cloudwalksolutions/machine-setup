package forms

import "charm.land/huh/v2"

// WizardForm builds the picker of agent setups to run at the end of `tars init`,
// all pre-checked; collect returns the chosen names.
func WizardForm(offered []string) (*huh.Form, func() []string) {
	selected := append([]string{}, offered...)
	options := make([]huh.Option[string], len(offered))
	for i, name := range offered {
		options[i] = huh.NewOption(name, name).Selected(true)
	}
	f := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Agent setups to run now").
			Description("Each one asks its own questions; skip any and run `tars init <name>` later.").
			Options(options...).
			Value(&selected),
	))
	return f, func() []string { return selected }
}
