package validator

import (
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cmdName   string
		path      string
		wantErrs  int
		wantField string
		wantMsg   string
	}{
		// Valid names
		{
			name:     "valid simple name",
			cmdName:  "review",
			path:     "review.md",
			wantErrs: 0,
		},
		{
			name:     "valid hyphenated name",
			cmdName:  "code-review",
			path:     "code-review.md",
			wantErrs: 0,
		},
		{
			name:     "valid long hyphenated name",
			cmdName:  "my-long-command-name",
			path:     "my-long-command-name.md",
			wantErrs: 0,
		},
		{
			name:     "valid single character name",
			cmdName:  "a",
			path:     "a.md",
			wantErrs: 0,
		},
		{
			name:     "valid name with numbers",
			cmdName:  "review2",
			path:     "review2.md",
			wantErrs: 0,
		},
		{
			name:     "valid max length name 64 chars",
			cmdName:  strings.Repeat("a", 64),
			path:     "maxlen.md",
			wantErrs: 0,
		},
		// Missing name scenarios
		{
			name:     "missing name with valid path inferred passes",
			cmdName:  "",
			path:     "inferred-name.md",
			wantErrs: 0, // Name will be inferred from path
		},
		{
			name:      "missing name with empty path fails",
			cmdName:   "",
			path:      "",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "required when path is not provided",
		},
		{
			name:      "missing name with path that infers empty fails",
			cmdName:   "",
			path:      ".md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "could not be inferred",
		},
		// Name length validation
		{
			name:      "name exceeding 64 characters fails",
			cmdName:   strings.Repeat("a", 65),
			path:      "toolong.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "exceeds maximum length",
		},
		// Case validation
		{
			name:      "name with uppercase fails",
			cmdName:   "MyCommand",
			path:      "MyCommand.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase",
		},
		{
			name:      "name with mixed case fails",
			cmdName:   "myCommand",
			path:      "myCommand.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase",
		},
		{
			name:      "name all uppercase fails",
			cmdName:   "REVIEW",
			path:      "REVIEW.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase",
		},
		// Hyphen validation
		{
			name:      "name starting with hyphen fails",
			cmdName:   "-review",
			path:      "-review.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "cannot start or end with a hyphen",
		},
		{
			name:      "name ending with hyphen fails",
			cmdName:   "review-",
			path:      "review-.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "cannot start or end with a hyphen",
		},
		{
			name:      "name with consecutive hyphens fails",
			cmdName:   "my--command",
			path:      "my--command.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "consecutive hyphens",
		},
		{
			name:      "name with triple hyphens fails",
			cmdName:   "my---command",
			path:      "my---command.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "consecutive hyphens",
		},
		// Special characters validation
		{
			name:      "name with underscore fails",
			cmdName:   "my_command",
			path:      "my_command.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase alphanumeric",
		},
		{
			name:      "name with space fails",
			cmdName:   "my command",
			path:      "my command.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase alphanumeric",
		},
		{
			name:      "name with dot fails",
			cmdName:   "my.command",
			path:      "my.command.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase alphanumeric",
		},
		{
			name:      "name with at symbol fails",
			cmdName:   "my@command",
			path:      "my@command.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase alphanumeric",
		},
		{
			name:      "name with slash fails",
			cmdName:   "my/command",
			path:      "mycommand.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase alphanumeric",
		},
		// Start with number fails
		{
			name:      "name starting with number fails",
			cmdName:   "123review",
			path:      "123review.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "start with a letter",
		},
		{
			name:      "name starting with hyphen then number fails",
			cmdName:   "-123",
			path:      "test.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "cannot start or end with a hyphen",
		},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &claude.Command{Name: tt.cmdName}
			result := v.Validate(cmd, tt.path)

			if len(result.Errors) != tt.wantErrs {
				t.Errorf("Validate() got %d errors, want %d; errors: %v",
					len(result.Errors), tt.wantErrs, result.Errors)
				return
			}

			if tt.wantErrs > 0 && tt.wantField != "" {
				found := false
				for _, issue := range result.Errors {
					if issue.Field == tt.wantField && strings.Contains(issue.Message, tt.wantMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error for field %q with message containing %q, got: %v",
						tt.wantField, tt.wantMsg, result.Errors)
				}
			}
		})
	}
}

func TestValidator_Validate_OpenCode(t *testing.T) {
	// Verify that validator works with OpenCode commands too (same interface)
	tests := []struct {
		name     string
		cmdName  string
		path     string
		wantErrs int
	}{
		{
			name:     "valid opencode command",
			cmdName:  "deploy",
			path:     "deploy.md",
			wantErrs: 0,
		},
		{
			name:     "invalid opencode command name",
			cmdName:  "Deploy",
			path:     "deploy.md",
			wantErrs: 1,
		},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &opencode.Command{Name: tt.cmdName}
			result := v.Validate(cmd, tt.path)

			if len(result.Errors) != tt.wantErrs {
				t.Errorf("Validate() got %d errors, want %d; errors: %v",
					len(result.Errors), tt.wantErrs, result.Errors)
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
					{Level: Error, Field: "name", Message: "test error"},
				},
			},
			want: true,
		},
		{
			name: "result with only warnings has no errors",
			result: &Result{
				Warnings: []Issue{
					{Level: Warning, Field: "name", Message: "test warning"},
				},
			},
			want: false,
		},
		{
			name: "result with both errors and warnings returns true",
			result: &Result{
				Errors: []Issue{
					{Level: Error, Field: "name", Message: "test error"},
				},
				Warnings: []Issue{
					{Level: Warning, Field: "description", Message: "test warning"},
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
		issue *Issue
		want  string
	}{
		{
			name: "with value",
			issue: &Issue{
				Level:   Error,
				Field:   "name",
				Message: "name is required",
				Value:   "bad-name",
			},
			want: `name: name is required (got "bad-name")`,
		},
		{
			name: "without value",
			issue: &Issue{
				Level:   Error,
				Field:   "name",
				Message: "name is required",
			},
			want: "name: name is required",
		},
		{
			name: "warning level",
			issue: &Issue{
				Level:   Warning,
				Field:   "description",
				Message: "description is recommended",
			},
			want: "description: description is recommended",
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
	v := New()
	if v == nil {
		t.Error("New() returned nil")
	}
}

func TestValidator_EdgeCases(t *testing.T) {
	v := New()

	t.Run("path with only directory returns dot for name", func(t *testing.T) {
		cmd := &claude.Command{Name: ""}
		result := v.Validate(cmd, "/some/path/")
		// filepath.Base("/some/path/") returns "path", not "."
		// So this should actually infer "path" successfully
		if result.HasErrors() {
			t.Errorf("expected no errors for path with directory, got: %v", result.Errors)
		}
	})

	t.Run("name exactly at max length is valid", func(t *testing.T) {
		cmd := &claude.Command{Name: strings.Repeat("a", 64)}
		result := v.Validate(cmd, "test.md")
		if result.HasErrors() {
			t.Errorf("expected no errors for max length name, got: %v", result.Errors)
		}
	})

	t.Run("name one over max length fails", func(t *testing.T) {
		cmd := &claude.Command{Name: strings.Repeat("a", 65)}
		result := v.Validate(cmd, "test.md")
		if !result.HasErrors() {
			t.Error("expected error for name exceeding max length")
		}
	})

	t.Run("valid name with hyphen in middle", func(t *testing.T) {
		cmd := &claude.Command{Name: "code-review"}
		result := v.Validate(cmd, "test.md")
		if result.HasErrors() {
			t.Errorf("expected no errors for valid hyphenated name, got: %v", result.Errors)
		}
	})

	t.Run("valid name with multiple hyphens non-consecutive", func(t *testing.T) {
		cmd := &claude.Command{Name: "my-long-command-name"}
		result := v.Validate(cmd, "test.md")
		if result.HasErrors() {
			t.Errorf("expected no errors for multiple non-consecutive hyphens, got: %v", result.Errors)
		}
	})
}

func TestLevel_Constants(t *testing.T) {
	// Ensure level constants have distinct values
	if Warning == Error {
		t.Error("Warning and Error should have different values")
	}

	// Ensure Warning < Error (common convention)
	if Warning >= Error {
		t.Error("Warning should be less than Error")
	}
}
