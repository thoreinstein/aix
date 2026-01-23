package commands

import (
	"errors"

	"github.com/spf13/cobra"
)

func init() {
	agentCmd.AddCommand(agentEditCmd)
}

var agentEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Open agent file in $EDITOR",
	Long: `Open the agent file in your default editor.

Uses the $EDITOR environment variable. If not set, defaults to 'vi'.
If the agent is installed on multiple platforms, uses the first one found
unless --platform is specified.

Examples:
  # Open installed agent
  aix agent edit code-reviewer

  # Open agent for specific platform
  aix agent edit code-reviewer --platform claude`,
	Args: cobra.ExactArgs(1),
	RunE: runAgentEdit,
}

func runAgentEdit(_ *cobra.Command, _ []string) error {
	// Stub implementation - resolution logic comes in aix-m92, editor launch in aix-ijn
	return errors.New("not yet implemented: agent resolution (aix-m92) and editor launch (aix-ijn)")
}
