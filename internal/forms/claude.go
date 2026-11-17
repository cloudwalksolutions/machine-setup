package forms

import (
	"charm.land/huh/v2"

	"tars/internal/config"
)

// ClaudeForm asks about the hook and the settings fragment; both start on.
func ClaudeForm() (*huh.Form, func() config.ClaudeConfig) {
	hook, settings := true, true
	f := huh.NewForm(
		huh.NewGroup(
			confirm("Install the hooks?",
				"The edit guard denies sed -i / heredoc / interpreter writes so every change is a reviewable Edit or Write; the PR-link check keeps a turn from ending without the open PR's link.",
				&hook),
			confirm("Merge settings into ~/.claude/settings.json?",
				"Model, theme, plugins and the hook entries. Machine-local keys such as autoMode and permissions are kept.",
				&settings),
		),
	)
	return f, func() config.ClaudeConfig {
		return config.ClaudeConfig{Hook: &hook, Settings: &settings}
	}
}
