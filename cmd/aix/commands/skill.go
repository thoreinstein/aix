package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(skillCmd)
}

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage skills across platforms",
	Long:  `Manage reusable AI assistant skills defined as Markdown with YAML frontmatter, enabling you to list, inspect, install, remove, validate, and initialize them across supported platforms`,
	// RunE shows help when invoked without a subcommand.
	// This ensures 'skill' appears as a command, not just a help topic.
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}
