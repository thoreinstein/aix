package commands

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(commandCmd)
}

var commandCmd = &cobra.Command{
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
