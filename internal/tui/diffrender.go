package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/aymanbagabas/go-udiff"
)

type diffStyles struct {
	file, hunk, add, del, addEm, delEm, ctx, noEOL lipgloss.Style
}

func newDiffStyles(dark bool) diffStyles {
	ld := lipgloss.LightDark(dark)
	add := lipgloss.NewStyle().Foreground(ld(lipgloss.Color("#1F6B3A"), lipgloss.Color("#9ECE6A"))).Background(ld(lipgloss.Color("#D6F5D6"), lipgloss.Color("#1C2F1C")))
	del := lipgloss.NewStyle().Foreground(ld(lipgloss.Color("#9B2240"), red)).Background(ld(lipgloss.Color("#FBE0E6"), lipgloss.Color("#3A1C25")))
	return diffStyles{
		file:  lipgloss.NewStyle().Bold(true).Foreground(accent),
		hunk:  lipgloss.NewStyle().Bold(true).Foreground(ld(lipgloss.Color("#3B5BA9"), accent)),
		add:   add,
		del:   del,
		addEm: add.Bold(true).Background(ld(lipgloss.Color("#A8E6A8"), lipgloss.Color("#2F5A2F"))),
		delEm: del.Bold(true).Background(ld(lipgloss.Color("#F5B7C4"), lipgloss.Color("#6B2A3C"))),
		ctx:   lipgloss.NewStyle().Faint(true),
		noEOL: lipgloss.NewStyle().Faint(true),
	}
}

// RenderDiff colors a unified diff one row per line, filling each row to width
// and emphasizing the words that changed within a removed/added pair, and
// returns the rows plus the index of every hunk header.
func RenderDiff(d udiff.UnifiedDiff, width int, dark bool) (rows []string, hunks []int) {
	s := newDiffStyles(dark)
	rows = append(rows, s.file.Render("--- "+d.From), s.file.Render("+++ "+d.To))
	for _, h := range d.Hunks {
		hunks = append(hunks, len(rows))
		rows = append(rows, s.hunk.Render(hunkHeader(h)))
		var dels, ins []udiff.Line
		flush := func() {
			rows = append(rows, s.pairRows(dels, ins, width)...)
			dels, ins = nil, nil
		}
		for _, l := range h.Lines {
			switch l.Kind {
			case udiff.Delete:
				if len(ins) > 0 {
					flush()
				}
				dels = append(dels, l)
			case udiff.Insert:
				ins = append(ins, l)
			default:
				flush()
				rows = append(rows, s.row(s.ctx, s.ctx, " ", []segment{{text: strings.TrimSuffix(l.Content, "\n")}}, width))
				rows = append(rows, s.eol(l)...)
			}
		}
		flush()
	}
	return rows, hunks
}

// pairRows renders a run of removed lines then the added lines that follow,
// emphasizing what changed between the k-th removed and k-th added line.
func (s diffStyles) pairRows(dels, ins []udiff.Line, width int) []string {
	oldSegs := make([][]segment, len(dels))
	newSegs := make([][]segment, len(ins))
	for i := range dels {
		oldSegs[i] = []segment{{text: strings.TrimSuffix(dels[i].Content, "\n")}}
	}
	for i := range ins {
		newSegs[i] = []segment{{text: strings.TrimSuffix(ins[i].Content, "\n")}}
	}
	for i := 0; i < len(dels) && i < len(ins); i++ {
		if o, n, ok := emphasize(oldSegs[i][0].text, newSegs[i][0].text); ok {
			oldSegs[i], newSegs[i] = o, n
		}
	}
	var rows []string
	for i, l := range dels {
		rows = append(rows, s.row(s.del, s.delEm, "-", oldSegs[i], width))
		rows = append(rows, s.eol(l)...)
	}
	for i, l := range ins {
		rows = append(rows, s.row(s.add, s.addEm, "+", newSegs[i], width))
		rows = append(rows, s.eol(l)...)
	}
	return rows
}

func (s diffStyles) eol(l udiff.Line) []string {
	if strings.HasSuffix(l.Content, "\n") {
		return nil
	}
	return []string{s.noEOL.Render(`\ No newline at end of file`)}
}

type segment struct {
	text string
	em   bool
}

// emphasize splits old and new into plain and changed segments, widened to
// whole words; ok is false when most of the line changed, where emphasis
// would only add noise.
func emphasize(old, new string) (oldSegs, newSegs []segment, ok bool) {
	var oldRanges, newRanges [][2]int
	changed, delta := 0, 0
	for _, e := range udiff.Strings(old, new) {
		changed += e.End - e.Start + len(e.New)
		oldRanges = append(oldRanges, [2]int{e.Start, e.End})
		newRanges = append(newRanges, [2]int{e.Start + delta, e.Start + delta + len(e.New)})
		delta += len(e.New) - (e.End - e.Start)
	}
	if changed*10 > 6*max(len(old), len(new)) {
		return nil, nil, false
	}
	return segments(old, oldRanges), segments(new, newRanges), true
}

// segments cuts s into plain and emphasized segments, each range widened to word boundaries.
func segments(s string, ranges [][2]int) []segment {
	var segs []segment
	last := 0
	for _, r := range ranges {
		start, end := r[0], r[1]
		for start > last && start < len(s) && s[start-1] != ' ' {
			start--
		}
		for end < len(s) && s[end] != ' ' {
			end++
		}
		if start < last {
			start = last
		}
		if end <= start {
			continue
		}
		segs = append(segs, segment{text: s[last:start]}, segment{text: s[start:end], em: true})
		last = end
	}
	return append(segs, segment{text: s[last:]})
}

// row renders prefix, segs and the padding to width, coalescing neighbours of
// the same style so a plain row is one styled run.
func (s diffStyles) row(base, em lipgloss.Style, prefix string, segs []segment, width int) string {
	all := append([]segment{{text: prefix}}, segs...)
	cells := 0
	for _, seg := range all {
		cells += lipgloss.Width(seg.text)
	}
	if pad := width - cells; pad > 0 {
		all = append(all, segment{text: strings.Repeat(" ", pad)})
	}
	var b strings.Builder
	for i := 0; i < len(all); {
		run := all[i]
		for i++; i < len(all) && all[i].em == run.em; i++ {
			run.text += all[i].text
		}
		if run.text == "" {
			continue
		}
		if run.em {
			b.WriteString(em.Render(run.text))
		} else {
			b.WriteString(base.Render(run.text))
		}
	}
	return b.String()
}

func hunkHeader(h *udiff.Hunk) string {
	from, to := 0, 0
	for _, l := range h.Lines {
		switch l.Kind {
		case udiff.Delete:
			from++
		case udiff.Insert:
			to++
		default:
			from++
			to++
		}
	}
	return fmt.Sprintf("@@ %s %s @@", span("-", h.FromLine, from), span("+", h.ToLine, to))
}

func span(sign string, line, count int) string {
	switch {
	case count > 1:
		return fmt.Sprintf("%s%d,%d", sign, line, count)
	case line == 1 && count == 0:
		return sign + "0,0"
	default:
		return fmt.Sprintf("%s%d", sign, line)
	}
}
