package sessions

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Runner runs byobu with the given args. The real impl wires the process's
// std streams so attach gets a tty; fakes record argv in specs.
type Runner func(args ...string) error

// DefaultRunner returns the production Runner that shells out to `byobu`
// (which forwards args to tmux), wiring std streams so attach gets a tty.
func DefaultRunner() Runner {
	return func(args ...string) error {
		cmd := exec.Command("byobu", args...)
		if len(args) > 0 && args[0] == "has-session" {
			return cmd.Run() // existence probe: silence "can't find session" noise
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

// Launcher drives byobu to create/attach configured sessions.
type Launcher struct {
	Run       Runner
	LookupEnv func(key string) (string, bool)
	Home      string
	Stdout    io.Writer
	Stderr    io.Writer
}

// Fresh opens a new byobu session named name (rooted at home), or an
// unnamed attached one when name is empty.
func (l Launcher) Fresh(name string) error {
	if name != "" {
		return l.Open(Session{Name: name, Dirs: []string{l.Home}})
	}
	if _, inside := l.LookupEnv("TMUX"); inside {
		return errors.New("already inside tmux/byobu — pass a session name to switch to a fresh one")
	}
	return l.Run("new-session")
}

// Open ensures the session exists, then attaches (or switches when inside tmux).
func (l Launcher) Open(s Session) error {
	if err := l.Ensure(s); err != nil {
		return err
	}
	return l.attach(s.Name)
}

// OpenAll ensures every configured session (tolerating per-session failures),
// then attaches to the first one that succeeded.
func (l Launcher) OpenAll(f File) error {
	first := ""
	for _, s := range f.Sessions {
		fmt.Fprintf(l.Stdout, "  → %s\n", s.Name)
		if err := l.Ensure(s); err != nil {
			fmt.Fprintf(l.Stderr, "  %s: %v\n", s.Name, err)
			continue
		}
		if first == "" {
			first = s.Name
		}
	}
	if first == "" {
		return errors.New("no sessions could be opened")
	}
	return l.attach(first)
}

// attach picks switch-client inside tmux; attaching there would nest sessions.
func (l Launcher) attach(name string) error {
	if _, inside := l.LookupEnv("TMUX"); inside {
		return l.Run("switch-client", "-t", "="+name)
	}
	return l.Run("attach-session", "-t", "="+name)
}

// Ensure creates the session detached (one window per dir) if it is missing.
func (l Launcher) Ensure(s Session) error {
	if err := l.Run("has-session", "-t", "="+s.Name); err == nil {
		return nil
	}
	first := s.Dirs[0]
	if err := l.Run("new-session", "-d", "-s", s.Name, "-n", WindowName(first), "-c", first); err != nil {
		return err
	}
	for _, dir := range s.Dirs[1:] {
		if err := l.Run("new-window", "-t", "="+s.Name+":", "-n", WindowName(dir), "-c", dir); err != nil {
			return err
		}
	}
	return nil
}
