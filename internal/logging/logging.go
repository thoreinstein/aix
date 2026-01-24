package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
)

type contextKey struct{}

// FromContext returns the logger stored in the context, or the default logger if none is set.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// NewContext returns a new context with the given logger attached.
func NewContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// Format specifies the output format for log messages.
type Format string

const (
	// FormatText produces human-readable text output.
	FormatText Format = "text"
	// FormatJSON produces machine-readable JSON output.
	FormatJSON Format = "json"
)

// LevelTrace is the log level for trace messages.
// It is lower than Debug (-4) to allow for very verbose output.
const LevelTrace = slog.LevelDebug - 4

// LevelFromVerbosity returns the log level for the given verbosity count.
// 0 -> Warn
// 1 -> Info
// 2 -> Debug
// 3+ -> Trace
func LevelFromVerbosity(v int) slog.Level {
	switch {
	case v >= 3:
		return LevelTrace
	case v == 2:
		return slog.LevelDebug
	case v == 1:
		return slog.LevelInfo
	default:
		return slog.LevelWarn
	}
}

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
// It logs at Warn level in text format to stderr, matching the CLI's default verbosity.
func Default() *slog.Logger {
	return New(Config{
		Level:  slog.LevelWarn,
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
