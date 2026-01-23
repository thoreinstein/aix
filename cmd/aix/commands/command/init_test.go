package command

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple name",
			input:   "review",
			wantErr: false,
		},
		{
			name:    "valid hyphenated name",
			input:   "code-review",
			wantErr: false,
		},
		{
			name:    "valid multi-hyphenated name",
			input:   "my-code-review",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "review2",
			wantErr: false,
		},
		{
			name:    "valid hyphen with numbers",
			input:   "review-v2",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
			errMsg:  "command name is required",
		},
		{
			name:    "uppercase not allowed",
			input:   "Review",
			wantErr: true,
			errMsg:  "lowercase",
		},
		{
			name:    "starts with number",
			input:   "2review",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "starts with hyphen",
			input:   "-review",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "ends with hyphen",
			input:   "review-",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "consecutive hyphens",
			input:   "code--review",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "special characters",
			input:   "code_review",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "spaces not allowed",
			input:   "code review",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "too long name (65 chars)",
			input:   "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklm",
			wantErr: true,
			errMsg:  "at most 64 characters",
		},
		{
			name:    "exactly 64 chars is ok",
			input:   "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateName(%q) expected error, got nil", tt.input)
				} else if tt.errMsg != "" && !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("validateName(%q) error = %q, want to contain %q", tt.input, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateName(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

func TestSanitizeDefaultName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple lowercase",
			input: "review",
			want:  "review",
		},
		{
			name:  "uppercase converted",
			input: "Review",
			want:  "review",
		},
		{
			name:  "spaces replaced with hyphens",
			input: "code review",
			want:  "code-review",
		},
		{
			name:  "underscores replaced with hyphens",
			input: "code_review",
			want:  "code-review",
		},
		{
			name:  "mixed special characters",
			input: "My Code_Review!",
			want:  "my-code-review",
		},
		{
			name:  "leading special chars trimmed",
			input: "---review",
			want:  "review",
		},
		{
			name:  "trailing special chars trimmed",
			input: "review---",
			want:  "review",
		},
		{
			name:  "empty string falls back to default",
			input: "",
			want:  "new-command",
		},
		{
			name:  "only special chars falls back to default",
			input: "!!!",
			want:  "new-command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeDefaultName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeDefaultName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single word",
			input: "review",
			want:  "Review",
		},
		{
			name:  "hyphenated",
			input: "code-review",
			want:  "Code Review",
		},
		{
			name:  "multiple hyphens",
			input: "my-code-review",
			want:  "My Code Review",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single char",
			input: "a",
			want:  "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTitle(tt.input)
			if got != tt.want {
				t.Errorf("formatTitle(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestInitCommand_Metadata(t *testing.T) {
	if initCmd.Use != "init [path]" {
		t.Errorf("Use = %q, want %q", initCmd.Use, "init [path]")
	}

	if initCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// MaximumNArgs(1) allows 0 or 1 args
	if initCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestNameRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid lowercase",
			input: "review",
			want:  true,
		},
		{
			name:  "valid with hyphen",
			input: "code-review",
			want:  true,
		},
		{
			name:  "valid with numbers",
			input: "v2",
			want:  true,
		},
		{
			name:  "invalid uppercase",
			input: "Review",
			want:  false,
		},
		{
			name:  "invalid starts with number",
			input: "2code",
			want:  false,
		},
		{
			name:  "invalid double hyphen",
			input: "code--review",
			want:  false,
		},
		{
			name:  "invalid trailing hyphen",
			input: "code-",
			want:  false,
		},
		{
			name:  "invalid leading hyphen",
			input: "-code",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nameRegex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("nameRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// containsStr checks if substr is in s.
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
