// Package logging provides structured logging for the aix CLI using slog.
//
// The package supports both text and JSON output formats, configurable log
// levels, and helpers for testing. All loggers are based on the standard
// library's [log/slog] package.
//
// # Basic Usage
//
//	logger := logging.New(logging.Config{
//		Level:  slog.LevelInfo,
//		Format: logging.FormatText,
//		Output: os.Stderr,
//	})
//	logger.Info("starting", "version", "1.0.0")
//
// # Testing
//
// For tests, use [ForTest] to capture log output via the testing framework:
//
//	func TestSomething(t *testing.T) {
//		logger := logging.ForTest(t)
//		// logs appear in test output on failure
//	}
//
// # Quiet Mode
//
// Use [NewDiscard] when log output should be suppressed entirely:
//
//	logger := logging.NewDiscard()
package logging
