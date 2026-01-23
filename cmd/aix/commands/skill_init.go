package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/pkg/frontmatter"
)

var (
	skillInitName         string
	skillInitDescription  string
	skillInitLicense      string
	skillInitVersion      string
	skillInitAuthor       string
	skillInitAllowedTools string
	skillInitDirs         string
	skillInitForce        bool
)

func init() {
	skillInitCmd.Flags().StringVar(&skillInitName, "name", "", "skill name (required)")
	skillInitCmd.Flags().StringVarP(&skillInitDescription, "description", "d", "", "skill description")
	skillInitCmd.Flags().StringVar(&skillInitLicense, "license", "", "license (e.g. MIT)")
	skillInitCmd.Flags().StringVar(&skillInitVersion, "version", "", "skill version")
	skillInitCmd.Flags().StringVar(&skillInitAuthor, "author", "", "skill author")
	skillInitCmd.Flags().StringVar(&skillInitAllowedTools, "allowed-tools", "", "comma-separated list of allowed tools")
	skillInitCmd.Flags().StringVar(&skillInitDirs, "dirs", "", "comma-separated list of optional directories to create (docs, tests, bin, data)")
	skillInitCmd.Flags().BoolVarP(&skillInitForce, "force", "f", false, "overwrite existing directory")
	skillCmd.AddCommand(skillInitCmd)
}

var skillInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create a new skill interactively",
	Long: `Create a new skill directory with a scaffolded SKILL.md file.

If [path] is provided, the skill is created in that directory.
If no path is provided, the current directory is used.

The command is interactive and will prompt for skill details unless they are
provided via flags.`,
	Example: `  # Create in current directory, interactive prompts
  aix skill init

  # Create in specific directory with optional folders
  aix skill init my-skill --dirs docs,tests

  # Non-interactive creation
  aix skill init my-skill --name my-skill --description "My Skill" --license MIT

  See Also:
    aix skill validate - Validate a skill
    aix skill edit     - Edit a skill`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSkillInit,
}

// skillNameRegex validates skill names per the Agent Skills Specification.
// Must start with a lowercase letter, followed by lowercase alphanumeric characters,
// optionally followed by hyphen-separated segments. No leading, trailing, or
// consecutive hyphens are allowed.
var skillNameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// skillNameSanitizer matches characters that are not allowed in a skill name.
var skillNameSanitizer = regexp.MustCompile(`[^a-z0-9-]+`)

// errInitFailed is a sentinel error that signals non-zero exit.
var errInitFailed = errors.New("skill initialization failed")

func sanitizeDefaultName(name string) string {
	// Normalize to lowercase and replace invalid characters with hyphens.
	sanitized := strings.ToLower(name)
	sanitized = skillNameSanitizer.ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")

	// Fallback to a safe default if the result is empty or still invalid.
	if sanitized == "" || !skillNameRegex.MatchString(sanitized) {
		return "new-skill"
	}

	return sanitized
}

func runSkillInit(_ *cobra.Command, args []string) error {
	// Determine default name
	defaultName := "my-skill"
	if len(args) > 0 {
		defaultName = sanitizeDefaultName(filepath.Base(args[0]))
	}

	// Interactive prompts
	scanner := bufio.NewScanner(os.Stdin)

	name := skillInitName
	if name == "" {
		name = prompt(scanner, "Skill Name", defaultName)
	}

	// Validate name immediately
	if err := validateSkillName(name); err != nil {
		fmt.Printf("Error: %s\n", err)
		return errInitFailed
	}

	// Determine final path
	var absPath string
	var err error
	if len(args) > 0 {
		// User provided a path, use it directly
		absPath, err = filepath.Abs(args[0])
	} else {
		// User provided no path, create subdirectory with skill name
		absPath, err = filepath.Abs(name)
	}
	if err != nil {
		return errors.Wrap(err, "resolving path")
	}
	targetDir := absPath // for display purposes

	description := skillInitDescription
	if description == "" {
		description = prompt(scanner, "Description", "A helpful AI skill")
	}

	license := skillInitLicense
	if license == "" {
		license = prompt(scanner, "License", "MIT")
	}

	version := skillInitVersion
	if version == "" {
		version = prompt(scanner, "Version", "1.0.0")
	}

	author := skillInitAuthor
	if author == "" {
		author = prompt(scanner, "Author", "")
	}

	toolsStr := skillInitAllowedTools
	if toolsStr == "" {
		toolsStr = prompt(scanner, "Allowed Tools (comma separated)", "Read, Glob, Grep")
	}

	// Parse allowed tools
	var allowedTools []string
	for _, t := range strings.Split(toolsStr, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			allowedTools = append(allowedTools, t)
		}
	}

	// Determine optional directories
	knownDirs := []string{"docs", "tests", "bin", "data"}
	selectedDirs := make(map[string]bool)

	if skillInitDirs != "" {
		for _, d := range strings.Split(skillInitDirs, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				selectedDirs[d] = true
			}
		}
	} else {
		fmt.Println("\nOptional Directories:")
		for _, d := range knownDirs {
			if promptBool(scanner, fmt.Sprintf("Create '%s' directory?", d), false) {
				selectedDirs[d] = true
			}
		}
	}

	// Check if directory exists
	// If the user specified a path like "my-skill", we likely need to create it.
	// If they specified "." or an existing dir, we check for SKILL.md collision.

	// Create directory if it doesn't exist
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Creating directory %s...\n", targetDir)
		if err := os.MkdirAll(absPath, 0o755); err != nil {
			return errors.Wrap(err, "creating directory")
		}
	}

	skillFile := filepath.Join(absPath, "SKILL.md")
	if _, err := os.Stat(skillFile); err == nil {
		if !skillInitForce {
			fmt.Printf("Error: %s/SKILL.md already exists (use --force to overwrite)\n", targetDir)
			return errInitFailed
		}
	}

	fmt.Println("Writing SKILL.md...")

	// Build metadata
	metadata := make(map[string]string)
	if version != "" {
		metadata["version"] = version
	}
	if author != "" {
		metadata["author"] = author
	}

	// Generate content
	body := `# Instructions

You are a helpful assistant for [describe purpose].

## Guidelines

- Guideline 1
- Guideline 2

## Examples

When the user asks to [do something], you should...
`

	// Prepare struct for formatting
	meta := struct {
		Name          string            `yaml:"name"`
		Description   string            `yaml:"description"`
		License       string            `yaml:"license,omitempty"`
		Compatibility []string          `yaml:"compatibility,omitempty"`
		Metadata      map[string]string `yaml:"metadata,omitempty"`
		AllowedTools  []string          `yaml:"allowed-tools,omitempty"`
	}{
		Name:        name,
		Description: description,
		License:     license,
		Compatibility: []string{
			"claude-code >=1.0",
			"opencode >=0.1",
		},
		Metadata:     metadata,
		AllowedTools: allowedTools,
	}

	content, err := frontmatter.Format(meta, body)
	if err != nil {
		return errors.Wrap(err, "generating template")
	}

	if err := os.WriteFile(skillFile, content, 0o644); err != nil {
		return errors.Wrap(err, "writing SKILL.md")
	}

	// Create optional directories
	if len(selectedDirs) > 0 {
		fmt.Println("Creating optional directories...")
		for dir := range selectedDirs {
			fullPath := filepath.Join(absPath, dir)
			if err := os.MkdirAll(fullPath, 0o755); err != nil {
				return errors.Wrapf(err, "creating %s", dir)
			}
			keepFile := filepath.Join(fullPath, ".keep")
			if err := os.WriteFile(keepFile, []byte(""), 0o644); err != nil {
				return errors.Wrapf(err, "creating .keep in %s", dir)
			}
		}
	}

	// Print success message
	fmt.Printf("✓ Skill '%s' created at %s\n", name, skillFile)
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    1. Edit %s with your skill's instructions\n", skillFile)
	fmt.Printf("    2. Run: aix skill validate %s\n", targetDir)
	fmt.Printf("    3. Run: aix skill install %s\n", targetDir)

	return nil
}

func prompt(scanner *bufio.Scanner, label, def string) string {
	fmt.Printf("%s", label)
	if def != "" {
		fmt.Printf(" [%s]", def)
	}
	fmt.Print(": ")

	if !scanner.Scan() {
		return def
	}
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return def
	}
	return input
}

func promptBool(scanner *bufio.Scanner, label string, def bool) bool {
	defStr := "y/N"
	if def {
		defStr = "Y/n"
	}
	fmt.Printf("%s [%s]: ", label, defStr)

	if !scanner.Scan() {
		return def
	}
	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if input == "" {
		return def
	}
	return input == "y" || input == "yes"
}

// validateSkillName checks if a skill name conforms to the specification.
func validateSkillName(name string) error {
	if name == "" {
		return errors.New("skill name is required")
	}

	if len(name) > 64 {
		return errors.Newf("skill name must be at most 64 characters (got %d)", len(name))
	}

	if !skillNameRegex.MatchString(name) {
		return errors.New("skill name must be lowercase alphanumeric with hyphens, starting with a letter")
	}

	return nil
}
