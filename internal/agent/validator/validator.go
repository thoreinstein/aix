// Package validator provides validation for agent structs.
package validator

import "fmt"

// Level represents the severity of a validation issue.
type Level int

const (
	// Warning indicates a recommended but non-blocking issue.
	Warning Level = iota
	// Error indicates a blocking validation failure.
	Error
)

// Agentable is implemented by agent types that can be validated.
type Agentable interface {
	GetName() string
	GetDescription() string
	GetInstructions() string
}

// Issue represents a validation problem found in an agent.
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

// Validator validates agent structs.
type Validator struct {
	strict bool
}

// New creates a new Validator.
// When strict is true, additional warnings are generated for missing optional fields.
func New(strict bool) *Validator {
	return &Validator{strict: strict}
}

// Validate checks an agent for compliance.
// The path parameter is used for context in error messages.
// Returns a Result containing errors and warnings.
func (v *Validator) Validate(agent Agentable, path string) *Result {
	result := &Result{}

	v.validateName(agent.GetName(), result)
	v.validateDescription(agent.GetDescription(), result)
	v.validateInstructions(agent.GetInstructions(), agent.GetName(), result)

	return result
}

// validateName checks the name field for compliance.
func (v *Validator) validateName(name string, result *Result) {
	if name == "" {
		result.Errors = append(result.Errors, Issue{
			Level:   Error,
			Field:   "name",
			Message: "name is required",
		})
	}
}

// validateDescription checks the description field.
// In strict mode, missing description generates a warning.
func (v *Validator) validateDescription(description string, result *Result) {
	if v.strict && description == "" {
		result.Warnings = append(result.Warnings, Issue{
			Level:   Warning,
			Field:   "description",
			Message: "description is recommended for agent discoverability",
		})
	}
}

// validateInstructions checks the instructions (body content).
// An empty file (no instructions and no meaningful frontmatter) is an error.
func (v *Validator) validateInstructions(instructions, name string, result *Result) {
	// Empty file check: no instructions AND no name means completely empty
	if instructions == "" && name == "" {
		result.Errors = append(result.Errors, Issue{
			Level:   Error,
			Field:   "instructions",
			Message: "agent file is empty (no frontmatter and no body content)",
		})
	}
}
