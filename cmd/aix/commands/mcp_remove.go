package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
)

var mcpRemoveForce bool

func init() {
	mcpRemoveCmd.Flags().BoolVar(&mcpRemoveForce, "force", false, "Skip confirmation prompt")
	mcpCmd.AddCommand(mcpRemoveCmd)
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an MCP server configuration",
	Long: `Remove an MCP server configuration from one or more platforms.

By default, removes the server from all detected platforms where it is configured.
Use the --platform flag to target specific platforms.

A confirmation prompt is shown before removal unless --force is specified.`,
	Example: `  # Remove MCP server from all platforms (with confirmation)
  aix mcp remove github

  # Remove MCP server without confirmation
  aix mcp remove github --force

  # Remove MCP server from a specific platform only
  aix mcp remove github --platform claude

  See Also:
    aix mcp add      - Add a new server
    aix mcp list     - List configured servers`,
	Args: cobra.ExactArgs(1),
	RunE: runMCPRemove,
}

func runMCPRemove(_ *cobra.Command, args []string) error {
	return runMCPRemoveWithIO(args, os.Stdout, os.Stdin)
}

// runMCPRemoveWithIO allows injecting writers for testing.
func runMCPRemoveWithIO(args []string, w io.Writer, r io.Reader) error {
	name := args[0]

	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	// Find platforms that have this MCP server configured
	configuredOn := findPlatformsWithMCP(platforms, name)
	if len(configuredOn) == 0 {
		return errors.Newf("server %q not found on any platform", name)
	}

	// Confirm removal unless --force is specified
	if !mcpRemoveForce {
		if !confirmMCPRemoval(w, r, name, configuredOn) {
			fmt.Fprintln(w, "removal cancelled")
			return nil
		}
	}

	// Remove from each platform
	fmt.Fprintf(w, "Removing MCP server %q...\n", name)

	var failed []string
	for _, p := range platforms {
		// Check if server exists on this platform
		_, err := p.GetMCP(name)
		if err != nil {
			fmt.Fprintf(w, "  %s: not found (skipped)\n", p.Name())
			continue
		}

		if err := p.RemoveMCP(name); err != nil {
			fmt.Fprintf(w, "  %s: failed\n", p.Name())
			failed = append(failed, fmt.Sprintf("%s: %v", p.DisplayName(), err))
			continue
		}
		fmt.Fprintf(w, "  %s: removed\n", p.Name())
	}

	if len(failed) > 0 {
		return errors.New("failed to remove from some platforms:\n  " + strings.Join(failed, "\n  "))
	}

	return nil
}

// findPlatformsWithMCP returns platforms where the MCP server is configured.
func findPlatformsWithMCP(platforms []cli.Platform, name string) []cli.Platform {
	var result []cli.Platform
	for _, p := range platforms {
		_, err := p.GetMCP(name)
		if err == nil {
			result = append(result, p)
		}
	}
	return result
}

// confirmMCPRemoval prompts the user to confirm MCP server removal.
// Returns true only if the user enters "y" or "yes" (case-insensitive).
func confirmMCPRemoval(w io.Writer, r io.Reader, name string, platforms []cli.Platform) bool {
	platformNames := make([]string, len(platforms))
	for i, p := range platforms {
		platformNames[i] = p.DisplayName()
	}
	fmt.Fprintf(w, "Remove MCP server %q from %s? [y/N]: ", name, strings.Join(platformNames, ", "))

	reader := bufio.NewReader(r)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
