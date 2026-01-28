package agent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/editor"
	"github.com/thoreinstein/aix/internal/errors"
)

func init() {
	Cmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Open agent file in $EDITOR",
	Long: `Open the agent file in your default editor.

Uses the $EDITOR environment variable. If not set, defaults to 'nano'.
If the agent is installed on multiple platforms, uses the first one found
unless --platform is specified.

Examples:
  # Open installed agent
  aix agent edit code-reviewer

  # Open agent for specific platform
  aix agent edit code-reviewer --platform claude`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	target := args[0]

	// 1. Check if target is a local file path
	info, err := os.Stat(target)
	if err == nil && !info.IsDir() {
		// It's a local file, use it directly
		absPath, err := filepath.Abs(target)
		if err != nil {
			return errors.Wrap(err, "getting absolute path")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Opening local agent file: %s\n", absPath)
		if err := editor.Open(absPath); err != nil {
			return errors.Wrap(err, "opening editor")
		}
		return validateAfterEdit(absPath, cmd)
	}

	// 2. Lookup as installed agent name
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	scope := cli.ParseScope(flags.GetScopeFlag())

	var agentPath string
	var foundPlatform cli.Platform

	for _, p := range platforms {
		_, err := p.GetAgent(target, scope)
		if err == nil {
			foundPlatform = p
			// Agents are .md files, not directories
			agentPath = filepath.Join(p.AgentDir(), target+".md")
			break
		}
	}

	if foundPlatform == nil {
		return errors.Newf("agent %q not found (checked local path and installed platforms)", target)
	}

	// Verify file exists
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		return errors.Newf("agent file not found at %s", agentPath)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Opening %s agent %q...\n", foundPlatform.DisplayName(), target)
	if err := editor.Open(agentPath); err != nil {
		return errors.Wrap(err, "opening editor")
	}
	return validateAfterEdit(agentPath, cmd)
}

// validateAfterEdit runs validation on the edited agent file.
// Validation errors are reported but don't fail the command since the user
// already saved their changes.
func validateAfterEdit(path string, cmd *cobra.Command) error {
	w := cmd.OutOrStdout()
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Validating agent...")
	// Ignore validation errors - user already saved their file
	_ = runValidate(path, w)
	return nil
}
