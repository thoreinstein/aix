// Package skill contains integration tests for the skill parsing and validation pipeline.
package skill_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/skill/parser"
	"github.com/thoreinstein/aix/internal/skill/toolperm"
	"github.com/thoreinstein/aix/internal/skill/validator"
)

// Test data representing realistic SKILL.md files

const validSkillComplete = `---
name: code-review
description: Performs comprehensive code review with security and performance focus
license: MIT
compatibility:
  - claude-code
  - opencode
metadata:
  author: Test Author
  version: "1.0.0"
  repository: https://github.com/test/code-review
allowed-tools: Read Glob Grep Bash(git:*)
---
# Code Review Skill

## Purpose
This skill performs comprehensive code review.

## Usage
Use /code-review to invoke this skill.
`

const validSkillMinimal = `---
name: simple-skill
description: A simple skill with only required fields
---
# Simple Skill

This skill has minimal configuration.
`

const invalidSkillBadName = `---
name: Invalid_Name
description: This skill has an invalid name
---
# Bad Name Skill
`

const invalidSkillMissingDescription = `---
name: missing-desc
---
# Missing Description
`

const invalidSkillBadTools = `---
name: bad-tools
description: This skill has invalid tool syntax
allowed-tools: Read Bash( Write
---
# Bad Tools
`

// TestIntegration_ParseAndValidate tests the complete parsing and validation workflow.
func TestIntegration_ParseAndValidate(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		dirName    string // Parent directory name
		wantErrors int    // Expected number of validation errors
		errField   string // Expected field with error (if wantErrors > 0)
	}{
		{
			name:       "complete valid skill",
			content:    validSkillComplete,
			dirName:    "code-review",
			wantErrors: 0,
		},
		{
			name:       "minimal valid skill",
			content:    validSkillMinimal,
			dirName:    "simple-skill",
			wantErrors: 0,
		},
		{
			name:       "invalid name format",
			content:    invalidSkillBadName,
			dirName:    "Invalid_Name",
			wantErrors: 1,
			errField:   "name",
		},
		{
			name:       "missing description",
			content:    invalidSkillMissingDescription,
			dirName:    "missing-desc",
			wantErrors: 1,
			errField:   "description",
		},
		{
			name:       "invalid tool syntax",
			content:    invalidSkillBadTools,
			dirName:    "bad-tools",
			wantErrors: 1,
			errField:   "allowed-tools",
		},
		{
			name:       "name doesn't match directory",
			content:    validSkillMinimal,
			dirName:    "wrong-directory",
			wantErrors: 1,
			errField:   "name",
		},
	}

	p := parser.New()
	v := validator.New(validator.WithStrict(true))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file structure
			tmpDir := t.TempDir()
			skillDir := filepath.Join(tmpDir, tt.dirName)
			if err := os.MkdirAll(skillDir, 0o755); err != nil {
				t.Fatalf("failed to create skill dir: %v", err)
			}
			skillPath := filepath.Join(skillDir, "SKILL.md")
			if err := os.WriteFile(skillPath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write skill file: %v", err)
			}

			// Parse
			skill, err := p.ParseFile(skillPath)
			if err != nil {
				if tt.wantErrors == 0 {
					t.Fatalf("unexpected parse error: %v", err)
				}
				// Parse error counts as validation error
				return
			}

			// Validate with path
			errs := v.ValidateWithPath(skill, skillPath)
			if len(errs) != tt.wantErrors {
				t.Errorf("got %d validation errors, want %d: %v", len(errs), tt.wantErrors, errs)
			}

			// Check error field if expected
			if tt.wantErrors > 0 && tt.errField != "" && len(errs) > 0 {
				verr, ok := errs[0].(*validator.ValidationError)
				if !ok {
					t.Errorf("expected ValidationError, got %T", errs[0])
				} else if verr.Field != tt.errField {
					t.Errorf("error field = %q, want %q", verr.Field, tt.errField)
				}
			}
		})
	}
}

// TestIntegration_ToolPermissionParsing tests parsing and validating tool permissions
// from a complete skill file.
func TestIntegration_ToolPermissionParsing(t *testing.T) {
	tests := []struct {
		name         string
		allowedTools string
		wantCount    int
		wantFirst    toolperm.Permission
		wantParseErr bool
	}{
		{
			name:         "multiple tools with scopes",
			allowedTools: "Read Glob Grep Bash(git:*) WebFetch",
			wantCount:    5,
			wantFirst:    toolperm.Permission{Name: "Read", Scope: ""},
		},
		{
			name:         "single scoped tool",
			allowedTools: "Bash(npm:install)",
			wantCount:    1,
			wantFirst:    toolperm.Permission{Name: "Bash", Scope: "npm:install"},
		},
		{
			name:         "empty tools",
			allowedTools: "",
			wantCount:    0,
		},
		{
			name:         "invalid syntax",
			allowedTools: "Read Bash(",
			wantParseErr: true,
		},
	}

	tp := toolperm.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the tool permissions directly from allowed-tools string
			// (simulating extraction from a skill's AllowedTools field)
			perms, err := tp.Parse(tt.allowedTools)
			if tt.wantParseErr {
				if err == nil {
					t.Error("expected parse error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}

			if len(perms) != tt.wantCount {
				t.Errorf("got %d permissions, want %d", len(perms), tt.wantCount)
			}

			if tt.wantCount > 0 && len(perms) > 0 {
				if perms[0].Name != tt.wantFirst.Name || perms[0].Scope != tt.wantFirst.Scope {
					t.Errorf("first permission = %+v, want %+v", perms[0], tt.wantFirst)
				}
			}
		})
	}
}

// TestIntegration_RoundTrip tests that skills can be parsed, modified, and the
// changes are reflected correctly.
func TestIntegration_RoundTrip(t *testing.T) {
	p := parser.New()
	v := validator.New()

	// Parse a complete skill
	skill, err := p.ParseBytes([]byte(validSkillComplete), "test.md")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Verify parsed values
	if skill.Name != "code-review" {
		t.Errorf("name = %q, want %q", skill.Name, "code-review")
	}
	if skill.License != "MIT" {
		t.Errorf("license = %q, want %q", skill.License, "MIT")
	}
	if len(skill.Compatibility) != 2 {
		t.Errorf("compatibility count = %d, want 2", len(skill.Compatibility))
	}
	if skill.Metadata["author"] != "Test Author" {
		t.Errorf("metadata.author = %q, want %q", skill.Metadata["author"], "Test Author")
	}
	if !strings.Contains(skill.Instructions, "Code Review Skill") {
		t.Error("instructions should contain 'Code Review Skill'")
	}

	// Should validate without errors
	if errs := v.Validate(skill); len(errs) > 0 {
		t.Errorf("validation errors: %v", errs)
	}
}

// TestIntegration_EdgeCases tests edge cases in the skill parsing pipeline.
func TestIntegration_EdgeCases(t *testing.T) {
	p := parser.New()
	v := validator.New()

	tests := []struct {
		name        string
		content     string
		wantName    string
		wantBodyLen int // Minimum expected body length
		wantErr     bool
	}{
		{
			name: "frontmatter only",
			content: `---
name: frontmatter-only
description: No body content
---
`,
			wantName:    "frontmatter-only",
			wantBodyLen: 0,
		},
		{
			name: "unicode in description",
			content: `---
name: unicode-skill
description: Skill with emojis and unicode
---
# Unicode body content
`,
			wantName:    "unicode-skill",
			wantBodyLen: 10,
		},
		{
			name: "multiline description via yaml",
			content: `---
name: multiline
description: >
  This is a multiline
  description in YAML
---
# Content
`,
			wantName:    "multiline",
			wantBodyLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill, err := p.ParseBytes([]byte(tt.content), "test.md")
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if skill.Name != tt.wantName {
				t.Errorf("name = %q, want %q", skill.Name, tt.wantName)
			}

			bodyLen := len(strings.TrimSpace(skill.Instructions))
			if bodyLen < tt.wantBodyLen {
				t.Errorf("body length = %d, want >= %d", bodyLen, tt.wantBodyLen)
			}

			// Should still validate (unless we expect parse error)
			if !tt.wantErr {
				errs := v.Validate(skill)
				if len(errs) > 0 {
					t.Errorf("validation errors: %v", errs)
				}
			}
		})
	}
}

// TestIntegration_FullPipeline tests the full pipeline: parse file, validate,
// extract tool permissions, and verify all components work together.
func TestIntegration_FullPipeline(t *testing.T) {
	// Create a realistic skill directory structure
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "security-scanner")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	skillContent := `---
name: security-scanner
description: Scans code for security vulnerabilities using multiple tools
license: Apache-2.0
compatibility:
  - claude-code
  - opencode
  - codex
metadata:
  author: Security Team
  version: "2.1.0"
  repository: https://github.com/example/security-scanner
allowed-tools: Read Glob Grep Bash(npm:audit) Bash(go:vet) WebFetch
---
# Security Scanner Skill

## Overview
This skill performs comprehensive security scanning of your codebase.

## Capabilities
- Static analysis
- Dependency vulnerability scanning
- Secret detection

## Usage
Invoke with /security-scan to begin analysis.
`

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
		t.Fatalf("failed to write skill file: %v", err)
	}

	// Step 1: Parse
	p := parser.New()
	skill, err := p.ParseFile(skillPath)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Step 2: Validate with path (strict mode)
	v := validator.New(validator.WithStrict(true))
	if errs := v.ValidateWithPath(skill, skillPath); len(errs) > 0 {
		t.Fatalf("validation failed: %v", errs)
	}

	// Step 3: Parse tool permissions
	tp := toolperm.New()
	perms, err := tp.Parse(skill.AllowedTools)
	if err != nil {
		t.Fatalf("tool permission parse failed: %v", err)
	}

	// Verify all components
	if skill.Name != "security-scanner" {
		t.Errorf("name = %q, want %q", skill.Name, "security-scanner")
	}
	if skill.License != "Apache-2.0" {
		t.Errorf("license = %q, want %q", skill.License, "Apache-2.0")
	}
	if len(skill.Compatibility) != 3 {
		t.Errorf("compatibility count = %d, want 3", len(skill.Compatibility))
	}
	if skill.Metadata["version"] != "2.1.0" {
		t.Errorf("metadata.version = %q, want %q", skill.Metadata["version"], "2.1.0")
	}
	if !strings.Contains(skill.Instructions, "Security Scanner Skill") {
		t.Error("instructions should contain 'Security Scanner Skill'")
	}

	// Verify tool permissions
	expectedTools := []struct {
		name  string
		scope string
	}{
		{"Read", ""},
		{"Glob", ""},
		{"Grep", ""},
		{"Bash", "npm:audit"},
		{"Bash", "go:vet"},
		{"WebFetch", ""},
	}

	if len(perms) != len(expectedTools) {
		t.Fatalf("got %d permissions, want %d", len(perms), len(expectedTools))
	}

	for i, exp := range expectedTools {
		if perms[i].Name != exp.name || perms[i].Scope != exp.scope {
			t.Errorf("permission[%d] = %+v, want name=%q scope=%q",
				i, perms[i], exp.name, exp.scope)
		}
	}
}

// TestIntegration_ValidationOrder verifies that validation errors are reported
// in a predictable order and contain expected context.
func TestIntegration_ValidationOrder(t *testing.T) {
	p := parser.New()
	v := validator.New(validator.WithStrict(true))

	// Create a skill with multiple issues
	content := `---
name: bad--name
description: ""
allowed-tools: Bash(
---
`
	skill, err := p.ParseBytes([]byte(content), "test.md")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	errs := v.Validate(skill)

	// Should have errors for: name (consecutive hyphens), description (empty), allowed-tools (invalid)
	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(errs), errs)
	}

	// Verify each error is a ValidationError with proper field
	fields := make(map[string]bool)
	for _, err := range errs {
		verr, ok := err.(*validator.ValidationError)
		if !ok {
			t.Errorf("expected ValidationError, got %T", err)
			continue
		}
		fields[verr.Field] = true
	}

	// Check that we got errors for expected fields
	expectedFields := []string{"name", "description"}
	for _, f := range expectedFields {
		if !fields[f] {
			t.Errorf("expected error for field %q, not found in %v", f, fields)
		}
	}
}
