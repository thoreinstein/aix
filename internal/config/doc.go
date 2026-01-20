// Package config provides configuration management for the aix CLI.
//
// This package handles loading, saving, and validating the aix tool's own
// configuration file. It is distinct from platform-specific configurations
// which are managed by the platform adapters.
//
// # Configuration File
//
// The default configuration file location is ~/.config/aix/config.yaml.
// The configuration file uses YAML format with the following structure:
//
//	version: 1
//	default_platforms:
//	  - claude
//	  - opencode
//	skills_dir: /custom/skills   # optional
//	commands_dir: /custom/commands # optional
//
// # Loading Configuration
//
// Use [LoadDefault] to load from the default location with graceful fallback
// to default values:
//
//	cfg, err := config.LoadDefault()
//	if err != nil {
//	    return fmt.Errorf("loading config: %w", err)
//	}
//
// Use [Load] when you need to load from a specific path:
//
//	cfg, err := config.Load("/path/to/config.yaml")
//	if err != nil {
//	    if errors.Is(err, aixerrors.ErrNotFound) {
//	        // file doesn't exist
//	    }
//	    return err
//	}
//
// # Validation
//
// All loaded configurations are validated automatically. You can also
// validate a configuration manually:
//
//	errs := config.Validate(cfg)
//	if len(errs) > 0 {
//	    for _, e := range errs {
//	        fmt.Println(e)
//	    }
//	}
//
// # Default Values
//
// The [Default] function returns a configuration with sensible defaults:
//
//	cfg := config.Default()
//	// cfg.Version = 1
//	// cfg.DefaultPlatforms = ["claude", "opencode", "codex", "gemini"]
package config
