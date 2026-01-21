package claude

import (
	"errors"
	"testing"
)

func TestTranslateVariables(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "preserves $ARGUMENTS",
			content: "Run with $ARGUMENTS",
			want:    "Run with $ARGUMENTS",
		},
		{
			name:    "preserves $SELECTION",
			content: "Selected: $SELECTION",
			want:    "Selected: $SELECTION",
		},
		{
			name:    "preserves both variables",
			content: "Command: $ARGUMENTS\nContext: $SELECTION",
			want:    "Command: $ARGUMENTS\nContext: $SELECTION",
		},
		{
			name:    "preserves non-variable content",
			content: "Just plain text without variables",
			want:    "Just plain text without variables",
		},
		{
			name:    "preserves mixed content",
			content: "Before $ARGUMENTS middle $SELECTION after",
			want:    "Before $ARGUMENTS middle $SELECTION after",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
		{
			name:    "preserves unknown variables unchanged",
			content: "Value: $UNKNOWN",
			want:    "Value: $UNKNOWN",
		},
		{
			name:    "preserves special characters",
			content: "$ARGUMENTS with special chars: @#%^&*()",
			want:    "$ARGUMENTS with special chars: @#%^&*()",
		},
		{
			name:    "preserves newlines and whitespace",
			content: "Line 1: $ARGUMENTS\n\nLine 3: $SELECTION\t\ttabs",
			want:    "Line 1: $ARGUMENTS\n\nLine 3: $SELECTION\t\ttabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateVariables(tt.content)
			if got != tt.want {
				t.Errorf("TranslateVariables() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTranslateToCanonical(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "pass-through for $ARGUMENTS",
			content: "Run with $ARGUMENTS",
			want:    "Run with $ARGUMENTS",
		},
		{
			name:    "pass-through for $SELECTION",
			content: "Selected: $SELECTION",
			want:    "Selected: $SELECTION",
		},
		{
			name:    "pass-through for mixed content",
			content: "Before $ARGUMENTS middle $SELECTION after",
			want:    "Before $ARGUMENTS middle $SELECTION after",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
		{
			name:    "no variables",
			content: "Plain text content",
			want:    "Plain text content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToCanonical(tt.content)
			if got != tt.want {
				t.Errorf("TranslateToCanonical() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateVariables(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid $ARGUMENTS",
			content: "Run with $ARGUMENTS",
			wantErr: false,
		},
		{
			name:    "valid $SELECTION",
			content: "Selected: $SELECTION",
			wantErr: false,
		},
		{
			name:    "valid both variables",
			content: "$ARGUMENTS and $SELECTION together",
			wantErr: false,
		},
		{
			name:    "no variables is valid",
			content: "Plain text without variables",
			wantErr: false,
		},
		{
			name:    "empty content is valid",
			content: "",
			wantErr: false,
		},
		{
			name:    "unknown variable",
			content: "Value: $UNKNOWN",
			wantErr: true,
			errMsg:  "$UNKNOWN",
		},
		{
			name:    "multiple unknown variables",
			content: "$FOO and $BAR and $BAZ",
			wantErr: true,
			errMsg:  "$FOO",
		},
		{
			name:    "mixed valid and invalid",
			content: "$ARGUMENTS with $INVALID",
			wantErr: true,
			errMsg:  "$INVALID",
		},
		{
			name:    "repeated unknown variable",
			content: "$UNKNOWN appears twice: $UNKNOWN",
			wantErr: true,
			errMsg:  "$UNKNOWN",
		},
		{
			name:    "lowercase is not a variable",
			content: "$arguments is not matched by pattern",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVariables(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !errors.Is(err, ErrUnsupportedVariable) {
					t.Errorf("ValidateVariables() error should wrap ErrUnsupportedVariable, got %v", err)
				}
				if tt.errMsg != "" && err != nil {
					errStr := err.Error()
					if !containsSubstring(errStr, tt.errMsg) {
						t.Errorf("ValidateVariables() error = %q, should contain %q", errStr, tt.errMsg)
					}
				}
			}
		})
	}
}

func TestListVariables(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "single $ARGUMENTS",
			content: "Run with $ARGUMENTS",
			want:    []string{"$ARGUMENTS"},
		},
		{
			name:    "single $SELECTION",
			content: "Selected: $SELECTION",
			want:    []string{"$SELECTION"},
		},
		{
			name:    "both variables",
			content: "$ARGUMENTS and $SELECTION",
			want:    []string{"$ARGUMENTS", "$SELECTION"},
		},
		{
			name:    "no variables",
			content: "Plain text content",
			want:    []string{},
		},
		{
			name:    "empty content",
			content: "",
			want:    []string{},
		},
		{
			name:    "unknown variable included",
			content: "$ARGUMENTS with $UNKNOWN",
			want:    []string{"$ARGUMENTS", "$UNKNOWN"},
		},
		{
			name:    "duplicate variables deduplicated",
			content: "$ARGUMENTS then $SELECTION then $ARGUMENTS again",
			want:    []string{"$ARGUMENTS", "$SELECTION"},
		},
		{
			name:    "preserves first occurrence order",
			content: "$SELECTION first, then $ARGUMENTS",
			want:    []string{"$SELECTION", "$ARGUMENTS"},
		},
		{
			name:    "multiple unknown variables",
			content: "$FOO $BAR $BAZ",
			want:    []string{"$FOO", "$BAR", "$BAZ"},
		},
		{
			name:    "variables in multiline content",
			content: "Line 1: $ARGUMENTS\nLine 2: $SELECTION\nLine 3: $ARGUMENTS",
			want:    []string{"$ARGUMENTS", "$SELECTION"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ListVariables(tt.content)
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("ListVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVarPattern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches []string
	}{
		{
			name:    "matches uppercase variable",
			input:   "$ARGUMENTS",
			matches: []string{"$ARGUMENTS"},
		},
		{
			name:    "matches variable with underscore",
			input:   "$MY_VAR",
			matches: []string{"$MY_VAR"},
		},
		{
			name:    "does not match lowercase",
			input:   "$arguments",
			matches: nil,
		},
		{
			name:    "does not match mixed case",
			input:   "$Arguments",
			matches: nil,
		},
		{
			name:    "does not match numbers",
			input:   "$VAR123",
			matches: nil,
		},
		{
			name:    "matches at word boundary",
			input:   "text$ARGUMENTS more",
			matches: []string{"$ARGUMENTS"},
		},
		{
			name:    "matches multiple",
			input:   "$FOO $BAR",
			matches: []string{"$FOO", "$BAR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := varPattern.FindAllString(tt.input, -1)
			if !stringSlicesEqual(got, tt.matches) {
				t.Errorf("varPattern.FindAllString(%q) = %v, want %v", tt.input, got, tt.matches)
			}
		})
	}
}

// stringSlicesEqual compares two string slices for equality.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	return len(substr) <= len(s) && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
