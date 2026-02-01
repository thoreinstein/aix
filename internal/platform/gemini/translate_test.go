package gemini

import (
	"strings"
	"testing"
)

func TestTranslateVariables(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Arguments",
			input: "Run command with $ARGUMENTS",
			want:  "Run command with {{argument}}",
		},
		{
			name:  "Selection",
			input: "Process $SELECTION now",
			want:  "Process {{selection}} now",
		},
		{
			name:  "Both",
			input: "$ARGUMENTS and $SELECTION",
			want:  "{{argument}} and {{selection}}",
		},
		{
			name:  "No variables",
			input: "Just plain text",
			want:  "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateVariables(tt.input)
			if got != tt.want {
				t.Errorf("TranslateVariables(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTranslateToCanonical(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Argument",
			input: "Run command with {{argument}}",
			want:  "Run command with $ARGUMENTS",
		},
		{
			name:  "Args (legacy)",
			input: "Run command with {{args}}",
			want:  "Run command with $ARGUMENTS",
		},
		{
			name:  "Selection",
			input: "Process {{selection}} now",
			want:  "Process $SELECTION now",
		},
		{
			name:  "No variables",
			input: "Just plain text",
			want:  "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToCanonical(tt.input)
			if got != tt.want {
				t.Errorf("TranslateToCanonical(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateVariables(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid arguments",
			input:   "Use $ARGUMENTS here",
			wantErr: false,
		},
		{
			name:    "Valid selection",
			input:   "Use $SELECTION here",
			wantErr: false,
		},
		{
			name:    "Invalid variable",
			input:   "Use $INVALID_VAR here",
			wantErr: true,
		},
		{
			name:    "No variables",
			input:   "Plain text",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVariables(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVariables(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), "unsupported variable") {
				t.Errorf("Expected error to contain 'unsupported variable', got: %v", err)
			}
		})
	}
}
