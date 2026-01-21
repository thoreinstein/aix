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
	Long: `Manage skills across AI coding assistant platforms.

Skills are reusable instructions that extend the capabilities of AI coding
assistants. They are defined as markdown files with YAML frontmatter and
can be installed to multiple platforms simultaneously.

Available subcommands:
  list      List installed skills
  show      Display skill details
  install   Install a skill from a file or URL
  remove    Remove an installed skill
  validate  Validate a skill file
  init      Create a new skill from a template

Use "aix skill [command] --help" for more information about a command.`,
	// RunE shows help when invoked without a subcommand.
	// This ensures 'skill' appears as a command, not just a help topic.
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}
