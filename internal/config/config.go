package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	aixerrors "github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/paths"
)

// CurrentVersion is the current configuration schema version.
// Increment this when making breaking changes to the config format.
const CurrentVersion = 1

// Config holds the aix CLI configuration settings.
type Config struct {
	// Version is the configuration schema version for future migrations.
	Version int `yaml:"version"`

	// DefaultPlatforms lists the platforms to target by default when
	// no explicit platforms are specified.
	DefaultPlatforms []string `yaml:"default_platforms"`

	// SkillsDir overrides the default skills directory location.
	// If empty, the platform-specific default is used.
	SkillsDir string `yaml:"skills_dir,omitempty"`

	// CommandsDir overrides the default commands directory location.
	// If empty, the platform-specific default is used.
	CommandsDir string `yaml:"commands_dir,omitempty"`
}

// DefaultConfigPath returns the default configuration file path.
// On Unix: ~/.config/aix/config.yaml
// On macOS: ~/Library/Application Support/aix/config.yaml
// On Windows: %LOCALAPPDATA%/aix/config.yaml
func DefaultConfigPath() string {
	return filepath.Join(paths.ConfigHome(), "aix", "config.yaml")
}

// Default returns a Config with sensible default values.
func Default() *Config {
	return &Config{
		Version:          CurrentVersion,
		DefaultPlatforms: paths.Platforms(),
	}
}

// Load reads and validates a configuration from the specified path.
// Returns ErrNotFound if the file doesn't exist.
// Returns ErrInvalidConfig if YAML is malformed or validation fails.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", aixerrors.ErrNotFound, path)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", aixerrors.ErrInvalidConfig, err)
	}

	if errs := Validate(&cfg); len(errs) > 0 {
		// Return the first validation error wrapped with ErrInvalidConfig
		return nil, fmt.Errorf("%w: %v", aixerrors.ErrInvalidConfig, errs[0])
	}

	return &cfg, nil
}

// LoadDefault loads configuration from the default location.
// If the file doesn't exist, returns a default Config (not an error).
// Returns an error only if the file exists but cannot be read or is invalid.
func LoadDefault() (*Config, error) {
	cfg, err := Load(DefaultConfigPath())
	if err != nil {
		if isNotFound(err) {
			return Default(), nil
		}
		return nil, err
	}
	return cfg, nil
}

// Save writes the configuration to the specified path.
// Creates parent directories if they don't exist.
// The file is written with mode 0644.
func Save(cfg *Config, path string) error {
	if errs := Validate(cfg); len(errs) > 0 {
		return fmt.Errorf("%w: %v", aixerrors.ErrInvalidConfig, errs[0])
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// isNotFound checks if an error is or wraps ErrNotFound.
func isNotFound(err error) bool {
	return err != nil && containsError(err, aixerrors.ErrNotFound)
}

// containsError checks if target is in the error chain.
func containsError(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}
