package shell

import (
	"io"
	"os"
	"os/exec"
)

// p10kRepo is the canonical Powerlevel10k git remote.
const p10kRepo = "https://github.com/romkatv/powerlevel10k.git"

// Powerlevel10kInstaller installs the Powerlevel10k oh-my-zsh theme by
// cloning its git repo into the oh-my-zsh custom themes directory.
type Powerlevel10kInstaller struct {
	// Dir is the destination clone path (typically ~/.oh-my-zsh/custom/themes/powerlevel10k).
	Dir string
	// Runner is the side-effect; tests replace it.
	Runner func(stdout, stderr io.Writer) error
	Stdout io.Writer
	Stderr io.Writer
}

// Install clones the repo if Dir does not exist; otherwise no-ops.
func (i Powerlevel10kInstaller) Install() error {
	if _, err := os.Stat(i.Dir); err == nil {
		return nil
	}
	return i.Runner(i.Stdout, i.Stderr)
}

// DefaultP10kRunner returns the production Runner: `git clone --depth=1` of
// the Powerlevel10k repo into the configured Dir. The Dir is closed over from
// the installer at construction time via the wrapper in cmd/setup.go.
func DefaultP10kRunner(dir string) func(stdout, stderr io.Writer) error {
	return func(stdout, stderr io.Writer) error {
		cmd := exec.Command("git", "clone", "--depth=1", p10kRepo, dir)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
}
