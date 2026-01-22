package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
)

func init() {
	mcpCmd.AddCommand(mcpEnableCmd)
	mcpCmd.AddCommand(mcpDisableCmd)
}

var mcpEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a disabled MCP server",
	Long: `Enable a previously disabled MCP server.

The server will become active and available to the AI coding assistant.

Examples:
  # Enable on all platforms
  aix mcp enable github

  # Enable on specific platform
  aix mcp enable github --platform=claude`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runMCPSetEnabledWithIO(args[0], true, os.Stdout)
	},
}

var mcpDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable an MCP server without removing it",
	Long: `Disable an MCP server without removing its configuration.

The server will remain in the config but won't be loaded by the AI coding assistant.
Use 'aix mcp enable' to re-enable it later.

Examples:
  # Disable on all platforms
  aix mcp disable github

  # Disable on specific platform
  aix mcp disable github --platform=opencode`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runMCPSetEnabledWithIO(args[0], false, os.Stdout)
	},
}

// runMCPSetEnabledWithIO enables or disables an MCP server across platforms.
// The enabled parameter controls whether to enable (true) or disable (false).
func runMCPSetEnabledWithIO(name string, enabled bool, w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	action := "Enabling"
	pastTense := "enabled"
	if !enabled {
		action = "Disabling"
		pastTense = "disabled"
	}

	fmt.Fprintf(w, "%s MCP server %q...\n", action, name)

	var foundAny bool
	for _, plat := range platforms {
		if !plat.IsAvailable() {
			continue
		}

		// Check if server exists
		_, err := plat.GetMCP(name)
		if err != nil {
			fmt.Fprintf(w, "  %s: not found\n", plat.Name())
			continue
		}

		foundAny = true

		if enabled {
			err = plat.EnableMCP(name)
		} else {
			err = plat.DisableMCP(name)
		}

		if err != nil {
			fmt.Fprintf(w, "  %s: error: %v\n", plat.Name(), err)
			continue
		}

		fmt.Fprintf(w, "  %s: %s\n", plat.Name(), pastTense)
	}

	if !foundAny {
		return fmt.Errorf("server %q not found on any platform", name)
	}

	return nil
}
