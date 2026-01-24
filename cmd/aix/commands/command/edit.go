package command

import (
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/editor"
)

func init() {
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
		return err
	}

	return runEditWithPlatforms(name, platforms, editor.Open)
}

func runEditWithPlatforms(name string, platforms []cli.Platform, opener func(string) error) error {
	var cmdPath string
	for _, p := range platforms {
		_, err := p.GetCommand(name)
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
