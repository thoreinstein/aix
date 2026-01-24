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
			setupLogging(rootCmd)

			logger := slog.Default()
			if !logger.Enabled(t.Context(), tt.wantLevel) {
				t.Errorf("expected level %v to be enabled", tt.wantLevel)
			}
			if tt.wantLevel > logging.LevelTrace {
				// Check that a lower level is NOT enabled (e.g. if we want Info, Debug shouldn't be enabled)

				// Wait, slog levels are: Debug(-4), Info(0), Warn(4), Error(8)
				// My LevelTrace is Debug-4 (-8).
				// If I want Info (0), Debug (-4) should NOT be enabled.
				// But wait, enabled checks if >= level.
				// If logger is at Info, Enabled(Info) is true. Enabled(Debug) is false.

				// Let's verify exact match by checking boundary
				shouldBeDisabled := tt.wantLevel - 4 // approximate next lower standard level
				if logger.Enabled(t.Context(), shouldBeDisabled) {
					t.Errorf("expected level %v to be disabled", shouldBeDisabled)
				}
			}
		})
	}
}

func TestSetupLogging_EnvVar(t *testing.T) {
	// Save/Restore original state
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
			verbosity = 0 // ensure flag doesn't override
			t.Setenv("AIX_DEBUG", tt.envVal)

			setupLogging(rootCmd)

			logger := slog.Default()
			if !logger.Enabled(t.Context(), tt.wantLevel) {
				t.Errorf("expected level %v to be enabled", tt.wantLevel)
			}

			// Verify it's not too verbose (e.g. 1 shouldn't give Trace)
			if tt.wantLevel == slog.LevelDebug {
				if logger.Enabled(t.Context(), logging.LevelTrace) {
					t.Error("expected Trace level to be disabled when AIX_DEBUG=1")
				}
			}
		})
	}
}

func TestSetupLogging_FlagPrecedence(t *testing.T) {
	// Flag should override Env Var
	origVerbosity := verbosity
	defer func() { verbosity = origVerbosity }()

	t.Setenv("AIX_DEBUG", "2") // Set to Trace
	verbosity = 1              // Set flag to Info

	setupLogging(rootCmd)

	logger := slog.Default()
	if !logger.Enabled(t.Context(), slog.LevelInfo) {
		t.Error("expected Info level to be enabled")
	}
	if logger.Enabled(t.Context(), slog.LevelDebug) {
		t.Error("expected Debug level to be disabled (flag should override env var)")
	}
}
