package commands

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/logging"
)

func TestSetupLogging_VerbosityFlags(t *testing.T) {
	// Save/Restore original state
	origVerbosity := verbosity
	defer func() { verbosity = origVerbosity }()

	tests := []struct {
		name      string
		verbosity int
		wantLevel slog.Level
	}{
		{"default (0)", 0, slog.LevelWarn},
		{"verbose (1)", 1, slog.LevelInfo},
		{"debug (2)", 2, slog.LevelDebug},
		{"trace (3)", 3, logging.LevelTrace},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verbosity = tt.verbosity
			if err := setupLogging(rootCmd); err != nil {
				t.Fatalf("setupLogging failed: %v", err)
			}

			logger := slog.Default()
			if !logger.Enabled(t.Context(), tt.wantLevel) {
				t.Errorf("expected level %v to be enabled", tt.wantLevel)
			}
			if tt.wantLevel > logging.LevelTrace {
				shouldBeDisabled := tt.wantLevel - 4
				if logger.Enabled(t.Context(), shouldBeDisabled) {
					t.Errorf("expected level %v to be disabled", shouldBeDisabled)
				}
			}
		})
	}
}

func TestSetupLogging_EnvVar(t *testing.T) {
	origVerbosity := verbosity
	defer func() { verbosity = origVerbosity }()

	tests := []struct {
		name      string
		envVal    string
		wantLevel slog.Level
	}{
		{"AIX_DEBUG=1", "1", slog.LevelDebug},
		{"AIX_DEBUG=true", "true", slog.LevelDebug},
		{"AIX_DEBUG=2", "2", logging.LevelTrace},
		{"AIX_DEBUG=0", "0", slog.LevelWarn},
		{"AIX_DEBUG=unknown", "foo", slog.LevelWarn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verbosity = 0
			t.Setenv("AIX_DEBUG", tt.envVal)

			if err := setupLogging(rootCmd); err != nil {
				t.Fatalf("setupLogging failed: %v", err)
			}

			logger := slog.Default()
			if !logger.Enabled(t.Context(), tt.wantLevel) {
				t.Errorf("expected level %v to be enabled", tt.wantLevel)
			}

			if tt.wantLevel == slog.LevelDebug {
				if logger.Enabled(t.Context(), logging.LevelTrace) {
					t.Error("expected Trace level to be disabled when AIX_DEBUG=1")
				}
			}
		})
	}
}

func TestSetupLogging_FlagPrecedence(t *testing.T) {
	origVerbosity := verbosity
	defer func() { verbosity = origVerbosity }()

	t.Setenv("AIX_DEBUG", "2")
	verbosity = 1

	if err := setupLogging(rootCmd); err != nil {
		t.Fatalf("setupLogging failed: %v", err)
	}

	logger := slog.Default()
	if !logger.Enabled(t.Context(), slog.LevelInfo) {
		t.Error("expected Info level to be enabled")
	}
	if logger.Enabled(t.Context(), slog.LevelDebug) {
		t.Error("expected Debug level to be disabled (flag should override env var)")
	}
}

func TestSetupLogging_Quiet(t *testing.T) {
	origQuiet := quiet
	origVerbosity := verbosity
	defer func() {
		quiet = origQuiet
		verbosity = origVerbosity
	}()

	quiet = true
	verbosity = 0

	if err := setupLogging(rootCmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger := slog.Default()
	if !logger.Enabled(t.Context(), slog.LevelError) {
		t.Error("expected Error level to be enabled")
	}
	if logger.Enabled(t.Context(), slog.LevelWarn) {
		t.Error("expected Warn level to be disabled")
	}
}

func TestSetupLogging_QuietMutualExclusion(t *testing.T) {
	origVerbosity := verbosity
	origQuiet := quiet
	defer func() {
		verbosity = origVerbosity
		quiet = origQuiet
	}()

	verbosity = 1
	quiet = true

	if err := setupLogging(rootCmd); err == nil {
		t.Error("expected error when both quiet and verbose are set")
	}
}

func TestSetupLogging_LogFile(t *testing.T) {
	origLogFile := logFile
	origVerbosity := verbosity
	defer func() {
		logFile = origLogFile
		verbosity = origVerbosity
	}()

	tmpDir := t.TempDir()
	logFile = filepath.Join(tmpDir, "test.log")
	verbosity = 1 // Info level

	if err := setupLogging(rootCmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slog.Info("test message", "foo", "bar")

	// Verify file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), `"msg":"test message"`) {
		t.Errorf("log file missing message: %s", string(content))
	}
	if !strings.Contains(string(content), `"foo":"bar"`) {
		t.Errorf("log file missing attribute: %s", string(content))
	}
	if !strings.Contains(string(content), `"level":"INFO"`) {
		t.Errorf("log file missing level: %s", string(content))
	}
}

func TestSetupLogging_LogFormat(t *testing.T) {
	origLogFormat := logFormat
	origVerbosity := verbosity
	defer func() {
		logFormat = origLogFormat
		verbosity = origVerbosity
	}()

	tests := []struct {
		format string
		want   string // partial content to check
	}{
		{"text", "INFO  test message foo=bar"},
		{"json", `"level":"INFO","msg":"test message","foo":"bar"`},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			logFormat = tt.format
			verbosity = 1 // Info

			var buf strings.Builder
			rootCmd.SetErr(&buf)
			defer rootCmd.SetErr(os.Stderr)

			if err := setupLogging(rootCmd); err != nil {
				t.Fatalf("setupLogging failed: %v", err)
			}

			slog.Info("test message", "foo", "bar")

			output := buf.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("expected output to contain %q, got %q", tt.want, output)
			}
		})
	}
}

func TestSetupLogging_ContextInjection(t *testing.T) {
	if err := setupLogging(rootCmd); err != nil {
		t.Fatalf("setupLogging failed: %v", err)
	}

	ctx := rootCmd.Context()
	logger := logging.FromContext(ctx)
	if logger == nil {
		t.Fatal("logger not found in context")
	}

	// We can't strictly compare slog.Logger directly as it's a struct with pointers,
	// but we can check if it works.
	logger.Info("context logging works")
}
