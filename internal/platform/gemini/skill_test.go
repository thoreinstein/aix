package gemini

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillManager_Install_WithSourceDir(t *testing.T) {
	srcDir := t.TempDir()
	files := map[string]string{
		"SKILL.md":     "---\nname: src-skill\ndescription: raw\n---\n\nRaw $ARGUMENTS",
		"helper.sh":    "#!/bin/bash\necho hi",
		"sub/data.txt": "some data",
	}
	for relPath, content := range files {
		fullPath := filepath.Join(srcDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	paths := NewGeminiPaths(ScopeProject, t.TempDir())
	mgr := NewSkillManager(paths)

	skill := &Skill{
		Name:         "src-skill",
		Description:  "formatted description",
		Instructions: "Use $ARGUMENTS here",
		SourceDir:    srcDir,
	}

	if err := mgr.Install(skill); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	skillDir := filepath.Dir(paths.SkillPath("src-skill"))

	// Extra files were copied
	for _, relPath := range []string{"helper.sh", "sub/data.txt"} {
		if _, err := os.Stat(filepath.Join(skillDir, relPath)); err != nil {
			t.Errorf("expected file %q to exist after install: %v", relPath, err)
		}
	}

	// SKILL.md should have Gemini-translated variables, not the raw source
	data, err := os.ReadFile(paths.SkillPath("src-skill"))
	if err != nil {
		t.Fatalf("failed to read SKILL.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "{{argument}}") {
		t.Errorf("SKILL.md missing Gemini variable translation, got: %s", content)
	}
	if strings.Contains(content, "Raw $ARGUMENTS") {
		t.Errorf("SKILL.md contains raw source content, formatted write did not overwrite")
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

		if !strings.Contains(string(data), "{{argument}}") {
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
