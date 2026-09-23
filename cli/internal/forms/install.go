package forms

import (
	"os"

	"github.com/charmbracelet/huh"
)

var packageCategories = map[string]string{
	"claude-code":   "Agentic Coding Tools",
	"gemini-cli":    "Agentic Coding Tools",
	"pi":            "Agentic Coding Tools",

	"neovim":        "Terminal Utilities & Editors",
	"byobu":         "Terminal Utilities & Editors",
	"fzf":           "Terminal Utilities & Editors",
	"ripgrep":       "Terminal Utilities & Editors",
	"bat":           "Terminal Utilities & Editors",
	"eza":           "Terminal Utilities & Editors",
	"jq":            "Terminal Utilities & Editors",
	"gh":            "Terminal Utilities & Editors",
	"lazygit":       "Terminal Utilities & Editors",
	"lazydocker":    "Terminal Utilities & Editors",
	"k9s":           "Terminal Utilities & Editors",
	"k3d":           "Terminal Utilities & Editors",
	"golangci-lint": "Terminal Utilities & Editors",

	"go":            "Languages & Runtimes",
	"node":          "Languages & Runtimes",
	"python":        "Languages & Runtimes",
	"yarn":          "Languages & Runtimes",
	"n":             "Languages & Runtimes",
	"rustup":        "Languages & Runtimes",
	"ghcup":         "Languages & Runtimes",
	"ruby":          "Languages & Runtimes",
	"rvm":           "Languages & Runtimes",

	"terraform":     "DevOps & Infrastructure",
	"ansible":       "DevOps & Infrastructure",
	"gcloud-cli":    "DevOps & Infrastructure",
	"gcloud":        "DevOps & Infrastructure",
}

// ShowInstallForm displays a multi-select with all dev tool names grouped by category.
// When TARS_NO_FORM=1 it returns toolNames unmodified (tests/CI).
func ShowInstallForm(toolNames []string) ([]string, error) {
	if os.Getenv("TARS_NO_FORM") != "" {
		result := make([]string, len(toolNames))
		copy(result, toolNames)
		return result, nil
	}

	categoriesOrder := []string{
		"Agentic Coding Tools",
		"Terminal Utilities & Editors",
		"Languages & Runtimes",
		"DevOps & Infrastructure",
		"Other Tools",
	}

	// Map each tool name to its category
	categorizedTools := make(map[string][]string)
	for _, name := range toolNames {
		cat := packageCategories[name]
		if cat == "" {
			cat = "Other Tools"
		}
		categorizedTools[cat] = append(categorizedTools[cat], name)
	}

	var groups []*huh.Group
	selectedMap := make(map[string]*[]string)

	for _, cat := range categoriesOrder {
		tools := categorizedTools[cat]
		if len(tools) == 0 {
			continue
		}

		selected := make([]string, len(tools))
		copy(selected, tools)
		selectedMap[cat] = &selected

		options := make([]huh.Option[string], len(tools))
		for i, name := range tools {
			options[i] = huh.NewOption(name, name).Selected(true)
		}

		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(cat).
				Description("Select which ones to install. Space to toggle, Enter to confirm.").
				Options(options...).
				Value(&selected),
		))
	}

	err := huh.NewForm(groups...).Run()
	if err != nil {
		return nil, err
	}

	// Merge all selected slices preserving the categorized order
	var finalSelected []string
	for _, cat := range categoriesOrder {
		if sel, ok := selectedMap[cat]; ok {
			finalSelected = append(finalSelected, *sel...)
		}
	}

	return finalSelected, nil
}
