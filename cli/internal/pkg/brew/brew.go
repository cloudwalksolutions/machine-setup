package brew

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// Brew is a Homebrew-backed package manager adapter.
type Brew struct {
	stdout io.Writer
	stderr io.Writer
}

// New returns a Brew adapter that streams output to stdout and stderr.
func New(stdout, stderr io.Writer) *Brew {
	return &Brew{stdout: stdout, stderr: stderr}
}

func (b *Brew) Install(name string) error {
	return b.run("install", name)
}

func (b *Brew) Uninstall(name string) error {
	// HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK skips the dependent scan that
	// iterates all installed casks — avoiding failures from unrelated broken casks.
	cmd := exec.Command("brew", "uninstall", name)
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_INSTALLED_DEPENDENTS_CHECK=1")
	cmd.Stdout = b.stdout
	cmd.Stderr = b.stderr
	if err := cmd.Run(); err != nil {
		// Still verify actual state in case brew exits non-zero for other reasons.
		if installed, checkErr := b.IsInstalled(name); checkErr == nil && !installed {
			return nil
		}
		return err
	}
	return nil
}

func (b *Brew) Update(name string) error {
	return b.run("upgrade", name)
}

// IsInstalled returns true if name is present in `brew list --formula`.
// A non-zero exit code from brew means the package is not installed.
func (b *Brew) IsInstalled(name string) (bool, error) {
	out, err := exec.Command("brew", "list", "--formula", name).Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func (b *Brew) run(subcommand, name string) error {
	cmd := exec.Command("brew", subcommand, name)
	cmd.Stdout = b.stdout
	cmd.Stderr = b.stderr
	return cmd.Run()
}
