package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"tars/internal/forms"
)

// formHost embeds one huh form in a model: styled, sized to the terminal, fed every message.
type formHost struct {
	form *huh.Form
}

// openForm styles and starts f, fitting it to size when the terminal size is already known.
func openForm(f *huh.Form, size *tea.WindowSizeMsg) (formHost, tea.Cmd) {
	h := formHost{form: forms.Styled(f)}
	cmd := h.form.Init()
	if size != nil {
		s := *size
		s.Height-- // huh leaves the blank line above its help footer out of the group height
		next, sizeCmd := h.form.Update(s)
		h.form, cmd = next.(*huh.Form), tea.Batch(cmd, sizeCmd)
	}
	return h, cmd
}

// update forwards msg to the form and reports where it stands.
func (h formHost) update(msg tea.Msg) (formHost, tea.Cmd, huh.FormState) {
	next, cmd := h.form.Update(msg)
	h.form = next.(*huh.Form)
	return h, cmd, h.form.State
}

func (h formHost) view() string { return h.form.View() }
