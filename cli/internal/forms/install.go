package forms

import (
	"os"

	"github.com/charmbracelet/huh"
)

// ShowInstallForm displays a multi-select with all dev tool names pre-checked.
// When MACHINE_SETUP_NO_FORM=1 it returns toolNames unmodified (tests/CI).
func ShowInstallForm(toolNames []string) ([]string, error) {
	if os.Getenv("MACHINE_SETUP_NO_FORM") != "" {
		result := make([]string, len(toolNames))
		copy(result, toolNames)
		return result, nil
	}

	selected := make([]string, len(toolNames))
	copy(selected, toolNames)

	options := make([]huh.Option[string], len(toolNames))
	for i, name := range toolNames {
		options[i] = huh.NewOption(name, name).Selected(true)
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select dev tools to install").
				Description("All CloudWalk defaults are pre-selected. Space to toggle, Enter to confirm.").
				Options(options...).
				Value(&selected),
		),
	).Run()

	return selected, err
}
