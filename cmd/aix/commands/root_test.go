package commands

import (
	"log/slog"
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
