package agent

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestEditCmd_Registration(t *testing.T) {
	// Verify command is registered
	cmd, _, err := Cmd.Find([]string{"edit"})
	if err != nil {
		t.Fatalf("agent edit command not registered: %v", err)
	}
	if cmd.Use != "edit <name>" {
		t.Errorf("unexpected Use: got %q, want %q", cmd.Use, "edit <name>")
	}
}

func TestEditCmd_RequiresArg(t *testing.T) {
	// Verify command requires exactly one argument
	if editCmd.Args == nil {
		t.Fatal("Args validator not set")
	}

	// Test with no args - should error
	err := editCmd.Args(editCmd, []string{})
	if err == nil {
		t.Error("expected error with no args, got nil")
	}

	// Test with one arg - should pass
	err = editCmd.Args(editCmd, []string{"test-agent"})
	if err != nil {
		t.Errorf("expected no error with one arg, got: %v", err)
	}

	// Test with two args - should error
	err = editCmd.Args(editCmd, []string{"arg1", "arg2"})
	if err == nil {
		t.Error("expected error with two args, got nil")
	}
}

func TestEditCmd_Metadata(t *testing.T) {
	if editCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if editCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestEdit_LocalPathResolution(t *testing.T) {
	// Set EDITOR to a command that exits immediately without blocking
	t.Setenv("EDITOR", "true")

	// Create a temp file to test local path resolution
	tmpFile, err := os.CreateTemp(t.TempDir(), "test-agent-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	// The command should resolve and "edit" the local path
	// (with EDITOR=true, it just returns success immediately)
	err = runEdit(editCmd, []string{tmpFile.Name()})
	if err != nil {
		t.Errorf("expected success with EDITOR=true, got: %v", err)
	}
}

func TestEdit_NotFound(t *testing.T) {
	// Test with non-existent agent
	err := runEdit(editCmd, []string{"nonexistent-agent-xyz"})
	if err == nil {
		t.Error("expected error for non-existent agent")
	}
	// Accept either "not found" (when platforms exist) or "no platforms available" (CI environment)
	// Both indicate the agent couldn't be found/accessed.
	errStr := err.Error()
	if !strings.Contains(errStr, "not found") && !strings.Contains(errStr, "no platforms available") {
		t.Errorf("expected 'not found' or 'no platforms available' error, got: %v", err)
	}
}

func TestEdit_PostEditValidationRuns(t *testing.T) {
	// Set EDITOR to a command that exits immediately without blocking
	t.Setenv("EDITOR", "true")

	// Create a valid agent file
	tmpDir := t.TempDir()
	agentPath := tmpDir + "/valid-agent.md"
	validContent := `---
name: test-agent
description: A test agent
---

You are a helpful assistant.
`
	if err := os.WriteFile(agentPath, []byte(validContent), 0o644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	cmd := *editCmd
	cmd.SetOut(&buf)

	// Run the edit command
	err := runEdit(&cmd, []string{agentPath})
	if err != nil {
		t.Errorf("expected success, got: %v", err)
	}

	// Verify validation output is present
	output := buf.String()
	if !strings.Contains(output, "Validating agent...") {
		t.Errorf("expected output to contain 'Validating agent...', got: %s", output)
	}
	if !strings.Contains(output, "✓") {
		t.Errorf("expected output to contain validation success marker, got: %s", output)
	}
}

func TestEdit_PostEditValidationShowsErrors(t *testing.T) {
	// Set EDITOR to a command that exits immediately without blocking
	t.Setenv("EDITOR", "true")

	// Create an invalid agent file (missing required fields)
	tmpDir := t.TempDir()
	agentPath := tmpDir + "/invalid-agent.md"
	invalidContent := `---
---

Just some text without proper metadata.
`
	if err := os.WriteFile(agentPath, []byte(invalidContent), 0o644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	cmd := *editCmd
	cmd.SetOut(&buf)

	// Run the edit command - should succeed even with invalid agent
	err := runEdit(&cmd, []string{agentPath})
	if err != nil {
		t.Errorf("expected success even with invalid agent, got: %v", err)
	}

	// Verify validation output shows errors
	output := buf.String()
	if !strings.Contains(output, "Validating agent...") {
		t.Errorf("expected output to contain 'Validating agent...', got: %s", output)
	}
	if !strings.Contains(output, "✗") {
		t.Errorf("expected output to contain validation error marker, got: %s", output)
	}
}

func TestEdit_ValidationDoesNotRunWhenEditorFails(t *testing.T) {
	// Set EDITOR to a command that fails (returns non-zero exit code)
	t.Setenv("EDITOR", "false")

	// Create a valid agent file
	tmpDir := t.TempDir()
	agentPath := tmpDir + "/test-agent.md"
	validContent := `---
name: test-agent
description: A test agent
---

You are a helpful assistant.
`
	if err := os.WriteFile(agentPath, []byte(validContent), 0o644); err != nil {
		t.Fatalf("failed to write agent file: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	cmd := *editCmd
	cmd.SetOut(&buf)

	// Run the edit command - should fail because editor fails
	err := runEdit(&cmd, []string{agentPath})
	if err == nil {
		t.Error("expected error when editor fails, got nil")
	}

	// Verify validation was NOT run (since editor failed)
	output := buf.String()
	if strings.Contains(output, "Validating agent...") {
		t.Errorf("validation should not run when editor fails, got: %s", output)
	}
}
