package components

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/cloudwalk/machine-setup/internal/fsutil"
	"github.com/cloudwalk/machine-setup/internal/paths"
)

// terminalProfileName is the Terminal.app window-settings profile the CLI manages.
const terminalProfileName = "CloudWalk"

// Terminal configures the terminal emulators (iTerm2, Terminal.app) to use the
// Nerd Font shipped by the Fonts component. macOS-only; a no-op elsewhere.
//
// System access is behind function seams (like Fonts.CopyFn) so specs can run
// without touching the real prefs: CurrentFontFn/SetFontFn drive iTerm2's plist
// via PlistBuddy; ApplyFn/ExportFn drive Terminal.app via open + defaults.
type Terminal struct {
	opts Options
	p    paths.TerminalPaths
	goos string

	CurrentFontFn func() (string, error)               // iTerm2: read default profile font
	SetFontFn     func(font string) error              // iTerm2: set default profile font
	ApplyFn       func(profilePath, name string) error // Terminal.app: import + set default
	ExportFn      func(name, dst string) error         // Terminal.app: export default profile
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
		ApplyFn:       terminalAppApply,
		ExportFn:      terminalAppExport,
	}
}

// Name returns "terminal".
func (t *Terminal) Name() string { return "terminal" }

// Pull applies the repo's font to iTerm2 and Terminal.app (idempotent).
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
	if cur != want {
		if err := t.archiveFont(cur, "terminal"); err != nil {
			return err
		}
		if err := t.SetFontFn(want); err != nil {
			return err
		}
	}
	return t.ApplyFn(t.p.ProfileRepo, terminalProfileName)
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
	if err := os.WriteFile(t.p.FontRepo, []byte(cur+"\n"), 0o644); err != nil {
		return err
	}
	return t.ExportFn(terminalProfileName, t.p.ProfileRepo)
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
	return plistBuddy("-c", "Set :'New Bookmarks':"+idx+":'Use Non-ASCII Font' true").Run()
}

func terminalAppApply(profilePath, name string) error {
	if err := exec.Command("open", profilePath).Run(); err != nil {
		return err
	}
	for _, key := range []string{"Default Window Settings", "Startup Window Settings"} {
		if err := exec.Command("defaults", "write", "com.apple.Terminal", key, "-string", name).Run(); err != nil {
			return err
		}
	}
	return nil
}

func terminalAppExport(name, dst string) error {
	out, err := exec.Command("/usr/libexec/PlistBuddy", "-x", "-c",
		"Print :'Window Settings':'"+name+"'", terminalDefaultsPlist()).Output()
	if err != nil {
		return err
	}
	return os.WriteFile(dst, out, 0o644)
}

func terminalDefaultsPlist() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Preferences", "com.apple.Terminal.plist")
}
