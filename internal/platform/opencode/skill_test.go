package opencode

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testPaths creates an OpenCodePaths instance rooted in a temp directory.
func testPaths(t *testing.T) *OpenCodePaths {
	t.Helper()
	return NewOpenCodePaths(ScopeProject, t.TempDir())
}

func TestSkillManager_List(t *testing.T) {
	t.Run("empty directory returns empty slice", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		skills, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		// nil slice is valid Go for empty result
		if len(skills) != 0 {
			t.Errorf("List() returned %d skills, want 0", len(skills))
		}
	})

	t.Run("missing directory returns empty slice", func(t *testing.T) {
		paths := NewOpenCodePaths(ScopeProject, "/nonexistent/path")
		mgr := NewSkillManager(paths)

		skills, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if len(skills) != 0 {
			t.Errorf("List() returned %d skills, want 0", len(skills))
		}
	})

	t.Run("returns skills from directory", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		// Create test skills
		skill1 := &Skill{
			Name:         "debug",
			Description:  "Debug skill",
			Instructions: "Debug instructions",
		}
		skill2 := &Skill{
			Name:         "refactor",
			Description:  "Refactor skill",
			Instructions: "Refactor instructions",
		}

		if err := mgr.Install(skill1); err != nil {
			t.Fatalf("Install(skill1) error = %v", err)
		}
		if err := mgr.Install(skill2); err != nil {
			t.Fatalf("Install(skill2) error = %v", err)
		}

		skills, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(skills) != 2 {
			t.Errorf("List() returned %d skills, want 2", len(skills))
		}

		// Verify skill names are present
		names := make(map[string]bool)
		for _, s := range skills {
			names[s.Name] = true
		}
		if !names["debug"] {
			t.Error("List() missing 'debug' skill")
		}
		if !names["refactor"] {
			t.Error("List() missing 'refactor' skill")
		}
	})

	t.Run("ignores non-directory entries", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		// Create skill directory
		skillDir := paths.SkillDir()
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatalf("failed to create skill directory: %v", err)
		}

		// Create a file (not a directory) in skill dir
		randomFile := filepath.Join(skillDir, "random.txt")
		if err := os.WriteFile(randomFile, []byte("content"), 0o644); err != nil {
			t.Fatalf("failed to create random file: %v", err)
		}

		// Create a valid skill
		skill := &Skill{Name: "valid", Description: "Valid skill", Instructions: "test"}
		if err := mgr.Install(skill); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		skills, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(skills) != 1 {
			t.Errorf("List() returned %d skills, want 1", len(skills))
		}
	})
}

func TestSkillManager_Get(t *testing.T) {
	t.Run("returns skill by name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		installed := &Skill{
			Name:         "test-skill",
			Description:  "Test description",
			Version:      "1.0.0",
			Author:       "Test Author",
			Tools:        []string{"bash", "read"},
			Instructions: "These are instructions",
		}
		if err := mgr.Install(installed); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		got, err := mgr.Get("test-skill")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Name != installed.Name {
			t.Errorf("Get().Name = %q, want %q", got.Name, installed.Name)
		}
		if got.Description != installed.Description {
			t.Errorf("Get().Description = %q, want %q", got.Description, installed.Description)
		}
		if got.Instructions != installed.Instructions {
			t.Errorf("Get().Instructions = %q, want %q", got.Instructions, installed.Instructions)
		}
	})

	t.Run("returns ErrSkillNotFound for nonexistent skill", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		_, err := mgr.Get("nonexistent")
		if !errors.Is(err, ErrSkillNotFound) {
			t.Errorf("Get() error = %v, want %v", err, ErrSkillNotFound)
		}
	})

	t.Run("returns ErrInvalidSkill for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		_, err := mgr.Get("")
		if !errors.Is(err, ErrInvalidSkill) {
			t.Errorf("Get() error = %v, want %v", err, ErrInvalidSkill)
		}
	})
}

func TestSkillManager_Install(t *testing.T) {
	t.Run("creates directory and file", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		skill := &Skill{
			Name:         "new-skill",
			Description:  "A new skill",
			Instructions: "Instructions here",
		}

		if err := mgr.Install(skill); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Verify file exists
		skillPath := paths.SkillPath("new-skill")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Errorf("Install() did not create file at %q", skillPath)
		}

		// Verify file contents
		data, err := os.ReadFile(skillPath)
		if err != nil {
			t.Fatalf("failed to read skill file: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, "name: new-skill") {
			t.Error("Install() file missing skill name in frontmatter")
		}
		if !strings.Contains(content, "Instructions here") {
			t.Error("Install() file missing instructions body")
		}
	})

	t.Run("overwrites existing skill", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		original := &Skill{
			Name:         "overwrite-test",
			Description:  "Original description",
			Instructions: "Original instructions",
		}
		if err := mgr.Install(original); err != nil {
			t.Fatalf("Install(original) error = %v", err)
		}

		updated := &Skill{
			Name:         "overwrite-test",
			Description:  "Updated description",
			Instructions: "Updated instructions",
		}
		if err := mgr.Install(updated); err != nil {
			t.Fatalf("Install(updated) error = %v", err)
		}

		got, err := mgr.Get("overwrite-test")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Description != "Updated description" {
			t.Errorf("Install() did not overwrite: Description = %q, want %q", got.Description, "Updated description")
		}
	})

	t.Run("returns ErrInvalidSkill for nil skill", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		err := mgr.Install(nil)
		if !errors.Is(err, ErrInvalidSkill) {
			t.Errorf("Install(nil) error = %v, want %v", err, ErrInvalidSkill)
		}
	})

	t.Run("returns ErrInvalidSkill for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		err := mgr.Install(&Skill{Name: ""})
		if !errors.Is(err, ErrInvalidSkill) {
			t.Errorf("Install(empty name) error = %v, want %v", err, ErrInvalidSkill)
		}
	})
}

func TestSkillManager_Uninstall(t *testing.T) {
	t.Run("removes skill directory", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		skill := &Skill{
			Name:         "to-remove",
			Description:  "Will be removed",
			Instructions: "test",
		}
		if err := mgr.Install(skill); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Verify skill exists
		skillPath := paths.SkillPath("to-remove")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Fatal("skill file should exist before uninstall")
		}

		if err := mgr.Uninstall("to-remove"); err != nil {
			t.Fatalf("Uninstall() error = %v", err)
		}

		// Verify skill directory is gone
		skillDir := filepath.Dir(skillPath)
		if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
			t.Errorf("Uninstall() did not remove skill directory %q", skillDir)
		}
	})

	t.Run("idempotent - no error if not exists", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		// Uninstall a skill that was never installed
		err := mgr.Uninstall("never-existed")
		if err != nil {
			t.Errorf("Uninstall(nonexistent) error = %v, want nil", err)
		}
	})

	t.Run("returns nil for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewSkillManager(paths)

		err := mgr.Uninstall("")
		if err != nil {
			t.Errorf("Uninstall('') error = %v, want nil", err)
		}
	})
}

func TestSkillManager_RoundTrip(t *testing.T) {
	paths := testPaths(t)
	mgr := NewSkillManager(paths)

	original := &Skill{
		Name:          "full-skill",
		Description:   "A complete skill",
		Version:       "2.0.0",
		Author:        "Test Author",
		Tools:         []string{"bash", "read", "write"},
		AllowedTools:  []string{"read"},
		Triggers:      []string{"/skill", "activate"},
		Compatibility: map[string]string{"claude": ">=1.0"},
		Metadata:      map[string]any{"custom": "value"},
		Instructions:  "Detailed instructions\nwith multiple lines",
	}

	if err := mgr.Install(original); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	got, err := mgr.Get("full-skill")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Verify all fields round-trip correctly
	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if got.Description != original.Description {
		t.Errorf("Description = %q, want %q", got.Description, original.Description)
	}
	if got.Version != original.Version {
		t.Errorf("Version = %q, want %q", got.Version, original.Version)
	}
	if got.Author != original.Author {
		t.Errorf("Author = %q, want %q", got.Author, original.Author)
	}
	if got.Instructions != original.Instructions {
		t.Errorf("Instructions = %q, want %q", got.Instructions, original.Instructions)
	}
	if len(got.Tools) != len(original.Tools) {
		t.Errorf("len(Tools) = %d, want %d", len(got.Tools), len(original.Tools))
	}
	if len(got.Triggers) != len(original.Triggers) {
		t.Errorf("len(Triggers) = %d, want %d", len(got.Triggers), len(original.Triggers))
	}
}
