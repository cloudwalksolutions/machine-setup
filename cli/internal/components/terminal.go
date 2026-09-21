package components

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/cloudwalk/machine-setup/internal/fsutil"
	"github.com/cloudwalk/machine-setup/internal/paths"
)

// Terminal points iTerm2's default profile at the font shipped by Fonts.
// macOS-only; a no-op elsewhere.
type Terminal struct {
	opts Options
	p    paths.TerminalPaths
	goos string

	CurrentFontFn func() (string, error)
	SetFontFn     func(font string) error
	IsRunningFn   func() bool
}

// NewTerminal returns a Terminal component with platform defaults for the current OS.
func NewTerminal(opts Options) *Terminal {
	return NewTerminalForOS(opts, runtime.GOOS)
}

// NewTerminalForOS is the OS-explicit form, useful for tests.
func NewTerminalForOS(opts Options, goos string) *Terminal {
	return &Terminal{
		opts:          opts,
		p:             paths.ForOS(opts.RepoRoot, opts.Home, goos).Terminal,
		goos:          goos,
		CurrentFontFn: itermCurrentFont,
		SetFontFn:     itermSetFont,
		IsRunningFn:   itermIsRunning,
	}
}

// Name returns "terminal".
func (t *Terminal) Name() string { return "terminal" }

// Pull applies the repo's font to iTerm2 (idempotent).
func (t *Terminal) Pull() error {
	if t.goos != "darwin" {
		return nil
	}
	wantBytes, err := os.ReadFile(t.p.FontRepo)
	if err != nil {
		return err
	}
	want := strings.TrimSpace(string(wantBytes))

	cur, err := t.CurrentFontFn()
	if err != nil {
		return err
	}
	if cur == want {
		return nil
	}
	if t.opts.DryRun {
		fmt.Fprintf(t.opts.Stdout, "    would set terminal font  %s → %s\n", cur, want)
		return nil
	}
	if err := t.archiveFont(cur, "terminal"); err != nil {
		return err
	}
	if err := t.SetFontFn(want); err != nil {
		return err
	}
	// A running iTerm2 rewrites its prefs from memory on quit, reverting this.
	if t.IsRunningFn != nil && t.IsRunningFn() {
		fmt.Fprintln(t.opts.Stderr,
			"  terminal: iTerm2 is running — quit and reopen it for the font change to stick")
	}
	return nil
}

// Push captures the live terminal font settings back into the repo.
func (t *Terminal) Push() error {
	if t.goos != "darwin" {
		return nil
	}
	cur, err := t.CurrentFontFn()
	if err != nil {
		return err
	}
	if _, err := os.Stat(t.p.FontRepo); err == nil {
		if _, err := fsutil.Backup(t.p.FontRepo, "terminal-repo", t.opts.BackupRoot); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(t.p.FontRepo), 0o755); err != nil {
		return err
	}
	return os.WriteFile(t.p.FontRepo, []byte(cur+"\n"), 0o644)
}

// archiveFont snapshots the current font string into a versioned backup dir.
func (t *Terminal) archiveFont(value, component string) error {
	dir, err := os.MkdirTemp("", "term-archive")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	f := filepath.Join(dir, "font")
	if err := os.WriteFile(f, []byte(value), 0o644); err != nil {
		return err
	}
	_, err = fsutil.Backup(f, component, t.opts.BackupRoot)
	return err
}

// --- default (untested) system implementations ---

const itermPlist = "com.googlecode.iterm2"

func plistBuddy(args ...string) *exec.Cmd {
	home, _ := os.UserHomeDir()
	plist := filepath.Join(home, "Library", "Preferences", itermPlist+".plist")
	return exec.Command("/usr/libexec/PlistBuddy", append(args, plist)...)
}

// itermDefaultProfileIndex returns the index of the default profile in New Bookmarks.
func itermDefaultProfileIndex() (string, error) {
	guidOut, err := plistBuddy("-c", "Print :'Default Bookmark Guid'").Output()
	if err != nil {
		return "", err
	}
	guid := strings.TrimSpace(string(guidOut))
	for i := 0; ; i++ {
		idx := strconv.Itoa(i)
		out, err := plistBuddy("-c", "Print :'New Bookmarks':"+idx+":Guid").Output()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(string(out)) == guid {
			return idx, nil
		}
	}
}

func itermIsRunning() bool {
	return exec.Command("pgrep", "-x", "iTerm2").Run() == nil
}

func itermCurrentFont() (string, error) {
	idx, err := itermDefaultProfileIndex()
	if err != nil {
		return "", err
	}
	out, err := plistBuddy("-c", "Print :'New Bookmarks':"+idx+":'Normal Font'").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func itermSetFont(font string) error {
	idx, err := itermDefaultProfileIndex()
	if err != nil {
		return err
	}
	for _, key := range []string{"Normal Font", "Non Ascii Font"} {
		base := "New Bookmarks':" + idx + ":'" + key
		// Set if present, otherwise Add.
		if err := plistBuddy("-c", "Set :'"+base+"' "+font).Run(); err != nil {
			if err := plistBuddy("-c", "Add :'"+base+"' string "+font).Run(); err != nil {
				return err
			}
		}
	}
	return nil
}
