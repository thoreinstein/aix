package agent

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/pkg/frontmatter"
)

var (
	initName        string
	initDescription string
	initModel       string
	initForce       bool
)

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "agent name (required)")
	initCmd.Flags().StringVarP(&initDescription, "description", "d", "", "short description")
	initCmd.Flags().StringVar(&initModel, "model", "", "AI model to use")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "overwrite existing file")
	Cmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create a new agent interactively",
	Long: `Create a new agent directory with a scaffolded AGENT.md file.

If [path] is provided, the agent is created in that directory.
If no path is provided, a directory named after the agent is created.

The command is interactive and will prompt for details unless they are
provided via flags.

Examples:
  # Interactive creation
  aix agent init

  # Create in specific directory
  aix agent init my-agent

  # Non-interactive creation
  aix agent init my-agent --name my-agent --description "Review code"

  # Specify model
  aix agent init review --name review --model claude-3-5-sonnet`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

// nameRegex validates agent names.
// Must start with a lowercase letter, followed by lowercase alphanumeric characters,
// optionally followed by hyphen-separated segments. No leading, trailing, or
// consecutive hyphens are allowed.
var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// nameSanitizer matches characters that are not allowed in an agent name.
var nameSanitizer = regexp.MustCompile(`[^a-z0-9-]+`)

// errInitFailed is a sentinel error that signals non-zero exit.
var errInitFailed = errors.New("agent initialization failed")

func sanitizeDefaultName(name string) string {
	// Normalize to lowercase and replace invalid characters with hyphens.
	sanitized := strings.ToLower(name)
	sanitized = nameSanitizer.ReplaceAllString(sanitized, "-")
	sanitized = strings.Trim(sanitized, "-")

	// Fallback to a safe default if the result is empty or still invalid.
	if sanitized == "" || !nameRegex.MatchString(sanitized) {
		return "new-agent"
	}

	return sanitized
}

// validateName checks if an agent name conforms to the specification.
func validateName(name string) error {
	if name == "" {
		return errors.New("agent name is required")
	}

	if len(name) > 64 {
		return fmt.Errorf("agent name must be at most 64 characters (got %d)", len(name))
	}

	if !nameRegex.MatchString(name) {
		return errors.New("agent name must be lowercase alphanumeric with hyphens, starting with a letter")
	}

	return nil
}

func runInit(_ *cobra.Command, args []string) error {
	// Determine default name
	defaultName := "my-agent"
	if len(args) > 0 {
		defaultName = sanitizeDefaultName(filepath.Base(args[0]))
	}

	// Interactive prompts
	scanner := bufio.NewScanner(os.Stdin)

	name := initName
	if name == "" {
		name = prompt(scanner, "Agent Name", defaultName)
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
		// User provided no path, create subdirectory with agent name
		absPath, err = filepath.Abs(name)
	}
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}
	targetDir := absPath // for display purposes

	description := initDescription
	if description == "" {
		description = prompt(scanner, "Description", "A helpful AI agent")
	}

	model := initModel
	if model == "" {
		model = prompt(scanner, "Model (optional)", "")
	}

	// Create directory if it doesn't exist
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("Creating directory %s...\n", targetDir)
		if err := os.MkdirAll(absPath, 0o755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
	}

	agentFile := filepath.Join(absPath, "AGENT.md")
	if _, err := os.Stat(agentFile); err == nil {
		if !initForce {
			fmt.Printf("Error: file already exists: %s. Use --force to overwrite.\n", agentFile)
			return errInitFailed
		}
	}

	fmt.Println("Writing agent file...")

	// Generate title from name (capitalize first letter of each word)
	title := formatTitle(name)

	// Generate body content
	body := fmt.Sprintf(`# %s Agent

%s

## Instructions

Your agent instructions go here.

<!-- Add your agent's system prompt below -->
`, title, description)

	// Prepare struct for formatting
	type agentMeta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Model       string `yaml:"model,omitempty"`
	}

	meta := agentMeta{
		Name:        name,
		Description: description,
		Model:       model,
	}

	content, err := frontmatter.Format(meta, body)
	if err != nil {
		return fmt.Errorf("generating template: %w", err)
	}

	if err := os.WriteFile(agentFile, content, 0o644); err != nil {
		return fmt.Errorf("writing agent file: %w", err)
	}

	// Print success message
	fmt.Printf("[OK] Agent '%s' created at %s\n", name, agentFile)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s with your agent's instructions\n", agentFile)
	fmt.Printf("  2. Run: aix agent validate %s\n", targetDir)
	fmt.Printf("  3. Run: aix agent install %s\n", targetDir)

	return nil
}

// prompt reads user input with a default value.
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

// formatTitle converts a hyphenated name to a title case string.
// e.g., "my-agent" -> "My Agent"
func formatTitle(name string) string {
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}
