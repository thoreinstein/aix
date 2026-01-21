// Package commands implements the CLI commands for aix.
package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/paths"
)

// version is set at build time via ldflags.
// Default to a development version for local builds.
const version = "0.1.0"

// platformFlag holds the value of the --platform flag.
var platformFlag []string

func init() {
	cobra.OnInitialize(initConfig)

	// Add persistent flags
	rootCmd.PersistentFlags().StringSliceVarP(&platformFlag, "platform", "p", nil,
		`target platform(s): claude, opencode (default: all detected)`)

	// Add version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("aix version {{.Version}}\n")
}

func initConfig() {
	config.Init()
	// Ignore load errors - defaults will be used if no config file
	_ = config.Load()
}

var rootCmd = &cobra.Command{
	Use:   "aix",
	Short: "Unified CLI for AI coding assistant configurations",
	Long: `aix is a unified CLI for managing AI coding assistant configurations
across multiple platforms including Claude Code, OpenCode, Codex CLI,
and Gemini CLI.

It manages skills, slash commands, agents, and MCP server configurations.
Write once, deploy everywhere. Define your configurations in a
platform-agnostic format and let aix handle the translation to each
platform's native format.

Use the --platform flag to target specific platforms, or omit it to
target all detected/installed platforms.`,
	PersistentPreRunE: validatePlatformFlag,
}

// validatePlatformFlag checks that all specified platforms are valid.
func validatePlatformFlag(cmd *cobra.Command, _ []string) error {
	// Skip validation for help and version commands
	if cmd.Name() == "help" || cmd.Name() == "version" {
		return nil
	}

	// If no platforms specified, that's fine - we'll use detected platforms
	if len(platformFlag) == 0 {
		return nil
	}

	// Validate each specified platform
	var invalid []string
	for _, p := range platformFlag {
		if !paths.ValidPlatform(p) {
			invalid = append(invalid, p)
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("invalid platform(s): %s (valid: %s)",
			strings.Join(invalid, ", "),
			strings.Join(paths.Platforms(), ", "))
	}

	return nil
}

// GetPlatformFlag returns the current value of the --platform flag.
// This is used by subcommands to access the flag value.
func GetPlatformFlag() []string {
	return platformFlag
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
