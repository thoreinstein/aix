package skill

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCommand_ValidSkill(t *testing.T) {
	// Create a temp directory with a valid skill
	skillDir := setupValidSkill(t, "test-skill")

	// Reset flags for test
	validateStrict = false
	validateJSON = false

	// Run validate command directly
	err := validateCmd.RunE(validateCmd, []string{skillDir})
	if err != nil {
		t.Fatalf("expected no error for valid skill, got: %v", err)
	}
}

func TestValidateCommand_InvalidSkill_MissingName(t *testing.T) {
	skillDir := setupSkillWithContent(t, "invalid-skill", `---
description: Missing name field
---
Instructions here.
`)

	// Reset flags for test
	validateStrict = false
	validateJSON = false

	err := validateCmd.RunE(validateCmd, []string{skillDir})
	if err == nil {
		t.Fatal("expected error for invalid skill, got nil")
	}

	if !errors.Is(err, errValidationFailed) {
		t.Errorf("expected errValidationFailed, got: %v", err)
	}
}

func TestValidateCommand_JSONOutput_Valid(t *testing.T) {
	skillDir := setupValidSkill(t, "json-test")

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Reset flags for test
	validateStrict = false
	validateJSON = true

	err := validateCmd.RunE(validateCmd, []string{skillDir})
	if err != nil {
		w.Close()
		os.Stdout = old
		t.Fatalf("expected no error for valid skill, got: %v", err)
	}

	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	os.Stdout = old

	output := buf.String()

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
}

func TestValidateCommand_NotFound(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "nonexistent")

	// Reset flags for test
	validateStrict = false
	validateJSON = false

	err := validateCmd.RunE(validateCmd, []string{nonexistent})
	if err == nil {
		t.Fatal("expected error for nonexistent skill, got nil")
	}
}

func TestValidateCommand_CommandMetadata(t *testing.T) {
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
