package command

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/editor"
	"github.com/thoreinstein/aix/internal/errors"
)

func init() {
	flags.AddScopeFlag(editCmd)
	Cmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Open a command definition in $EDITOR",
	Long: `Open the source definition of an installed slash command in your default editor.

Searches for the command across all detected platforms (or only the specified
--platform). If found on multiple platforms, opens the first one found.

Uses the $EDITOR environment variable, falling back to $VISUAL, then nano, then vi.`,
	Example: `  # Edit a command
  aix command edit review

  # Edit a command on a specific platform
  aix command edit review --platform claude`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func runEdit(_ *cobra.Command, args []string) error {
	name := args[0]

	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	scope := cli.ParseScope(flags.GetScopeFlag())

	return runEditWithPlatforms(name, platforms, scope, editor.Open)
}

func runEditWithPlatforms(name string, platforms []cli.Platform, scope cli.Scope, opener func(string) error) error {
	var cmdPath string
	for _, p := range platforms {
		_, err := p.GetCommand(name, scope)
		if err != nil {
			// Command not found on this platform, continue to next
			continue
		}

		// Found the command, construct path
		cmdPath = filepath.Join(p.CommandDir(), name+".md")
		break
	}

	if cmdPath == "" {
		return errors.Newf("command %q not found on any platform", name)
	}

	if err := opener(cmdPath); err != nil {
		return errors.Wrapf(err, "opening command %q", name)
	}

	return nil
}
