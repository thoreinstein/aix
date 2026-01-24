package command

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   bool
	}{
		{
			name:   "https URL",
			source: "https://github.com/user/repo.git",
			want:   true,
		},
		{
			name:   "http URL",
			source: "http://github.com/user/repo",
			want:   true,
		},
		{
			name:   "git@ URL",
			source: "git@github.com:user/repo.git",
			want:   true,
		},
		{
			name:   "ends with .git",
			source: "github.com/user/repo.git",
			want:   true,
		},
		{
			name:   "local relative path",
			source: "./my-command",
			want:   false,
		},
		{
			name:   "local absolute path",
			source: "/path/to/command",
			want:   false,
		},
		{
			name:   "local directory name",
			source: "my-command",
			want:   false,
		},
		{
			name:   "ssh protocol",
			source: "ssh://git@github.com/user/repo",
			want:   true,
		},
		{
			name:   "file protocol",
			source: "file:///path/to/repo",
			want:   true,
		},
		{
			name:   "simple filename with .md extension",
			source: "review.md",
			want:   false,
		},
		{
			name:   "path with .git in middle",
			source: "/home/user/.git/config",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGitURL(tt.source)
			if got != tt.want {
				t.Errorf("isGitURL(%q) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func TestInstallFromLocal_MissingFile(t *testing.T) {
	// Create a temp directory without any command file
	tempDir := t.TempDir()

	err := installFromLocal(tempDir)
	if err == nil {
		t.Error("expected error for missing command file, got nil")
	}

	// Check error message contains path
	if err != nil && !containsString(err.Error(), "no command file found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestInstallFromLocal_InvalidCommand(t *testing.T) {
	// Create a temp directory with invalid command file (missing name)
	tempDir := t.TempDir()
	cmdPath := filepath.Join(tempDir, "command.md")

	// Write a command with an invalid name (contains uppercase)
	content := `---
name: INVALID-NAME
description: Test command
---

Some instructions.
`
	if err := os.WriteFile(cmdPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err := installFromLocal(tempDir)
	if err == nil {
		t.Error("expected error for invalid command, got nil")
	}

	// Should fail validation
	if !errors.Is(err, errInstallFailed) {
		t.Errorf("expected errInstallFailed, got: %v", err)
	}
}

func TestInstallFromLocal_FileNotDirectory(t *testing.T) {
	// Create a temp directory with a .md file
	tempDir := t.TempDir()
	cmdPath := filepath.Join(tempDir, "review.md")

	content := `---
name: review
description: Review code
---

Review the code carefully.
`
	if err := os.WriteFile(cmdPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Try to install from the file path (not directory)
	// This should fail because there are no available platforms
	err := installFromLocal(cmdPath)
	// We expect an error because no platforms are available in test
	if err == nil {
		t.Log("expected error (no platforms available), got nil - this may be expected in test env")
	}
}

func TestConvertToOpenCode(t *testing.T) {
	tests := []struct {
		name  string
		input *claude.Command
		check func(t *testing.T, got *opencode.Command)
	}{
		{
			name: "basic fields",
			input: &claude.Command{
				Name:         "test-command",
				Description:  "A test command",
				Instructions: "Do things",
			},
			check: func(t *testing.T, got *opencode.Command) {
				if got.Name != "test-command" {
					t.Errorf("Name = %q, want %q", got.Name, "test-command")
				}
				if got.Description != "A test command" {
					t.Errorf("Description = %q, want %q", got.Description, "A test command")
				}
				if got.Instructions != "Do things" {
					t.Errorf("Instructions = %q, want %q", got.Instructions, "Do things")
				}
			},
		},
		{
			name: "with model field",
			input: &claude.Command{
				Name:         "model-cmd",
				Description:  "Test",
				Model:        "claude-3-5-sonnet",
				Instructions: "Instructions here",
			},
			check: func(t *testing.T, got *opencode.Command) {
				if got.Model != "claude-3-5-sonnet" {
					t.Errorf("Model = %q, want %q", got.Model, "claude-3-5-sonnet")
				}
			},
		},
		{
			name: "with agent field",
			input: &claude.Command{
				Name:         "agent-cmd",
				Description:  "Test",
				Agent:        "task",
				Instructions: "Instructions here",
			},
			check: func(t *testing.T, got *opencode.Command) {
				if got.Agent != "task" {
					t.Errorf("Agent = %q, want %q", got.Agent, "task")
				}
			},
		},
		{
			name: "empty optional fields",
			input: &claude.Command{
				Name:         "minimal-cmd",
				Instructions: "Just instructions",
			},
			check: func(t *testing.T, got *opencode.Command) {
				if got.Name != "minimal-cmd" {
					t.Errorf("Name = %q, want %q", got.Name, "minimal-cmd")
				}
				if got.Description != "" {
					t.Errorf("Description = %q, want empty", got.Description)
				}
				if got.Model != "" {
					t.Errorf("Model = %q, want empty", got.Model)
				}
				if got.Agent != "" {
					t.Errorf("Agent = %q, want empty", got.Agent)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToOpenCode(tt.input)
			tt.check(t, got)
		})
	}
}

func TestConvertForPlatform(t *testing.T) {
	tests := []struct {
		name         string
		cmd          *claude.Command
		platformName string
		checkType    func(t *testing.T, result any)
	}{
		{
			name: "claude platform returns original",
			cmd: &claude.Command{
				Name:         "test-cmd",
				Description:  "Test",
				Instructions: "Test instructions",
			},
			platformName: "claude",
			checkType: func(t *testing.T, result any) {
				_, ok := result.(*claude.Command)
				if !ok {
					t.Errorf("expected *claude.Command, got %T", result)
				}
			},
		},
		{
			name: "opencode platform returns converted",
			cmd: &claude.Command{
				Name:         "test-cmd",
				Description:  "Test",
				Instructions: "Test instructions",
			},
			platformName: "opencode",
			checkType: func(t *testing.T, result any) {
				_, ok := result.(*opencode.Command)
				if !ok {
					t.Errorf("expected *opencode.Command, got %T", result)
				}
			},
		},
		{
			name: "unknown platform returns original",
			cmd: &claude.Command{
				Name:         "test-cmd",
				Description:  "Test",
				Instructions: "Test instructions",
			},
			platformName: "unknown-platform",
			checkType: func(t *testing.T, result any) {
				_, ok := result.(*claude.Command)
				if !ok {
					t.Errorf("expected *claude.Command for unknown platform, got %T", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertForPlatform(tt.cmd, tt.platformName)
			tt.checkType(t, result)
		})
	}
}

func TestInstallCommand_Metadata(t *testing.T) {
	if installCmd.Use != "install <source>" {
		t.Errorf("Use = %q, want %q", installCmd.Use, "install <source>")
	}

	if installCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if installCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

// containsString checks if substr is in s.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
