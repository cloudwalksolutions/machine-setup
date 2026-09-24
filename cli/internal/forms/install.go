package forms

import (
	"os"

	"github.com/charmbracelet/huh"
)

var categorizedPackages = map[string][]string{
	"Agentic Coding Tools": {
		"claude-code", "gemini-cli", "pi",
	},
	"Terminal Utilities & Editors": {
		"neovim", "byobu", "fzf", "ripgrep", "bat", "eza",
		"jq", "gh", "lazygit", "lazydocker", "k9s", "k3d", "golangci-lint",
	},
	"Languages & Runtimes": {
		"go", "node", "python", "yarn", "n", "rustup", "ghcup", "ruby", "rvm",
	},
	"DevOps & Infrastructure": {
		"terraform", "ansible", "gcloud-cli", "gcloud",
	},
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

	// Invert categorizedPackages map for fast O(1) runtime lookups
	packageToCategory := make(map[string]string)
	for cat, pkgs := range categorizedPackages {
		for _, pkgName := range pkgs {
			packageToCategory[pkgName] = cat
		}
	}

	// Map each tool name to its category
	categorizedTools := make(map[string][]string)
	for _, name := range toolNames {
		cat := packageToCategory[name]
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
