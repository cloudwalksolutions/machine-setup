package fsutil

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aymanbagabas/go-udiff"
)

// Kind is what a Change does to its path.
type Kind int

const (
	Create Kind = iota
	Overwrite
	Remove
)

// Change is one write a pull would make, or made.
type Change struct {
	Path   string
	Kind   Kind
	Backup string
	Binary bool
	Diff   udiff.UnifiedDiff
}

func (c Change) String() string {
	switch c.Kind {
	case Overwrite:
		if c.Binary {
			return fmt.Sprintf("    would overwrite  %s (backup %s, binary)", c.Path, c.Backup)
		}
		return fmt.Sprintf("    would overwrite  %s (backup %s)", c.Path, c.Backup)
	case Remove:
		return "    would remove  " + c.Path
	default:
		return "    would create  " + c.Path
	}
}

// classify compares src against dst and returns the Change that would make
// dst match src; changed is false when they already match.
func classify(src, dst string) (ch Change, changed bool, err error) {
	want, err := os.ReadFile(src)
	if os.IsNotExist(err) {
		return Change{Path: dst, Kind: Remove}, true, nil
	} else if err != nil {
		return Change{}, false, err
	}
	have, err := os.ReadFile(dst)
	if os.IsNotExist(err) {
		return Change{Path: dst, Kind: Create}, true, nil
	} else if err != nil {
		return Change{}, false, err
	}
	if bytes.Equal(want, have) {
		return Change{}, false, nil
	}
	ch = Change{Path: dst, Kind: Overwrite}
	if bytes.IndexByte(want, 0) >= 0 || bytes.IndexByte(have, 0) >= 0 {
		ch.Binary = true
		return ch, true, nil
	}
	ch.Diff, err = udiff.ToUnifiedDiff(dst, dst, string(have), udiff.Lines(string(have), string(want)), udiff.DefaultContextLines)
	return ch, true, err
}
