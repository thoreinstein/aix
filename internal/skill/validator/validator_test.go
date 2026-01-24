package validator

import (
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/platform/claude"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		skill     *claude.Skill
		strict    bool
		wantErrs  int
		wantField string
		wantMsg   string
	}{
		{
			name: "valid skill with all fields",
			skill: &claude.Skill{
				Name:         "my-skill",
				Description:  "A test skill",
				AllowedTools: claude.ToolList{"Read", "Write", "Bash(git:*)"},
			},
			strict:   true,
			wantErrs: 0,
		},
		{
			name: "valid skill with minimal fields",
			skill: &claude.Skill{
				Name:        "test",
				Description: "A test skill",
			},
			strict:   false,
			wantErrs: 0,
		},
		{
			name: "valid skill single char name",
			skill: &claude.Skill{
				Name:        "a",
				Description: "Single character name",
			},
			wantErrs: 0,
		},
		{
			name: "valid skill max length name",
			skill: &claude.Skill{
				Name:        strings.Repeat("a", 64),
				Description: "Max length name",
			},
			wantErrs: 0,
		},
		// Name validation
		{
			name: "missing name",
			skill: &claude.Skill{
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "required",
		},
		{
			name: "empty name",
			skill: &claude.Skill{
				Name:        "",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "required",
		},
		{
			name: "name too long",
			skill: &claude.Skill{
				Name:        strings.Repeat("a", 65),
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "exceeds maximum length",
		},
		{
			name: "name with uppercase",
			skill: &claude.Skill{
				Name:        "MySkill",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase",
		},
		{
			name: "name starts with hyphen",
			skill: &claude.Skill{
				Name:        "-myskill",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "cannot start or end with a hyphen",
		},
		{
			name: "name ends with hyphen",
			skill: &claude.Skill{
				Name:        "myskill-",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "cannot start or end with a hyphen",
		},
		{
			name: "name with consecutive hyphens",
			skill: &claude.Skill{
				Name:        "my--skill",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "consecutive hyphens",
		},
		{
			name: "name with invalid chars underscore",
			skill: &claude.Skill{
				Name:        "my_skill",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase alphanumeric",
		},
		{
			name: "name with invalid chars space",
			skill: &claude.Skill{
				Name:        "my skill",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase alphanumeric",
		},
		{
			name: "name with invalid chars dot",
			skill: &claude.Skill{
				Name:        "my.skill",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "lowercase alphanumeric",
		},
		{
			name: "name starts with number",
			skill: &claude.Skill{
				Name:        "123skill",
				Description: "A test skill",
			},
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "start with a letter",
		},
		// Description validation
		{
			name: "missing description",
			skill: &claude.Skill{
				Name: "myskill",
			},
			wantErrs:  1,
			wantField: "description",
			wantMsg:   "required",
		},
		{
			name: "empty description",
			skill: &claude.Skill{
				Name:        "myskill",
				Description: "",
			},
			wantErrs:  1,
			wantField: "description",
			wantMsg:   "required",
		},
		{
			name: "whitespace only description",
			skill: &claude.Skill{
				Name:        "myskill",
				Description: "   \t\n  ",
			},
			wantErrs:  1,
			wantField: "description",
			wantMsg:   "whitespace",
		},
		// AllowedTools validation (strict mode)
		{
			name: "valid allowed tools strict",
			skill: &claude.Skill{
				Name:         "myskill",
				Description:  "A test skill",
				AllowedTools: claude.ToolList{"Read", "Write", "Bash(git:*)", "Glob"},
			},
			strict:   true,
			wantErrs: 0,
		},
		{
			name: "invalid allowed tools syntax strict",
			skill: &claude.Skill{
				Name:         "myskill",
				Description:  "A test skill",
				AllowedTools: claude.ToolList{"Read", "Write", "Bash(git:*", "Glob"},
			},
			strict:    true,
			wantErrs:  1,
			wantField: "allowed-tools",
			wantMsg:   "invalid",
		},
		{
			name: "invalid allowed tools ignored non-strict",
			skill: &claude.Skill{
				Name:         "myskill",
				Description:  "A test skill",
				AllowedTools: claude.ToolList{"Read", "Write", "Bash(git:*", "Glob"},
			},
			strict:   false,
			wantErrs: 0,
		},
		// Multiple errors
		{
			name: "multiple errors",
			skill: &claude.Skill{
				Name:        "",
				Description: "",
			},
			wantErrs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(WithStrict(tt.strict))
			result := v.Validate(tt.skill)

			if len(result.Issues) != tt.wantErrs {
				t.Errorf("Validate() got %d issues, want %d; issues: %v", len(result.Issues), tt.wantErrs, result.Issues)
				return
			}

			if tt.wantErrs > 0 && tt.wantField != "" {
				found := false
				for _, issue := range result.Issues {
					if issue.Field == tt.wantField && strings.Contains(issue.Message, tt.wantMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue for field %q with message containing %q, got: %v",
						tt.wantField, tt.wantMsg, result.Issues)
				}
			}
		})
	}
}

func TestValidator_ValidateWithPath(t *testing.T) {
	tests := []struct {
		name      string
		skill     *claude.Skill
		path      string
		wantErrs  int
		wantField string
		wantMsg   string
	}{
		{
			name: "name matches directory",
			skill: &claude.Skill{
				Name:        "my-skill",
				Description: "A test skill",
			},
			path:     "/path/to/my-skill/SKILL.md",
			wantErrs: 0,
		},
		{
			name: "name does not match directory",
			skill: &claude.Skill{
				Name:        "my-skill",
				Description: "A test skill",
			},
			path:      "/path/to/other-skill/SKILL.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "must match directory name",
		},
		{
			name: "missing name still reports required error",
			skill: &claude.Skill{
				Description: "A test skill",
			},
			path:      "/path/to/my-skill/SKILL.md",
			wantErrs:  1,
			wantField: "name",
			wantMsg:   "required",
		},
		{
			name: "name mismatch and other errors",
			skill: &claude.Skill{
				Name:        "my-skill",
				Description: "",
			},
			path:     "/path/to/other-skill/SKILL.md",
			wantErrs: 2, // description required + name mismatch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			result := v.ValidateWithPath(tt.skill, tt.path)

			if len(result.Issues) != tt.wantErrs {
				t.Errorf("ValidateWithPath() got %d issues, want %d; issues: %v", len(result.Issues), tt.wantErrs, result.Issues)
				return
			}

			if tt.wantErrs > 0 && tt.wantField != "" {
				found := false
				for _, issue := range result.Issues {
					if issue.Field == tt.wantField && strings.Contains(issue.Message, tt.wantMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected issue for field %q with message containing %q, got: %v",
						tt.wantField, tt.wantMsg, result.Issues)
				}
			}
		})
	}
}

func TestNew_Options(t *testing.T) {
	t.Run("default is non-strict", func(t *testing.T) {
		v := New()
		// Non-strict mode should not validate AllowedTools
		skill := &claude.Skill{
			Name:         "test",
			Description:  "Test",
			AllowedTools: claude.ToolList{"Invalid(("},
		}
		result := v.Validate(skill)
		if len(result.Issues) != 0 {
			t.Errorf("non-strict mode should not validate AllowedTools, got issues: %v", result.Issues)
		}
	})

	t.Run("strict mode validates AllowedTools", func(t *testing.T) {
		v := New(WithStrict(true))
		skill := &claude.Skill{
			Name:         "test",
			Description:  "Test",
			AllowedTools: claude.ToolList{"Invalid(("},
		}
		result := v.Validate(skill)
		if len(result.Issues) != 1 {
			t.Errorf("strict mode should validate AllowedTools, got %d issues", len(result.Issues))
		}
	})
}
