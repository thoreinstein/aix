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
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}
