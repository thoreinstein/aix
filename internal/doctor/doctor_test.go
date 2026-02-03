package doctor

import (
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner()
	if r == nil {
		t.Fatal("NewRunner returned nil")
	}
	if len(r.checks) != 0 {
		t.Errorf("NewRunner().checks = %d, want 0", len(r.checks))
	}
}

func TestRunner_AddCheck(t *testing.T) {
	t.Run("single check", func(t *testing.T) {
		r := NewRunner()
		check := NewMockCheck(t)
		check.EXPECT().Name().Return("test-1").Maybe()
		r.AddCheck(check)

		if len(r.checks) != 1 {
			t.Errorf("AddCheck: checks count = %d, want 1", len(r.checks))
		}
		if r.checks[0].Name() != "test-1" {
			t.Errorf("AddCheck: check name = %q, want %q", r.checks[0].Name(), "test-1")
		}
	})

	t.Run("multiple checks", func(t *testing.T) {
		r := NewRunner()

		for range 3 {
			check := NewMockCheck(t)
			r.AddCheck(check)
		}

		if len(r.checks) != 3 {
			t.Errorf("AddCheck: checks count = %d, want 3", len(r.checks))
		}
	})

	t.Run("order preserved", func(t *testing.T) {
		r := NewRunner()
		names := []string{"first", "second", "third"}

		for _, name := range names {
			check := NewMockCheck(t)
			check.EXPECT().Name().Return(name).Maybe()
			r.AddCheck(check)
		}

		for i, want := range names {
			if r.checks[i].Name() != want {
				t.Errorf("AddCheck order: checks[%d].Name() = %q, want %q", i, r.checks[i].Name(), want)
			}
		}
	})
}

func TestRunner_Run(t *testing.T) {
	tests := []struct {
		name            string
		results         []*CheckResult
		wantResultCount int
		wantPassed      int
		wantInfo        int
		wantWarnings    int
		wantErrors      int
	}{
		{
			name:            "empty runner",
			results:         nil,
			wantResultCount: 0,
			wantPassed:      0,
			wantInfo:        0,
			wantWarnings:    0,
			wantErrors:      0,
		},
		{
			name: "single pass",
			results: []*CheckResult{
				{Status: SeverityPass},
			},
			wantResultCount: 1,
			wantPassed:      1,
			wantInfo:        0,
			wantWarnings:    0,
			wantErrors:      0,
		},
		{
			name: "single info",
			results: []*CheckResult{
				{Status: SeverityInfo},
			},
			wantResultCount: 1,
			wantPassed:      0,
			wantInfo:        1,
			wantWarnings:    0,
			wantErrors:      0,
		},
		{
			name: "single warning",
			results: []*CheckResult{
				{Status: SeverityWarning},
			},
			wantResultCount: 1,
			wantPassed:      0,
			wantInfo:        0,
			wantWarnings:    1,
			wantErrors:      0,
		},
		{
			name: "single error",
			results: []*CheckResult{
				{Status: SeverityError},
			},
			wantResultCount: 1,
			wantPassed:      0,
			wantInfo:        0,
			wantWarnings:    0,
			wantErrors:      1,
		},
		{
			name: "mixed severities",
			results: []*CheckResult{
				{Status: SeverityPass},
				{Status: SeverityPass},
				{Status: SeverityInfo},
				{Status: SeverityWarning},
				{Status: SeverityWarning},
				{Status: SeverityError},
			},
			wantResultCount: 6,
			wantPassed:      2,
			wantInfo:        1,
			wantWarnings:    2,
			wantErrors:      1,
		},
		{
			name: "all pass",
			results: []*CheckResult{
				{Status: SeverityPass},
				{Status: SeverityPass},
				{Status: SeverityPass},
			},
			wantResultCount: 3,
			wantPassed:      3,
			wantInfo:        0,
			wantWarnings:    0,
			wantErrors:      0,
		},
		{
			name: "all errors",
			results: []*CheckResult{
				{Status: SeverityError},
				{Status: SeverityError},
			},
			wantResultCount: 2,
			wantPassed:      0,
			wantInfo:        0,
			wantWarnings:    0,
			wantErrors:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRunner()
			for _, result := range tt.results {
				check := NewMockCheck(t)
				check.EXPECT().Run().Return(result)
				r.AddCheck(check)
			}

			before := time.Now().UTC()
			report := r.Run()
			after := time.Now().UTC()

			// Verify timestamp is recent and between before and after
			if report.Timestamp.Before(before) || report.Timestamp.After(after) {
				t.Errorf("Timestamp %v not in expected range [%v, %v]",
					report.Timestamp, before, after)
			}

			// Verify results count
			if len(report.Results) != tt.wantResultCount {
				t.Errorf("Results count = %d, want %d", len(report.Results), tt.wantResultCount)
			}

			// Verify summary counts
			if report.Summary.Passed != tt.wantPassed {
				t.Errorf("Summary.Passed = %d, want %d", report.Summary.Passed, tt.wantPassed)
			}
			if report.Summary.Info != tt.wantInfo {
				t.Errorf("Summary.Info = %d, want %d", report.Summary.Info, tt.wantInfo)
			}
			if report.Summary.Warnings != tt.wantWarnings {
				t.Errorf("Summary.Warnings = %d, want %d", report.Summary.Warnings, tt.wantWarnings)
			}
			if report.Summary.Errors != tt.wantErrors {
				t.Errorf("Summary.Errors = %d, want %d", report.Summary.Errors, tt.wantErrors)
			}
		})
	}
}

func TestRunner_Run_ResultsOrder(t *testing.T) {
	r := NewRunner()
	names := []string{"first", "second", "third"}
	statuses := []Severity{SeverityPass, SeverityWarning, SeverityError}

	for i, name := range names {
		check := NewMockCheck(t)
		check.EXPECT().Run().Return(&CheckResult{Name: name, Status: statuses[i]})
		r.AddCheck(check)
	}

	report := r.Run()

	// Results should be in the same order as checks were added
	for i, want := range names {
		if report.Results[i].Name != want {
			t.Errorf("Results[%d].Name = %q, want %q", i, report.Results[i].Name, want)
		}
	}
}

func TestDoctorReport_HasErrors(t *testing.T) {
	tests := []struct {
		name   string
		errors int
		want   bool
	}{
		{"no errors", 0, false},
		{"one error", 1, true},
		{"multiple errors", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DoctorReport{Summary: Summary{Errors: tt.errors}}
			if got := r.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoctorReport_HasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		warnings int
		want     bool
	}{
		{"no warnings", 0, false},
		{"one warning", 1, true},
		{"multiple warnings", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DoctorReport{Summary: Summary{Warnings: tt.warnings}}
			if got := r.HasWarnings(); got != tt.want {
				t.Errorf("HasWarnings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoctorReport_HasErrors_IndependentOfWarnings(t *testing.T) {
	// Verify HasErrors only checks errors, not warnings
	r := &DoctorReport{Summary: Summary{Warnings: 10, Errors: 0}}
	if r.HasErrors() {
		t.Error("HasErrors() = true when only warnings present, want false")
	}

	r = &DoctorReport{Summary: Summary{Warnings: 10, Errors: 1}}
	if !r.HasErrors() {
		t.Error("HasErrors() = false when errors present, want true")
	}
}

func TestDoctorReport_HasWarnings_IndependentOfErrors(t *testing.T) {
	// Verify HasWarnings only checks warnings, not errors
	r := &DoctorReport{Summary: Summary{Warnings: 0, Errors: 10}}
	if r.HasWarnings() {
		t.Error("HasWarnings() = true when only errors present, want false")
	}

	r = &DoctorReport{Summary: Summary{Warnings: 1, Errors: 10}}
	if !r.HasWarnings() {
		t.Error("HasWarnings() = false when warnings present, want true")
	}
}

func TestDoctorReport_ZeroValue(t *testing.T) {
	// Test that zero-value report behaves correctly
	var r DoctorReport

	if r.HasErrors() {
		t.Error("zero-value HasErrors() = true, want false")
	}
	if r.HasWarnings() {
		t.Error("zero-value HasWarnings() = true, want false")
	}
	if r.Timestamp != (time.Time{}) {
		t.Error("zero-value Timestamp should be zero time")
	}
	if r.Results != nil {
		t.Error("zero-value Results should be nil")
	}
}
