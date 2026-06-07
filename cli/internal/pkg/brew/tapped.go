package brew

import "io"

// TappedFormula is a brew formula whose source lives in a non-core tap (e.g.
// `hashicorp/tap/terraform`). Install runs `brew tap <tap>` first, then
// `brew install <tap>/<name>` so the tap is fetched on first use and the
// install resolves to the qualified formula.
type TappedFormula struct {
	name string
	tap  string
	run  Runner
}

// NewTappedFormula binds a name + its tap to a runner.
func NewTappedFormula(name, tap string, run Runner) TappedFormula {
	return TappedFormula{name: name, tap: tap, run: run}
}

// Name returns the formula's unqualified name (what the user sees).
func (t TappedFormula) Name() string { return t.name }

// Install taps then installs.
func (t TappedFormula) Install(stdout, stderr io.Writer) error {
	if err := t.run([]string{"tap", t.tap}, stdout, stderr); err != nil {
		return err
	}
	return t.run([]string{"install", t.tap + "/" + t.name}, stdout, stderr)
}
