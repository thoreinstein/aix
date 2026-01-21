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
show, and remove commands across supported platforms.

Available subcommands:
  install   Install a slash command from local file or git repository
  list      List installed slash commands
  show      Show details of a specific slash command
  remove    Remove a slash command from platforms
  init      Scaffold a new slash command
  validate  Validate a slash command file

Examples:
  # List all commands
  aix command list

  # Install a command from local file
  aix command install ./my-command.md

  # Show command details
  aix command show review`,
}
