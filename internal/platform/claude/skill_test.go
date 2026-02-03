package claude

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSkillManager_List_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	skills, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(skills) != 0 {
		t.Errorf("List() = %d skills, want 0", len(skills))
	}
}

func TestSkillManager_List_NonExistentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist")
	paths := NewClaudePaths(ScopeProject, nonExistent)
	mgr := NewSkillManager(paths)

	skills, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(skills) != 0 {
		t.Errorf("List() = %d skills, want 0", len(skills))
	}
}

func TestSkillManager_List_MultipleSkills(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	// Install multiple skills
	skills := []*Skill{
		{
			Name:         "skill-a",
			Description:  "First skill",
			Metadata:     map[string]string{"version": "1.0.0"},
			Instructions: "Instructions for A",
		},
		{
			Name:         "skill-b",
			Description:  "Second skill",
			AllowedTools: ToolList{"Read", "Write"},
			Instructions: "Instructions for B",
		},
		{
			Name:         "skill-c",
			Description:  "Third skill",
			Metadata:     map[string]string{"author": "test"},
			Instructions: "Instructions for C",
		},
	}

	for _, s := range skills {
		if err := mgr.Install(s); err != nil {
			t.Fatalf("Install(%s) error = %v", s.Name, err)
		}
	}

	// List skills
	listed, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(listed) != len(skills) {
		t.Errorf("List() = %d skills, want %d", len(listed), len(skills))
	}

	// Verify all skills are present
	found := make(map[string]bool)
	for _, s := range listed {
		found[s.Name] = true
	}

	for _, s := range skills {
		if !found[s.Name] {
			t.Errorf("skill %q not found in list", s.Name)
		}
	}
}

func TestSkillManager_List_IgnoresNonSkillDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	// Create skill directory
	skillDir := paths.SkillDir()
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill directory: %v", err)
	}

	// Create a directory without SKILL.md
	emptyDir := filepath.Join(skillDir, "not-a-skill")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatalf("failed to create empty directory: %v", err)
	}

	// Create a regular file (not a directory)
	regularFile := filepath.Join(skillDir, "regular-file.txt")
	if err := os.WriteFile(regularFile, []byte("not a skill"), 0o644); err != nil {
		t.Fatalf("failed to create regular file: %v", err)
	}

	// Install one valid skill
	validSkill := &Skill{
		Name:        "valid-skill",
		Description: "A valid skill",
	}
	if err := mgr.Install(validSkill); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// List should only return the valid skill
	skills, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(skills) != 1 {
		t.Errorf("List() = %d skills, want 1", len(skills))
	}

	if len(skills) > 0 && skills[0].Name != "valid-skill" {
		t.Errorf("List()[0].Name = %q, want %q", skills[0].Name, "valid-skill")
	}
}

func TestSkillManager_Get_ExistingSkill(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	// Install a skill
	original := &Skill{
		Name:          "test-skill",
		Description:   "A test skill",
		License:       "MIT",
		Compatibility: []string{"claude-code", "opencode"},
		Metadata:      map[string]string{"version": "2.0.0", "author": "test-author"},
		AllowedTools:  ToolList{"Read", "Write", "Bash"},
		Instructions:  "These are the instructions.\n\nWith multiple paragraphs.",
	}

	if err := mgr.Install(original); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Get the skill
	got, err := mgr.Get("test-skill")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Verify all fields
	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if got.Description != original.Description {
		t.Errorf("Description = %q, want %q", got.Description, original.Description)
	}
	if got.License != original.License {
		t.Errorf("License = %q, want %q", got.License, original.License)
	}
	if !reflect.DeepEqual(got.Compatibility, original.Compatibility) {
		t.Errorf("Compatibility = %v, want %v", got.Compatibility, original.Compatibility)
	}
	if !reflect.DeepEqual(got.Metadata, original.Metadata) {
		t.Errorf("Metadata = %v, want %v", got.Metadata, original.Metadata)
	}
	if !reflect.DeepEqual(got.AllowedTools, original.AllowedTools) {
		t.Errorf("AllowedTools = %q, want %q", got.AllowedTools, original.AllowedTools)
	}
	if got.Instructions != original.Instructions {
		t.Errorf("Instructions = %q, want %q", got.Instructions, original.Instructions)
	}
}

func TestSkillManager_Get_NonExistentSkill(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	_, err := mgr.Get("does-not-exist")
	if !errors.Is(err, ErrSkillNotFound) {
		t.Errorf("Get() error = %v, want %v", err, ErrSkillNotFound)
	}
}

func TestSkillManager_Get_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	_, err := mgr.Get("")
	if !errors.Is(err, ErrInvalidSkill) {
		t.Errorf("Get() error = %v, want %v", err, ErrInvalidSkill)
	}
}

func TestSkillManager_Install_NewSkill(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	skill := &Skill{
		Name:         "new-skill",
		Description:  "A brand new skill",
		Metadata:     map[string]string{"version": "1.0.0"},
		Instructions: "New skill instructions",
	}

	if err := mgr.Install(skill); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Verify file exists
	skillPath := paths.SkillPath("new-skill")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("skill file not created: %v", err)
	}

	// Verify content can be retrieved
	got, err := mgr.Get("new-skill")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name != skill.Name {
		t.Errorf("Name = %q, want %q", got.Name, skill.Name)
	}
	if got.Description != skill.Description {
		t.Errorf("Description = %q, want %q", got.Description, skill.Description)
	}
}

func TestSkillManager_Install_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	// Install original
	original := &Skill{
		Name:         "overwrite-skill",
		Description:  "Original description",
		Metadata:     map[string]string{"version": "1.0.0"},
		Instructions: "Original instructions",
	}

	if err := mgr.Install(original); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Overwrite with new version
	updated := &Skill{
		Name:         "overwrite-skill",
		Description:  "Updated description",
		Metadata:     map[string]string{"version": "2.0.0", "author": "new-author"},
		AllowedTools: ToolList{"Read"},
		Instructions: "Updated instructions",
	}

	if err := mgr.Install(updated); err != nil {
		t.Fatalf("Install() overwrite error = %v", err)
	}

	// Verify updated content
	got, err := mgr.Get("overwrite-skill")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Description != updated.Description {
		t.Errorf("Description = %q, want %q", got.Description, updated.Description)
	}
	if !reflect.DeepEqual(got.Metadata, updated.Metadata) {
		t.Errorf("Metadata = %v, want %v", got.Metadata, updated.Metadata)
	}
	if !reflect.DeepEqual(got.AllowedTools, updated.AllowedTools) {
		t.Errorf("AllowedTools = %q, want %q", got.AllowedTools, updated.AllowedTools)
	}
	if got.Instructions != updated.Instructions {
		t.Errorf("Instructions = %q, want %q", got.Instructions, updated.Instructions)
	}
}

func TestSkillManager_Install_NilSkill(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	err := mgr.Install(nil)
	if !errors.Is(err, ErrInvalidSkill) {
		t.Errorf("Install(nil) error = %v, want %v", err, ErrInvalidSkill)
	}
}

func TestSkillManager_Install_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	err := mgr.Install(&Skill{Description: "No name"})
	if !errors.Is(err, ErrInvalidSkill) {
		t.Errorf("Install() error = %v, want %v", err, ErrInvalidSkill)
	}
}

func TestSkillManager_Uninstall_ExistingSkill(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	// Install a skill
	skill := &Skill{
		Name:         "to-remove",
		Description:  "Will be removed",
		Instructions: "Goodbye",
	}

	if err := mgr.Install(skill); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Verify it exists
	if _, err := mgr.Get("to-remove"); err != nil {
		t.Fatalf("Get() before uninstall error = %v", err)
	}

	// Uninstall
	if err := mgr.Uninstall("to-remove"); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	// Verify it's gone
	_, err := mgr.Get("to-remove")
	if !errors.Is(err, ErrSkillNotFound) {
		t.Errorf("Get() after uninstall error = %v, want %v", err, ErrSkillNotFound)
	}

	// Verify directory is gone
	skillDir := filepath.Dir(paths.SkillPath("to-remove"))
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Errorf("skill directory still exists after uninstall")
	}
}

func TestSkillManager_Uninstall_NonExistentSkill(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	// Uninstalling non-existent skill should be idempotent (no error)
	err := mgr.Uninstall("does-not-exist")
	if err != nil {
		t.Errorf("Uninstall() error = %v, want nil", err)
	}
}

func TestSkillManager_Uninstall_EmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	// Uninstalling empty name should be idempotent (no error)
	err := mgr.Uninstall("")
	if err != nil {
		t.Errorf("Uninstall() error = %v, want nil", err)
	}
}

func TestParseSkillFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *Skill
		wantErr bool
	}{
		{
			name: "minimal skill",
			content: `---
name: test
description: A test skill
---
`,
			want: &Skill{
				Name:         "test",
				Description:  "A test skill",
				Instructions: "",
			},
		},
		{
			name: "skill with instructions",
			content: `---
name: test
description: A test skill
---

These are the instructions.
`,
			want: &Skill{
				Name:         "test",
				Description:  "A test skill",
				Instructions: "These are the instructions.",
			},
		},
		{
			name: "skill with all fields",
			content: `---
name: full-skill
description: Full skill description
license: MIT
compatibility:
  - claude-code
  - opencode
metadata:
  version: "1.2.3"
  author: test-author
allowed-tools: Read Write
---

Multi-line
instructions
here.
`,
			want: &Skill{
				Name:          "full-skill",
				Description:   "Full skill description",
				License:       "MIT",
				Compatibility: []string{"claude-code", "opencode"},
				Metadata:      map[string]string{"version": "1.2.3", "author": "test-author"},
				AllowedTools:  ToolList{"Read", "Write"},
				Instructions:  "Multi-line\ninstructions\nhere.",
			},
		},
		{
			name:    "missing opening delimiter",
			content: "name: test\ndescription: test\n---\n",
			wantErr: true,
		},
		{
			name:    "missing closing delimiter",
			content: "---\nname: test\ndescription: test\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSkillFile([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSkillFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.License != tt.want.License {
				t.Errorf("License = %q, want %q", got.License, tt.want.License)
			}
			if !reflect.DeepEqual(got.Compatibility, tt.want.Compatibility) {
				t.Errorf("Compatibility = %v, want %v", got.Compatibility, tt.want.Compatibility)
			}
			if !reflect.DeepEqual(got.Metadata, tt.want.Metadata) {
				t.Errorf("Metadata = %v, want %v", got.Metadata, tt.want.Metadata)
			}
			if !reflect.DeepEqual(got.AllowedTools, tt.want.AllowedTools) {
				t.Errorf("AllowedTools = %q, want %q", got.AllowedTools, tt.want.AllowedTools)
			}
			if got.Instructions != tt.want.Instructions {
				t.Errorf("Instructions = %q, want %q", got.Instructions, tt.want.Instructions)
			}
		})
	}
}

func TestFormatSkillFile(t *testing.T) {
	tests := []struct {
		name  string
		skill *Skill
	}{
		{
			name: "minimal skill",
			skill: &Skill{
				Name:        "minimal",
				Description: "A minimal skill",
			},
		},
		{
			name: "skill with instructions",
			skill: &Skill{
				Name:         "with-instructions",
				Description:  "Has instructions",
				Instructions: "These are the instructions.",
			},
		},
		{
			name: "full skill",
			skill: &Skill{
				Name:          "full",
				Description:   "Full skill",
				License:       "Apache-2.0",
				Compatibility: []string{"claude-code"},
				Metadata:      map[string]string{"version": "1.0.0", "author": "author"},
				AllowedTools:  ToolList{"Read", "Write"},
				Instructions:  "Full instructions\nwith newlines.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Format the skill
			data, err := formatSkillFile(tt.skill)
			if err != nil {
				t.Fatalf("formatSkillFile() error = %v", err)
			}

			// Parse it back
			got, err := parseSkillFile(data)
			if err != nil {
				t.Fatalf("parseSkillFile() error = %v", err)
			}

			// Verify round-trip
			if got.Name != tt.skill.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.skill.Name)
			}
			if got.Description != tt.skill.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.skill.Description)
			}
			if got.License != tt.skill.License {
				t.Errorf("License = %q, want %q", got.License, tt.skill.License)
			}
			if !reflect.DeepEqual(got.Compatibility, tt.skill.Compatibility) {
				t.Errorf("Compatibility = %v, want %v", got.Compatibility, tt.skill.Compatibility)
			}
			if !reflect.DeepEqual(got.Metadata, tt.skill.Metadata) {
				t.Errorf("Metadata = %v, want %v", got.Metadata, tt.skill.Metadata)
			}
			if !reflect.DeepEqual(got.AllowedTools, tt.skill.AllowedTools) {
				t.Errorf("AllowedTools = %q, want %q", got.AllowedTools, tt.skill.AllowedTools)
			}
			if got.Instructions != tt.skill.Instructions {
				t.Errorf("Instructions = %q, want %q", got.Instructions, tt.skill.Instructions)
			}
		})
	}
}

func TestSkillManager_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, tmpDir)
	mgr := NewSkillManager(paths)

	// Full skill with all fields populated
	original := &Skill{
		Name:          "round-trip-skill",
		Description:   "Tests complete round-trip",
		License:       "MIT",
		Compatibility: []string{"claude-code", "opencode", "codex"},
		Metadata:      map[string]string{"version": "3.2.1", "author": "test-author", "repository": "github.com/test/repo"},
		AllowedTools:  ToolList{"Read", "Write", "Bash", "Glob"},
		Instructions:  "Complex instructions\n\nWith multiple paragraphs.\n\n- And bullet points\n- Like this one",
	}

	// Install
	if err := mgr.Install(original); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Retrieve
	retrieved, err := mgr.Get(original.Name)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Verify complete equality
	if !reflect.DeepEqual(retrieved, original) {
		t.Errorf("Round-trip mismatch:\ngot:  %+v\nwant: %+v", retrieved, original)
	}

	// List should contain it
	skills, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	found := false
	for _, s := range skills {
		if s.Name == original.Name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("skill not found in List()")
	}

	// Uninstall
	if err := mgr.Uninstall(original.Name); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	// Verify gone
	_, err = mgr.Get(original.Name)
	if !errors.Is(err, ErrSkillNotFound) {
		t.Errorf("Get() after uninstall = %v, want %v", err, ErrSkillNotFound)
	}
}

func TestSkillManager_CheckCollision(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := t.TempDir()

	// Mock HOME for user scope resolution
	t.Setenv("HOME", homeDir)

	// User scope paths (aware of project root)
	userPaths := NewClaudePaths(ScopeUser, projectDir)
	userMgr := NewSkillManager(userPaths)

	// Project scope paths
	projectPaths := NewClaudePaths(ScopeProject, projectDir)
	projectMgr := NewSkillManager(projectPaths)

	// Case 1: No collision initially
	found, err := projectMgr.CheckCollision("collision-skill")
	if err != nil {
		t.Fatalf("CheckCollision() error = %v", err)
	}
	if found {
		t.Error("CheckCollision() = true, want false")
	}

	// Case 2: Create skill in User scope, check from Project scope
	skill := &Skill{Name: "collision-skill", Description: "User skill"}
	if err := userMgr.Install(skill); err != nil {
		t.Fatalf("userMgr.Install() error = %v", err)
	}

	found, err = projectMgr.CheckCollision("collision-skill")
	if err != nil {
		t.Fatalf("CheckCollision() error = %v", err)
	}
	if !found {
		t.Error("CheckCollision() = false, want true (collision with user scope)")
	}

	// Case 3: Create skill in Project scope, check from User scope
	projSkill := &Skill{Name: "project-skill", Description: "Project skill"}
	if err := projectMgr.Install(projSkill); err != nil {
		t.Fatalf("projectMgr.Install() error = %v", err)
	}

	found, err = userMgr.CheckCollision("project-skill")
	if err != nil {
		t.Fatalf("CheckCollision() error = %v", err)
	}
	if !found {
		t.Error("CheckCollision() = false, want true (collision with project scope)")
	}

	// Case 4: Opposing scope unavailable (User scope without project root)
	orphanUserPaths := NewClaudePaths(ScopeUser, "")
	orphanMgr := NewSkillManager(orphanUserPaths)

	found, err = orphanMgr.CheckCollision("any-skill")
	if err != nil {
		t.Fatalf("CheckCollision() error = %v", err)
	}
	if found {
		t.Error("CheckCollision() = true, want false (no opposing scope)")
	}
}
