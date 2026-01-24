package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/command/parser"
	"github.com/thoreinstein/aix/internal/command/validator"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/platform/claude"
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
    aix command install  - Install the command
    aix command edit     - Edit the command definition
    aix command show     - Show command details`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runValidate(args[0], os.Stdout)
	},
}

// validateResult represents the JSON output structure.
type validateResult struct {
	Valid      bool     `json:"valid"`
	Command    *info    `json:"command,omitempty"`
	Errors     []string `json:"errors,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	ParseError string   `json:"parse_error,omitempty"`
	Path       string   `json:"path"`
	StrictMode bool     `json:"strict_mode"`
}

// info contains command metadata for display.
type info struct {
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

	// Parse the command file
	p := parser.New[*claude.Command]()
	cmd, parseErr := p.ParseFile(absPath)
	if parseErr != nil {
		result.ParseError = formatParseError(parseErr)
		return outputValidateResult(w, result)
	}

	result.Command = &info{
		Name:        (*cmd).Name,
		Description: (*cmd).Description,
	}

	// Validate
	v := validator.New()
	valResult := v.Validate(*cmd, absPath)

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
			return errors.Wrap(err, "encoding JSON")
		}
		if !result.Valid || result.ParseError != "" {
			return errValidationFailed
		}
		return nil
	}

	// Human-readable output
	if result.ParseError != "" {
		fmt.Fprintln(w, "✗ Command validation failed")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "  Parse error:")
		fmt.Fprintf(w, "    - %s\n", result.ParseError)
		return errValidationFailed
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
		return errValidationFailed
	}
	return nil
}

// formatParseError extracts a user-friendly message from parse errors.
func formatParseError(err error) string {
	var parseErr *parser.ParseError
	if errors.As(err, &parseErr) {
		if os.IsNotExist(parseErr.Err) {
			return "command file not found"
		}
		return parseErr.Err.Error()
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
