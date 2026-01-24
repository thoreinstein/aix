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
	logger.Info("hello world", "key", "value")

	output := buf.String()

	// Check format: Time Level Message Attributes
	// Example: 10:00PM INFO  hello world key=value

	if !strings.Contains(output, "INFO") {
		t.Errorf("expected level INFO in output, got: %q", output)
	}
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected message in output, got: %q", output)
	}
	if !strings.Contains(output, "key=value") {
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
