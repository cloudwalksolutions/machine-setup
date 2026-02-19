package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// Config is the top-level machine-setup configuration.
type Config struct {
	Architecture string    `mapstructure:"architecture" yaml:"architecture"`
	Sources      []string  `mapstructure:"sources"      yaml:"sources"`
	Packages     []Package `mapstructure:"packages"     yaml:"packages"`
	Apps         []App     `mapstructure:"apps"         yaml:"apps"`
}

// Package represents a managed package abstracted over package managers.
type Package struct {
	Name    string `mapstructure:"name"    yaml:"name"`
	Manager string `mapstructure:"manager" yaml:"manager"` // "brew" | "apt"
}

// App represents a desktop application to track.
type App struct {
	Name string `mapstructure:"name" yaml:"name"`
}

// DefaultConfigPath returns ~/.config/.machine-setup/config.yaml.
// The MACHINE_SETUP_CONFIG_PATH env var overrides this (used by tests).
func DefaultConfigPath() string {
	if envPath := os.Getenv("MACHINE_SETUP_CONFIG_PATH"); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".machine-setup/config.yaml"
	}
	return filepath.Join(home, ".config", ".machine-setup", "config.yaml")
}

// Init writes defaults to path if the file does not exist, or loads and
// re-writes an existing config preserving all user-set values. Idempotent.
func Init(path string) (*Config, error) {
	v := viper.New()

	v.SetDefault("architecture", runtime.GOARCH)
	v.SetDefault("sources", []string{})
	v.SetDefault("packages", []map[string]string{})
	v.SetDefault("apps", []map[string]string{})

	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Only read the existing file if it exists; SetConfigFile+ReadInConfig
	// returns an *os.PathError (not ConfigFileNotFoundError) when missing.
	if _, err := os.Stat(path); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	if err := v.WriteConfigAs(path); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
