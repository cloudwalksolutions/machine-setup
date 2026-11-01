package components

import (
	"fmt"
	"io"
	"strings"

	"tars/internal/fsutil"
)

// Plan is what pulling one component would change.
type Plan struct {
	Component string
	Changes   []fsutil.Change
}

// Summary counts the changes by kind, e.g. "1 create, 2 modify".
func (p Plan) Summary() string {
	return describe(count(p.Changes), [...]string{"create", "modify", "remove", "setting"})
}

// PlanLine totals the plans before anything is written.
func PlanLine(plans []Plan) string {
	var all []fsutil.Change
	for _, p := range plans {
		all = append(all, p.Changes...)
	}
	return fmt.Sprintf("Plan: %s · %d components", describe(count(all), [...]string{"to create", "to modify", "to remove", "to set"}), len(plans))
}

// Recap totals what the plans wrote, plus the components that had nothing to do.
func Recap(plans []Plan, unchanged int) string {
	var all []fsutil.Change
	for _, p := range plans {
		all = append(all, p.Changes...)
	}
	done := "nothing to change"
	if len(all) > 0 {
		done = describe(count(all), [...]string{"created", "overwritten", "removed", "set"})
	}
	return fmt.Sprintf("Pull complete: %s, %d components unchanged.", done, unchanged)
}

// WritePlans prints the plans as a pipeable report ending in the Plan line.
func WritePlans(w io.Writer, plans []Plan) {
	for _, p := range plans {
		fmt.Fprintf(w, "== %s (%s)\n", p.Component, p.Summary())
		for _, ch := range p.Changes {
			fmt.Fprintln(w, ch)
			if len(ch.Diff.Hunks) > 0 {
				fmt.Fprint(w, ch.Diff.String())
			}
		}
	}
	fmt.Fprintf(w, "\n%s\n", PlanLine(plans))
}

func count(changes []fsutil.Change) [4]int {
	var counts [4]int
	for _, ch := range changes {
		counts[ch.Kind]++
	}
	return counts
}

func describe(counts [4]int, labels [4]string) string {
	var parts []string
	for kind, label := range labels {
		if counts[kind] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", counts[kind], label))
		}
	}
	return strings.Join(parts, ", ")
}
