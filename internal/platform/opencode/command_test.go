package opencode

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCommandManager_List(t *testing.T) {
	t.Run("empty directory returns empty slice", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		commands, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		// nil slice is valid Go for empty result
		if len(commands) != 0 {
			t.Errorf("List() returned %d commands, want 0", len(commands))
		}
	})

	t.Run("missing directory returns empty slice", func(t *testing.T) {
		paths := NewOpenCodePaths(ScopeProject, "/nonexistent/path")
		mgr := NewCommandManager(paths)

		commands, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if len(commands) != 0 {
			t.Errorf("List() returned %d commands, want 0", len(commands))
		}
	})

	t.Run("returns commands from directory", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		cmd1 := &Command{
			Name:         "build",
			Description:  "Build the project",
			Instructions: "Run go build",
		}
		cmd2 := &Command{
			Name:         "test",
			Description:  "Run tests",
			Instructions: "Run go test",
		}

		if err := mgr.Install(cmd1); err != nil {
			t.Fatalf("Install(cmd1) error = %v", err)
		}
		if err := mgr.Install(cmd2); err != nil {
			t.Fatalf("Install(cmd2) error = %v", err)
		}

		commands, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(commands) != 2 {
			t.Errorf("List() returned %d commands, want 2", len(commands))
		}

		// Verify command names
		names := make(map[string]bool)
		for _, c := range commands {
			names[c.Name] = true
		}
		if !names["build"] {
			t.Error("List() missing 'build' command")
		}
		if !names["test"] {
			t.Error("List() missing 'test' command")
		}
	})

	t.Run("ignores non-md files", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		// Create command directory with random file
		cmdDir := paths.CommandDir()
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatalf("failed to create command dir: %v", err)
		}
		if err := os.WriteFile(cmdDir+"/notes.txt", []byte("notes"), 0o644); err != nil {
			t.Fatalf("failed to create notes file: %v", err)
		}

		// Create valid command
		if err := mgr.Install(&Command{Name: "valid", Instructions: "test"}); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		commands, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(commands) != 1 {
			t.Errorf("List() returned %d commands, want 1", len(commands))
		}
	})
}

func TestCommandManager_Get(t *testing.T) {
	t.Run("returns command by name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		installed := &Command{
			Name:         "deploy",
			Description:  "Deploy the application",
			Instructions: "Deploy instructions here",
		}
		if err := mgr.Install(installed); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		got, err := mgr.Get("deploy")
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

	t.Run("returns ErrCommandNotFound for nonexistent command", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		_, err := mgr.Get("nonexistent")
		if !errors.Is(err, ErrCommandNotFound) {
			t.Errorf("Get() error = %v, want %v", err, ErrCommandNotFound)
		}
	})

	t.Run("returns ErrInvalidCommand for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		_, err := mgr.Get("")
		if !errors.Is(err, ErrInvalidCommand) {
			t.Errorf("Get() error = %v, want %v", err, ErrInvalidCommand)
		}
	})
}

func TestCommandManager_Install(t *testing.T) {
	t.Run("creates file with frontmatter when description present", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		cmd := &Command{
			Name:         "with-desc",
			Description:  "Has a description",
			Instructions: "Command body",
		}

		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Verify file contents include frontmatter
		data, err := os.ReadFile(paths.CommandPath("with-desc"))
		if err != nil {
			t.Fatalf("failed to read command file: %v", err)
		}
		content := string(data)
		if !strings.HasPrefix(content, "---") {
			t.Error("Install() with description should create frontmatter")
		}
		if !strings.Contains(content, "description: Has a description") {
			t.Error("Install() file missing description in frontmatter")
		}
		if !strings.Contains(content, "Command body") {
			t.Error("Install() file missing command body")
		}
	})

	t.Run("creates file without frontmatter when no description", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		cmd := &Command{
			Name:         "no-desc",
			Description:  "", // Empty description
			Instructions: "Just instructions",
		}

		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Verify no frontmatter
		data, err := os.ReadFile(paths.CommandPath("no-desc"))
		if err != nil {
			t.Fatalf("failed to read command file: %v", err)
		}
		content := string(data)
		if strings.HasPrefix(content, "---") {
			t.Error("Install() without description should not create frontmatter")
		}
		if !strings.Contains(content, "Just instructions") {
			t.Error("Install() file missing instructions")
		}
	})

	t.Run("overwrites existing command", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		original := &Command{
			Name:         "overwrite",
			Description:  "Original",
			Instructions: "Original instructions",
		}
		if err := mgr.Install(original); err != nil {
			t.Fatalf("Install(original) error = %v", err)
		}

		updated := &Command{
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

	t.Run("returns ErrInvalidCommand for nil command", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		err := mgr.Install(nil)
		if !errors.Is(err, ErrInvalidCommand) {
			t.Errorf("Install(nil) error = %v, want %v", err, ErrInvalidCommand)
		}
	})

	t.Run("returns ErrInvalidCommand for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		err := mgr.Install(&Command{Name: ""})
		if !errors.Is(err, ErrInvalidCommand) {
			t.Errorf("Install(empty name) error = %v, want %v", err, ErrInvalidCommand)
		}
	})
}

func TestCommandManager_Uninstall(t *testing.T) {
	t.Run("removes command file", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		cmd := &Command{
			Name:         "to-remove",
			Description:  "Will be removed",
			Instructions: "test",
		}
		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Verify file exists
		cmdPath := paths.CommandPath("to-remove")
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			t.Fatal("command file should exist before uninstall")
		}

		if err := mgr.Uninstall("to-remove"); err != nil {
			t.Fatalf("Uninstall() error = %v", err)
		}

		// Verify file is gone
		if _, err := os.Stat(cmdPath); !os.IsNotExist(err) {
			t.Errorf("Uninstall() did not remove command file %q", cmdPath)
		}
	})

	t.Run("idempotent - no error if not exists", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		err := mgr.Uninstall("never-existed")
		if err != nil {
			t.Errorf("Uninstall(nonexistent) error = %v, want nil", err)
		}
	})

	t.Run("returns ErrInvalidCommand for empty name", func(t *testing.T) {
		paths := testPaths(t)
		mgr := NewCommandManager(paths)

		err := mgr.Uninstall("")
		if !errors.Is(err, ErrInvalidCommand) {
			t.Errorf("Uninstall('') error = %v, want %v", err, ErrInvalidCommand)
		}
	})
}

func TestCommandManager_RoundTrip(t *testing.T) {
	paths := testPaths(t)
	mgr := NewCommandManager(paths)

	original := &Command{
		Name:         "complex-cmd",
		Description:  "A complex command",
		Instructions: "Multi-line\ninstructions\nhere",
	}

	if err := mgr.Install(original); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	got, err := mgr.Get("complex-cmd")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if got.Description != original.Description {
		t.Errorf("Description = %q, want %q", got.Description, original.Description)
	}
	if got.Instructions != original.Instructions {
		t.Errorf("Instructions = %q, want %q", got.Instructions, original.Instructions)
	}
}
