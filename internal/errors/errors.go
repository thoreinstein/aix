package errors

import (
	"errors"
	"fmt"

	pkgerrors "github.com/cockroachdb/errors"
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
	ErrMissingName = pkgerrors.New("name is required")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = pkgerrors.New("resource not found")

	// ErrInvalidConfig indicates configuration validation failed.
	ErrInvalidConfig = pkgerrors.New("invalid configuration")

	// ErrInvalidToolSyntax indicates a malformed tool permission string.
	ErrInvalidToolSyntax = pkgerrors.New("invalid tool syntax")

	// ErrUnknownTool indicates the tool is not in the known tool list.
	ErrUnknownTool = pkgerrors.New("unknown tool")

	// ErrPermissionDenied indicates the operation is not permitted.
	ErrPermissionDenied = pkgerrors.New("permission denied")

	// ErrNotImplemented indicates the requested feature is not yet implemented.
	ErrNotImplemented = pkgerrors.New("not implemented")

	// ErrNotSupported indicates the requested operation is not supported on the platform.
	ErrNotSupported = pkgerrors.New("not supported")

	// ErrTimeout indicates the operation timed out.
	ErrTimeout = pkgerrors.New("operation timed out")

	// ErrInternal indicates an unexpected internal error.
	ErrInternal = pkgerrors.New("internal error")

	// ErrValidation indicates a validation failure.
	ErrValidation = pkgerrors.New("validation failed")
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

// Newf creates a new error with a formatted message.
// This is a passthrough to cockroachdb/errors.Newf.
func Newf(format string, args ...interface{}) error {
	return pkgerrors.Newf(format, args...)
}

// Wrap wraps an error with a message.
// This is a passthrough to cockroachdb/errors.Wrap.
func Wrap(err error, msg string) error {
	return pkgerrors.Wrap(err, msg)
}

// Wrapf wraps an error with a formatted message.
// This is a passthrough to cockroachdb/errors.Wrapf.
func Wrapf(err error, format string, args ...interface{}) error {
	return pkgerrors.Wrapf(err, format, args...)
}

// Is reports whether any error in err's chain matches target.
// This is a passthrough to cockroachdb/errors.Is.
func Is(err, target error) bool {
	return pkgerrors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
// This is a passthrough to cockroachdb/errors.As.
func As(err error, target interface{}) bool {
	return pkgerrors.As(err, target)
}

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if every value in errs is nil.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each element.
//
// This is a passthrough to standard library errors.Join.
func Join(errs ...error) error {
	return errors.Join(errs...)
}
