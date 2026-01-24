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
//
// # Configuration Precedence
//
// Logging is configured in the following order of precedence:
//
//  1. CLI Flags: --verbose/-v, --quiet/-q, --log-format, --log-file.
//  2. Environment Variables: AIX_DEBUG=1 (Debug), AIX_DEBUG=2 (Trace).
//  3. Defaults: Warn level, Text format, Output to Stderr.
//
// Note that --log-file always uses JSON format regardless of --log-format.
// --log-format only controls the primary output (Stderr).
package logging
