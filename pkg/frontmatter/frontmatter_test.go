package frontmatter

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SkillMeta represents the frontmatter structure for skill files.
type SkillMeta struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tools       []string `yaml:"tools"`
}

// CommandMeta represents the frontmatter structure for command files.
type CommandMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Arguments   []struct {
		Name     string `yaml:"name"`
		Required bool   `yaml:"required"`
	} `yaml:"arguments"`
}

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMeta   *SkillMeta
		wantBody   string
		wantErr    error
		wantErrMsg string
	}{
		{
			name: "valid skill frontmatter",
			input: `---
name: skill-name
description: A brief description
tools:
  - tool1
  - tool2
---

# Skill instructions here
`,
			wantMeta: &SkillMeta{
				Name:        "skill-name",
				Description: "A brief description",
				Tools:       []string{"tool1", "tool2"},
			},
			wantBody: "\n# Skill instructions here\n",
			wantErr:  nil,
		},
		{
			name:       "no frontmatter",
			input:      "# Just a markdown file\n\nNo frontmatter here.",
			wantMeta:   nil,
			wantBody:   "",
			wantErr:    ErrNoFrontmatter,
			wantErrMsg: "no frontmatter found",
		},
		{
			name: "empty frontmatter",
			input: `---
---

Body content here.
`,
			wantMeta: &SkillMeta{},
			wantBody: "\nBody content here.\n",
			wantErr:  nil,
		},
		{
			name: "invalid YAML in frontmatter",
			input: `---
name: [invalid yaml
  this is broken
---

Body content.
`,
			wantMeta:   nil,
			wantBody:   "",
			wantErr:    ErrInvalidYAML,
			wantErrMsg: "invalid YAML",
		},
		{
			name: "empty body after frontmatter",
			input: `---
name: no-body-skill
description: Has no body content
---
`,
			wantMeta: &SkillMeta{
				Name:        "no-body-skill",
				Description: "Has no body content",
			},
			wantBody: "",
			wantErr:  nil,
		},
		{
			name: "frontmatter only no trailing newline",
			input: `---
name: minimal
---`,
			wantMeta: &SkillMeta{
				Name: "minimal",
			},
			wantBody: "",
			wantErr:  nil,
		},
		{
			name:  "Windows CRLF line endings",
			input: "---\r\nname: windows-skill\r\ndescription: Uses CRLF\r\n---\r\n\r\nBody with CRLF.\r\n",
			wantMeta: &SkillMeta{
				Name:        "windows-skill",
				Description: "Uses CRLF",
			},
			wantBody: "\nBody with CRLF.\n",
			wantErr:  nil,
		},
		{
			name: "partial frontmatter delimiter",
			input: `--
name: not-frontmatter
--

This doesn't have proper delimiters.
`,
			wantMeta:   nil,
			wantBody:   "",
			wantErr:    ErrNoFrontmatter,
			wantErrMsg: "no frontmatter found",
		},
		{
			name: "frontmatter with multiline description",
			input: `---
name: multiline-skill
description: |
  This is a multiline
  description with
  multiple lines
tools:
  - tool1
---

Instructions follow.
`,
			wantMeta: &SkillMeta{
				Name:        "multiline-skill",
				Description: "This is a multiline\ndescription with\nmultiple lines\n",
				Tools:       []string{"tool1"},
			},
			wantBody: "\nInstructions follow.\n",
			wantErr:  nil,
		},
		{
			name:       "empty input",
			input:      "",
			wantMeta:   nil,
			wantBody:   "",
			wantErr:    ErrNoFrontmatter,
			wantErrMsg: "no frontmatter found",
		},
		{
			name:       "only delimiter no closing",
			input:      "---\nname: unclosed\n",
			wantMeta:   nil,
			wantBody:   "",
			wantErr:    ErrNoFrontmatter,
			wantErrMsg: "no frontmatter found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			gotMeta, gotBody, err := Parse[SkillMeta](r)

			// Check error cases
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.wantErrMsg, err.Error())
				}
				return
			}

			// Check success cases
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotMeta == nil {
				t.Fatal("expected non-nil meta, got nil")
			}

			if gotMeta.Name != tt.wantMeta.Name {
				t.Errorf("name: got %q, want %q", gotMeta.Name, tt.wantMeta.Name)
			}
			if gotMeta.Description != tt.wantMeta.Description {
				t.Errorf("description: got %q, want %q", gotMeta.Description, tt.wantMeta.Description)
			}
			if len(gotMeta.Tools) != len(tt.wantMeta.Tools) {
				t.Errorf("tools length: got %d, want %d", len(gotMeta.Tools), len(tt.wantMeta.Tools))
			} else {
				for i, tool := range gotMeta.Tools {
					if tool != tt.wantMeta.Tools[i] {
						t.Errorf("tools[%d]: got %q, want %q", i, tool, tt.wantMeta.Tools[i])
					}
				}
			}

			if gotBody != tt.wantBody {
				t.Errorf("body: got %q, want %q", gotBody, tt.wantBody)
			}
		})
	}
}

func TestParse_CommandMeta(t *testing.T) {
	input := `---
name: /my-command
description: Does something useful
arguments:
  - name: arg1
    required: true
  - name: arg2
    required: false
---

Command template content $ARGUMENTS
`
	r := strings.NewReader(input)
	meta, body, err := Parse[CommandMeta](r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Name != "/my-command" {
		t.Errorf("name: got %q, want %q", meta.Name, "/my-command")
	}
	if meta.Description != "Does something useful" {
		t.Errorf("description: got %q, want %q", meta.Description, "Does something useful")
	}
	if len(meta.Arguments) != 2 {
		t.Fatalf("arguments length: got %d, want 2", len(meta.Arguments))
	}
	if meta.Arguments[0].Name != "arg1" || !meta.Arguments[0].Required {
		t.Errorf("arg1: got %+v, want {Name:arg1 Required:true}", meta.Arguments[0])
	}
	if meta.Arguments[1].Name != "arg2" || meta.Arguments[1].Required {
		t.Errorf("arg2: got %+v, want {Name:arg2 Required:false}", meta.Arguments[1])
	}

	wantBody := "\nCommand template content $ARGUMENTS\n"
	if body != wantBody {
		t.Errorf("body: got %q, want %q", body, wantBody)
	}
}

func TestParseFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "skill.md")

		content := `---
name: file-skill
description: Parsed from file
tools:
  - fileutil
---

File body content.
`
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		meta, body, err := ParseFile[SkillMeta](path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if meta.Name != "file-skill" {
			t.Errorf("name: got %q, want %q", meta.Name, "file-skill")
		}
		if meta.Description != "Parsed from file" {
			t.Errorf("description: got %q, want %q", meta.Description, "Parsed from file")
		}
		if len(meta.Tools) != 1 || meta.Tools[0] != "fileutil" {
			t.Errorf("tools: got %v, want [fileutil]", meta.Tools)
		}

		wantBody := "\nFile body content.\n"
		if body != wantBody {
			t.Errorf("body: got %q, want %q", body, wantBody)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, _, err := ParseFile[SkillMeta]("/nonexistent/path/to/file.md")
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("expected os.ErrNotExist, got %v", err)
		}
	})

	t.Run("file with Windows CRLF", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "crlf.md")

		content := "---\r\nname: crlf-skill\r\n---\r\n\r\nCRLF body.\r\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		meta, body, err := ParseFile[SkillMeta](path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if meta.Name != "crlf-skill" {
			t.Errorf("name: got %q, want %q", meta.Name, "crlf-skill")
		}

		wantBody := "\nCRLF body.\n"
		if body != wantBody {
			t.Errorf("body: got %q, want %q", body, wantBody)
		}
	})

	t.Run("file without frontmatter", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "nofm.md")

		content := "# No Frontmatter\n\nJust content."
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, _, err := ParseFile[SkillMeta](path)
		if err == nil {
			t.Fatal("expected error for file without frontmatter")
		}
		if !errors.Is(err, ErrNoFrontmatter) {
			t.Errorf("expected ErrNoFrontmatter, got %v", err)
		}
	})
}

func TestErrorsAreCorrectlySentinel(t *testing.T) {
	t.Run("ErrNoFrontmatter is identifiable", func(t *testing.T) {
		_, _, err := Parse[SkillMeta](strings.NewReader("no frontmatter"))
		if !errors.Is(err, ErrNoFrontmatter) {
			t.Errorf("expected errors.Is(err, ErrNoFrontmatter) to be true, got false for: %v", err)
		}
	})

	t.Run("ErrInvalidYAML is identifiable", func(t *testing.T) {
		input := "---\ninvalid: [broken\n---\nbody"
		_, _, err := Parse[SkillMeta](strings.NewReader(input))
		if !errors.Is(err, ErrInvalidYAML) {
			t.Errorf("expected errors.Is(err, ErrInvalidYAML) to be true, got false for: %v", err)
		}
	})
}
