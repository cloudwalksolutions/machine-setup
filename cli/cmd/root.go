package cmd

import (
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "machine-setup",
	Short: "CloudWalk machine setup CLI",
	Long:  "Provision and manage a CloudWalk development machine.",
}

// Execute is the single public entry point called by main.go.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "",
		"config file (default: ~/.config/.machine-setup/config.yaml)",
	)
	rootCmd.AddCommand(setupCmd)
}
