package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Config is the top-level machine-setup configuration.
type Config struct {
	Architecture string       `mapstructure:"architecture" yaml:"architecture"`
	Sources      []string     `mapstructure:"sources"      yaml:"sources"`
	Packages     []Package    `mapstructure:"packages"     yaml:"packages"`
	Apps         []App        `mapstructure:"apps"         yaml:"apps"`
	Claude       ClaudeConfig `mapstructure:"claude"       yaml:"claude"`
	Pi           PiConfig     `mapstructure:"pi"           yaml:"pi"`
}

// PiConfig records the `tars pi init` choices; providers are per-machine and rendered
// into ~/.pi/agent/models.json, never stored in the repo.
type PiConfig struct {
	Packages        []string     `mapstructure:"packages"         yaml:"packages"`
	Providers       []PiProvider `mapstructure:"providers"        yaml:"providers"`
	DefaultProvider string       `mapstructure:"default_provider" yaml:"default_provider"`
	DefaultModel    string       `mapstructure:"default_model"    yaml:"default_model"`
}

// KeyEnvFor derives the env var holding a provider's API key: remote-llama → REMOTE_LLAMA_API_KEY.
func KeyEnvFor(provider string) string {
	return strings.ToUpper(strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(provider)) + "_API_KEY"
}

// PiProvider is one model provider: kind ollama, llamacpp, or openai (any OpenAI-compatible endpoint).
type PiProvider struct {
	Name    string   `mapstructure:"name"     yaml:"name"`
	Kind    string   `mapstructure:"kind"     yaml:"kind"`
	BaseURL string   `mapstructure:"base_url" yaml:"base_url"`
	KeyEnv  string   `mapstructure:"key_env"  yaml:"key_env"`
	Models  []string `mapstructure:"models"   yaml:"models"`
}

// ClaudeConfig records the `tars claude init` choices; nil/empty means "all on".
type ClaudeConfig struct {
	Hook     *bool    `mapstructure:"hook"     yaml:"hook"`
	Settings *bool    `mapstructure:"settings" yaml:"settings"`
	Rules    []string `mapstructure:"rules"    yaml:"rules"`
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

// DefaultConfigPath is ~/.config/tars/config.yaml unless TARS_CONFIG_PATH overrides it.
func DefaultConfigPath() string {
	if envPath := os.Getenv("TARS_CONFIG_PATH"); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "tars/config.yaml"
	}
	return filepath.Join(home, ".config", "tars", "config.yaml")
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

// Save writes cfg back to path, preserving any unrecognized keys already in the file.
func Save(path string, cfg *Config) error {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if _, err := os.Stat(path); err == nil {
		_ = v.ReadInConfig()
	}
	v.Set("architecture", cfg.Architecture)
	v.Set("sources", cfg.Sources)
	v.Set("packages", cfg.Packages)
	v.Set("apps", cfg.Apps)
	v.Set("claude", cfg.Claude)
	v.Set("pi", cfg.Pi)
	return v.WriteConfigAs(path)
}
