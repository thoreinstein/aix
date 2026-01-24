package validator

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestReporter_Report(t *testing.T) {
	result := &Result{}
	result.AddError("name", "is required", nil)
	result.AddWarning("desc", "missing", "some val")
	result.Issues[0].Context = map[string]string{"file": "test.md"}

	t.Run("text format", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := NewReporter(&buf, FormatText)
		if err := reporter.Report(result); err != nil {
			t.Fatalf("Report() error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "1 error(s)") {
			t.Error("output missing error summary")
		}
		if !strings.Contains(output, "name: is required") {
			t.Error("output missing error details")
		}
		if !strings.Contains(output, "(file=test.md)") {
			t.Error("output missing context")
		}
		if !strings.Contains(output, "[some val]") {
			t.Error("output missing value")
		}
	})

	t.Run("json format", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := NewReporter(&buf, FormatJSON)
		if err := reporter.Report(result); err != nil {
			t.Fatalf("Report() error: %v", err)
		}

		var decoded Result
		if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatalf("failed to decode JSON output: %v", err)
		}

		if len(decoded.Issues) != 2 {
			t.Errorf("decoded issues count = %d, want 2", len(decoded.Issues))
		}
		if decoded.Issues[0].Field != "name" {
			t.Errorf("first issue field = %q, want name", decoded.Issues[0].Field)
		}
	})

	t.Run("empty result text", func(t *testing.T) {
		var buf bytes.Buffer
		reporter := NewReporter(&buf, FormatText)
		if err := reporter.Report(&Result{}); err != nil {
			t.Fatalf("Report() error: %v", err)
		}
		if !strings.Contains(buf.String(), "Validation passed") {
			t.Error("output missing success message")
		}
	})
}
