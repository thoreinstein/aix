package validator

import "testing"

// mockAgent implements Agentable for testing.
type mockAgent struct {
	name         string
	description  string
	instructions string
}

func (m *mockAgent) GetName() string         { return m.name }
func (m *mockAgent) GetDescription() string  { return m.description }
func (m *mockAgent) GetInstructions() string { return m.instructions }

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		agent      Agentable
		strict     bool
		wantErrors int
		wantWarns  int
	}{
		{
			name: "valid agent with all fields",
			agent: &mockAgent{
				name:         "test-agent",
				description:  "A test agent",
				instructions: "Do something useful",
			},
			strict:     false,
			wantErrors: 0,
			wantWarns:  0,
		},
		{
			name: "valid agent minimal (name and instructions only)",
			agent: &mockAgent{
				name:         "minimal-agent",
				instructions: "Just instructions",
			},
			strict:     false,
			wantErrors: 0,
			wantWarns:  0,
		},
		{
			name: "strict mode missing description generates warning",
			agent: &mockAgent{
				name:         "strict-agent",
				instructions: "Instructions present",
			},
			strict:     true,
			wantErrors: 0,
			wantWarns:  1,
		},
		{
			name: "strict mode with description no warning",
			agent: &mockAgent{
				name:         "strict-agent",
				description:  "Has description",
				instructions: "Instructions present",
			},
			strict:     true,
			wantErrors: 0,
			wantWarns:  0,
		},
		{
			name: "missing name is error",
			agent: &mockAgent{
				description:  "Has description",
				instructions: "Has instructions",
			},
			strict:     false,
			wantErrors: 1,
			wantWarns:  0,
		},
		{
			name: "empty file is error",
			agent: &mockAgent{
				name:         "",
				description:  "",
				instructions: "",
			},
			strict:     false,
			wantErrors: 2, // missing name + empty file
			wantWarns:  0,
		},
		{
			name: "empty file strict mode",
			agent: &mockAgent{
				name:         "",
				description:  "",
				instructions: "",
			},
			strict:     true,
			wantErrors: 2, // missing name + empty file
			wantWarns:  1, // missing description warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(tt.strict)
			result := v.Validate(tt.agent, "test.md")

			if got := len(result.Errors); got != tt.wantErrors {
				t.Errorf("Validate() errors = %d, want %d", got, tt.wantErrors)
				for _, err := range result.Errors {
					t.Logf("  error: %s", err.Error())
				}
			}

			if got := len(result.Warnings); got != tt.wantWarns {
				t.Errorf("Validate() warnings = %d, want %d", got, tt.wantWarns)
				for _, warn := range result.Warnings {
					t.Logf("  warning: %s", warn.Error())
				}
			}
		})
	}
}

func TestResult_HasErrors(t *testing.T) {
	tests := []struct {
		name   string
		result *Result
		want   bool
	}{
		{
			name:   "empty result has no errors",
			result: &Result{},
			want:   false,
		},
		{
			name: "result with errors returns true",
			result: &Result{
				Errors: []Issue{
					{Level: Error, Field: "name", Message: "name is required"},
				},
			},
			want: true,
		},
		{
			name: "result with only warnings returns false",
			result: &Result{
				Warnings: []Issue{
					{Level: Warning, Field: "description", Message: "description is recommended"},
				},
			},
			want: false,
		},
		{
			name: "result with both errors and warnings returns true",
			result: &Result{
				Errors: []Issue{
					{Level: Error, Field: "name", Message: "name is required"},
				},
				Warnings: []Issue{
					{Level: Warning, Field: "description", Message: "description is recommended"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIssue_Error(t *testing.T) {
	tests := []struct {
		name  string
		issue Issue
		want  string
	}{
		{
			name: "issue without value",
			issue: Issue{
				Level:   Error,
				Field:   "name",
				Message: "name is required",
			},
			want: "name: name is required",
		},
		{
			name: "issue with value",
			issue: Issue{
				Level:   Error,
				Field:   "name",
				Message: "invalid format",
				Value:   "bad-name!",
			},
			want: `name: invalid format (got "bad-name!")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.issue.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("creates non-strict validator", func(t *testing.T) {
		if v := New(false); v == nil {
			t.Fatal("New(false) returned nil")
		} else if v.strict {
			t.Error("New(false) created strict validator")
		}
	})

	t.Run("creates strict validator", func(t *testing.T) {
		if v := New(true); v == nil {
			t.Fatal("New(true) returned nil")
		} else if !v.strict {
			t.Error("New(true) created non-strict validator")
		}
	})
}
