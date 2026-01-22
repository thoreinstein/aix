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
//   - ExitUser (1): User-related error (invalid input, configuration, etc.)
//   - ExitSystem (2): System-related error (I/O, network, permissions, etc.)
//
// # ExitError
//
// [ExitError] wraps an underlying error with an exit code and optional suggestion
// for CLI applications. It supports error unwrapping via [errors.Unwrap] and
// [errors.As]:
//
//	err := aixerrors.NewUserError(aixerrors.ErrInvalidConfig, "Check your config file")
//	var exitErr *aixerrors.ExitError
//	if errors.As(err, &exitErr) {
//	    if exitErr.Suggestion != "" {
//	        fmt.Println("Suggestion:", exitErr.Suggestion)
//	    }
//	    os.Exit(exitErr.Code)
//	}
package errors
