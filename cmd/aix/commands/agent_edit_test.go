package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestAgentEditCmd_Registration(t *testing.T) {
	// Verify command is registered
	cmd, _, err := agentCmd.Find([]string{"edit"})
	if err != nil {
		t.Fatalf("agent edit command not registered: %v", err)
	}
	if cmd.Use != "edit <name>" {
		t.Errorf("unexpected Use: got %q, want %q", cmd.Use, "edit <name>")
	}
}

func TestAgentEditCmd_RequiresArg(t *testing.T) {
	// Verify command requires exactly one argument
	if agentEditCmd.Args == nil {
		t.Fatal("Args validator not set")
	}

	// Test with no args - should error
	err := agentEditCmd.Args(agentEditCmd, []string{})
	if err == nil {
		t.Error("expected error with no args, got nil")
	}

	// Test with one arg - should pass
	err = agentEditCmd.Args(agentEditCmd, []string{"test-agent"})
	if err != nil {
		t.Errorf("expected no error with one arg, got: %v", err)
	}

	// Test with two args - should error
	err = agentEditCmd.Args(agentEditCmd, []string{"arg1", "arg2"})
	if err == nil {
		t.Error("expected error with two args, got nil")
	}
}

func TestAgentEditCmd_Metadata(t *testing.T) {
	if agentEditCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if agentEditCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestAgentEdit_LocalPathResolution(t *testing.T) {
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
	err = runAgentEdit(agentEditCmd, []string{tmpFile.Name()})
	if err != nil {
		t.Errorf("expected success with EDITOR=true, got: %v", err)
	}
}

func TestAgentEdit_NotFound(t *testing.T) {
	// Test with non-existent agent
	err := runAgentEdit(agentEditCmd, []string{"nonexistent-agent-xyz"})
	if err == nil {
		t.Error("expected error for non-existent agent")
	}
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestAgentEdit_PostEditValidationRuns(t *testing.T) {
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
	cmd := *agentEditCmd
	cmd.SetOut(&buf)

	// Run the edit command
	err := runAgentEdit(&cmd, []string{agentPath})
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

func TestAgentEdit_PostEditValidationShowsErrors(t *testing.T) {
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
	cmd := *agentEditCmd
	cmd.SetOut(&buf)

	// Run the edit command - should succeed even with invalid agent
	err := runAgentEdit(&cmd, []string{agentPath})
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

func TestAgentEdit_ValidationDoesNotRunWhenEditorFails(t *testing.T) {
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
	cmd := *agentEditCmd
	cmd.SetOut(&buf)

	// Run the edit command - should fail because editor fails
	err := runAgentEdit(&cmd, []string{agentPath})
	if err == nil {
		t.Error("expected error when editor fails, got nil")
	}

	// Verify validation was NOT run (since editor failed)
	output := buf.String()
	if strings.Contains(output, "Validating agent...") {
		t.Errorf("validation should not run when editor fails, got: %s", output)
	}
}
