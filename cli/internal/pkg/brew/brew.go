package brew

import (
	"io"
	"os/exec"
)

// DefaultRunner returns the production Runner that shells out to `brew`.
// It streams stdout/stderr to the given writers.
func DefaultRunner() Runner {
	return func(args []string, stdout, stderr io.Writer) error {
		cmd := exec.Command("brew", args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		return cmd.Run()
	}
}
