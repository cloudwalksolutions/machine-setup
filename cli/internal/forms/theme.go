package forms

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var accent = lipgloss.Color("#7AA2F7")

// theme makes the focused field unmistakable: accent bar and bold title on the
// active field, accent cursor and check marks, everything inactive dimmed.
func theme() *huh.Theme {
	t := huh.ThemeCharm()
	t.Focused.Base = t.Focused.Base.BorderStyle(lipgloss.ThickBorder()).BorderLeft(true).BorderForeground(accent)
	t.Focused.Title = t.Focused.Title.Bold(true).Foreground(accent)
	t.Focused.SelectSelector = lipgloss.NewStyle().Foreground(accent).SetString("▶ ")
	t.Focused.MultiSelectSelector = lipgloss.NewStyle().Foreground(accent).SetString("▶ ")
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(accent).SetString("✓ ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().SetString("  ")
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(accent)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Background(accent)
	t.Blurred.Base = t.Blurred.Base.BorderLeft(true).BorderForeground(lipgloss.Color("240"))
	t.Blurred.Title = t.Blurred.Title.Faint(true)
	t.Blurred.Description = t.Blurred.Description.Faint(true)
	t.Blurred.SelectedOption = t.Blurred.SelectedOption.Faint(true)
	t.Blurred.UnselectedOption = t.Blurred.UnselectedOption.Faint(true)
	return t
}

// keymap toggles multi-select rows with space only (huh's default also binds x).
func keymap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.MultiSelect.Toggle = key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle"))
	return km
}

// run applies the shared theme and keymap so every tars form looks and behaves the same.
func run(f *huh.Form) error {
	return f.WithTheme(theme()).WithKeyMap(keymap()).Run()
}
