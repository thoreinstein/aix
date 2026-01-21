package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSkillName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple name",
			input:   "my-skill",
			wantErr: false,
		},
		{
			name:    "valid single letter",
			input:   "a",
			wantErr: false,
		},
		{
			name:    "valid alphanumeric",
			input:   "skill123",
			wantErr: false,
		},
		{
			name:    "valid with hyphens",
			input:   "my-cool-skill-v2",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
			errMsg:  "skill name is required",
		},
		{
			name:    "starts with number",
			input:   "123skill",
			wantErr: true,
			errMsg:  "lowercase alphanumeric with hyphens",
		},
		{
			name:    "starts with hyphen",
			input:   "-skill",
			wantErr: true,
			errMsg:  "lowercase alphanumeric with hyphens",
		},
		{
			name:    "uppercase letters",
			input:   "MySkill",
			wantErr: true,
			errMsg:  "lowercase alphanumeric with hyphens",
		},
		{
			name:    "contains underscore",
			input:   "my_skill",
			wantErr: true,
			errMsg:  "lowercase alphanumeric with hyphens",
		},
		{
			name:    "contains space",
			input:   "my skill",
			wantErr: true,
			errMsg:  "lowercase alphanumeric with hyphens",
		},
		{
			name:    "too long (65 chars)",
			input:   strings.Repeat("a", 65),
			wantErr: true,
			errMsg:  "at most 64 characters",
		},
		{
			name:    "exactly 64 chars is valid",
			input:   "a" + strings.Repeat("b", 63),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSkillName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateSkillName(%q) = nil, want error containing %q", tt.input, tt.errMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateSkillName(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateSkillName(%q) = %v, want nil", tt.input, err)
			}
		})
	}
}

func TestSkillInitCommand_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Reset flags
	skillInitDescription = ""
	skillInitForce = false

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "init", "test-skill", "--name", "test-skill", "--dirs", "docs"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("skill init failed: %v", err)
		}
	})

	// Verify directory was created
	skillDir := filepath.Join(tmpDir, "test-skill")
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		t.Error("skill directory was not created")
	}

	// Verify SKILL.md was created
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Error("SKILL.md was not created")
	}

	// Verify content
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}

	if !strings.Contains(string(content), "name: test-skill") {
		t.Error("SKILL.md does not contain skill name")
	}

	// Verify output messages
	if !strings.Contains(output, "✓ Skill 'test-skill' created") {
		t.Error("output missing success message")
	}
	if !strings.Contains(output, "Next steps:") {
		t.Error("output missing next steps")
	}
}

func TestSkillInitCommand_WithDescriptionFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Reset flags
	skillInitDescription = ""
	skillInitForce = false

	_ = captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "init", "-d", "My custom description", "my-skill", "--name", "my-skill", "--dirs", "docs"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("skill init failed: %v", err)
		}
	})

	skillFile := filepath.Join(tmpDir, "my-skill", "SKILL.md")
	content, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}

	if !strings.Contains(string(content), "description: My custom description") {
		t.Errorf("SKILL.md does not contain custom description, got:\n%s", string(content))
	}
}

func TestSkillInitCommand_ExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create existing directory
	existingDir := filepath.Join(tmpDir, "existing-skill")
	if err := os.MkdirAll(existingDir, 0o755); err != nil {
		t.Fatalf("failed to create existing directory: %v", err)
	}
	// Create SKILL.md to force collision
	if err := os.WriteFile(filepath.Join(existingDir, "SKILL.md"), []byte("exists"), 0o644); err != nil {
		t.Fatalf("failed to create existing SKILL.md: %v", err)
	}

	// Reset flags - importantly no force
	skillInitDescription = ""
	skillInitForce = false

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "init", "existing-skill", "--name", "existing-skill", "--dirs", "docs"})
		err := rootCmd.Execute()
		// Should fail without --force
		if err == nil {
			t.Error("expected error when directory exists, got nil")
		}
	})

	if !strings.Contains(output, "already exists") {
		t.Errorf("expected 'already exists' in output, got:\n%s", output)
	}
}

func TestSkillInitCommand_ForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create existing directory with some content
	existingDir := filepath.Join(tmpDir, "force-skill")
	if err := os.MkdirAll(existingDir, 0o755); err != nil {
		t.Fatalf("failed to create existing directory: %v", err)
	}
	oldFile := filepath.Join(existingDir, "old-file.txt")
	if err := os.WriteFile(oldFile, []byte("old content"), 0o644); err != nil {
		t.Fatalf("failed to write old file: %v", err)
	}

	// Reset flags - with force
	skillInitDescription = ""
	skillInitForce = false

	_ = captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "init", "--force", "force-skill", "--name", "force-skill", "--dirs", "docs"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("skill init with --force failed: %v", err)
		}
	})

	// Verify SKILL.md was created
	skillFile := filepath.Join(existingDir, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Error("SKILL.md was not created with --force")
	}
}

func TestSkillInitCommand_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	tests := []struct {
		name    string
		wantErr string
	}{
		{"Invalid-Caps", "lowercase alphanumeric"},
		{"123-starts-with-number", "lowercase alphanumeric"},
		{"has_underscore", "lowercase alphanumeric"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			skillInitDescription = ""
			skillInitForce = false

			output := captureStdout(t, func() {
				rootCmd.SetArgs([]string{"skill", "init", ".", "--name", tt.name})
				err := rootCmd.Execute()
				if err == nil {
					t.Errorf("expected error for invalid name %q, got nil", tt.name)
				}
			})

			if !strings.Contains(output, tt.wantErr) {
				t.Errorf("expected error containing %q for name %q, got:\n%s", tt.wantErr, tt.name, output)
			}
		})
	}
}

func TestSkillInitCommand_CommandMetadata(t *testing.T) {
	if skillInitCmd.Use != "init [path]" {
		t.Errorf("Use = %q, want %q", skillInitCmd.Use, "init [path]")
	}

	if skillInitCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if skillInitCmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify flags are registered
	descFlag := skillInitCmd.Flags().Lookup("description")
	if descFlag == nil {
		t.Fatal("--description flag not registered")
	}
	if descFlag.Shorthand != "d" {
		t.Errorf("--description shorthand = %q, want %q", descFlag.Shorthand, "d")
	}

	forceFlag := skillInitCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("--force flag not registered")
	}
	if forceFlag.Shorthand != "f" {
		t.Errorf("--force shorthand = %q, want %q", forceFlag.Shorthand, "f")
	}
}
