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
}
