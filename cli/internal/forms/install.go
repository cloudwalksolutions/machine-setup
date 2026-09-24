package forms

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"

	"tars/internal/pkg"
)

var categorizedPackages = map[string][]string{
	"Agentic Coding Tools": {
		"claude-code", "gemini-cli", "pi",
	},
	"Terminal Utilities & Editors": {
		"neovim", "byobu", "gh", "lazygit", "jq", "bat", "eza",
		"k9s", "lazydocker", "k3d", "golangci-lint", "fzf", "ripgrep",
	},
	"Languages & Runtimes": {
		"go", "node", "python", "ruby", "rustup", "ghcup", "yarn", "n", "rvm",
	},
	"DevOps & Infrastructure": {
		"terraform", "ansible", "gcloud-cli", "gcloud",
	},
}

var categoriesOrder = []string{
	"Agentic Coding Tools",
	"Terminal Utilities & Editors",
	"Languages & Runtimes",
	"DevOps & Infrastructure",
	"Other Tools",
}

// ShowInstallForm displays a multi-select over the catalog grouped by category, in
// catalog order. Installed tools show their version and start unchecked; the rest start
// checked. When TARS_NO_FORM=1 it selects every tool that is not installed yet (tests/CI).
func ShowInstallForm(tools []pkg.ToolInfo) ([]string, error) {
	if os.Getenv("TARS_NO_FORM") != "" {
		var missing []string
		for _, t := range tools {
			if !t.Installed {
				missing = append(missing, t.Name)
			}
		}
		return missing, nil
	}

	byCategory := groupByCategory(tools)
	var groups []*huh.Group
	selections := make(map[string]*[]string)
	for _, cat := range categoriesOrder {
		catTools := byCategory[cat]
		if len(catTools) == 0 {
			continue
		}
		selected := []string{}
		options := make([]huh.Option[string], len(catTools))
		for i, t := range catTools {
			options[i] = huh.NewOption(row(t), t.Name).Selected(!t.Installed)
			if !t.Installed {
				selected = append(selected, t.Name)
			}
		}
		selections[cat] = &selected
		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(cat).
				Description("Installed tools show their version and start unchecked. Space to toggle, Enter to confirm.").
				Options(options...).
				Value(&selected),
		))
	}

	if err := run(huh.NewForm(groups...)); err != nil {
		return nil, err
	}

	var final []string
	for _, cat := range categoriesOrder {
		if sel, ok := selections[cat]; ok {
			final = append(final, *sel...)
		}
	}
	return final, nil
}

func groupByCategory(tools []pkg.ToolInfo) map[string][]pkg.ToolInfo {
	categoryOf := make(map[string]string)
	for cat, names := range categorizedPackages {
		for _, n := range names {
			categoryOf[n] = cat
		}
	}
	grouped := make(map[string][]pkg.ToolInfo)
	for _, t := range tools {
		cat := categoryOf[t.Name]
		if cat == "" {
			cat = "Other Tools"
		}
		grouped[cat] = append(grouped[cat], t)
	}
	return grouped
}

// row renders `name  description  ✓ version` as one aligned option label.
func row(t pkg.ToolInfo) string {
	status := ""
	if t.Installed {
		status = "✓ " + t.Version
	}
	return fmt.Sprintf("%-14s %-44s %s", t.Name, t.Description, status)
}
