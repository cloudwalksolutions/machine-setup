package forms

import (
	"fmt"

	"charm.land/huh/v2"

	"tars/internal/pkg"
)

var categorizedPackages = map[string][]string{
	"Agentic Coding Tools": {
		"claude-code", "gemini-cli", "pi",
	},
	"Terminal Utilities & Editors": {
		"neovim", "tree-sitter-cli", "byobu", "gh", "lazygit", "jq", "bat", "eza",
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

// InstallForm builds the tool picker over the catalog grouped by category, in catalog
// order. Installed tools show their version and start unchecked; the rest start checked.
// collect returns the chosen names in category order.
func InstallForm(tools []pkg.ToolInfo) (*huh.Form, func() []string) {
	byCategory := groupByCategory(tools)
	var groups []*huh.Group
	var selections []*[]string
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
		selections = append(selections, &selected)
		groups = append(groups, huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(cat).
				Description("Installed tools show their version and start unchecked.").
				Options(options...).
				Height(len(options)+2). // huh subtracts the title and description rows from the options
				Value(&selected),
		))
	}
	collect := func() []string {
		var final []string
		for _, sel := range selections {
			final = append(final, *sel...)
		}
		return final
	}
	return huh.NewForm(groups...), collect
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
