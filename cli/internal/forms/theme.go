package forms

import (
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

var (
	accent = lipgloss.Color("#7AA2F7")
	dim    = lipgloss.Color("240")
)

// theme makes the focused field unmistakable: accent bar and a pointer on the
// active field, accent cursor and check marks, everything inactive dimmed.
// huh's own themes copy Focused into Blurred, so every blurred style is set here.
func theme() huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		t := huh.ThemeCharm(isDark)
		t.Focused.Base = t.Focused.Base.BorderStyle(lipgloss.ThickBorder()).BorderLeft(true).BorderForeground(accent)
		t.Focused.Card = t.Focused.Base
		t.Focused.Title = t.Focused.Title.Bold(true).Foreground(accent).SetString("▶")
		t.Focused.SelectSelector = lipgloss.NewStyle().Foreground(accent).SetString("▶ ")
		t.Focused.MultiSelectSelector = lipgloss.NewStyle().Foreground(accent).SetString("▶ ")
		t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(accent).SetString("✓ ")
		t.Focused.UnselectedPrefix = lipgloss.NewStyle().SetString("  ")
		t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(accent)
		t.Focused.FocusedButton = t.Focused.FocusedButton.Background(accent)

		t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).BorderForeground(dim)
		t.Blurred.Card = t.Blurred.Base
		t.Blurred.Title = lipgloss.NewStyle().Foreground(dim)
		t.Blurred.Description = lipgloss.NewStyle().Foreground(dim)
		t.Blurred.SelectSelector = lipgloss.NewStyle().SetString("  ")
		t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")
		t.Blurred.SelectedPrefix = lipgloss.NewStyle().Foreground(dim).SetString("✓ ")
		t.Blurred.UnselectedPrefix = lipgloss.NewStyle().SetString("  ")
		t.Blurred.SelectedOption = lipgloss.NewStyle().Foreground(dim)
		t.Blurred.UnselectedOption = lipgloss.NewStyle().Foreground(dim)
		t.Blurred.FocusedButton = t.Focused.FocusedButton.Foreground(lipgloss.Color("252")).Background(dim)
		t.Blurred.BlurredButton = t.Focused.BlurredButton.Foreground(dim).Background(lipgloss.Color("236"))
		t.Blurred.TextInput.Text = lipgloss.NewStyle().Foreground(dim)
		t.Blurred.TextInput.Prompt = lipgloss.NewStyle().Foreground(dim)
		return t
	})
}

// keymap toggles multi-select rows with space only (huh's default also binds x).
func keymap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.MultiSelect.Toggle = key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle"))
	return km
}

// confirm is a yes/no question with its buttons under the title, not centered.
func confirm(title, description string, value *bool) *huh.Confirm {
	return huh.NewConfirm().Title(title).Description(description).WithButtonAlignment(lipgloss.Left).Value(value)
}

// run applies the shared theme and keymap so every tars form looks and behaves the same.
func run(f *huh.Form) error {
	return f.WithTheme(theme()).WithKeyMap(keymap()).Run()
}
