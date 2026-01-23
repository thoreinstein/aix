package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/editor"
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

func runAgentEdit(_ *cobra.Command, args []string) error {
	target := args[0]

	// 1. Check if target is a local file path
	info, err := os.Stat(target)
	if err == nil && !info.IsDir() {
		// It's a local file, use it directly
		absPath, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		fmt.Printf("Opening local agent file: %s\n", absPath)
		return editor.Open(absPath)
	}

	// 2. Lookup as installed agent name
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	var agentPath string
	var foundPlatform cli.Platform

	for _, p := range platforms {
		_, err := p.GetAgent(target)
		if err == nil {
			foundPlatform = p
			// Agents are .md files, not directories
			agentPath = filepath.Join(p.AgentDir(), target+".md")
			break
		}
	}

	if foundPlatform == nil {
		return fmt.Errorf("agent %q not found (checked local path and installed platforms)", target)
	}

	// Verify file exists
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		return fmt.Errorf("agent file not found at %s", agentPath)
	}

	fmt.Printf("Opening %s agent %q...\n", foundPlatform.DisplayName(), target)
	return editor.Open(agentPath)
}
