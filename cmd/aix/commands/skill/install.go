package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
	"github.com/thoreinstein/aix/internal/cli"
	cliprompt "github.com/thoreinstein/aix/internal/cli/prompt"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/git"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
	"github.com/thoreinstein/aix/internal/resource"
	"github.com/thoreinstein/aix/internal/skill/parser"
	skillvalidator "github.com/thoreinstein/aix/internal/skill/validator"
	"github.com/thoreinstein/aix/internal/validator"
)

var (
	installForce bool
	installFile  bool
)

func init() {
	installCmd.Flags().BoolVar(&installForce, "force", false,
		"overwrite existing skill without confirmation")
	installCmd.Flags().BoolVarP(&installFile, "file", "f", false,
		"treat argument as a file path instead of searching repos")
	Cmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install <source>",
	Short: "Install a skill from a repository, local path, or git URL",
	Long: `Install a skill from a configured repository, local directory, or git URL.

The source can be:
  - A skill name to search in configured repositories
  - A local path to a directory containing SKILL.md
  - A git URL (https://, git@, or .git suffix)

When given a name (not a path), aix searches configured repositories first.
If the skill exists in multiple repositories, you will be prompted to select one.
Use --file to skip repo search and treat the argument as a file path.

For git URLs, the repository is cloned to a temporary directory, the skill
is installed, and the temporary directory is cleaned up.`,
	Example: `  # Install by name from configured repos
  aix skill install code-review

  # Install from local directory
  aix skill install ./my-skill
  aix skill install --file my-skill  # Force file path interpretation

  # Install from absolute path
  aix skill install /path/to/skill-dir

  # Install from git repository
  aix skill install https://github.com/user/skill-repo.git

  # Install from git SSH URL
  aix skill install git@github.com:user/skill-repo.git

  # Force overwrite existing skill
  aix skill install code-review --force`,
	Args: cobra.ExactArgs(1),
	RunE: runInstall,
}

// Sentinel errors for install operations.
var (
	errInstallFailed = errors.New("installation failed")
)

func runInstall(_ *cobra.Command, args []string) error {
	source := args[0]

	// If --file flag is set, treat argument as file path or URL (old behavior)
	if installFile {
		if git.IsURL(source) {
			return installFromGit(source)
		}
		return installFromLocal(source)
	}

	// If source is clearly a path or URL, use direct install
	if git.IsURL(source) || looksLikePath(source) {
		if git.IsURL(source) {
			return installFromGit(source)
		}
		return installFromLocal(source)
	}

	// Try repo lookup first
	matches, err := resource.FindByName(source, resource.TypeSkill)
	if err != nil {
		if errors.Is(err, resource.ErrNoReposConfigured) {
			return errors.New("no repositories configured. Run 'aix repo add <url>' to add one")
		}
		return errors.Wrap(err, "searching repositories")
	}

	if len(matches) > 0 {
		return installFromRepo(source, matches)
	}

	// No matches in repos - check if input might be a forgotten path
	if mightBePath(source) {
		if _, statErr := os.Stat(source); statErr == nil {
			return errors.Newf("skill %q not found in repositories, but a local file exists at that path.\nDid you mean: aix skill install --file %s", source, source)
		}
	}

	// Check if it's a local path that exists
	if _, err := os.Stat(source); err == nil {
		return installFromLocal(source)
	}

	return errors.Newf("skill %q not found in any configured repository", source)
}

// looksLikePath returns true if the source appears to be a file path.
func looksLikePath(source string) bool {
	// Starts with ./ or ../ or /
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") || strings.HasPrefix(source, "/") {
		return true
	}
	// Contains path separator
	if strings.Contains(source, string(filepath.Separator)) {
		return true
	}
	// On Windows, also check for backslash
	if filepath.Separator != '/' && strings.Contains(source, "/") {
		return true
	}
	return false
}

// mightBePath returns true if the input might be a path the user forgot the --file flag for.
// This catches edge cases not handled by looksLikePath, like Windows-style paths on Unix
// or files with .md extension.
func mightBePath(s string) bool {
	// Ends with .md extension
	if strings.HasSuffix(strings.ToLower(s), ".md") {
		return true
	}
	// Contains backslash (Windows paths, even on Unix)
	if strings.Contains(s, "\\") {
		return true
	}
	return false
}

// installFromRepo installs a skill from a configured repository.
func installFromRepo(name string, matches []resource.Resource) error {
	var selected *resource.Resource

	if len(matches) == 1 {
		selected = &matches[0]
	} else {
		// Multiple matches - prompt user to select
		choice, err := cliprompt.SelectResourceDefault(name, matches)
		if err != nil {
			return errors.Wrap(err, "selecting resource")
		}
		selected = choice
	}

	// Copy from cache to temp directory
	// For directory resources, tempDir is the resource subdirectory (e.g., /tmp/aix-install-xyz/implement/)
	// We need to clean up the parent temp directory
	tempDir, err := resource.CopyToTemp(selected)
	if err != nil {
		return errors.Wrap(err, "copying from repository cache")
	}
	defer func() {
		// For directory resources, tempDir is a subdirectory; clean up the parent
		// For flat files, tempDir is the temp directory itself
		parentDir := filepath.Dir(tempDir)
		if strings.Contains(filepath.Base(parentDir), "aix-install-") {
			if removeErr := os.RemoveAll(parentDir); removeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp dir: %v\n", removeErr)
			}
		} else {
			// Flat file case: tempDir is the temp directory
			if removeErr := os.RemoveAll(tempDir); removeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp dir: %v\n", removeErr)
			}
		}
	}()

	fmt.Printf("Installing from repository: %s\n", selected.RepoName)
	return installFromLocal(tempDir)
}

// installFromGit clones a git repository and installs the skill from it.
func installFromGit(url string) error {
	fmt.Println("Cloning repository...")

	// Create temp directory for clone
	tempDir, err := os.MkdirTemp("", "aix-skill-*")
	if err != nil {
		return errors.Wrap(err, "creating temp directory")
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temp dir: %v\n", removeErr)
		}
	}()

	// Clone the repository
	if err := git.Clone(url, tempDir, 1); err != nil {
		return errors.Wrap(err, "cloning repository")
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
		return errors.Newf("SKILL.md not found at %s", absPath)
	}

	fmt.Println("Validating skill...")

	// Parse the skill
	p := parser.New()
	skill, err := p.ParseFile(skillFile)
	if err != nil {
		return errors.Wrap(err, "parsing skill")
	}

	// Validate the skill
	v := skillvalidator.New()
	result := v.ValidateWithPath(skill, skillFile)
	if result.HasErrors() {
		fmt.Println("Skill validation failed:")
		reporter := validator.NewReporter(os.Stdout, validator.FormatText)
		_ = reporter.Report(result)
		return errInstallFailed
	}

	// Get target platforms
	platforms, err := cli.ResolvePlatforms(flags.GetPlatformFlag())
	if err != nil {
		return errors.Wrap(err, "resolving platforms")
	}

	// Check for existing skills (unless --force)
	if !installForce {
		for _, plat := range platforms {
			if _, err := plat.GetSkill(skill.Name); err == nil {
				return errors.Newf("skill %q already exists on %s (use --force to overwrite)",
					skill.Name, plat.DisplayName())
			}
		}
	}

	// Install to each platform
	var installedCount int
	for _, plat := range platforms {
		// Ensure backup exists before modifying
		if err := backup.EnsureBackedUp(plat.Name(), plat.BackupPaths()); err != nil {
			return errors.Wrapf(err, "backing up %s before install", plat.DisplayName())
		}

		fmt.Printf("Installing '%s' to %s... ", skill.Name, plat.DisplayName())

		// Convert skill to platform-specific type
		platformSkill := convertSkillForPlatform(skill, plat.Name())

		if err := plat.InstallSkill(platformSkill); err != nil {
			fmt.Println("failed")
			return errors.Wrapf(err, "failed to install to %s", plat.DisplayName())
		}

		fmt.Println("done")
		installedCount++
	}

	// Print summary
	platformWord := "platform"
	if installedCount != 1 {
		platformWord = "platforms"
	}
	fmt.Printf("[OK] Skill '%s' installed to %d %s\n", skill.Name, installedCount, platformWord)

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
