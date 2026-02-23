package cmd

import (
	"fmt"

	"github.com/cloudwalk/machine-setup/internal/config"
	"github.com/cloudwalk/machine-setup/internal/forms"
	"github.com/cloudwalk/machine-setup/internal/pkg"
	"github.com/spf13/cobra"
)

// newManagerFn is the factory for the package manager. Tests override this to
// inject a ManagerSpy without touching the real package manager.
var newManagerFn = pkg.NewManager

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initialize this machine with CloudWalk defaults",
	Long: `Display a welcome greeting, select dev tools to install, initialize
the machine-setup config, and install selected packages.`,
	RunE: runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	// 1. Welcome form.
	if err := forms.ShowWelcome(); err != nil {
		if err.Error() == "user aborted" {
			return nil
		}
		return fmt.Errorf("welcome form: %w", err)
	}

	// 2. Resolve config path.
	cfgPath := config.DefaultConfigPath()
	if cfgFile != "" {
		cfgPath = cfgFile
	}

	// 3. Init config (creates file with defaults if missing).
	cfg, err := config.Init(cfgPath)
	if err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	// 4. Package selection form (MACHINE_SETUP_NO_FORM=1 returns all defaults).
	selected, err := forms.ShowInstallForm(pkg.DevToolNames())
	if err != nil && err.Error() != "user aborted" {
		return fmt.Errorf("install form: %w", err)
	}

	// 5. Persist selected packages to config.
	cfg.Packages = pkg.NamesToPackages(selected)
	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Config written to %s\n", cfgPath)
	fmt.Fprintf(cmd.OutOrStdout(), "Detected architecture: %s\n", cfg.Architecture)

	// 6. Install selected packages.
	manager, err := newManagerFn(cmd.OutOrStdout(), cmd.ErrOrStderr())
	if err != nil {
		return fmt.Errorf("detecting package manager: %w", err)
	}

	for _, name := range selected {
		fmt.Fprintf(cmd.OutOrStdout(), "Installing %s...\n", name)
		if installErr := manager.Install(name); installErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  error: %s: %v\n", name, installErr)
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nSetup complete.\n")
	return nil
}

