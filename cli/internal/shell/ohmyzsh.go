// Package shell wraps shell-out installers (oh-my-zsh, etc.) behind
// injectable runners so the orchestrator can drive them and tests can spy.
package shell

import (
	"io"
	"os"
	"os/exec"
)

// installerScript is the official oh-my-zsh install line. Piped through sh.
const installerScript = `sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"`

// DefaultRunner returns the production runner, which invokes the official
// installer with RUNZSH=no KEEP_ZSHRC=yes CHSH=no so it does not overwrite
// ~/.zshrc (the zsh component does that next) and does not chsh.
func DefaultRunner() func(stdout, stderr io.Writer) error {
	return func(stdout, stderr io.Writer) error {
		cmd := exec.Command("sh", "-c", installerScript)
		cmd.Env = append(os.Environ(), "RUNZSH=no", "KEEP_ZSHRC=yes", "CHSH=no")
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
}

// OhMyZshInstaller installs oh-my-zsh non-interactively.
type OhMyZshInstaller struct {
	// Dir is the path that signals "already installed" (typically ~/.oh-my-zsh).
	Dir string
	// Runner is the side-effect; tests replace it.
	Runner func(stdout, stderr io.Writer) error
	Stdout io.Writer
	Stderr io.Writer
}

// Install runs the installer if Dir does not exist; otherwise no-ops.
func (i OhMyZshInstaller) Install() error {
	if _, err := os.Stat(i.Dir); err == nil {
		return nil
	}
	return i.Runner(i.Stdout, i.Stderr)
}
