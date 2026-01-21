package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/skill/parser"
	"github.com/thoreinstein/aix/internal/skill/validator"
)

var (
	validateStrict bool
	validateJSON   bool
)

func init() {
	skillValidateCmd.Flags().BoolVar(&validateStrict, "strict", false,
		"enable strict validation (validates allowed-tools syntax)")
	skillValidateCmd.Flags().BoolVar(&validateJSON, "json", false,
		"output results as JSON")
	skillCmd.AddCommand(skillValidateCmd)
}

var skillValidateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a skill file",
	Long: `Validate a skill file without installing it.

Parses and validates the skill at the given path against the Agent Skills
Specification. The path should be a directory containing a SKILL.md file.

Use --strict to also validate allowed-tools syntax.
Use --json for machine-readable output.

Exit codes:
  0 - Skill is valid
  1 - Skill validation failed`,
	Args: cobra.ExactArgs(1),
	RunE: runSkillValidate,
}

// validateResult represents the JSON output structure.
type validateResult struct {
	Valid      bool       `json:"valid"`
	Skill      *skillInfo `json:"skill,omitempty"`
	Errors     []string   `json:"errors,omitempty"`
	ParseError string     `json:"parse_error,omitempty"`
	Path       string     `json:"path"`
	StrictMode bool       `json:"strict_mode"`
}

// skillInfo contains skill metadata for display.
type skillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	License     string `json:"license,omitempty"`
}

func runSkillValidate(_ *cobra.Command, args []string) error {
	skillPath := args[0]

	// Resolve to absolute path for consistent error messages
	absPath, err := filepath.Abs(skillPath)
	if err != nil {
		absPath = skillPath
	}

	// Construct path to SKILL.md
	skillFile := filepath.Join(absPath, "SKILL.md")

	// Parse the skill
	p := parser.New()
	skill, parseErr := p.ParseFile(skillFile)

	if parseErr != nil {
		return outputParseError(absPath, parseErr)
	}

	// Validate the skill
	v := validator.New(validator.WithStrict(validateStrict))
	validationErrs := v.ValidateWithPath(skill, skillFile)

	if len(validationErrs) > 0 {
		return outputValidationErrors(absPath, validationErrs)
	}

	// Success
	return outputSuccess(absPath, skill)
}

func outputParseError(path string, err error) error {
	if validateJSON {
		result := validateResult{
			Valid:      false,
			Path:       path,
			StrictMode: validateStrict,
			ParseError: formatParseError(err),
		}
		return outputValidateJSON(result)
	}

	fmt.Println("✗ Skill validation failed")
	fmt.Println()
	fmt.Printf("  Parse error:\n")
	fmt.Printf("    - %s\n", formatParseError(err))
	return errValidationFailed
}

func outputValidationErrors(path string, errs []error) error {
	if validateJSON {
		errStrings := make([]string, len(errs))
		for i, e := range errs {
			errStrings[i] = formatValidationError(e)
		}
		result := validateResult{
			Valid:      false,
			Path:       path,
			StrictMode: validateStrict,
			Errors:     errStrings,
		}
		return outputValidateJSON(result)
	}

	fmt.Println("✗ Skill validation failed")
	fmt.Println()
	fmt.Println("  Errors:")
	for _, e := range errs {
		fmt.Printf("    - %s\n", formatValidationError(e))
	}
	return errValidationFailed
}

func outputSuccess(path string, skill *claude.Skill) error {
	if validateJSON {
		result := validateResult{
			Valid:      true,
			Path:       path,
			StrictMode: validateStrict,
			Skill: &skillInfo{
				Name:        skill.Name,
				Description: skill.Description,
				License:     skill.License,
			},
		}
		return outputValidateJSON(result)
	}

	fmt.Printf("✓ Skill '%s' is valid\n", skill.Name)
	fmt.Println()
	fmt.Printf("  Name:        %s\n", skill.Name)
	fmt.Printf("  Description: %s\n", skill.Description)
	if skill.License != "" {
		fmt.Printf("  License:     %s\n", skill.License)
	}
	return nil
}

func outputValidateJSON(result validateResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
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
		// Check for file not found
		if os.IsNotExist(parseErr.Err) {
			return "SKILL.md not found in directory"
		}
		return parseErr.Err.Error()
	}
	return err.Error()
}

// formatValidationError extracts a user-friendly message from validation errors.
func formatValidationError(err error) string {
	var valErr *validator.ValidationError
	if errors.As(err, &valErr) {
		return fmt.Sprintf("%s: %s", valErr.Field, valErr.Message)
	}
	return err.Error()
}

// errValidationFailed is a sentinel error that signals non-zero exit.
var errValidationFailed = errors.New("validation failed")
