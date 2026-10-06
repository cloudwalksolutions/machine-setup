package forms

import (
	"errors"

	"charm.land/huh/v2"

	"tars/internal/projects"
)

// TemplateForm edits what a new session for project looks like, one `name | dir | command` window per line.
func TemplateForm(project string, t projects.Template) (*huh.Form, func() projects.Template) {
	text := t.String()
	f := huh.NewForm(huh.NewGroup(
		huh.NewText().
			Title("Session template for " + project).
			Description("One window per line: name | dir (relative to the project) | command").
			Validate(func(s string) error {
				_, err := projects.ParseTemplate(s)
				return err
			}).
			Value(&text),
	))
	return f, func() projects.Template {
		parsed, _ := projects.ParseTemplate(text) // the field only completes when the text parses
		return parsed
	}
}

// KillForm confirms killing a running session; it defaults to keeping it.
func KillForm(name string) (*huh.Form, func() bool) {
	kill := false
	f := huh.NewForm(huh.NewGroup(
		confirm("Kill session "+name+"?", "Its windows and whatever runs in them end.", &kill),
	))
	return f, func() bool { return kill }
}

// RenameForm asks for a running session's new name.
func RenameForm(current string) (*huh.Form, func() string) {
	name := current
	f := huh.NewForm(huh.NewGroup(
		huh.NewInput().Title("Rename session " + current).Value(&name),
	))
	return f, func() string { return name }
}

// ShowSessionPicker displays a single-select over the given session options.
// When TARS_NO_FORM=1 it returns the first option (tests/CI).
func ShowSessionPicker(options []string) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no sessions to pick from")
	}
	if headless() {
		return options[0], nil
	}

	selected := options[0]
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}

	err := run(huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Open a byobu session").
				Options(opts...).
				Value(&selected),
		),
	))

	return selected, err
}
