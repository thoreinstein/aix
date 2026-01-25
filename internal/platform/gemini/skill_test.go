package gemini

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTranslateVariables(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "translate $ARGUMENTS",
			content: "Run with $ARGUMENTS",
			want:    "Run with {{argument}}",
		},
		{
			name:    "translate $SELECTION",
			content: "Context: $SELECTION",
			want:    "Context: {{selection}}",
		},
		{
			name:    "translate both",
			content: "$ARGUMENTS and $SELECTION",
			want:    "{{argument}} and {{selection}}",
		},
		{
			name:    "no variables",
			content: "Pure text",
			want:    "Pure text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TranslateVariables(tt.content); got != tt.want {
				t.Errorf("TranslateVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslateToCanonical(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "translate {{args}}",
			content: "Run with {{args}}",
			want:    "Run with $ARGUMENTS",
		},
		{
			name:    "translate {{argument}}",
			content: "Run with {{argument}}",
			want:    "Run with $ARGUMENTS",
		},
		{
			name:    "translate {{selection}}",
			content: "Context: {{selection}}",
			want:    "Context: $SELECTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TranslateToCanonical(tt.content); got != tt.want {
				t.Errorf("TranslateToCanonical() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkillManager(t *testing.T) {
	tmpDir := t.TempDir()
	geminiDir := filepath.Join(tmpDir, ".gemini")
	if err := os.Mkdir(geminiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Mock GeminiPaths
	// Since GeminiPaths uses paths.GlobalConfigDir which uses Home(),
	// we'll just use a Project scope to control the base dir.
	paths := NewGeminiPaths(ScopeProject, tmpDir)

	mgr := NewSkillManager(paths)
	skill := &Skill{
		Name:         "test-skill",
		Description:  "A test skill",
		Instructions: "Use $ARGUMENTS",
	}

	// Test Install
	t.Run("Install", func(t *testing.T) {
		err := mgr.Install(skill)
		if err != nil {
			t.Fatalf("Install failed: %v", err)
		}

		// Verify file exists and content is translated
		skillPath := paths.SkillPath(skill.Name)
		data, err := os.ReadFile(skillPath)
		if err != nil {
			t.Fatalf("Failed to read skill file: %v", err)
		}

		if !contains(string(data), "{{argument}}") {
			t.Errorf("Skill content not translated: %s", string(data))
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		skills, err := mgr.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(skills) != 1 {
			t.Errorf("Expected 1 skill, got %d", len(skills))
		}

		if skills[0].Name != "test-skill" {
			t.Errorf("Expected test-skill, got %s", skills[0].Name)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		s, err := mgr.Get("test-skill")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if s.Name != "test-skill" {
			t.Errorf("Expected test-skill, got %s", s.Name)
		}

		// Verify instructions are back to canonical
		if s.Instructions != "Use $ARGUMENTS" {
			t.Errorf("Instructions not canonical: %s", s.Instructions)
		}
	})

	// Test Uninstall
	t.Run("Uninstall", func(t *testing.T) {
		err := mgr.Uninstall("test-skill")
		if err != nil {
			t.Fatalf("Uninstall failed: %v", err)
		}

		// Verify directory is gone
		skillPath := paths.SkillPath("test-skill")
		skillDir := filepath.Dir(skillPath)
		if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
			t.Errorf("Skill directory still exists after Uninstall")
		}
	})
}

func contains(s, substr string) bool {
	return filepath.Base(s) != s || (len(s) >= len(substr) && s[0:len(substr)] == substr) || (len(s) > 0 && contains(s[1:], substr))
}
