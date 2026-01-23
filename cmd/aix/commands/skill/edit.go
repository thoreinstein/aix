package skill

import (
	"fmt"
	"os"
	"os/exec"
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
	Use:   "edit <name|path>",
	Short: "Open skill directory in $EDITOR",
	Long: `Open the directory containing the skill in your default editor.

You can provide either:
  - The name of an installed skill (e.g. "my-skill")
  - A path to a local skill directory (e.g. "./my-skill" or ".")

Uses the $EDITOR environment variable. If not set, defaults to 'vi'.
If the skill is installed on multiple platforms, uses the first one found
unless --platform is specified.`,
	Example: `  # Open installed skill
  aix skill edit my-skill

  # Open local skill directory
  aix skill edit ./my-new-skill

  # Open current directory
  aix skill edit .

  # Open skill on specific platform
  aix skill edit my-skill --platform claude

  See Also:
    aix skill show     - Show skill details
    aix skill list     - List installed skills`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func runEdit(_ *cobra.Command, args []string) error {
	target := args[0]

	// 1. Check if target is a local path
	info, err := os.Stat(target)
	if err == nil {
		path := target
		if !info.IsDir() {
			// If it's a file, open the parent directory
			path = filepath.Dir(target)
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		fmt.Printf("Opening local skill at %s...\n", absPath)
		return editor.Open(absPath)
	}
	// 2. Lookup as installed skill name
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return err
	}

	// Find the skill on available platforms
	var skillPath string
	var foundPlatform cli.Platform

	for _, p := range platforms {
		_, err := p.GetSkill(target)
		if err == nil {
			// Found it
			foundPlatform = p
			// Construct path to the skill directory
			skillPath = filepath.Join(p.SkillDir(), target)
			break
		}
	}

	if foundPlatform == nil {
		return errors.Newf("skill %q not found (checked local path and installed platforms)", target)
	}

	// Check if directory exists (sanity check)
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return errors.Newf("skill directory not found at %s", skillPath)
	}

	fmt.Printf("Opening %s skill %q...\n", foundPlatform.DisplayName(), target)
	return openInEditor(skillPath)
}

func openInEditor(path string) error {
	// Determine editor
	editorName := os.Getenv("EDITOR")
	if editorName == "" {
		editorName = "vi"
	}

	fmt.Printf("Location: %s\n", path)

	// Launch editor
	cmd := exec.Command(editorName, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "running editor")
	}

	return nil
}
