package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/agent/validator"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

var (
	validateStrict bool
	validateJSON   bool
)

var errValidationFailed = errors.New("validation failed")

func init() {
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false,
		"enable strict validation mode")
	validateCmd.Flags().BoolVar(&validateJSON, "json", false,
		"output results as JSON")
	Cmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate an agent file",
	Long: `Validate an agent definition file for required fields and format.

Checks for required fields, valid YAML frontmatter, and common issues.
Use --strict for additional checks beyond the basic requirements.

Exit codes:
  0 - Valid agent (warnings OK)
  1 - Invalid agent or validation errors

Examples:
  # Validate an agent file
  aix agent validate ./AGENT.md

  # Strict validation
  aix agent validate ./AGENT.md --strict

  # JSON output for CI/CD
  aix agent validate ./AGENT.md --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runValidate(args[0], os.Stdout)
	},
}

// validateResult represents the JSON output structure.
type validateResult struct {
	Valid      bool       `json:"valid"`
	Agent      *agentInfo `json:"agent,omitempty"`
	Errors     []string   `json:"errors,omitempty"`
	Warnings   []string   `json:"warnings,omitempty"`
	ParseError string     `json:"parse_error,omitempty"`
	Path       string     `json:"path"`
	StrictMode bool       `json:"strict_mode"`
}

// agentInfo contains agent metadata for display.
type agentInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func runValidate(path string, w io.Writer) error {
	// Resolve to absolute path for consistent error messages
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	result := &validateResult{
		Path:       absPath,
		StrictMode: validateStrict,
	}

	// Read file
	content, readErr := os.ReadFile(absPath)
	if readErr != nil {
		result.ParseError = formatReadError(readErr, absPath)
		return outputValidateResult(w, result)
	}

	// Check for empty file
	if len(bytes.TrimSpace(content)) == 0 {
		result.ParseError = "agent file is empty"
		return outputValidateResult(w, result)
	}

	// Parse frontmatter
	var meta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	body, parseErr := frontmatter.Parse(bytes.NewReader(content), &meta)
	if parseErr != nil {
		result.ParseError = fmt.Sprintf("invalid YAML frontmatter: %v", parseErr)
		return outputValidateResult(w, result)
	}

	// Create agent for validation (using claude.Agent as canonical)
	agent := &claude.Agent{
		Name:         meta.Name,
		Description:  meta.Description,
		Instructions: string(body),
	}

	result.Agent = &agentInfo{
		Name:        agent.Name,
		Description: agent.Description,
	}

	// Validate
	v := validator.New(validateStrict)
	valResult := v.Validate(agent, absPath)

	// Collect errors and warnings
	for _, e := range valResult.Errors {
		result.Errors = append(result.Errors, formatValidationIssue(e))
	}
	for _, warning := range valResult.Warnings {
		result.Warnings = append(result.Warnings, formatValidationIssue(warning))
	}

	result.Valid = !valResult.HasErrors()
	return outputValidateResult(w, result)
}

func outputValidateResult(w io.Writer, result *validateResult) error {
	if validateJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return fmt.Errorf("encoding JSON: %w", err)
		}
		if !result.Valid || result.ParseError != "" {
			return errValidationFailed
		}
		return nil
	}

	// Human-readable output
	if result.ParseError != "" {
		fmt.Fprintln(w, "✗ Agent validation failed")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  Parse error:")
		fmt.Fprintf(w, "    - %s\n", result.ParseError)
		return errValidationFailed
	}

	if !result.Valid {
		name := result.Agent.Name
		if name == "" {
			name = "(unknown)"
		}
		fmt.Fprintf(w, "✗ Agent '%s' is invalid\n", name)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  Errors:")
		for _, e := range result.Errors {
			fmt.Fprintf(w, "    - %s\n", e)
		}
	} else {
		fmt.Fprintf(w, "✓ Agent '%s' is valid\n", result.Agent.Name)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  Name:        %s\n", result.Agent.Name)
		if result.Agent.Description != "" {
			fmt.Fprintf(w, "  Description: %s\n", result.Agent.Description)
		}
	}

	// Always show warnings
	if len(result.Warnings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Fprintf(w, "    ⚠ %s\n", warning)
		}
	}

	if !result.Valid {
		return errValidationFailed
	}
	return nil
}

// formatReadError extracts a user-friendly message from read errors.
func formatReadError(err error, path string) string {
	if os.IsNotExist(err) {
		return "file not found: " + path
	}
	if os.IsPermission(err) {
		return "permission denied: " + path
	}
	return err.Error()
}

// formatValidationIssue formats a validation issue for display.
func formatValidationIssue(issue validator.Issue) string {
	if issue.Value == "" {
		return fmt.Sprintf("%s: %s", issue.Field, issue.Message)
	}
	return fmt.Sprintf("%s: %s (got %q)", issue.Field, issue.Message, issue.Value)
}
