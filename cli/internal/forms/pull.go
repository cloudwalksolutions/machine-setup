package forms

import (
	"fmt"
	"io"

	"charm.land/huh/v2"
)

// PullConfirm prints a component's pending diff to Out, then asks whether to apply it; Enter applies.
type PullConfirm struct{ Out io.Writer }

func (p PullConfirm) ConfirmApply(name, diff string) (bool, error) {
	fmt.Fprint(p.Out, diff)
	apply := true
	err := run(huh.NewForm(huh.NewGroup(
		confirm("Apply "+name+"?", "Enter applies the diff above; choose No to skip "+name+".", &apply),
	)))
	return apply, err
}
