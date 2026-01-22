// Package config provides configuration management for aix using Viper.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/thoreinstein/aix/internal/paths"
)

// AppName is the application name used for config file naming.
const AppName = "aix"

// Config represents the top-level configuration structure.
type Config struct {
	Version          int                         `mapstructure:"version" yaml:"version"`
	DefaultPlatforms []string                    `mapstructure:"default_platforms" yaml:"default_platforms"`
	Platforms        map[string]PlatformOverride `mapstructure:"platforms" yaml:"platforms"`
}

// PlatformOverride contains configuration overrides for a specific platform.
type PlatformOverride struct {
	ConfigDir string `mapstructure:"config_dir" yaml:"config_dir"`
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version: %d", c.Version)
	}

	for _, p := range c.DefaultPlatforms {
		if !paths.ValidPlatform(p) {
			return fmt.Errorf("invalid default platform: %s", p)
		}
	}

	for p := range c.Platforms {
		if !paths.ValidPlatform(p) {
			return fmt.Errorf("invalid platform override key: %s", p)
		}
	}

	return nil
}

// Init initializes Viper with default configuration.
// Call this once at application startup before accessing config values.
func Init() {
	// Config file settings
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Search paths (in order of precedence)
	if envDir := os.Getenv("AIX_CONFIG_DIR"); envDir != "" {
		viper.AddConfigPath(envDir)
	} else {
		viper.AddConfigPath(".") // Current directory
		viper.AddConfigPath(filepath.Join(paths.ConfigHome(), AppName))
	}

	// Environment variable support
	viper.SetEnvPrefix("AIX")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("version", 1)
	viper.SetDefault("default_platforms", paths.Platforms())
}

// Load reads the configuration file.
// If path is provided, it reads from that specific file.
// If path is empty, it searches in the default locations.
// Returns the loaded configuration or default values if no file is found (when path is empty).
func Load(path string) (*Config, error) {
	if path != "" {
		viper.SetConfigFile(path)
	}

	if err := viper.ReadInConfig(); err != nil {
		// If config file not found...
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// If user specified a path, this is an error
			if path != "" {
				return nil, fmt.Errorf("config file not found at %s: %w", path, err)
			}
			// Otherwise (implicit load), it's fine to use defaults
		} else {
			// Real read error (parsing, permissions, etc)
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}
