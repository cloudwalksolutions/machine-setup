package pkg

import (
	"bytes"
	"io"
	"os/exec"
	"regexp"
)

var versionToken = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)

// PathProbe detects a tool by its binary on PATH, independent of any package manager.
type PathProbe struct {
	LookPath func(file string) (string, error)
	Run      func(cmd []string, stdout, stderr io.Writer) error
}

// Status reports the first version token `<binary> --version` prints; a binary that
// exists but fails to report a version still counts as installed.
func (p PathProbe) Status(binary string) (InstallStatus, string, error) {
	lookPath := p.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	path, err := lookPath(binary)
	if err != nil {
		return StatusNotInstalled, "", nil
	}
	run := p.Run
	if run == nil {
		run = runCommand
	}
	var out bytes.Buffer
	if err := run([]string{path, "--version"}, &out, &out); err != nil {
		return StatusUpToDate, "", nil
	}
	return StatusUpToDate, versionToken.FindString(out.String()), nil
}

func runCommand(cmd []string, stdout, stderr io.Writer) error {
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdout = stdout
	c.Stderr = stderr
	return c.Run()
}
