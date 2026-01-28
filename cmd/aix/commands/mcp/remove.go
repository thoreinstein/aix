package mcp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
)

var removeForce bool

func init() {
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "Skip confirmation prompt")
	Cmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
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
	RunE: runRemove,
}

func runRemove(_ *cobra.Command, args []string) error {
	return runRemoveWithIO(args, os.Stdout, os.Stdin)
}

// runRemoveWithIO allows injecting writers for testing.
func runRemoveWithIO(args []string, w io.Writer, r io.Reader) error {
	name := args[0]

	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	// Find platforms that have this MCP server configured
	configuredOn := findPlatformsWithMCP(platforms, name)
	if len(configuredOn) == 0 {
		return errors.Newf("server %q not found on any platform", name)
	}

	// Confirm removal unless --force is specified
	if !removeForce {
		if !confirmRemoval(w, r, name, configuredOn) {
			fmt.Fprintln(w, "removal cancelled")
			return nil
		}
	}

	// Remove from each platform
	fmt.Fprintf(w, "Removing MCP server %q...\n", name)

	var failed []string
	for _, p := range platforms {
		// Ensure backup exists before modifying
		if err := backup.EnsureBackedUp(p.Name(), p.BackupPaths()); err != nil {
			failed = append(failed, fmt.Sprintf("%s: backup failed: %v", p.DisplayName(), err))
			continue
		}

		// Check if server exists on this platform
		_, err := p.GetMCP(name)
		if err != nil {
			fmt.Fprintf(w, "  %s: not found (skipped)\n", p.Name())
			continue
		}

		if err := p.RemoveMCP(name, cli.ScopeUser); err != nil {
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

// confirmRemoval prompts the user to confirm MCP server removal.
// Returns true only if the user enters "y" or "yes" (case-insensitive).
func confirmRemoval(w io.Writer, r io.Reader, name string, platforms []cli.Platform) bool {
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
