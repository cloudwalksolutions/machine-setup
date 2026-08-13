package main

import (
	"os"

	"github.com/cloudwalk/machine-setup/cmd"
)

// Build metadata, injected by GoReleaser via -ldflags -X main.version=... etc.
var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
