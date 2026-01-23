// Package agent provides commands for managing AI coding agents.
package agent

import "github.com/spf13/cobra"

// Cmd is the parent command for all agent subcommands.
var Cmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage AI coding agents",
	Long:  `Commands for managing AI coding agents across Claude Code and OpenCode.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}
