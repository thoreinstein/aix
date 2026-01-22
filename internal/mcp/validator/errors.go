// Package validator provides validation for canonical MCP configurations.
package validator

import (
	"errors"
	"fmt"
)

// Sentinel errors for validation failures.
var (
	// ErrEmptyConfig indicates the config has no servers defined.
	ErrEmptyConfig = errors.New("config has no servers")

	// ErrMissingServerName indicates a server has no name.
	ErrMissingServerName = errors.New("server name is required")

	// ErrMissingCommand indicates a local server has no command.
	ErrMissingCommand = errors.New("local server requires command")

	// ErrMissingURL indicates a remote server has no URL.
	ErrMissingURL = errors.New("remote server requires URL")

	// ErrInvalidTransport indicates an unrecognized transport value.
	ErrInvalidTransport = errors.New("invalid transport value")

	// ErrInvalidPlatform indicates an unrecognized platform value.
	ErrInvalidPlatform = errors.New("invalid platform value")

	// ErrEmptyEnvKey indicates an environment variable has an empty key.
	ErrEmptyEnvKey = errors.New("environment variable key is empty")

	// ErrEmptyHeaderKey indicates an HTTP header has an empty key.
	ErrEmptyHeaderKey = errors.New("header key is empty")
)

// Severity indicates whether a validation issue is an error or warning.
type Severity int

const (
	// SeverityError indicates a validation issue that makes the config invalid.
	SeverityError Severity = iota

	// SeverityWarning indicates a validation issue that doesn't prevent usage
	// but may indicate a configuration problem.
	SeverityWarning
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return "unknown"
	}
}

// ValidationError represents a single validation issue with context.
type ValidationError struct {
	// ServerName identifies which server has the issue.
	// Empty for config-level issues.
	ServerName string

	// Field identifies which field has the issue.
	Field string

	// Message is a human-readable description of the problem.
	Message string

	// Severity indicates whether this is an error or warning.
	Severity Severity

	// Err is the underlying sentinel error, if any.
	Err error
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	prefix := "error"
	if e.Severity == SeverityWarning {
		prefix = "warning"
	}

	if e.ServerName != "" && e.Field != "" {
		return fmt.Sprintf("%s: server %q field %q: %s", prefix, e.ServerName, e.Field, e.Message)
	}
	if e.ServerName != "" {
		return fmt.Sprintf("%s: server %q: %s", prefix, e.ServerName, e.Message)
	}
	if e.Field != "" {
		return fmt.Sprintf("%s: field %q: %s", prefix, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", prefix, e.Message)
}

// Unwrap returns the underlying sentinel error.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// Is reports whether the error matches the target.
func (e *ValidationError) Is(target error) bool {
	return e.Err != nil && errors.Is(e.Err, target)
}

// HasErrors returns true if any of the validation errors have error severity.
func HasErrors(errs []*ValidationError) bool {
	for _, err := range errs {
		if err.Severity == SeverityError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any of the validation errors have warning severity.
func HasWarnings(errs []*ValidationError) bool {
	for _, err := range errs {
		if err.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

// Errors returns only the validation errors with error severity.
func Errors(errs []*ValidationError) []*ValidationError {
	var result []*ValidationError
	for _, err := range errs {
		if err.Severity == SeverityError {
			result = append(result, err)
		}
	}
	return result
}

// Warnings returns only the validation errors with warning severity.
func Warnings(errs []*ValidationError) []*ValidationError {
	var result []*ValidationError
	for _, err := range errs {
		if err.Severity == SeverityWarning {
			result = append(result, err)
		}
	}
	return result
}
