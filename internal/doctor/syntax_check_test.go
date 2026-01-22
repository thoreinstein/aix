package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

func TestConfigSyntaxCheck_Name(t *testing.T) {
	c := NewConfigSyntaxCheck()
	if got := c.Name(); got != "config-syntax" {
		t.Errorf("Name() = %q, want %q", got, "config-syntax")
	}
}

func TestConfigSyntaxCheck_Category(t *testing.T) {
	c := NewConfigSyntaxCheck()
	if got := c.Category(); got != "config" {
		t.Errorf("Category() = %q, want %q", got, "config")
	}
}

func TestConfigSyntaxCheck_ImplementsCheck(t *testing.T) {
	var _ Check = (*ConfigSyntaxCheck)(nil)
}

func TestConfigSyntaxCheck_validateFile(t *testing.T) {
	c := NewConfigSyntaxCheck()

	tests := []struct {
		name       string
		filename   string
		content    string
		wantStatus string
		wantHasMsg bool
	}{
		{
			name:       "valid JSON",
			filename:   "test.json",
			content:    `{"key": "value"}`,
			wantStatus: "pass",
			wantHasMsg: false,
		},
		{
			name:       "valid JSON with nested objects",
			filename:   "test.json",
			content:    `{"servers": {"github": {"command": "npx"}}}`,
			wantStatus: "pass",
			wantHasMsg: false,
		},
		{
			name:       "valid JSON array",
			filename:   "test.json",
			content:    `[1, 2, 3]`,
			wantStatus: "pass",
			wantHasMsg: false,
		},
		{
			name:       "invalid JSON - missing closing brace",
			filename:   "test.json",
			content:    `{"key": "value"`,
			wantStatus: "error",
			wantHasMsg: true,
		},
		{
			name:       "invalid JSON - trailing comma",
			filename:   "test.json",
			content:    `{"key": "value",}`,
			wantStatus: "error",
			wantHasMsg: true,
		},
		{
			name:       "valid TOML",
			filename:   "test.toml",
			content:    "[section]\nkey = \"value\"",
			wantStatus: "pass",
			wantHasMsg: false,
		},
		{
			name:       "valid TOML with arrays",
			filename:   "test.toml",
			content:    "[servers]\nports = [8080, 8443]",
			wantStatus: "pass",
			wantHasMsg: false,
		},
		{
			name:       "invalid TOML - missing value",
			filename:   "test.toml",
			content:    "[section]\nkey = ",
			wantStatus: "error",
			wantHasMsg: true,
		},
		{
			name:       "invalid TOML - unclosed bracket",
			filename:   "test.toml",
			content:    "[section\nkey = \"value\"",
			wantStatus: "error",
			wantHasMsg: true,
		},
		{
			name:       "empty file",
			filename:   "test.json",
			content:    "",
			wantStatus: "pass",
			wantHasMsg: true, // "empty file" message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)

			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			result := c.validateFile(path)

			if result.Status != tt.wantStatus {
				t.Errorf("validateFile() status = %q, want %q (message: %s)", result.Status, tt.wantStatus, result.Message)
			}

			if tt.wantHasMsg && result.Message == "" {
				t.Error("validateFile() expected a message, got empty string")
			}
			if !tt.wantHasMsg && result.Message != "" {
				t.Errorf("validateFile() expected no message, got %q", result.Message)
			}
		})
	}
}

func TestConfigSyntaxCheck_validateFile_nonExistent(t *testing.T) {
	c := NewConfigSyntaxCheck()

	result := c.validateFile("/nonexistent/path/config.json")

	if result.Status != "info" {
		t.Errorf("validateFile() for non-existent file status = %q, want %q", result.Status, "info")
	}
	if result.Message == "" {
		t.Error("validateFile() expected message for non-existent file")
	}
}

func TestConfigSyntaxCheck_validateFile_permission(t *testing.T) {
	// Skip on systems where we can't reliably create unreadable files
	if os.Getuid() == 0 {
		t.Skip("skipping permission test as root")
	}

	c := NewConfigSyntaxCheck()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "unreadable.json")

	if err := os.WriteFile(path, []byte(`{"key": "value"}`), 0000); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := c.validateFile(path)

	if result.Status != "error" {
		t.Errorf("validateFile() for unreadable file status = %q, want %q", result.Status, "error")
	}
	if result.Message == "" {
		t.Error("validateFile() expected message for unreadable file")
	}
}

func TestConfigSyntaxCheck_validateJSON(t *testing.T) {
	c := NewConfigSyntaxCheck()

	tests := []struct {
		name       string
		input      string
		wantStatus string
	}{
		{
			name:       "valid object",
			input:      `{"key": "value"}`,
			wantStatus: "pass",
		},
		{
			name:       "valid array",
			input:      `[1, 2, 3]`,
			wantStatus: "pass",
		},
		{
			name:       "valid string",
			input:      `"hello"`,
			wantStatus: "pass",
		},
		{
			name:       "valid number",
			input:      `42`,
			wantStatus: "pass",
		},
		{
			name:       "valid null",
			input:      `null`,
			wantStatus: "pass",
		},
		{
			name:       "invalid - truncated",
			input:      `{"key": `,
			wantStatus: "error",
		},
		{
			name:       "invalid - extra comma",
			input:      `{"a": 1,}`,
			wantStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := syntaxFileResult{Path: "test.json"}
			result := c.validateJSON([]byte(tt.input), fr)

			if result.Status != tt.wantStatus {
				t.Errorf("validateJSON() status = %q, want %q (message: %s)", result.Status, tt.wantStatus, result.Message)
			}
		})
	}
}

func TestConfigSyntaxCheck_validateTOML(t *testing.T) {
	c := NewConfigSyntaxCheck()

	tests := []struct {
		name       string
		input      string
		wantStatus string
	}{
		{
			name:       "valid simple",
			input:      "key = \"value\"",
			wantStatus: "pass",
		},
		{
			name:       "valid with section",
			input:      "[section]\nkey = \"value\"",
			wantStatus: "pass",
		},
		{
			name:       "valid with array",
			input:      "ports = [80, 443]",
			wantStatus: "pass",
		},
		{
			name:       "invalid - no value",
			input:      "key = ",
			wantStatus: "error",
		},
		{
			name:       "invalid - unclosed string",
			input:      "key = \"unclosed",
			wantStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := syntaxFileResult{Path: "test.toml"}
			result := c.validateTOML([]byte(tt.input), fr)

			if result.Status != tt.wantStatus {
				t.Errorf("validateTOML() status = %q, want %q (message: %s)", result.Status, tt.wantStatus, result.Message)
			}
		})
	}
}

func TestFormatJSONError(t *testing.T) {
	// Test that formatJSONError produces meaningful error messages
	data := []byte(`{"key": "value",}`)
	var v any
	err := json.Unmarshal(data, &v)
	if err == nil {
		t.Fatal("expected JSON error for invalid input")
	}

	msg := formatJSONError(err, data)

	// Should contain line/column info
	if !contains(msg, "line") || !contains(msg, "column") {
		t.Errorf("formatJSONError() = %q, expected to contain line and column info", msg)
	}
}

func TestFormatTOMLError(t *testing.T) {
	// Test that formatTOMLError produces meaningful error messages
	data := []byte("key = ")
	var v any
	err := toml.Unmarshal(data, &v)
	if err == nil {
		t.Fatal("expected TOML error for invalid input")
	}

	msg := formatTOMLError(err)

	// Should contain line/column info
	if !contains(msg, "line") || !contains(msg, "column") {
		t.Errorf("formatTOMLError() = %q, expected to contain line and column info", msg)
	}
}

func TestOffsetToLineCol(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		offset   int
		wantLine int
		wantCol  int
	}{
		{
			name:     "first char",
			data:     "hello\nworld",
			offset:   0,
			wantLine: 1,
			wantCol:  1,
		},
		{
			name:     "end of first line",
			data:     "hello\nworld",
			offset:   5,
			wantLine: 1,
			wantCol:  6,
		},
		{
			name:     "start of second line",
			data:     "hello\nworld",
			offset:   6,
			wantLine: 2,
			wantCol:  1,
		},
		{
			name:     "middle of second line",
			data:     "hello\nworld",
			offset:   8,
			wantLine: 2,
			wantCol:  3,
		},
		{
			name:     "offset beyond length",
			data:     "hello",
			offset:   100,
			wantLine: 1,
			wantCol:  6, // clamped to end
		},
		{
			name:     "negative offset",
			data:     "hello",
			offset:   -5,
			wantLine: 1,
			wantCol:  1,
		},
		{
			name:     "empty data",
			data:     "",
			offset:   0,
			wantLine: 1,
			wantCol:  1,
		},
		{
			name:     "multiple newlines",
			data:     "a\nb\nc\nd",
			offset:   6, // 'd'
			wantLine: 4,
			wantCol:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col := offsetToLineCol([]byte(tt.data), tt.offset)
			if line != tt.wantLine || col != tt.wantCol {
				t.Errorf("offsetToLineCol() = (line=%d, col=%d), want (line=%d, col=%d)", line, col, tt.wantLine, tt.wantCol)
			}
		})
	}
}

func TestConfigSyntaxCheck_getGlobalConfigPath(t *testing.T) {
	c := NewConfigSyntaxCheck()

	tests := []struct {
		platform string
		wantSfx  string
	}{
		{"claude", "settings.json"},
		{"opencode", "config.toml"},
		{"codex", "settings.json"},
		{"gemini", "settings.toml"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			got := c.getGlobalConfigPath(tt.platform)
			if tt.wantSfx == "" {
				if got != "" {
					t.Errorf("getGlobalConfigPath(%q) = %q, want empty", tt.platform, got)
				}
			} else {
				if !hasSuffix(got, tt.wantSfx) {
					t.Errorf("getGlobalConfigPath(%q) = %q, want suffix %q", tt.platform, got, tt.wantSfx)
				}
			}
		})
	}
}

// contains reports whether substr is within s.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr) >= 0))
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// hasSuffix tests whether the string s ends with suffix.
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
