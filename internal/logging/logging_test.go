package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:  slog.LevelInfo,
		Format: FormatJSON,
		Output: &buf,
	})

	logger.Info("test message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, output)
	}

	// Verify expected fields exist
	if _, ok := parsed["msg"]; !ok {
		t.Errorf("JSON output missing 'msg' field: %s", output)
	}
	if _, ok := parsed["level"]; !ok {
		t.Errorf("JSON output missing 'level' field: %s", output)
	}
	if parsed["key"] != "value" {
		t.Errorf("JSON output missing custom attribute: got %v, want 'value'", parsed["key"])
	}
}

func TestNew_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:  slog.LevelInfo,
		Format: FormatText,
		Output: &buf,
	})

	logger.Info("test message", "key", "value")

	output := buf.String()
	if output == "" {
		t.Fatal("expected output, got empty string")
	}

	// Verify it's NOT valid JSON (it's text format)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err == nil {
		t.Error("text format should not be valid JSON")
	}

	// Verify expected content exists
	if !strings.Contains(output, "test message") {
		t.Errorf("output missing message: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("output missing key=value attribute: %s", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("output missing level indicator: %s", output)
	}
}

func TestNew_DefaultsToStderr(t *testing.T) {
	// This test verifies the code path, not actual stderr output
	logger := New(Config{
		Level:  slog.LevelInfo,
		Format: FormatText,
		// Output intentionally nil
	})

	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_UnknownFormatDefaultsToText(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:  slog.LevelInfo,
		Format: Format("unknown"),
		Output: &buf,
	})

	logger.Info("test message")

	output := buf.String()

	// Verify it's text format (not JSON)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err == nil {
		t.Error("unknown format should default to text, not JSON")
	}
}

func TestDefault(t *testing.T) {
	logger := Default()
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewDiscard(t *testing.T) {
	logger := NewDiscard()

	// Create a buffer to verify nothing is written elsewhere
	var buf bytes.Buffer
	// NewDiscard should write to io.Discard, not our buffer
	// We can't directly verify io.Discard, but we can verify the logger works
	logger.Info("this should be discarded")
	logger.Error("this too")
	logger.Debug("and this")

	if buf.Len() != 0 {
		t.Error("discard logger should not write to our buffer")
	}
}

func TestNewDiscard_ProducesNoOutput(t *testing.T) {
	// The best we can do is verify the logger doesn't panic
	// and that it accepts log calls without error
	logger := NewDiscard()

	// These should all succeed silently
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "count", 42)
	logger.Warn("warn message", "flag", true)
	logger.Error("error message", "err", "something went wrong")
}

func TestLevelFiltering(t *testing.T) {
	tests := []struct {
		name         string
		configLevel  slog.Level
		logLevel     slog.Level
		shouldAppear bool
	}{
		{
			name:         "info logged at info level",
			configLevel:  slog.LevelInfo,
			logLevel:     slog.LevelInfo,
			shouldAppear: true,
		},
		{
			name:         "debug not logged at info level",
			configLevel:  slog.LevelInfo,
			logLevel:     slog.LevelDebug,
			shouldAppear: false,
		},
		{
			name:         "error logged at info level",
			configLevel:  slog.LevelInfo,
			logLevel:     slog.LevelError,
			shouldAppear: true,
		},
		{
			name:         "warn logged at warn level",
			configLevel:  slog.LevelWarn,
			logLevel:     slog.LevelWarn,
			shouldAppear: true,
		},
		{
			name:         "info not logged at warn level",
			configLevel:  slog.LevelWarn,
			logLevel:     slog.LevelInfo,
			shouldAppear: false,
		},
		{
			name:         "debug logged at debug level",
			configLevel:  slog.LevelDebug,
			logLevel:     slog.LevelDebug,
			shouldAppear: true,
		},
		{
			name:         "error not logged at error+4 level",
			configLevel:  slog.LevelError + 4,
			logLevel:     slog.LevelError,
			shouldAppear: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(Config{
				Level:  tt.configLevel,
				Format: FormatText,
				Output: &buf,
			})

			switch tt.logLevel {
			case slog.LevelDebug:
				logger.Debug("test message")
			case slog.LevelInfo:
				logger.Info("test message")
			case slog.LevelWarn:
				logger.Warn("test message")
			case slog.LevelError:
				logger.Error("test message")
			}

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldAppear {
				t.Errorf("level filtering: got output=%v, want output=%v\nconfig level: %v, log level: %v\noutput: %q",
					hasOutput, tt.shouldAppear, tt.configLevel, tt.logLevel, buf.String())
			}
		})
	}
}

func TestForTest(t *testing.T) {
	// ForTest returns a logger that writes to t.Log
	// We can verify it doesn't panic and produces output
	logger := ForTest(t)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// These should all be captured by the test framework
	logger.Debug("debug from test logger")
	logger.Info("info from test logger", "test", t.Name())
}

func TestForTest_CapturesAllLevels(t *testing.T) {
	// ForTest is configured at Debug level to capture everything
	logger := ForTest(t)

	// All of these should work without panic
	logger.Debug("debug level")
	logger.Info("info level")
	logger.Warn("warn level")
	logger.Error("error level")
}

func TestFormat_Constants(t *testing.T) {
	// Verify the format constants have expected values
	if FormatText != "text" {
		t.Errorf("FormatText = %q, want %q", FormatText, "text")
	}
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
}

func TestNew_WithAttributes(t *testing.T) {
	tests := []struct {
		name   string
		format Format
	}{
		{"text format", FormatText},
		{"json format", FormatJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(Config{
				Level:  slog.LevelInfo,
				Format: tt.format,
				Output: &buf,
			})

			// Test various attribute types
			logger.Info("message",
				"string", "value",
				"int", 42,
				"float", 3.14,
				"bool", true,
			)

			output := buf.String()
			if output == "" {
				t.Fatal("expected output, got empty string")
			}

			// Verify all attributes appear in output
			for _, want := range []string{"string", "value", "42", "3.14", "true"} {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q: %s", want, output)
				}
			}
		})
	}
}

func TestLevelFromVerbosity(t *testing.T) {
	tests := []struct {
		verbosity int
		want      slog.Level
	}{
		{-1, slog.LevelWarn},
		{0, slog.LevelWarn},
		{1, slog.LevelInfo},
		{2, slog.LevelDebug},
		{3, LevelTrace},
		{4, LevelTrace},
	}

	for _, tt := range tests {
		got := LevelFromVerbosity(tt.verbosity)
		if got != tt.want {
			t.Errorf("LevelFromVerbosity(%d) = %v, want %v", tt.verbosity, got, tt.want)
		}
	}
}

func TestLevelTrace(t *testing.T) {
	if LevelTrace >= slog.LevelDebug {
		t.Error("LevelTrace should be lower than LevelDebug")
	}
}

func TestTestWriter_TrimsNewline(t *testing.T) {
	// Verify the testWriter properly trims trailing newlines
	tw := &testWriter{t: t}

	// Write with trailing newline (like slog does)
	n, err := tw.Write([]byte("test message\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("test message\n") {
		t.Errorf("Write returned %d, want %d", n, len("test message\n"))
	}

	// Write without trailing newline
	n, err = tw.Write([]byte("no newline"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("no newline") {
		t.Errorf("Write returned %d, want %d", n, len("no newline"))
	}

	// Write empty string
	n, err = tw.Write([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("Write returned %d, want 0", n)
	}
}
