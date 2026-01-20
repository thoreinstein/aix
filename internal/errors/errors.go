package errors

import (
	"errors"
	"fmt"
)

// Exit codes for CLI applications following Unix conventions.
const (
	// ExitSuccess indicates the command completed successfully.
	ExitSuccess = 0

	// ExitGeneral indicates a general error occurred.
	ExitGeneral = 1

	// ExitUsage indicates a usage or argument error.
	ExitUsage = 2

	// ExitMisuse indicates the command was used incorrectly.
	// This follows the BSD sysexits.h EX_USAGE convention.
	ExitMisuse = 64
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

// ExitError wraps an error with an exit code for CLI applications.
// It implements the error interface and supports unwrapping via errors.Unwrap.
type ExitError struct {
	// Err is the underlying error that caused the exit.
	Err error

	// Code is the exit code to return to the operating system.
	Code int
}

// NewExitError creates an ExitError with the given underlying error and exit code.
// If err is nil, the returned ExitError will have a nil Err field.
func NewExitError(err error, code int) *ExitError {
	return &ExitError{
		Err:  err,
		Code: code,
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
