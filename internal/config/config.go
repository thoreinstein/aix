// Package config provides configuration management for aix using Viper.
package config

import (
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/thoreinstein/aix/internal/paths"
)

// AppName is the application name used for config file naming.
const AppName = "aix"

// Init initializes Viper with default configuration.
// Call this once at application startup before accessing config values.
func Init() {
	// Config file settings
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Search paths (in order of precedence)
	viper.AddConfigPath(".") // Current directory
	viper.AddConfigPath(filepath.Join(paths.ConfigHome(), AppName))

	// Environment variable support
	viper.SetEnvPrefix("AIX")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("version", 1)
	viper.SetDefault("default_platforms", paths.Platforms())
}

// Load reads the configuration file. Returns nil if no config file found
// (defaults will be used). Returns error only for parse failures.
func Load() error {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use defaults - this is fine
			return nil
		}
		return err
	}
	return nil
}
