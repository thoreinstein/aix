package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillValidateCommand_ValidSkill(t *testing.T) {
	// Create a temp directory with a valid skill
	skillDir := setupValidSkill(t, "test-skill")

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", skillDir})
		validateStrict = false
		validateJSON = false
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("expected no error for valid skill, got: %v", err)
		}
	})

	if !strings.Contains(output, "✓ Skill 'test-skill' is valid") {
		t.Errorf("expected success message in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Name:        test-skill") {
		t.Errorf("expected name in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Description: A test skill") {
		t.Errorf("expected description in output, got:\n%s", output)
	}
}

func TestSkillValidateCommand_ValidSkillWithLicense(t *testing.T) {
	// Create a temp directory with a valid skill that has a license
	skillDir := setupSkillWithContent(t, "licensed-skill", `---
name: licensed-skill
description: A skill with a license
license: MIT
---
Instructions here.
`)

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", skillDir})
		validateStrict = false
		validateJSON = false
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("expected no error for valid skill, got: %v", err)
		}
	})

	if !strings.Contains(output, "License:     MIT") {
		t.Errorf("expected license in output, got:\n%s", output)
	}
}

func TestSkillValidateCommand_InvalidSkill_MissingName(t *testing.T) {
	skillDir := setupSkillWithContent(t, "invalid-skill", `---
description: Missing name field
---
Instructions here.
`)

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", skillDir})
		validateStrict = false
		validateJSON = false
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid skill, got nil")
		}
	})

	if !strings.Contains(output, "✗ Skill validation failed") {
		t.Errorf("expected failure message in output, got:\n%s", output)
	}
	if !strings.Contains(output, "name: name is required") {
		t.Errorf("expected name error in output, got:\n%s", output)
	}
}

func TestSkillValidateCommand_InvalidSkill_MissingDescription(t *testing.T) {
	skillDir := setupSkillWithContent(t, "no-desc", `---
name: no-desc
---
Instructions here.
`)

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", skillDir})
		validateStrict = false
		validateJSON = false
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid skill, got nil")
		}
	})

	if !strings.Contains(output, "description: description is required") {
		t.Errorf("expected description error in output, got:\n%s", output)
	}
}

func TestSkillValidateCommand_InvalidSkill_NameMismatch(t *testing.T) {
	// Create a skill where the name doesn't match the directory
	skillDir := setupSkillWithContent(t, "mismatch-dir", `---
name: different-name
description: Name doesn't match directory
---
Instructions here.
`)

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", skillDir})
		validateStrict = false
		validateJSON = false
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error for name mismatch, got nil")
		}
	})

	if !strings.Contains(output, "name: skill name must match directory name") {
		t.Errorf("expected name mismatch error in output, got:\n%s", output)
	}
}

func TestSkillValidateCommand_NotFound(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "nonexistent")

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", nonexistent})
		validateStrict = false
		validateJSON = false
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error for nonexistent skill, got nil")
		}
	})

	if !strings.Contains(output, "SKILL.md not found in directory") {
		t.Errorf("expected not found error in output, got:\n%s", output)
	}
}

func TestSkillValidateCommand_JSONOutput_Valid(t *testing.T) {
	skillDir := setupValidSkill(t, "json-test")

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", "--json", skillDir})
		validateStrict = false
		validateJSON = true
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("expected no error for valid skill, got: %v", err)
		}
	})

	var result validateResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput:\n%s", err, output)
	}

	if !result.Valid {
		t.Error("expected valid=true")
	}
	if result.Skill == nil {
		t.Fatal("expected skill info in result")
	}
	if result.Skill.Name != "json-test" {
		t.Errorf("expected name 'json-test', got %q", result.Skill.Name)
	}
	if result.Skill.Description != "A test skill" {
		t.Errorf("expected description 'A test skill', got %q", result.Skill.Description)
	}
}

func TestSkillValidateCommand_JSONOutput_Invalid(t *testing.T) {
	skillDir := setupSkillWithContent(t, "json-invalid", `---
description: Missing name
---
Instructions here.
`)

	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", "--json", skillDir})
		validateStrict = false
		validateJSON = true
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid skill, got nil")
		}
	})

	var result validateResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput:\n%s", err, output)
	}

	if result.Valid {
		t.Error("expected valid=false")
	}
	if len(result.Errors) == 0 {
		t.Error("expected errors in result")
	}
}

func TestSkillValidateCommand_StrictMode(t *testing.T) {
	// Create a skill with invalid allowed-tools syntax
	skillDir := setupSkillWithContent(t, "strict-test", `---
name: strict-test
description: Test strict validation
allowed-tools: InvalidToolSyntax!!!
---
Instructions here.
`)

	// Without --strict, should pass
	output := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", skillDir})
		validateStrict = false
		validateJSON = false
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("expected no error without --strict, got: %v", err)
		}
	})
	if !strings.Contains(output, "✓ Skill") {
		t.Errorf("expected success without --strict, got:\n%s", output)
	}

	// With --strict, should fail
	output = captureStdout(t, func() {
		rootCmd.SetArgs([]string{"skill", "validate", "--strict", skillDir})
		validateStrict = true
		validateJSON = false
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error with --strict, got nil")
		}
	})
	if !strings.Contains(output, "✗ Skill validation failed") {
		t.Errorf("expected failure with --strict, got:\n%s", output)
	}
	if !strings.Contains(output, "allowed-tools:") {
		t.Errorf("expected allowed-tools error with --strict, got:\n%s", output)
	}
}

func TestSkillValidateCommand_CommandMetadata(t *testing.T) {
	if skillValidateCmd.Use != "validate <path>" {
		t.Errorf("Use = %q, want %q", skillValidateCmd.Use, "validate <path>")
	}

	if skillValidateCmd.Short == "" {
		t.Error("Short should not be empty")
	}

	if skillValidateCmd.Long == "" {
		t.Error("Long should not be empty")
	}

	// Verify flags are registered
	strictFlag := skillValidateCmd.Flags().Lookup("strict")
	if strictFlag == nil {
		t.Error("--strict flag not registered")
	}

	jsonFlag := skillValidateCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("--json flag not registered")
	}
}

// setupValidSkill creates a temp directory with a valid SKILL.md file.
func setupValidSkill(t *testing.T, name string) string {
	t.Helper()
	return setupSkillWithContent(t, name, `---
name: `+name+`
description: A test skill
---
Test instructions.
`)
}

// setupSkillWithContent creates a temp directory with the given SKILL.md content.
func setupSkillWithContent(t *testing.T, dirName, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	skillDir := filepath.Join(tempDir, dirName)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	return skillDir
}
