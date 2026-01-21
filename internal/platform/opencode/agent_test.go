package opencode

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestAgentManager_List(t *testing.T) {
	t.Run("empty directory returns empty slice", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		agents, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		// nil slice is valid Go for empty result
		if len(agents) != 0 {
			t.Errorf("List() returned %d agents, want 0", len(agents))
		}
	})

	t.Run("missing directory returns empty slice", func(t *testing.T) {
		paths := NewOpenCodePaths(ScopeProject, "/nonexistent/path")
		mgr := NewAgentManager(paths)

		agents, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if len(agents) != 0 {
			t.Errorf("List() returned %d agents, want 0", len(agents))
		}
	})

	t.Run("returns agents from directory", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		agent1 := &Agent{
			Name:         "reviewer",
			Description:  "Code reviewer",
			Instructions: "Review code",
		}
		agent2 := &Agent{
			Name:         "planner",
			Description:  "Project planner",
			Instructions: "Plan tasks",
		}

		if err := mgr.Install(agent1); err != nil {
			t.Fatalf("Install(agent1) error = %v", err)
		}
		if err := mgr.Install(agent2); err != nil {
			t.Fatalf("Install(agent2) error = %v", err)
		}

		agents, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(agents) != 2 {
			t.Errorf("List() returned %d agents, want 2", len(agents))
		}

		names := make(map[string]bool)
		for _, a := range agents {
			names[a.Name] = true
		}
		if !names["reviewer"] {
			t.Error("List() missing 'reviewer' agent")
		}
		if !names["planner"] {
			t.Error("List() missing 'planner' agent")
		}
	})

	t.Run("ignores non-md files", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		// Create agent directory with random file
		agentDir := paths.AgentDir()
		if err := os.MkdirAll(agentDir, 0o755); err != nil {
			t.Fatalf("failed to create agent dir: %v", err)
		}
		if err := os.WriteFile(agentDir+"/notes.txt", []byte("notes"), 0o644); err != nil {
			t.Fatalf("failed to create notes file: %v", err)
		}

		// Create valid agent
		if err := mgr.Install(&Agent{Name: "valid", Instructions: "test"}); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		agents, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(agents) != 1 {
			t.Errorf("List() returned %d agents, want 1", len(agents))
		}
	})
}

func TestAgentManager_Get(t *testing.T) {
	t.Run("returns agent by name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		installed := &Agent{
			Name:         "test-agent",
			Description:  "Test agent description",
			Mode:         "chat",
			Temperature:  0.7,
			Instructions: "Agent instructions here",
		}
		if err := mgr.Install(installed); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		got, err := mgr.Get("test-agent")
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

	t.Run("returns ErrAgentNotFound for nonexistent agent", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		_, err := mgr.Get("nonexistent")
		if !errors.Is(err, ErrAgentNotFound) {
			t.Errorf("Get() error = %v, want %v", err, ErrAgentNotFound)
		}
	})

	t.Run("returns ErrInvalidAgent for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		_, err := mgr.Get("")
		if !errors.Is(err, ErrInvalidAgent) {
			t.Errorf("Get() error = %v, want %v", err, ErrInvalidAgent)
		}
	})
}

func TestAgentManager_Install(t *testing.T) {
	t.Run("creates file with frontmatter when metadata present", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		agent := &Agent{
			Name:         "with-meta",
			Description:  "Has metadata",
			Mode:         "edit",
			Temperature:  0.5,
			Instructions: "Agent body",
		}

		if err := mgr.Install(agent); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		data, err := os.ReadFile(paths.AgentPath("with-meta"))
		if err != nil {
			t.Fatalf("failed to read agent file: %v", err)
		}
		content := string(data)
		if !strings.HasPrefix(content, "---") {
			t.Error("Install() with metadata should create frontmatter")
		}
		if !strings.Contains(content, "description: Has metadata") {
			t.Error("Install() file missing description in frontmatter")
		}
		if !strings.Contains(content, "mode: edit") {
			t.Error("Install() file missing mode in frontmatter")
		}
		if !strings.Contains(content, "Agent body") {
			t.Error("Install() file missing agent body")
		}
	})

	t.Run("creates file without frontmatter when no metadata", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		agent := &Agent{
			Name:         "no-meta",
			Description:  "", // Empty
			Mode:         "", // Empty
			Temperature:  0,  // Zero value
			Instructions: "Just instructions",
		}

		if err := mgr.Install(agent); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		data, err := os.ReadFile(paths.AgentPath("no-meta"))
		if err != nil {
			t.Fatalf("failed to read agent file: %v", err)
		}
		content := string(data)
		if strings.HasPrefix(content, "---") {
			t.Error("Install() without metadata should not create frontmatter")
		}
		if !strings.Contains(content, "Just instructions") {
			t.Error("Install() file missing instructions")
		}
	})

	t.Run("overwrites existing agent", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		original := &Agent{
			Name:         "overwrite",
			Description:  "Original",
			Instructions: "Original instructions",
		}
		if err := mgr.Install(original); err != nil {
			t.Fatalf("Install(original) error = %v", err)
		}

		updated := &Agent{
			Name:         "overwrite",
			Description:  "Updated",
			Instructions: "Updated instructions",
		}
		if err := mgr.Install(updated); err != nil {
			t.Fatalf("Install(updated) error = %v", err)
		}

		got, err := mgr.Get("overwrite")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Description != "Updated" {
			t.Errorf("Install() did not overwrite: Description = %q, want %q", got.Description, "Updated")
		}
	})

	t.Run("returns ErrInvalidAgent for nil agent", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		err := mgr.Install(nil)
		if !errors.Is(err, ErrInvalidAgent) {
			t.Errorf("Install(nil) error = %v, want %v", err, ErrInvalidAgent)
		}
	})

	t.Run("returns ErrInvalidAgent for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		err := mgr.Install(&Agent{Name: ""})
		if !errors.Is(err, ErrInvalidAgent) {
			t.Errorf("Install(empty name) error = %v, want %v", err, ErrInvalidAgent)
		}
	})
}

func TestAgentManager_Uninstall(t *testing.T) {
	t.Run("removes agent file", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		agent := &Agent{
			Name:         "to-remove",
			Description:  "Will be removed",
			Instructions: "test",
		}
		if err := mgr.Install(agent); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		agentPath := paths.AgentPath("to-remove")
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			t.Fatal("agent file should exist before uninstall")
		}

		if err := mgr.Uninstall("to-remove"); err != nil {
			t.Fatalf("Uninstall() error = %v", err)
		}

		if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
			t.Errorf("Uninstall() did not remove agent file %q", agentPath)
		}
	})

	t.Run("idempotent - no error if not exists", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		err := mgr.Uninstall("never-existed")
		if err != nil {
			t.Errorf("Uninstall(nonexistent) error = %v, want nil", err)
		}
	})

	t.Run("returns ErrInvalidAgent for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		err := mgr.Uninstall("")
		if !errors.Is(err, ErrInvalidAgent) {
			t.Errorf("Uninstall('') error = %v, want %v", err, ErrInvalidAgent)
		}
	})
}

func TestAgentManager_Exists(t *testing.T) {
	t.Run("returns true for existing agent", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		agent := &Agent{Name: "exists", Instructions: "test"}
		if err := mgr.Install(agent); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		if !mgr.Exists("exists") {
			t.Error("Exists() = false, want true")
		}
	})

	t.Run("returns false for nonexistent agent", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		if mgr.Exists("nonexistent") {
			t.Error("Exists() = true, want false")
		}
	})

	t.Run("returns false for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		if mgr.Exists("") {
			t.Error("Exists('') = true, want false")
		}
	})
}

func TestAgentManager_Names(t *testing.T) {
	t.Run("returns names without extension", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewAgentManager(paths)

		agents := []string{"alpha", "beta", "gamma"}
		for _, name := range agents {
			if err := mgr.Install(&Agent{Name: name, Instructions: "test"}); err != nil {
				t.Fatalf("Install(%q) error = %v", name, err)
			}
		}

		names, err := mgr.Names()
		if err != nil {
			t.Fatalf("Names() error = %v", err)
		}

		if len(names) != 3 {
			t.Errorf("Names() returned %d names, want 3", len(names))
		}

		nameSet := make(map[string]bool)
		for _, n := range names {
			nameSet[n] = true
			// Verify no .md extension
			if strings.HasSuffix(n, ".md") {
				t.Errorf("Names() returned name with extension: %q", n)
			}
		}

		for _, expected := range agents {
			if !nameSet[expected] {
				t.Errorf("Names() missing %q", expected)
			}
		}
	})

	t.Run("returns empty slice for missing directory", func(t *testing.T) {
		paths := NewOpenCodePaths(ScopeProject, "/nonexistent/path")
		mgr := NewAgentManager(paths)

		names, err := mgr.Names()
		if err != nil {
			t.Fatalf("Names() error = %v, want nil", err)
		}
		if len(names) != 0 {
			t.Errorf("Names() returned %d names, want 0", len(names))
		}
	})
}

func TestAgentManager_RoundTrip(t *testing.T) {
	paths := testPaths(t)
	mgr := NewAgentManager(paths)

	original := &Agent{
		Name:         "full-agent",
		Description:  "A complete agent",
		Mode:         "review",
		Temperature:  0.8,
		Instructions: "Multi-line\ninstructions\nhere",
	}

	if err := mgr.Install(original); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	got, err := mgr.Get("full-agent")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if got.Description != original.Description {
		t.Errorf("Description = %q, want %q", got.Description, original.Description)
	}
	if got.Mode != original.Mode {
		t.Errorf("Mode = %q, want %q", got.Mode, original.Mode)
	}
	if got.Temperature != original.Temperature {
		t.Errorf("Temperature = %f, want %f", got.Temperature, original.Temperature)
	}
	if got.Instructions != original.Instructions {
		t.Errorf("Instructions = %q, want %q", got.Instructions, original.Instructions)
	}
}
