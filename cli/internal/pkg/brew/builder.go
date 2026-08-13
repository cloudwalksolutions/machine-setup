package brew

// Builder mints brew installables (Formulas and Casks) bound to a single
// Runner. It exists so callers that need to build many installables don't
// repeat the runner argument on every constructor call.
type Builder struct {
	run Runner
}

// NewBuilder returns a Builder bound to the given runner.
func NewBuilder(run Runner) Builder {
	return Builder{run: run}
}

// Formulas returns one Formula per name, in input order.
func (b Builder) Formulas(names ...string) []Formula {
	out := make([]Formula, len(names))
	for i, n := range names {
		out[i] = NewFormula(n, b.run)
	}
	return out
}

// Casks returns one Cask per name, in input order.
func (b Builder) Casks(names ...string) []Cask {
	out := make([]Cask, len(names))
	for i, n := range names {
		out[i] = NewCask(n, b.run)
	}
	return out
}
