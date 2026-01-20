package config

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/thoreinstein/aix/internal/paths"
)

// Validation errors for configuration fields.
var (
	// ErrVersionTooLow indicates the version field is below the minimum.
	ErrVersionTooLow = errors.New("version must be >= 1")

	// ErrInvalidPlatform indicates an unrecognized platform name.
	ErrInvalidPlatform = errors.New("invalid platform")

	// ErrInvalidPath indicates a path value is malformed.
	ErrInvalidPath = errors.New("invalid path")
)

// Validate checks a Config for validity.
// Returns nil if valid, or a slice of validation errors.
func Validate(cfg *Config) []error {
	if cfg == nil {
		return []error{errors.New("config is nil")}
	}

	var errs []error

	// Version must be >= 1
	if cfg.Version < 1 {
		errs = append(errs, ErrVersionTooLow)
	}

	// Validate platform names
	for _, platform := range cfg.DefaultPlatforms {
		if !paths.ValidPlatform(platform) {
			errs = append(errs, &PlatformError{
				Platform: platform,
				Err:      ErrInvalidPlatform,
			})
		}
	}

	// Validate directory paths if set
	if cfg.SkillsDir != "" {
		if err := validatePath(cfg.SkillsDir); err != nil {
			errs = append(errs, &PathError{
				Field: "skills_dir",
				Path:  cfg.SkillsDir,
				Err:   err,
			})
		}
	}

	if cfg.CommandsDir != "" {
		if err := validatePath(cfg.CommandsDir); err != nil {
			errs = append(errs, &PathError{
				Field: "commands_dir",
				Path:  cfg.CommandsDir,
				Err:   err,
			})
		}
	}

	return errs
}

// validatePath checks if a path string is well-formed.
// It does not check if the path exists, only that it's syntactically valid.
func validatePath(path string) error {
	// Empty paths are valid (they mean "use default")
	if path == "" {
		return nil
	}

	// Check for null bytes which are never valid in paths
	if strings.ContainsRune(path, '\x00') {
		return ErrInvalidPath
	}

	// Clean the path and check it's not empty after cleaning
	cleaned := filepath.Clean(path)
	if cleaned == "" || cleaned == "." {
		return ErrInvalidPath
	}

	return nil
}

// PlatformError represents an error for a specific platform.
type PlatformError struct {
	Platform string
	Err      error
}

func (e *PlatformError) Error() string {
	return e.Err.Error() + ": " + e.Platform
}

func (e *PlatformError) Unwrap() error {
	return e.Err
}

// PathError represents an error for a specific path field.
type PathError struct {
	Field string
	Path  string
	Err   error
}

func (e *PathError) Error() string {
	return e.Field + ": " + e.Err.Error() + ": " + e.Path
}

func (e *PathError) Unwrap() error {
	return e.Err
}
