package command

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
)

var removeForce bool

func init() {
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "Skip confirmation prompt")
	Cmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an installed slash command",
	Long: `Remove an installed slash command from one or more platforms.

By default, removes the command from all detected platforms where it is installed.
Use the --platform flag to target specific platforms.

A confirmation prompt is shown before removal unless --force is specified.`,
	Example: `  # Remove command from all platforms (with confirmation)
  aix command remove review

  # Remove command without confirmation
  aix command remove review --force

  # Remove command from a specific platform only
  aix command remove review --platform claude

  See Also:
    aix command list     - List installed commands
    aix command edit     - Edit a command definition
    aix command install  - Install a new command`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func runRemove(_ *cobra.Command, args []string) error {
	return runRemoveWithIO(args, os.Stdout, os.Stdin)
}

// runRemoveWithIO allows injecting writers for testing.
func runRemoveWithIO(args []string, w io.Writer, r io.Reader) error {
	name := args[0]

	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return err
	}

	// Find platforms that have this command installed
	installedOn := findPlatformsWithCommand(platforms, name)
	if len(installedOn) == 0 {
		return errors.Newf("command %q not found on any platform", name)
	}

	// Confirm removal unless --force is specified
	if !removeForce {
		if !confirmRemoval(w, r, name, installedOn) {
			fmt.Fprintln(w, "Removal cancelled")
			return nil
		}
	}

	// Remove from each platform
	var failed []string
	for _, p := range installedOn {
		// Ensure backup exists before modifying
		if err := backup.EnsureBackedUp(p.Name(), p.BackupPaths()); err != nil {
			failed = append(failed, fmt.Sprintf("%s: backup failed: %v", p.DisplayName(), err))
			continue
		}

		fmt.Fprintf(w, "Removing from %s... ", p.DisplayName())
		if err := p.UninstallCommand(name); err != nil {
			fmt.Fprintln(w, "failed")
			failed = append(failed, fmt.Sprintf("%s: %v", p.DisplayName(), err))
			continue
		}
		fmt.Fprintln(w, "done")
	}

	if len(failed) > 0 {
		return errors.New("failed to remove from some platforms:\n  " + strings.Join(failed, "\n  "))
	}

	fmt.Fprintf(w, "\u2713 Command %q removed from %d platform(s)\n", name, len(installedOn))
	return nil
}

// findPlatformsWithCommand returns platforms where the command is installed.
func findPlatformsWithCommand(platforms []cli.Platform, name string) []cli.Platform {
	var result []cli.Platform
	for _, p := range platforms {
		_, err := p.GetCommand(name)
		if err == nil {
			result = append(result, p)
		}
	}
	return result
}

// confirmRemoval prompts the user to confirm command removal.
// Returns true only if the user enters "y" or "yes" (case-insensitive).
func confirmRemoval(w io.Writer, r io.Reader, name string, platforms []cli.Platform) bool {
	fmt.Fprintf(w, "Remove command %q from:\n", name)
	for _, p := range platforms {
		fmt.Fprintf(w, "  - %s\n", p.DisplayName())
	}
	fmt.Fprint(w, "Continue? [y/N]: ")

	reader := bufio.NewReader(r)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
