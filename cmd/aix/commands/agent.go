package commands

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage AI coding agents",
	Long:  `Commands for managing AI coding agents across Claude Code and OpenCode.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}
