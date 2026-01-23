package commands

import (
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
