// Package commands implements the CLI commands for aix.
package commands

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/logging"
	"github.com/thoreinstein/aix/internal/paths"
)

// version is set at build time via ldflags.
// Default to a development version for local builds.
const version = "0.1.0"

// platformFlag holds the value of the --platform flag.
var platformFlag []string

// verbosity holds the count of -v flags.
var verbosity int

// quiet holds the value of the -q/--quiet flag.
var quiet bool

// configLoadErr holds any error that occurred during config loading.
var configLoadErr error

func init() {
	cobra.OnInitialize(initConfig)

	// Add persistent flags
	rootCmd.PersistentFlags().StringSliceVarP(&platformFlag, "platform", "p", nil,
		`target platform(s): claude, opencode (default: all detected)`)
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v",
		"increase verbosity level (e.g., -v, -vv)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false,
		"suppress non-error output")

	// Add version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("aix version {{.Version}}\n")

	// Silence errors and usage so we can control error output
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
}

func initConfig() {
	config.Init()
	// Capture load errors for later reporting
	_, configLoadErr = config.Load("")
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
	Example: `  # Initialize configuration
  aix init

  # List installed skills
  aix skill list

  # Check system health
  aix doctor

  # Target specific platform
  aix skill list --platform claude

  See Also: aix init, aix doctor, aix config`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logging first
		if err := setupLogging(cmd); err != nil {
			return err
		}
		return validatePlatformFlag(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// setupLogging configures the default logger based on verbosity flags.
func setupLogging(cmd *cobra.Command) error {
	if quiet && verbosity > 0 {
		return errors.NewUserError(nil, "cannot use --quiet and --verbose together")
	}

	var level slog.Level
	if quiet {
		level = slog.LevelError
	} else {
		v := verbosity

		// CLI flags take precedence, but if not set, check env var
		if v == 0 {
			if val, ok := os.LookupEnv("AIX_DEBUG"); ok {
				switch val {
				case "1", "true":
					v = 2 // Debug
				case "2":
					v = 3 // Trace
				}
			}
		}
		level = logging.LevelFromVerbosity(v)
	}

	logger := logging.New(logging.Config{
		Level:  level,
		Format: logging.FormatText, // Default to text
		Output: cmd.ErrOrStderr(),
	})
	slog.SetDefault(logger)

	return nil
}

// validatePlatformFlag checks that all specified platforms are valid.
func validatePlatformFlag(cmd *cobra.Command, _ []string) error {
	// Skip validation for help and version commands
	if cmd.Name() == "help" || cmd.Name() == "version" {
		return nil
	}

	// Check for config load errors first
	if configLoadErr != nil {
		return errors.NewConfigError(configLoadErr)
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
		err := errors.Newf("invalid platform(s): %s (valid: %s)",
			strings.Join(invalid, ", "),
			strings.Join(paths.Platforms(), ", "))
		return errors.NewUserError(err, "Run 'aix --help' to see valid platforms")
	}

	return nil
}

// GetPlatformFlag returns the current value of the --platform flag.
// This is used by subcommands to access the flag value.
func GetPlatformFlag() []string {
	return platformFlag
}

// SetPlatformFlag sets the platform flag value.
// This is used for programmatic override (e.g., interactive mode).
func SetPlatformFlag(platforms []string) {
	platformFlag = platforms
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
