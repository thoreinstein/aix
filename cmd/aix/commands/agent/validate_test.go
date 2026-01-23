package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidate(t *testing.T) {
	tests := []struct {
		name        string
		content     string // file content, empty string means file doesn't exist
		strict      bool
		jsonOutput  bool
		wantErr     bool
		wantContain string // substring to check in output
	}{
		{
			name: "valid agent file",
			content: `---
name: test-agent
description: A test agent
---

Instructions for the agent.
`,
			strict:      false,
			jsonOutput:  false,
			wantErr:     false,
			wantContain: "✓ Agent 'test-agent' is valid",
		},
		{
			name: "valid agent with name only",
			content: `---
name: minimal-agent
---

Instructions here.
`,
			strict:      false,
			jsonOutput:  false,
			wantErr:     false,
			wantContain: "✓ Agent 'minimal-agent' is valid",
		},
		{
			name: "missing name returns error",
			content: `---
description: A test agent
---

Instructions for the agent.
`,
			strict:      false,
			jsonOutput:  false,
			wantErr:     true,
			wantContain: "✗ Agent '(unknown)' is invalid",
		},
		{
			name:        "file not found",
			content:     "", // empty means don't create file
			strict:      false,
			jsonOutput:  false,
			wantErr:     true,
			wantContain: "file not found",
		},
		{
			name:        "empty file",
			content:     "   \n\t\n   ", // whitespace only
			strict:      false,
			jsonOutput:  false,
			wantErr:     true,
			wantContain: "agent file is empty",
		},
		{
			name: "invalid YAML frontmatter",
			content: `---
name: [invalid yaml
---

Instructions here.
`,
			strict:      false,
			jsonOutput:  false,
			wantErr:     true,
			wantContain: "invalid YAML frontmatter",
		},
		{
			name: "JSON output format valid",
			content: `---
name: json-test
description: Testing JSON output
---

Instructions.
`,
			strict:      false,
			jsonOutput:  true,
			wantErr:     false,
			wantContain: `"valid": true`,
		},
		{
			name: "JSON output format invalid",
			content: `---
description: Missing name
---

Instructions.
`,
			strict:      false,
			jsonOutput:  true,
			wantErr:     true,
			wantContain: `"valid": false`,
		},
		{
			name: "strict mode with missing description generates warning",
			content: `---
name: strict-test
---

Instructions here.
`,
			strict:      true,
			jsonOutput:  false,
			wantErr:     false, // warnings don't cause errors
			wantContain: "⚠",   // warning indicator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			validateStrict = tt.strict
			validateJSON = tt.jsonOutput

			dir := t.TempDir()
			path := filepath.Join(dir, "AGENT.md")

			// Only create file if content is non-empty
			if tt.content != "" {
				if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			var buf bytes.Buffer
			err := runValidate(path, &buf)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runValidate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify the specific error type when expected
			if tt.wantErr && err != nil {
				if !errors.Is(err, errValidationFailed) {
					t.Errorf("expected errValidationFailed, got %v", err)
				}
			}

			// Check output contains expected string
			output := buf.String()
			if tt.wantContain != "" && !strings.Contains(output, tt.wantContain) {
				t.Errorf("output = %q, want contain %q", output, tt.wantContain)
			}
		})
	}
}

func TestRunValidate_JSONStructure(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		strict         bool
		wantValid      bool
		wantAgentName  string
		wantParseError bool
		wantErrors     bool
		wantWarnings   bool
	}{
		{
			name: "valid JSON structure",
			content: `---
name: json-struct-test
description: Test description
---

Instructions.
`,
			strict:        false,
			wantValid:     true,
			wantAgentName: "json-struct-test",
		},
		{
			name: "JSON with validation errors",
			content: `---
description: No name
---

Instructions.
`,
			strict:     false,
			wantValid:  false,
			wantErrors: true,
		},
		{
			name: "JSON with parse error",
			content: `---
name: [broken
---
`,
			strict:         false,
			wantValid:      false,
			wantParseError: true,
		},
		{
			name: "JSON with warnings in strict mode",
			content: `---
name: strict-json
---

Instructions.
`,
			strict:       true,
			wantValid:    true, // valid despite warnings
			wantWarnings: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateStrict = tt.strict
			validateJSON = true

			dir := t.TempDir()
			path := filepath.Join(dir, "AGENT.md")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			var buf bytes.Buffer
			_ = runValidate(path, &buf)

			var result validateResult
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Fatalf("failed to parse JSON output: %v\nOutput:\n%s", err, buf.String())
			}

			if result.Valid != tt.wantValid {
				t.Errorf("valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if tt.wantAgentName != "" {
				if result.Agent == nil {
					t.Fatal("expected agent info in result")
				}
				if result.Agent.Name != tt.wantAgentName {
					t.Errorf("agent name = %q, want %q", result.Agent.Name, tt.wantAgentName)
				}
			}

			if tt.wantParseError && result.ParseError == "" {
				t.Error("expected parse error in result")
			}

			if tt.wantErrors && len(result.Errors) == 0 {
				t.Error("expected errors in result")
			}

			if tt.wantWarnings && len(result.Warnings) == 0 {
				t.Error("expected warnings in result")
			}

			// Verify path is included
			if result.Path == "" {
				t.Error("expected path in result")
			}

			// Verify strict mode is correctly reported
			if result.StrictMode != tt.strict {
				t.Errorf("strict_mode = %v, want %v", result.StrictMode, tt.strict)
			}
		})
	}
}

func TestRunValidate_OutputDetails(t *testing.T) {
	t.Run("shows agent details on success", func(t *testing.T) {
		validateStrict = false
		validateJSON = false

		content := `---
name: detailed-agent
description: A detailed description
---

Instructions.
`
		dir := t.TempDir()
		path := filepath.Join(dir, "AGENT.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		var buf bytes.Buffer
		err := runValidate(path, &buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()

		// Check name is displayed
		if !strings.Contains(output, "Name:        detailed-agent") {
			t.Errorf("expected name in output, got:\n%s", output)
		}

		// Check description is displayed
		if !strings.Contains(output, "Description: A detailed description") {
			t.Errorf("expected description in output, got:\n%s", output)
		}
	})

	t.Run("shows errors list on validation failure", func(t *testing.T) {
		validateStrict = false
		validateJSON = false

		content := `---
description: No name field
---

Instructions.
`
		dir := t.TempDir()
		path := filepath.Join(dir, "AGENT.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		var buf bytes.Buffer
		_ = runValidate(path, &buf)

		output := buf.String()

		// Check for errors section
		if !strings.Contains(output, "Errors:") {
			t.Errorf("expected Errors section in output, got:\n%s", output)
		}

		// Check for specific error
		if !strings.Contains(output, "name: name is required") {
			t.Errorf("expected name required error in output, got:\n%s", output)
		}
	})

	t.Run("shows parse error details", func(t *testing.T) {
		validateStrict = false
		validateJSON = false

		content := `---
invalid: yaml: content: [
---
`
		dir := t.TempDir()
		path := filepath.Join(dir, "AGENT.md")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		var buf bytes.Buffer
		_ = runValidate(path, &buf)

		output := buf.String()

		// Check for parse error section
		if !strings.Contains(output, "Parse error:") {
			t.Errorf("expected Parse error section in output, got:\n%s", output)
		}
	})
}

func TestRunValidate_PermissionDenied(t *testing.T) {
	// Skip on systems where we can't reliably test permissions
	if os.Geteuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	validateStrict = false
	validateJSON = false

	dir := t.TempDir()
	path := filepath.Join(dir, "AGENT.md")

	// Create file then make it unreadable
	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("failed to chmod test file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(path, 0o644) // Restore for cleanup
	})

	var buf bytes.Buffer
	err := runValidate(path, &buf)

	if err == nil {
		t.Error("expected error for permission denied")
	}

	output := buf.String()
	if !strings.Contains(output, "permission denied") {
		t.Errorf("expected permission denied in output, got:\n%s", output)
	}
}

func TestValidateCommand_Metadata(t *testing.T) {
	if validateCmd.Use != "validate <path>" {
		t.Errorf("Use = %q, want %q", validateCmd.Use, "validate <path>")
	}

	if validateCmd.Short == "" {
		t.Error("Short should not be empty")
	}

	if validateCmd.Long == "" {
		t.Error("Long should not be empty")
	}

	// Verify flags are registered
	strictFlag := validateCmd.Flags().Lookup("strict")
	if strictFlag == nil {
		t.Error("--strict flag not registered")
	}

	jsonFlag := validateCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("--json flag not registered")
	}
}
