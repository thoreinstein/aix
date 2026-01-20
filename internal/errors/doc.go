// Package errors provides error handling conventions for the aix CLI.
//
// This package defines sentinel errors for common failure conditions,
// an ExitError type for CLI exit code handling, and exit code constants
// following standard Unix conventions.
//
// # Sentinel Errors
//
// Sentinel errors allow callers to check for specific error conditions
// using [errors.Is]:
//
//	if errors.Is(err, aixerrors.ErrNotFound) {
//	    // handle not found case
//	}
//
// # Exit Codes
//
// The package defines standard exit codes for CLI applications:
//
//   - ExitSuccess (0): Command completed successfully
//   - ExitGeneral (1): General error
//   - ExitUsage (2): Usage or argument error
//   - ExitMisuse (64): Command used incorrectly
//
// # ExitError
//
// [ExitError] wraps an underlying error with an exit code for CLI
// applications. It supports error unwrapping via [errors.Unwrap] and
// [errors.As]:
//
//	err := aixerrors.NewExitError(aixerrors.ErrInvalidConfig, aixerrors.ExitUsage)
//	var exitErr *aixerrors.ExitError
//	if errors.As(err, &exitErr) {
//	    os.Exit(exitErr.Code)
//	}
package errors
