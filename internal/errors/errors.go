package errors

import (
	"errors"
	"fmt"
)

// Exit codes for CLI applications.
const (
	// ExitSuccess indicates the command completed successfully.
	ExitSuccess = 0

	// ExitUser indicates a user-related error (invalid input, configuration, etc.).
	ExitUser = 1

	// ExitSystem indicates a system-related error (I/O, network, permissions, etc.).
	ExitSystem = 2
)

// Sentinel errors for common failure conditions.
var (
	// ErrMissingName indicates a required name field is missing.
	ErrMissingName = errors.New("name is required")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidConfig indicates configuration validation failed.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrInvalidToolSyntax indicates a malformed tool permission string.
	ErrInvalidToolSyntax = errors.New("invalid tool syntax")

	// ErrUnknownTool indicates the tool is not in the known tool list.
	ErrUnknownTool = errors.New("unknown tool")
)

// ExitError wraps an error with an exit code and optional suggestion for CLI applications.
// It implements the error interface and supports unwrapping via errors.Unwrap.
type ExitError struct {
	// Err is the underlying error that caused the exit.
	Err error

	// Code is the exit code to return to the operating system.
	Code int

	// Suggestion is an optional actionable suggestion for the user.
	Suggestion string
}

// NewExitError creates an ExitError with the given underlying error and exit code.
// If err is nil, the returned ExitError will have a nil Err field.
func NewExitError(err error, code int) *ExitError {
	return &ExitError{
		Err:  err,
		Code: code,
	}
}

// NewExitErrorWithSuggestion creates an ExitError with a suggestion.
func NewExitErrorWithSuggestion(err error, code int, suggestion string) *ExitError {
	return &ExitError{
		Err:        err,
		Code:       code,
		Suggestion: suggestion,
	}
}

// NewUserError creates an ExitError with ExitUser code and a suggestion.
func NewUserError(err error, suggestion string) *ExitError {
	return &ExitError{
		Err:        err,
		Code:       ExitUser,
		Suggestion: suggestion,
	}
}

// NewSystemError creates an ExitError with ExitSystem code and a suggestion.
func NewSystemError(err error, suggestion string) *ExitError {
	return &ExitError{
		Err:        err,
		Code:       ExitSystem,
		Suggestion: suggestion,
	}
}

// NewConfigError creates an ExitError with ExitUser code and a standard suggestion.
func NewConfigError(err error) *ExitError {
	return &ExitError{
		Err:        err,
		Code:       ExitUser,
		Suggestion: "Run: aix doctor",
	}
}

// Error returns the error message from the underlying error.
// If the underlying error is nil, it returns a generic message with the exit code.
func (e *ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit code %d", e.Code)
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error, enabling errors.Is and errors.As
// to examine the error chain.
func (e *ExitError) Unwrap() error {
	return e.Err
}
