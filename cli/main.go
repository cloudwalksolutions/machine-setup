package main

import (
	"os"

	"github.com/cloudwalk/machine-setup/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
