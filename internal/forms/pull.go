package forms

import (
	"fmt"
	"io"

	"charm.land/huh/v2"
)

// PullConfirm shows a component's diff and asks whether to apply it.
type PullConfirm struct{ Out io.Writer }

func (p PullConfirm) ConfirmApply(name, diff string) (bool, error) {
	fmt.Fprint(p.Out, diff)
	apply := true
	err := run(huh.NewForm(huh.NewGroup(
		confirm("Apply "+name+"?", "Enter applies the diff above; choose No to skip "+name+".", &apply),
	)))
	return apply, err
}
