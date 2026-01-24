package logging

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

// Format specifies the output format for log messages.
type Format string

const (
	// FormatText produces human-readable text output.
	FormatText Format = "text"
	// FormatJSON produces machine-readable JSON output.
	FormatJSON Format = "json"
)

// Config holds the configuration for creating a new logger.
type Config struct {
	// Level sets the minimum log level. Messages below this level are discarded.
	Level slog.Level
	// Format specifies the output format (text or JSON).
	Format Format
	// Output is where log messages are written. Defaults to os.Stderr if nil.
	Output io.Writer
}

// New creates a logger with the given configuration.
// If cfg.Output is nil, it defaults to os.Stderr.
// If cfg.Format is not recognized, it defaults to FormatText.
func New(cfg Config) *slog.Logger {
	output := cfg.Output
	if output == nil {
		output = os.Stderr
	}

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	var handler slog.Handler
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = NewHandler(output, opts)
	}

	return slog.New(handler)
}

// Default returns a sensible default logger configured for CLI use.
// It logs at Info level in text format to stderr.
func Default() *slog.Logger {
	return New(Config{
		Level:  slog.LevelInfo,
		Format: FormatText,
		Output: os.Stderr,
	})
}

// NewDiscard creates a logger that discards all output.
// Use this for quiet mode or when logging should be suppressed.
func NewDiscard() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// testWriter adapts testing.T to io.Writer for use with slog handlers.
type testWriter struct {
	t *testing.T
}

// Write implements io.Writer by logging to the test.
func (w *testWriter) Write(p []byte) (n int, err error) {
	w.t.Helper()
	// Trim trailing newline since t.Log adds its own
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.t.Log(msg)
	return len(p), nil
}

// ForTest creates a logger that writes to the test's log output.
// Log messages appear only when the test fails or when running with -v.
// The logger is configured at Debug level to capture all messages.
func ForTest(t *testing.T) *slog.Logger {
	t.Helper()
	return New(Config{
		Level:  slog.LevelDebug,
		Format: FormatText,
		Output: &testWriter{t: t},
	})
}
