// Package config provides configuration management for aix using Viper.
package config

import (
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/spf13/viper"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/paths"
)

// AppName is the application name used for config file naming.
const AppName = "aix"

// repoNamePattern validates repository names.
// Names must be lowercase alphanumeric with hyphens, starting with a letter.
var repoNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Config represents the top-level configuration structure.
type Config struct {
	Version          int                         `mapstructure:"version" yaml:"version"`
	DefaultPlatforms []string                    `mapstructure:"default_platforms" yaml:"default_platforms"`
	Platforms        map[string]PlatformOverride `mapstructure:"platforms" yaml:"platforms"`
	Repos            map[string]RepoConfig       `mapstructure:"repos" yaml:"repos"`
}

// PlatformOverride contains configuration overrides for a specific platform.
type PlatformOverride struct {
	ConfigDir string `mapstructure:"config_dir" yaml:"config_dir"`
}

// RepoConfig contains configuration for a skill repository.
type RepoConfig struct {
	URL     string    `mapstructure:"url" yaml:"url"`
	Name    string    `mapstructure:"name" yaml:"name"`
	Path    string    `mapstructure:"path" yaml:"path"`
	AddedAt time.Time `mapstructure:"added_at" yaml:"added_at"`
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return errors.Newf("unsupported config version: %d", c.Version)
	}

	for _, p := range c.DefaultPlatforms {
		if !paths.ValidPlatform(p) {
			return errors.Newf("invalid default platform: %s", p)
		}
	}

	for p := range c.Platforms {
		if !paths.ValidPlatform(p) {
			return errors.Newf("invalid platform override key: %s", p)
		}
	}

	for name := range c.Repos {
		if !repoNamePattern.MatchString(name) {
			return errors.Newf("invalid repo name: %s", name)
		}
	}

	return nil
}

// Init initializes Viper with default configuration.
// Call this once at application startup before accessing config values.
func Init() {
	viper.Reset()

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

// DefaultConfigPath returns the default path for the config file.
// This respects AIX_CONFIG_DIR if set, otherwise uses XDG config home.
func DefaultConfigPath() string {
	if envDir := os.Getenv("AIX_CONFIG_DIR"); envDir != "" {
		return filepath.Join(envDir, "config.yaml")
	}
	return filepath.Join(paths.ConfigHome(), AppName, "config.yaml")
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
		var configNotFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &configNotFoundErr) {
			// If user specified a path, this is an error
			if path != "" {
				return nil, errors.Wrapf(err, "config file not found at %s", path)
			}
			// Otherwise (implicit load), it's fine to use defaults
		} else {
			// Real read error (parsing, permissions, etc)
			return nil, errors.Wrap(err, "reading config file")
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "unmarshaling config")
	}

	// Apply defaults if not set by config
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if len(cfg.DefaultPlatforms) == 0 {
		cfg.DefaultPlatforms = paths.Platforms()
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "validating config")
	}

	return &cfg, nil
}
