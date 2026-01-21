package parser

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/platform/claude"
)

const validSkillFull = `---
name: test-skill
description: A test skill for comprehensive parsing
license: MIT
compatibility:
  - claude-code
  - opencode
metadata:
  author: Test Author
  version: "1.0.0"
  repository: https://github.com/test/repo
allowed-tools: Read Write Bash(git:*)
---
# Test Skill Instructions

This is the body content.

With multiple paragraphs.
`

const validSkillMinimal = `---
name: minimal-skill
description: A minimal test skill
---
`

const validSkillNoBody = `---
name: header-only
description: Skill with no body content
license: Apache-2.0
---
`

const validSkillBodyOnly = `# Just Instructions

No frontmatter here at all.
This should fail with MustParse.
`

const malformedYAML = `---
name: bad-yaml
description: [unclosed bracket
---
Body content.
`

const emptyFile = ``

const frontmatterOnly = `---
name: unclosed-frontmatter
description: Missing closing delimiter
body starts here without delimiter
`

func TestParser_ParseBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		checkSkill  func(t *testing.T, s *claude.Skill)
	}{
		{
			name:    "valid skill with all fields",
			input:   validSkillFull,
			wantErr: false,
			checkSkill: func(t *testing.T, s *claude.Skill) {
				t.Helper()
				if s.Name != "test-skill" {
					t.Errorf("Name = %q, want %q", s.Name, "test-skill")
				}
				if s.Description != "A test skill for comprehensive parsing" {
					t.Errorf("Description = %q, want %q", s.Description, "A test skill for comprehensive parsing")
				}
				if s.License != "MIT" {
					t.Errorf("License = %q, want %q", s.License, "MIT")
				}
				if len(s.Compatibility) != 2 {
					t.Errorf("Compatibility len = %d, want 2", len(s.Compatibility))
				}
				expectedTools := claude.ToolList{"Read", "Write", "Bash(git:*)"}
				if !reflect.DeepEqual(s.AllowedTools, expectedTools) {
					t.Errorf("AllowedTools = %q, want %q", s.AllowedTools, expectedTools)
				}
				if s.Metadata["author"] != "Test Author" {
					t.Errorf("Metadata[author] = %q, want %q", s.Metadata["author"], "Test Author")
				}
				if s.Metadata["version"] != "1.0.0" {
					t.Errorf("Metadata[version] = %q, want %q", s.Metadata["version"], "1.0.0")
				}
				if !strings.Contains(s.Instructions, "Test Skill Instructions") {
					t.Errorf("Instructions should contain 'Test Skill Instructions', got %q", s.Instructions)
				}
				if !strings.Contains(s.Instructions, "multiple paragraphs") {
					t.Errorf("Instructions should contain 'multiple paragraphs', got %q", s.Instructions)
				}
			},
		},
		{
			name:    "valid skill with only required fields",
			input:   validSkillMinimal,
			wantErr: false,
			checkSkill: func(t *testing.T, s *claude.Skill) {
				t.Helper()
				if s.Name != "minimal-skill" {
					t.Errorf("Name = %q, want %q", s.Name, "minimal-skill")
				}
				if s.Description != "A minimal test skill" {
					t.Errorf("Description = %q, want %q", s.Description, "A minimal test skill")
				}
				if s.License != "" {
					t.Errorf("License = %q, want empty", s.License)
				}
				if len(s.Compatibility) != 0 {
					t.Errorf("Compatibility len = %d, want 0", len(s.Compatibility))
				}
				if s.Instructions != "" {
					t.Errorf("Instructions = %q, want empty", s.Instructions)
				}
			},
		},
		{
			name:    "frontmatter only, no body",
			input:   validSkillNoBody,
			wantErr: false,
			checkSkill: func(t *testing.T, s *claude.Skill) {
				t.Helper()
				if s.Name != "header-only" {
					t.Errorf("Name = %q, want %q", s.Name, "header-only")
				}
				if s.License != "Apache-2.0" {
					t.Errorf("License = %q, want %q", s.License, "Apache-2.0")
				}
				if s.Instructions != "" {
					t.Errorf("Instructions = %q, want empty", s.Instructions)
				}
			},
		},
		{
			name:        "body only, no frontmatter",
			input:       validSkillBodyOnly,
			wantErr:     true,
			errContains: "missing frontmatter",
		},
		{
			name:        "malformed YAML",
			input:       malformedYAML,
			wantErr:     true,
			errContains: "",
		},
		{
			name:        "empty file",
			input:       emptyFile,
			wantErr:     true,
			errContains: "missing frontmatter",
		},
		{
			name:        "unclosed frontmatter",
			input:       frontmatterOnly,
			wantErr:     true,
			errContains: "",
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill, err := p.ParseBytes([]byte(tt.input), "test.md")

			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseBytes() expected error, got nil")
				}
				var parseErr *ParseError
				if !errors.As(err, &parseErr) {
					t.Errorf("expected *ParseError, got %T", err)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseBytes() error = %v", err)
			}
			if skill == nil {
				t.Fatal("ParseBytes() returned nil skill")
			}
			if tt.checkSkill != nil {
				tt.checkSkill(t, skill)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	p := New()

	t.Run("reads from reader successfully", func(t *testing.T) {
		r := bytes.NewReader([]byte(validSkillFull))
		skill, err := p.Parse(r, "reader-test.md")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if skill.Name != "test-skill" {
			t.Errorf("Name = %q, want %q", skill.Name, "test-skill")
		}
	})

	t.Run("includes path in error", func(t *testing.T) {
		r := bytes.NewReader([]byte(emptyFile))
		_, err := p.Parse(r, "my-path.md")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "my-path.md") {
			t.Errorf("error should contain path, got %q", err.Error())
		}
	})
}

func TestParser_ParseFile(t *testing.T) {
	p := New()

	t.Run("parses file from filesystem", func(t *testing.T) {
		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(validSkillFull), 0o644); err != nil {
			t.Fatalf("writing test file: %v", err)
		}

		skill, err := p.ParseFile(skillPath)
		if err != nil {
			t.Fatalf("ParseFile() error = %v", err)
		}
		if skill.Name != "test-skill" {
			t.Errorf("Name = %q, want %q", skill.Name, "test-skill")
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := p.ParseFile("/nonexistent/path/SKILL.md")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("expected *ParseError, got %T", err)
		}
		if parseErr.Path != "/nonexistent/path/SKILL.md" {
			t.Errorf("ParseError.Path = %q, want %q", parseErr.Path, "/nonexistent/path/SKILL.md")
		}
	})
}

func TestParser_ParseHeader(t *testing.T) {
	p := New()

	t.Run("parses only header", func(t *testing.T) {
		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(validSkillFull), 0o644); err != nil {
			t.Fatalf("writing test file: %v", err)
		}

		skill, err := p.ParseHeader(skillPath)
		if err != nil {
			t.Fatalf("ParseHeader() error = %v", err)
		}
		if skill.Name != "test-skill" {
			t.Errorf("Name = %q, want %q", skill.Name, "test-skill")
		}
		if skill.Description != "A test skill for comprehensive parsing" {
			t.Errorf("Description = %q, want %q", skill.Description, "A test skill for comprehensive parsing")
		}
		// Instructions should NOT be populated by ParseHeader
		if skill.Instructions != "" {
			t.Errorf("Instructions = %q, want empty (ParseHeader should not parse body)", skill.Instructions)
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := p.ParseHeader("/nonexistent/path/SKILL.md")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("expected *ParseError, got %T", err)
		}
	})

	t.Run("returns empty skill for file without frontmatter", func(t *testing.T) {
		tmpDir := t.TempDir()
		skillPath := filepath.Join(tmpDir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(validSkillBodyOnly), 0o644); err != nil {
			t.Fatalf("writing test file: %v", err)
		}

		skill, err := p.ParseHeader(skillPath)
		if err != nil {
			t.Fatalf("ParseHeader() error = %v", err)
		}
		// ParseHeader returns empty skill (no error) when no frontmatter
		if skill.Name != "" {
			t.Errorf("Name = %q, want empty", skill.Name)
		}
	})
}

func TestParseError(t *testing.T) {
	t.Run("formats with path", func(t *testing.T) {
		err := &ParseError{
			Path: "/some/path.md",
			Err:  errors.New("underlying error"),
		}
		expected := "parsing skill /some/path.md: underlying error"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("formats without path", func(t *testing.T) {
		err := &ParseError{
			Err: errors.New("underlying error"),
		}
		expected := "parsing skill: underlying error"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := &ParseError{
			Path: "/path.md",
			Err:  underlying,
		}
		if !errors.Is(err, underlying) {
			t.Error("Unwrap() should allow errors.Is to match underlying error")
		}
	})
}

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Error("New() returned nil")
	}
}
