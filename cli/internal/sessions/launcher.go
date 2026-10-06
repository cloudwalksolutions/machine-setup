package sessions

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

// Output runs byobu with the given args and returns its stdout; fakes return canned text.
type Output func(args ...string) (string, error)

// DefaultOutput returns the production Output that shells out to `byobu`.
func DefaultOutput() Output {
	return func(args ...string) (string, error) {
		out, err := exec.Command("byobu", args...).Output()
		var exit *exec.ExitError
		if errors.As(err, &exit) && len(exit.Stderr) > 0 {
			return string(out), errors.New(strings.TrimSpace(string(exit.Stderr)))
		}
		return string(out), err
	}
}

// Live is a running byobu session.
type Live struct {
	Name     string
	Windows  int
	Attached bool
}

// Window is one window of a new session, started in Dir and running Command if set.
type Window struct {
	Name    string `yaml:"name"`
	Dir     string `yaml:"dir,omitempty"`
	Command string `yaml:"command,omitempty"`
}

// Launcher drives byobu to create/attach configured sessions.
type Launcher struct {
	Run       Runner
	Output    Output
	LookupEnv func(key string) (string, bool)
}

// List returns the running sessions.
func (l Launcher) List() ([]Live, error) {
	out, err := l.Output("list-sessions", "-F", "#{session_name}\t#{session_windows}\t#{session_attached}")
	if err != nil && strings.Contains(err.Error(), "no server running") {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var live []Live
	for line := range strings.Lines(out) {
		fields := strings.Split(strings.TrimSpace(line), "\t")
		if len(fields) < 3 {
			continue
		}
		windows, _ := strconv.Atoi(fields[1])
		live = append(live, Live{Name: fields[0], Windows: windows, Attached: fields[2] != "0"})
	}
	return live, nil
}

// Kill ends a running session.
func (l Launcher) Kill(name string) error {
	return l.Run("kill-session", "-t", "="+name)
}

// Rename gives a running session a new name.
func (l Launcher) Rename(old, name string) error {
	if strings.ContainsAny(name, ".:") {
		return fmt.Errorf("invalid session name %q: '.' and ':' are not allowed", name)
	}
	return l.Run("rename-session", "-t", "="+old, name)
}

// Open ensures the session exists, then attaches (or switches when inside tmux).
func (l Launcher) Open(name string, windows []Window) error {
	if err := l.Ensure(name, windows); err != nil {
		return err
	}
	return l.Attach(name)
}

// Attach picks switch-client inside tmux; attaching there would nest sessions.
func (l Launcher) Attach(name string) error {
	if _, inside := l.LookupEnv("TMUX"); inside {
		return l.Run("switch-client", "-t", "="+name)
	}
	return l.Run("attach-session", "-t", "="+name)
}

// Ensure creates the session detached, one window per template window, if it is missing.
func (l Launcher) Ensure(name string, windows []Window) error {
	if err := l.Run("has-session", "-t", "="+name); err == nil {
		return nil
	}
	first := windows[0]
	if err := l.Run("new-session", "-d", "-s", name, "-n", first.Name, "-c", first.Dir); err != nil {
		return err
	}
	for _, w := range windows[1:] {
		if err := l.Run("new-window", "-t", "="+name+":", "-n", w.Name, "-c", w.Dir); err != nil {
			return err
		}
	}
	for _, w := range windows {
		if w.Command == "" {
			continue
		}
		if err := l.Run("send-keys", "-t", "="+name+":"+w.Name, w.Command, "Enter"); err != nil {
			return err
		}
	}
	return nil
}
