// Package commands provides CLI commands for the aix tool.
package commands

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(mcpCmd)
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP server configurations",
	Long: `Manage Model Context Protocol (MCP) server configurations across platforms.

MCP servers extend AI coding assistants with additional tools and capabilities.
This command group allows you to add, remove, list, and manage MCP server
configurations in Claude Code, OpenCode, and other supported platforms.`,
	Example: `  # Add a local MCP server
  aix mcp add github npx -y @modelcontextprotocol/server-github

  # Add a remote SSE server
  aix mcp add api-gateway --url=https://api.example.com/mcp

  # List all configured MCP servers
  aix mcp list

  # Show details of an MCP server
  aix mcp show github

  See Also:
    aix mcp add      - Add a new MCP server
    aix mcp list     - List configured servers
    aix mcp show     - Show server details
    aix mcp remove   - Remove a server
    aix mcp enable   - Enable a server
    aix mcp disable  - Disable a server`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}
