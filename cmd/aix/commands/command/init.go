package command

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
	initName        string
	initDescription string
	initModel       string
	initAgent       string
	initForce       bool
)

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "command name (required)")
	initCmd.Flags().StringVarP(&initDescription, "description", "d", "", "short description")
	initCmd.Flags().StringVar(&initModel, "model", "", "AI model to use")
	initCmd.Flags().StringVar(&initAgent, "agent", "task", "agent type")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "overwrite existing directory")
	Cmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create a new slash command interactively",
	Long: `Create a new slash command directory with a scaffolded command.md file.

If [path] is provided, the command is created in that directory.
If no path is provided, a directory named after the command is created.

The command is interactive and will prompt for details unless they are
provided via flags.`,
	Example: `  # Interactive creation
  aix command init

  # Create in specific directory
  aix command init my-command

  # Non-interactive creation
  aix command init my-command --name my-command --description "Review code"

  # Specify model and agent
  aix command init review --name review --model claude-3-5-sonnet --agent task

  See Also:
    aix command install  - Install the created command
    aix command edit     - Edit the command definition
    aix command validate - Validate the command file`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

// nameRegex validates command names.
// Must start with a lowercase letter, followed by lowercase alphanumeric characters,
// optionally followed by hyphen-separated segments. No leading, trailing, or
// consecutive hyphens are allowed.
var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// nameSanitizer matches characters that are not allowed in a command name.
var nameSanitizer = regexp.MustCompile(`[^a-z0-9-]+`)

// errInitFailed is a sentinel error that signals non-zero exit.
var errInitFailed = errors.New("command initialization failed")

func sanitizeDefaultName(name string) string {
	// Normalize to lowercase and replace invalid characters with hyphens.
	sanitized := strings.ToLower(name)
	sanitized = nameSanitizer.ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")

	// Fallback to a safe default if the result is empty or still invalid.
	if sanitized == "" || !nameRegex.MatchString(sanitized) {
		return "new-command"
	}

	return sanitized
}

// validateName checks if a command name conforms to the specification.
func validateName(name string) error {
	if name == "" {
		return errors.New("command name is required")
	}

	if len(name) > 64 {
		return errors.Newf("command name must be at most 64 characters (got %d)", len(name))
	}

	if !nameRegex.MatchString(name) {
		return errors.New("command name must be lowercase alphanumeric with hyphens, starting with a letter")
	}

	return nil
}

func runInit(_ *cobra.Command, args []string) error {
	// Determine default name
	defaultName := "my-command"
	if len(args) > 0 {
		defaultName = sanitizeDefaultName(filepath.Base(args[0]))
	}

	// Interactive prompts
	scanner := bufio.NewScanner(os.Stdin)

	name := initName
	if name == "" {
		name = prompt(scanner, "Command Name", defaultName)
	}

	// Validate name immediately
	if err := validateName(name); err != nil {
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
		// User provided no path, create subdirectory with command name
		absPath, err = filepath.Abs(name)
	}
	if err != nil {
		return errors.Wrap(err, "resolving path")
	}
	targetDir := absPath // for display purposes

	description := initDescription
	if description == "" {
		description = prompt(scanner, "Description", "A helpful slash command")
	}

	model := initModel
	if model == "" {
		model = prompt(scanner, "Model (optional)", "")
	}

	agent := initAgent
	if agent == "" {
		agent = prompt(scanner, "Agent", "task")
	}

	// Create directory if it doesn't exist
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Creating directory %s...\n", targetDir)
		if err := os.MkdirAll(absPath, 0o755); err != nil {
			return errors.Wrap(err, "creating directory")
		}
	}

	commandFile := filepath.Join(absPath, "command.md")
	if _, err := os.Stat(commandFile); err == nil {
		if !initForce {
			fmt.Printf("Error: %s/command.md already exists (use --force to overwrite)\n", targetDir)
			return errInitFailed
		}
	}

	fmt.Println("Writing command.md...")

	// Generate title from name (capitalize first letter of each word)
	title := formatTitle(name)

	// Generate body content
	body := fmt.Sprintf(`# %s Command

%s

## Instructions

Your command instructions go here.

<!-- Add your command logic, prompts, and workflows below -->
`, title, description)

	// Prepare struct for formatting
	type commandMeta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Model       string `yaml:"model,omitempty"`
		Agent       string `yaml:"agent"`
	}

	meta := commandMeta{
		Name:        name,
		Description: description,
		Model:       model,
		Agent:       agent,
	}

	content, err := frontmatter.Format(meta, body)
	if err != nil {
		return errors.Wrap(err, "generating template")
	}

	if err := os.WriteFile(commandFile, content, 0o644); err != nil {
		return errors.Wrap(err, "writing command.md")
	}

	// Print success message
	fmt.Printf("Command '%s' created at %s\n", name, commandFile)
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    1. Edit %s with your command's instructions\n", commandFile)
	fmt.Printf("    2. Run: aix command install %s\n", targetDir)

	return nil
}

// formatTitle converts a hyphenated name to a title case string.
// e.g., "my-command" -> "My Command"
func formatTitle(name string) string {
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// prompt displays a label with optional default value and reads input from scanner.
// If the user enters nothing, returns the default value.
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
