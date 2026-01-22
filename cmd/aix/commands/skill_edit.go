package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
)

func init() {
	skillCmd.AddCommand(skillEditCmd)
}

var skillEditCmd = &cobra.Command{
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
	RunE: runSkillEdit,
}

func runSkillEdit(_ *cobra.Command, args []string) error {
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
		return openInEditor(absPath)
	}
	// 2. Lookup as installed skill name
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
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
		return fmt.Errorf("skill %q not found (checked local path and installed platforms)", target)
	}

	// Check if directory exists (sanity check)
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return fmt.Errorf("skill directory not found at %s", skillPath)
	}

	fmt.Printf("Opening %s skill %q...\n", foundPlatform.DisplayName(), target)
	return openInEditor(skillPath)
}

func openInEditor(path string) error {
	// Determine editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	fmt.Printf("Location: %s\n", path)

	// Launch editor
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running editor: %w", err)
	}

	return nil
}
