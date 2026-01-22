package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
	"github.com/thoreinstein/aix/internal/skill/parser"
	"github.com/thoreinstein/aix/internal/skill/validator"
)

var installForce bool

func init() {
	skillInstallCmd.Flags().BoolVarP(&installForce, "force", "f", false,
		"overwrite existing skill without confirmation")
	skillCmd.AddCommand(skillInstallCmd)
}

var skillInstallCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a skill from a local path or git URL",
	Long: `Install a skill from a local directory or git repository.

The source can be:
  - A local path to a directory containing SKILL.md
  - A git URL (https://, git@, or .git suffix)

For git URLs, the repository is cloned to a temporary directory, the skill
is installed, and the temporary directory is cleaned up.

Flags:
  --force, -f     Overwrite existing skill without confirmation
  --platform, -p  Install to specific platform(s) only`,
	Example: `  # Install from local directory
  aix skill install ./my-skill

  # Install from absolute path
  aix skill install /path/to/skill-dir

  # Install from git repository
  aix skill install https://github.com/user/skill-repo.git

  # Install from git SSH URL
  aix skill install git@github.com:user/skill-repo.git

  # Force overwrite existing skill
  aix skill install ./my-skill --force

  See Also:
    aix skill remove   - Remove an installed skill
    aix skill list     - List installed skills
    aix skill validate - Validate a skill before installing`,
	Args: cobra.ExactArgs(1),
	RunE: runSkillInstall,
}

// Sentinel errors for install operations.
var (
	errInstallFailed = errors.New("installation failed")
)

func runSkillInstall(_ *cobra.Command, args []string) error {
	source := args[0]

	// Determine if source is a git URL or local path
	if isGitURL(source) {
		return installFromGit(source)
	}

	return installFromLocal(source)
}

// isGitURL returns true if the source looks like a git repository URL.
func isGitURL(source string) bool {
	// Check for common git URL patterns
	if strings.Contains(source, "://") {
		return true
	}
	if strings.HasSuffix(source, ".git") {
		return true
	}
	if strings.HasPrefix(source, "git@") {
		return true
	}
	return false
}

// installFromGit clones a git repository and installs the skill from it.
func installFromGit(url string) error {
	fmt.Println("Cloning repository...")

	// Create temp directory for clone
	tempDir, err := os.MkdirTemp("", "aix-skill-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp dir: %v\n", removeErr)
		}
	}()

	// Clone the repository
	cmd := exec.Command("git", "clone", "--depth=1", url, tempDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cloning repository: %w", err)
	}

	return installFromLocal(tempDir)
}

// installFromLocal installs a skill from a local directory.
func installFromLocal(skillPath string) error {
	// Resolve to absolute path for consistent error messages
	absPath, err := filepath.Abs(skillPath)
	if err != nil {
		absPath = skillPath
	}

	// Construct path to SKILL.md
	skillFile := filepath.Join(absPath, "SKILL.md")

	// Check if SKILL.md exists
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		return fmt.Errorf("SKILL.md not found at %s", absPath)
	}

	fmt.Println("Validating skill...")

	// Parse the skill
	p := parser.New()
	skill, err := p.ParseFile(skillFile)
	if err != nil {
		return fmt.Errorf("parsing skill: %w", err)
	}

	// Validate the skill
	v := validator.New()
	validationErrs := v.ValidateWithPath(skill, skillFile)
	if len(validationErrs) > 0 {
		fmt.Println("Skill validation failed:")
		for _, e := range validationErrs {
			fmt.Printf("  - %v\n", e)
		}
		return errInstallFailed
	}

	// Get target platforms
	platforms, err := cli.ResolvePlatforms(GetPlatformFlag())
	if err != nil {
		return err
	}

	// Check for existing skills (unless --force)
	if !installForce {
		for _, plat := range platforms {
			if _, err := plat.GetSkill(skill.Name); err == nil {
				return fmt.Errorf("skill %q already exists on %s (use --force to overwrite)",
					skill.Name, plat.DisplayName())
			}
		}
	}

	// Install to each platform
	var installedCount int
	for _, plat := range platforms {
		fmt.Printf("Installing '%s' to %s... ", skill.Name, plat.DisplayName())

		// Convert skill to platform-specific type
		platformSkill := convertSkillForPlatform(skill, plat.Name())

		if err := plat.InstallSkill(platformSkill); err != nil {
			fmt.Println("failed")
			return fmt.Errorf("failed to install to %s: %w", plat.DisplayName(), err)
		}

		fmt.Println("done")
		installedCount++
	}

	// Print summary
	platformWord := "platform"
	if installedCount != 1 {
		platformWord = "platforms"
	}
	fmt.Printf("✓ Skill '%s' installed to %d %s\n", skill.Name, installedCount, platformWord)

	return nil
}

// convertSkillForPlatform converts a canonical claude.Skill to the appropriate
// platform-specific skill type.
func convertSkillForPlatform(skill *claude.Skill, platformName string) any {
	switch platformName {
	case "claude":
		// Claude uses the canonical format, return as-is
		return skill
	case "opencode":
		// Convert to OpenCode skill format
		return convertToOpenCodeSkill(skill)
	default:
		// Unknown platform, return as-is and let the adapter handle it
		return skill
	}
}

// convertToOpenCodeSkill converts a Claude skill to an OpenCode skill.
func convertToOpenCodeSkill(s *claude.Skill) *opencode.Skill {
	// Convert ToolList to []string
	allowedTools := []string(s.AllowedTools)

	// Convert compatibility slice to map (OpenCode uses map format)
	var compatibility map[string]string
	if len(s.Compatibility) > 0 {
		compatibility = make(map[string]string, len(s.Compatibility))
		for _, c := range s.Compatibility {
			// Compatibility entries might be "platform version" or just "platform"
			parts := strings.SplitN(c, " ", 2)
			if len(parts) == 2 {
				compatibility[parts[0]] = parts[1]
			} else {
				compatibility[c] = ""
			}
		}
	}

	// Convert string metadata to any metadata
	var metadata map[string]any
	if len(s.Metadata) > 0 {
		metadata = make(map[string]any, len(s.Metadata))
		for k, v := range s.Metadata {
			metadata[k] = v
		}
	}

	// Extract version and author from metadata if present
	var version, author string
	if s.Metadata != nil {
		version = s.Metadata["version"]
		author = s.Metadata["author"]
	}

	return &opencode.Skill{
		Name:          s.Name,
		Description:   s.Description,
		Version:       version,
		Author:        author,
		AllowedTools:  allowedTools,
		Compatibility: opencode.CompatibilityMap(compatibility),
		Metadata:      metadata,
		Instructions:  s.Instructions,
	}
}
