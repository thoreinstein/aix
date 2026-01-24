package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestHandler_Handle(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(h)

	now := time.Now()
	logger.Info("hello world", "foo", "value")

	output := buf.String()

	// Check format: Time Level Message Attributes
	// Example: 10:00PM INFO  hello world foo=value

	if !strings.Contains(output, "INFO") {
		t.Errorf("expected level INFO in output, got: %q", output)
	}
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected message in output, got: %q", output)
	}
	if !strings.Contains(output, "foo=value") {
		t.Errorf("expected attribute in output, got: %q", output)
	}

	// Verify it contains the time (using Kitchen format as implemented)
	expectedTime := now.Format(time.Kitchen)
	if !strings.Contains(output, expectedTime) {
		t.Errorf("expected time %q in output, got: %q", expectedTime, output)
	}
}

func TestHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, nil)
	logger := slog.New(h).With("common", "attr")

	logger.Info("message", "local", "val")

	output := buf.String()
	if !strings.Contains(output, "common=attr") {
		t.Errorf("expected common attribute in output, got: %q", output)
	}
	if !strings.Contains(output, "local=val") {
		t.Errorf("expected local attribute in output, got: %q", output)
	}
}

func TestHandler_Enabled(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})

	ctx := t.Context()
	if h.Enabled(ctx, slog.LevelInfo) {
		t.Error("expected Info level to be disabled when min level is Warn")
	}
	if !h.Enabled(ctx, slog.LevelWarn) {
		t.Error("expected Warn level to be enabled")
	}
	if !h.Enabled(ctx, slog.LevelError) {
		t.Error("expected Error level to be enabled")
	}
}

func TestHandler_NoTime(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, nil)

	// Create a record without time
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "no time", 0)
	err := h.Handle(t.Context(), r)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	output := buf.String()
	// Should not start with a time-like pattern (Kitchen format usually has ':')
	if strings.Contains(output, ":") && strings.Index(output, ":") < 10 {
		t.Errorf("expected no time in output, got: %q", output)
	}
}

func TestHandler_Redaction(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(h)

	// Test case-insensitive key matching
	logger.Info("sensitive data", "api_key", "secret12345", "Token", "ghp_abcdef")

	output := buf.String()

	// Should be redacted
	if strings.Contains(output, "secret12345") {
		t.Error("api_key value should be redacted")
	}
	if strings.Contains(output, "ghp_abcdef") {
		t.Error("Token value should be redacted")
	}

	// Should contain masked values
	if !strings.Contains(output, "api_key=****2345") {
		t.Errorf("expected masked api_key, got: %q", output)
	}
	// "ghp_abcdef" length is 10. MaskValue: "****" + last 4 ("cdef") -> "****cdef"
	if !strings.Contains(output, "Token=****cdef") {
		t.Errorf("expected masked Token, got: %q", output)
	}

	// Test value prefix matching
	buf.Reset()
	logger.Info("token value", "foo", "ghp_secrettoken")
	output = buf.String()

	if strings.Contains(output, "ghp_secrettoken") {
		t.Error("value with token prefix should be redacted even if key is safe")
	}
	// "ghp_secrettoken" length 15. MaskValue: "****oken"
	if !strings.Contains(output, "foo=****oken") {
		t.Errorf("expected masked value based on prefix, got: %q", output)
	}
}
