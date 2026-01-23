package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/pkg/fileutil"
)

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configEditCmd)
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage aix configuration",
	Long: `Manage aix configuration stored in ~/.config/aix/config.yaml.

Without a subcommand, lists all configuration values.`,
	Example: `  # List all configuration
  aix config

  # Get a specific value
  aix config get default_platforms

  # Set a value
  aix config set default_platforms claude,opencode

See Also: aix init, aix doctor`,
	RunE: runConfigList,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a single configuration value by key.

Supports dot notation for nested keys. Array values are printed one per line.`,
	Example: `  # Get version
  aix config get version

  # Get default platforms
  aix config get default_platforms

See Also: aix config set, aix config list`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

For array values like default_platforms, use comma-separated values.
Platform names are validated against supported platforms.`,
	Example: `  # Set version
  aix config set version 2

  # Set default platforms
  aix config set default_platforms claude,opencode

See Also: aix config get, aix config list`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration",
	Long:  `List all configuration values in YAML format.`,
	Example: `  # List all configuration
  aix config list

See Also: aix config get, aix config set`,
	RunE: runConfigList,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration in $EDITOR",
	Long: `Open the configuration file in your default editor.

Uses $EDITOR environment variable, or falls back to vi.
If no configuration file exists, prints an error suggesting to run 'aix init'.`,
	Example: `  # Open config in default editor
  aix config edit

  # Open with specific editor
  EDITOR=nano aix config edit

See Also: aix config list, aix init`,
	RunE: runConfigEdit,
}

func runConfigGet(_ *cobra.Command, args []string) error {
	key := args[0]

	// Check if value exists
	if !viper.IsSet(key) {
		fmt.Println("not set")
		return nil
	}

	// Get the value and determine its type
	val := viper.Get(key)

	switch v := val.(type) {
	case []any:
		// Array values - print one per line
		for _, item := range v {
			fmt.Println(item)
		}
	case []string:
		// String slice - print one per line
		for _, item := range v {
			fmt.Println(item)
		}
	default:
		// Scalar values
		fmt.Println(viper.GetString(key))
	}

	return nil
}

func runConfigSet(_ *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Handle special keys
	switch key {
	case "default_platforms":
		platforms := parsePlatforms(value)
		if len(platforms) == 0 {
			return errors.New("no valid platforms specified")
		}

		// Validate all platforms
		var invalid []string
		for _, p := range platforms {
			if !paths.ValidPlatform(p) {
				invalid = append(invalid, p)
			}
		}
		if len(invalid) > 0 {
			return errors.Newf("invalid platform(s): %s (valid: %s)",
				strings.Join(invalid, ", "),
				strings.Join(paths.Platforms(), ", "))
		}

		viper.Set(key, platforms)
		if err := writeConfig(); err != nil {
			return err
		}
		fmt.Printf("Set %s = %v\n", key, platforms)

	case "version":
		viper.Set(key, value)
		if err := writeConfig(); err != nil {
			return err
		}
		fmt.Printf("Set %s = %s\n", key, value)

	default:
		viper.Set(key, value)
		if err := writeConfig(); err != nil {
			return err
		}
		fmt.Printf("Set %s = %s\n", key, value)
	}

	return nil
}

func runConfigList(_ *cobra.Command, _ []string) error {
	// Build config structure from viper
	cfg := map[string]any{
		"version":           viper.GetInt("version"),
		"default_platforms": viper.GetStringSlice("default_platforms"),
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "marshaling config")
	}

	fmt.Print(string(data))
	return nil
}

func runConfigEdit(_ *cobra.Command, _ []string) error {
	configPath := filepath.Join(paths.ConfigHome(), config.AppName, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return errors.Newf("config file not found at %s\nRun 'aix init' to create it", configPath)
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Launch editor
	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "running editor")
	}

	return nil
}

// parsePlatforms splits a comma-separated string into a slice of platform names.
func parsePlatforms(s string) []string {
	var platforms []string
	for p := range strings.SplitSeq(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			platforms = append(platforms, p)
		}
	}
	return platforms
}

// writeConfig writes the current viper configuration to the config file.
func writeConfig() error {
	configPath := filepath.Join(paths.ConfigHome(), config.AppName, "config.yaml")

	// Build config structure
	cfg := map[string]any{
		"version":           viper.GetInt("version"),
		"default_platforms": viper.GetStringSlice("default_platforms"),
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return errors.Wrap(err, "creating config directory")
	}

	if err := fileutil.AtomicWriteYAML(configPath, cfg); err != nil {
		return errors.Wrap(err, "writing config file")
	}

	return nil
}
