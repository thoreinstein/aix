package skill

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
	Short: "Remove an installed skill",
	Long: `Remove an installed skill from one or more platforms.

By default, removes the skill from all detected platforms where it is installed.
Use the --platform flag to target specific platforms.

A confirmation prompt is shown before removal unless --force is specified.`,
	Example: `  # Remove skill from all platforms (with confirmation)
  aix skill remove debug

  # Remove skill without confirmation
  aix skill remove debug --force

  # Remove skill from a specific platform only
  aix skill remove debug --platform claude

  See Also:
    aix skill install  - Install a skill
    aix skill list     - List installed skills`,
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

	// Find platforms that have this skill installed
	installedOn := findPlatformsWithSkill(platforms, name)
	if len(installedOn) == 0 {
		return errors.Newf("skill %q not found on any platform", name)
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
		if err := p.UninstallSkill(name); err != nil {
			fmt.Fprintln(w, "failed")
			failed = append(failed, fmt.Sprintf("%s: %v", p.DisplayName(), err))
			continue
		}
		fmt.Fprintln(w, "done")
	}

	if len(failed) > 0 {
		return errors.New("failed to remove from some platforms:\n  " + strings.Join(failed, "\n  "))
	}

	fmt.Fprintf(w, "\u2713 Skill %q removed from %d platform(s)\n", name, len(installedOn))
	return nil
}

// findPlatformsWithSkill returns platforms where the skill is installed.
func findPlatformsWithSkill(platforms []cli.Platform, name string) []cli.Platform {
	var result []cli.Platform
	for _, p := range platforms {
		_, err := p.GetSkill(name)
		if err == nil {
			result = append(result, p)
		}
	}
	return result
}

// confirmRemoval prompts the user to confirm skill removal.
// Returns true only if the user enters "y" or "yes" (case-insensitive).
func confirmRemoval(w io.Writer, r io.Reader, name string, platforms []cli.Platform) bool {
	fmt.Fprintf(w, "Remove skill %q from:\n", name)
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
