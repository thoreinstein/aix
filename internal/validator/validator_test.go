package validator

import (
	"testing"
)

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		s    Severity
		want string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{SeverityInfo, "info"},
		{Severity(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("Severity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssue_Error(t *testing.T) {
	tests := []struct {
		name string
		i    Issue
		want string
	}{
		{
			name: "error with field and value",
			i: Issue{
				Severity: SeverityError,
				Field:    "name",
				Message:  "is required",
				Value:    "",
			},
			want: "error: field \"name\": is required (got )",
		},
		{
			name: "warning without field",
			i: Issue{
				Severity: SeverityWarning,
				Message:  "recommended description",
			},
			want: "warning: recommended description",
		},
		{
			name: "info with field",
			i: Issue{
				Severity: SeverityInfo,
				Field:    "version",
				Message:  "is outdated",
			},
			want: "info: field \"version\": is outdated",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Error(); got != tt.want {
				t.Errorf("Issue.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResult_Helpers(t *testing.T) {
	r := &Result{}

	if r.HasErrors() {
		t.Error("expected no errors")
	}

	r.AddError("f1", "m1", "v1")
	if !r.HasErrors() {
		t.Error("expected errors")
	}
	if len(r.Errors()) != 1 {
		t.Errorf("expected 1 error, got %d", len(r.Errors()))
	}

	if r.HasWarnings() {
		t.Error("expected no warnings")
	}
	r.AddWarning("f2", "m2", "v2")
	if !r.HasWarnings() {
		t.Error("expected warnings")
	}
	if len(r.Warnings()) != 1 {
		t.Errorf("expected 1 warning, got %d", len(r.Warnings()))
	}

	r.AddInfo("f3", "m3", "v3")
	if len(r.Issues) != 3 {
		t.Errorf("expected 3 issues, got %d", len(r.Issues))
	}
}

func TestResult_NilSafety(t *testing.T) {
	var r *Result
	if r.HasErrors() {
		t.Error("expected no errors for nil result")
	}
	if r.HasWarnings() {
		t.Error("expected no warnings for nil result")
	}
	if r.Errors() != nil {
		t.Error("expected nil Errors() for nil result")
	}
	if r.Warnings() != nil {
		t.Error("expected nil Warnings() for nil result")
	}
}
