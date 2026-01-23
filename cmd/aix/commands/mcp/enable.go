package mcp

import (
	"fmt"
	"io"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
)

func init() {
	Cmd.AddCommand(enableCmd)
	Cmd.AddCommand(disableCmd)
}

var enableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a disabled MCP server",
	Long: `Enable a previously disabled MCP server.

The server will become active and available to the AI coding assistant.`,
	Example: `  # Enable on all platforms
  aix mcp enable github

  # Enable on specific platform
  aix mcp enable github --platform=claude

  See Also:
    aix mcp disable  - Disable a server
    aix mcp list     - List configured servers`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runSetEnabledWithIO(args[0], true, os.Stdout)
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable an MCP server without removing it",
	Long: `Disable an MCP server without removing its configuration.

The server will remain in the config but won't be loaded by the AI coding assistant.
Use 'aix mcp enable' to re-enable it later.`,
	Example: `  # Disable on all platforms
  aix mcp disable github

  # Disable on specific platform
  aix mcp disable github --platform=opencode

  See Also:
    aix mcp enable   - Enable a server
    aix mcp list     - List configured servers`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runSetEnabledWithIO(args[0], false, os.Stdout)
	},
}

// runSetEnabledWithIO enables or disables an MCP server across platforms.
// The enabled parameter controls whether to enable (true) or disable (false).
func runSetEnabledWithIO(name string, enabled bool, w io.Writer) error {
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
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

		// Ensure backup exists before modifying
		if err := backup.EnsureBackedUp(plat.Name(), plat.BackupPaths()); err != nil {
			fmt.Fprintf(w, "  %s: backup failed: %v\n", plat.Name(), err)
			continue
		}

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
		return errors.Newf("server %q not found on any platform", name)
	}

	return nil
}
