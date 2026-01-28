package agent

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
)

var removeForce bool

func init() {
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "Skip confirmation prompt")
	Cmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove an installed agent",
	Long: `Remove an installed agent from one or more platforms.

By default, removes the agent from all detected platforms where it is installed.
Use the --platform flag to target specific platforms.

A confirmation prompt is shown before removal unless --force is specified.

Examples:
  # Remove agent from all platforms (with confirmation)
  aix agent remove my-agent

  # Remove agent without confirmation
  aix agent remove my-agent --force

  # Remove agent from a specific platform only
  aix agent remove my-agent --platform claude`,
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
		return errors.Wrap(err, "resolving platforms")
	}

	// Find platforms that have this agent installed
	installedOn := findPlatformsWithAgent(platforms, name)
	if len(installedOn) == 0 {
		return fmt.Errorf("agent %q not found on any platform", name)
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
		if err := p.UninstallAgent(name, cli.ScopeUser); err != nil {
			fmt.Fprintln(w, "failed")
			failed = append(failed, fmt.Sprintf("%s: %v", p.DisplayName(), err))
			continue
		}
		fmt.Fprintln(w, "done")
	}

	if len(failed) > 0 {
		return errors.New("failed to remove from some platforms:\n  " + strings.Join(failed, "\n  "))
	}

	fmt.Fprintf(w, "\u2713 Agent %q removed from %d platform(s)\n", name, len(installedOn))
	return nil
}

// findPlatformsWithAgent returns platforms where the agent is installed.
func findPlatformsWithAgent(platforms []cli.Platform, name string) []cli.Platform {
	var result []cli.Platform
	for _, p := range platforms {
		_, err := p.GetAgent(name)
		if err == nil {
			result = append(result, p)
		}
	}
	return result
}

// confirmRemoval prompts the user to confirm agent removal.
// Returns true only if the user enters "y" or "yes" (case-insensitive).
func confirmRemoval(w io.Writer, r io.Reader, name string, platforms []cli.Platform) bool {
	fmt.Fprintf(w, "Remove agent %q from:\n", name)
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
