package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tars",
	Short: "tars — CloudWalk dev-machine setup CLI",
	Long:  "Provision and manage a CloudWalk development machine (dotfiles, tools, fonts, terminals).",
}

// Execute is the single public entry point called by main.go.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion wires the build metadata (injected via ldflags in main) into the
// root command so `tars --version` reports it.
func SetVersion(version, commit, date string) {
	if version == "" {
		version = "dev"
	}
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "",
		"config file (default: ~/.config/.machine-setup/config.yaml)",
	)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
}
