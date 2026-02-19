package cmd

import (
	"fmt"

	"github.com/cloudwalk/machine-setup/internal/config"
	"github.com/cloudwalk/machine-setup/internal/forms"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initialize this machine with CloudWalk defaults",
	Long: `Display a welcome greeting and initialize the machine-setup config
at ~/.config/.machine-setup/config.yaml with auto-detected defaults.`,
	RunE: runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	if err := forms.ShowWelcome(); err != nil {
		// huh.ErrUserAborted means the user pressed Ctrl+C — exit cleanly.
		if err.Error() == "user aborted" {
			return nil
		}
		return fmt.Errorf("welcome form: %w", err)
	}

	cfgPath := config.DefaultConfigPath()
	if cfgFile != "" {
		cfgPath = cfgFile
	}

	cfg, err := config.Init(cfgPath)
	if err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Config written to %s\n", cfgPath)
	fmt.Fprintf(cmd.OutOrStdout(), "Detected architecture: %s\n", cfg.Architecture)
	return nil
}
