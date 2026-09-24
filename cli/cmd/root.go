package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "tars",
	Short: "tars — dev-machine setup CLI",
	Long:  "Provision and manage a development machine (dotfiles, tools, fonts, terminals), open byobu sessions from a simple config, and switch git/GitHub/SSH identity per project dir.",
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
	binaryVersion = version
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "",
		"config file (default: ~/.config/tars/config.yaml; env: TARS_CONFIG_PATH)",
	)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(profilesCmd)
	pullCmd.Flags().BoolVar(&pullDryRun, "dry-run", false,
		"report the writes a pull would make, without changing anything")
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(claudeCmd)
	rootCmd.AddCommand(piCmd)
}
