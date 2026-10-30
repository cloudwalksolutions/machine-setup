package report

import "strings"

// LineWriter turns a child process's stream into whole lines for a renderer;
// carriage returns (progress bars) count as line breaks. Skipped empty lines
// keep the live pane dense.
type LineWriter struct {
	Emit func(line string)
	buf  strings.Builder
}

// Write buffers partial lines and emits each completed one.
func (w *LineWriter) Write(p []byte) (int, error) {
	for _, b := range p {
		if b == '\n' || b == '\r' {
			w.flush()
			continue
		}
		w.buf.WriteByte(b)
	}
	return len(p), nil
}

// Close emits whatever is left without a trailing newline.
func (w *LineWriter) Close() error {
	w.flush()
	return nil
}

func (w *LineWriter) flush() {
	if w.buf.Len() == 0 {
		return
	}
	w.Emit(w.buf.String())
	w.buf.Reset()
}
