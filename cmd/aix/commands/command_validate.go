package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/command/parser"
	"github.com/thoreinstein/aix/internal/command/validator"
	"github.com/thoreinstein/aix/internal/platform/claude"
)

var (
	commandValidateStrict bool
	commandValidateJSON   bool
)

var errCommandValidationFailed = errors.New("validation failed")

func init() {
	commandValidateCmd.Flags().BoolVar(&commandValidateStrict, "strict", false,
		"enable strict validation mode")
	commandValidateCmd.Flags().BoolVar(&commandValidateJSON, "json", false,
		"output results as JSON")
	commandCmd.AddCommand(commandValidateCmd)
}

var commandValidateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a slash command file",
	Long: `Validate a slash command definition file against the command specification.

Checks for required fields, valid syntax, and common issues. Use --strict for
additional checks beyond the basic requirements.

Exit codes:
  0 - Valid command
  1 - Invalid command or validation errors`,
	Example: `  # Validate a command file
  aix command validate ./review.md

  # Strict validation
  aix command validate ./review.md --strict

  # JSON output for CI/CD
  aix command validate ./review.md --json

  See Also:
    aix command init     - Create a new command
    aix command install  - Install a command`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runCommandValidate(args[0], os.Stdout)
	},
}

// commandValidateResult represents the JSON output structure.
type commandValidateResult struct {
	Valid      bool         `json:"valid"`
	Command    *commandInfo `json:"command,omitempty"`
	Errors     []string     `json:"errors,omitempty"`
	Warnings   []string     `json:"warnings,omitempty"`
	ParseError string       `json:"parse_error,omitempty"`
	Path       string       `json:"path"`
	StrictMode bool         `json:"strict_mode"`
}

// commandInfo contains command metadata for display.
type commandInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func runCommandValidate(path string, w io.Writer) error {
	// Resolve to absolute path for consistent error messages
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	result := &commandValidateResult{
		Path:       absPath,
		StrictMode: commandValidateStrict,
	}

	// Parse the command file
	p := parser.New[*claude.Command]()
	cmd, parseErr := p.ParseFile(absPath)
	if parseErr != nil {
		result.ParseError = formatCommandParseError(parseErr)
		return outputCommandValidateResult(w, result)
	}

	result.Command = &commandInfo{
		Name:        (*cmd).Name,
		Description: (*cmd).Description,
	}

	// Validate
	v := validator.New()
	valResult := v.Validate(*cmd, absPath)

	// Collect errors and warnings
	for _, e := range valResult.Errors {
		result.Errors = append(result.Errors, formatCommandValidationIssue(e))
	}
	for _, warning := range valResult.Warnings {
		result.Warnings = append(result.Warnings, formatCommandValidationIssue(warning))
	}

	result.Valid = !valResult.HasErrors()
	return outputCommandValidateResult(w, result)
}

func outputCommandValidateResult(w io.Writer, result *commandValidateResult) error {
	if commandValidateJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return errors.Wrap(err, "encoding JSON")
		}
		if !result.Valid || result.ParseError != "" {
			return errCommandValidationFailed
		}
		return nil
	}

	// Human-readable output
	if result.ParseError != "" {
		fmt.Fprintln(w, "✗ Command validation failed")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  Parse error:")
		fmt.Fprintf(w, "    - %s\n", result.ParseError)
		return errCommandValidationFailed
	}

	if !result.Valid {
		fmt.Fprintf(w, "✗ Command '/%s' is invalid\n", result.Command.Name)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  Errors:")
		for _, e := range result.Errors {
			fmt.Fprintf(w, "    - %s\n", e)
		}
	} else {
		fmt.Fprintf(w, "✓ Command '/%s' is valid\n", result.Command.Name)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  Name:        %s\n", result.Command.Name)
		if result.Command.Description != "" {
			fmt.Fprintf(w, "  Description: %s\n", result.Command.Description)
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
		return errCommandValidationFailed
	}
	return nil
}

// formatCommandParseError extracts a user-friendly message from parse errors.
func formatCommandParseError(err error) string {
	var parseErr *parser.ParseError
	if errors.As(err, &parseErr) {
		if os.IsNotExist(parseErr.Err) {
			return "command file not found"
		}
		return parseErr.Err.Error()
	}
	return err.Error()
}

// formatCommandValidationIssue formats a validation issue for display.
func formatCommandValidationIssue(issue validator.Issue) string {
	if issue.Value == "" {
		return fmt.Sprintf("%s: %s", issue.Field, issue.Message)
	}
	return fmt.Sprintf("%s: %s (got %q)", issue.Field, issue.Message, issue.Value)
}
