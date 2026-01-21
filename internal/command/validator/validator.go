// Package validator provides validation for command structs.
package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/thoreinstein/aix/internal/command"
	"github.com/thoreinstein/aix/internal/command/parser"
)

const (
	// maxNameLength is the maximum allowed length for command names.
	maxNameLength = 64
)

// nameRegex validates command names: must start with a letter, lowercase alphanumeric,
// single hyphens allowed between segments, no start/end hyphen, no consecutive hyphens.
var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Level represents the severity of a validation issue.
type Level int

const (
	// Warning indicates a recommended but non-blocking issue.
	Warning Level = iota
	// Error indicates a blocking validation failure.
	Error
)

// Issue represents a validation problem found in a command.
type Issue struct {
	Level   Level
	Field   string
	Message string
	Value   string
}

func (i *Issue) Error() string {
	if i.Value == "" {
		return fmt.Sprintf("%s: %s", i.Field, i.Message)
	}
	return fmt.Sprintf("%s: %s (got %q)", i.Field, i.Message, i.Value)
}

// Result contains the validation results.
type Result struct {
	Errors   []Issue
	Warnings []Issue
}

// HasErrors returns true if there are any validation errors.
func (r *Result) HasErrors() bool {
	return len(r.Errors) > 0
}

// Validator validates command structs.
type Validator struct{}

// New creates a new Validator.
func New() *Validator {
	return &Validator{}
}

// Validate checks a command for compliance.
// The path parameter is used for name inference if the command has no name.
// Returns a Result containing errors and warnings.
func (v *Validator) Validate(cmd parser.Commandable, path string) *Result {
	result := &Result{}

	v.validateName(cmd.GetName(), path, result)

	return result
}

// validateName checks the name field for compliance.
func (v *Validator) validateName(name, path string, result *Result) {
	// If no name and no path to infer from, that's an error
	if name == "" {
		if path == "" {
			result.Errors = append(result.Errors, Issue{
				Level:   Error,
				Field:   "name",
				Message: "name is required when path is not provided for inference",
			})
			return
		}
		// Name will be inferred, just warn that it's missing from frontmatter
		inferred := command.InferName(path)
		if inferred == "" || inferred == "." {
			result.Errors = append(result.Errors, Issue{
				Level:   Error,
				Field:   "name",
				Message: "name is required and could not be inferred from path",
				Value:   path,
			})
			return
		}
		// Name can be inferred - this is fine, no warning needed
		return
	}

	// Validate provided name
	if len(name) > maxNameLength {
		result.Errors = append(result.Errors, Issue{
			Level:   Error,
			Field:   "name",
			Message: "name exceeds maximum length of 64 characters",
			Value:   name,
		})
	}

	if !nameRegex.MatchString(name) {
		msg := "name must start with a letter, be lowercase alphanumeric with single hyphens between segments"
		if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
			msg = "name cannot start or end with a hyphen"
		} else if strings.Contains(name, "--") {
			msg = "name cannot contain consecutive hyphens"
		} else if strings.ToLower(name) != name {
			msg = "name must be lowercase"
		}
		result.Errors = append(result.Errors, Issue{
			Level:   Error,
			Field:   "name",
			Message: msg,
			Value:   name,
		})
	}
}
