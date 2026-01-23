package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/paths"
)

var (
	initYes       bool
	initPlatforms string
	initForce     bool
)

func init() {
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Non-interactive mode, accept all defaults")
	initCmd.Flags().StringVar(&initPlatforms, "platforms", "", "Comma-separated list of platforms to configure (overrides auto-detection)")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing configuration")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize aix configuration",
	Long: `Bootstrap aix configuration with automatic platform detection.

Creates ~/.config/aix/config.yaml with detected AI coding platforms.
Platforms are detected by checking if their config directories exist.`,
	Example: `  # Initialize with interactive prompts
  aix init

  # Initialize non-interactively, accepting defaults
  aix init --yes

  # Initialize for specific platforms
  aix init --platforms claude,opencode

  # Force overwrite existing configuration
  aix init --force

  See Also: aix config, aix doctor`,
	RunE: runInit,
}

// aixConfig represents the aix configuration file structure.
type aixConfig struct {
	Version          int      `yaml:"version"`
	DefaultPlatforms []string `yaml:"default_platforms"`
}

func runInit(_ *cobra.Command, _ []string) error {
	configPath := filepath.Join(paths.ConfigHome(), config.AppName, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !initForce {
		fmt.Printf("Configuration already exists at %s\n", configPath)
		fmt.Println("Use --force to overwrite")
		return nil
	}

	// Determine platforms to configure
	platforms := detectPlatforms()
	if initPlatforms != "" {
		platforms = parsePlatformList(initPlatforms)
	}

	// Interactive confirmation
	if !initYes {
		fmt.Printf("Detected platforms: %s\n", strings.Join(platforms, ", "))
		fmt.Println()
		fmt.Println("This will create:")
		fmt.Printf("  %s\n", configPath)
		fmt.Println()

		if !confirm("Proceed?") {
			fmt.Println("Aborted")
			return nil
		}
	} else {
		fmt.Printf("Detected platforms: %s\n", strings.Join(platforms, ", "))
	}

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return errors.Wrap(err, "creating config directory")
	}

	// Write config file
	cfg := aixConfig{
		Version:          1,
		DefaultPlatforms: platforms,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return errors.Wrap(err, "marshaling config")
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return errors.Wrap(err, "writing config file")
	}

	fmt.Printf("Created %s\n", configPath)
	return nil
}

// detectPlatforms returns a list of installed platform names.
func detectPlatforms() []string {
	var installed []string
	for _, p := range paths.Platforms() {
		if platformInstalled(p) {
			installed = append(installed, p)
		}
	}
	return installed
}

// parsePlatformList parses a comma-separated list of platform names.
// Invalid platform names are silently filtered out.
func parsePlatformList(s string) []string {
	var platforms []string
	for p := range strings.SplitSeq(s, ",") {
		p = strings.TrimSpace(p)
		if paths.ValidPlatform(p) {
			platforms = append(platforms, p)
		}
	}
	return platforms
}

// confirm prompts the user for a yes/no confirmation.
// Returns true only if the user enters "y" or "yes" (case-insensitive).
func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N] ", prompt)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
