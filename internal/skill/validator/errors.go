package validator

import "fmt"

// ValidationError represents a validation failure for a specific field.
type ValidationError struct {
	Field   string
	Message string
	Value   string
	Context map[string]string
}

func (e *ValidationError) Error() string {
	if e.Value == "" {
		return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s %q: %s", e.Field, e.Value, e.Message)
}
