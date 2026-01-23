// Package command provides the command group for managing AI assistant slash commands.
package command

import "github.com/spf13/cobra"

// Cmd is the command that groups all command-related subcommands.
var Cmd = &cobra.Command{
	Use:   "command",
	Short: "Manage slash commands across platforms",
	Long: `Manage reusable AI assistant slash commands.

Slash commands extend AI assistant functionality with custom workflows,
templates, and automated operations. Use subcommands to install, list,
show, and remove commands across supported platforms.`,
	Example: `  # List all installed commands
  aix command list

  # Install a command from a git repository
  aix command install https://github.com/user/my-command.git

  # Show details for a command
  aix command show review

  See Also:
    aix command list     - List installed commands
    aix command install  - Install a command
    aix command show     - Show command details
    aix command remove   - Remove a command`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}
