package skill

import (
	"os"
	"path/filepath"
	"strings"
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
			err := validateName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateName(%q) = nil, want error containing %q", tt.input, tt.errMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateName(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("validateName(%q) = %v, want nil", tt.input, err)
			}
		})
	}
}

func TestInitCommand_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Reset flags
	initDescription = ""
	initForce = false
	initName = "test-skill"
	initDirs = "docs"

	err := initCmd.RunE(initCmd, []string{"test-skill"})
	if err != nil {
		t.Fatalf("skill init failed: %v", err)
	}

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
}

func TestInitCommand_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	tests := []struct {
		testName  string
		skillName string
		wantErr   string
	}{
		{"Invalid-Caps", "Invalid-Caps", "lowercase alphanumeric"},
		{"starts with number", "123-starts-with-number", "lowercase alphanumeric"},
		{"has underscore", "has_underscore", "lowercase alphanumeric"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			// Reset flags
			initDescription = ""
			initForce = false
			initName = tt.skillName
			initDirs = ""

			err := initCmd.RunE(initCmd, []string{"."})
			if err == nil {
				t.Errorf("expected error for invalid name %q, got nil", tt.skillName)
			}
		})
	}
}

func TestInitCommand_CommandMetadata(t *testing.T) {
	if initCmd.Use != "init [path]" {
		t.Errorf("Use = %q, want %q", initCmd.Use, "init [path]")
	}

	if initCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if initCmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Verify flags are registered
	descFlag := initCmd.Flags().Lookup("description")
	if descFlag == nil {
		t.Fatal("--description flag not registered")
	}
	if descFlag.Shorthand != "d" {
		t.Errorf("--description shorthand = %q, want %q", descFlag.Shorthand, "d")
	}

	forceFlag := initCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("--force flag not registered")
	}
	if forceFlag.Shorthand != "f" {
		t.Errorf("--force shorthand = %q, want %q", forceFlag.Shorthand, "f")
	}
}
