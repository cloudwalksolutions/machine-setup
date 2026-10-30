package forms

import (
	"charm.land/huh/v2"

	"tars/internal/config"
)

// ClaudeForm builds the `tars init claude` questions: the edit-blocking hook and
// the settings fragment. Both start on.
func ClaudeForm() (*huh.Form, func() config.ClaudeConfig) {
	hook, settings := true, true
	f := huh.NewForm(
		huh.NewGroup(
			confirm("Install the edit-blocking hook?",
				"Denies sed -i / heredoc / interpreter writes so every change is a reviewable Edit or Write.",
				&hook),
			confirm("Merge settings into ~/.claude/settings.json?",
				"Model, theme, plugins and the hook entry. Machine-local keys such as autoMode and permissions are kept.",
				&settings),
		),
	)
	return f, func() config.ClaudeConfig {
		return config.ClaudeConfig{Hook: &hook, Settings: &settings}
	}
}
