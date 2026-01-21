package claude

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCommandManager_List(t *testing.T) {
	t.Run("empty directory returns empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		// Create empty commands directory
		cmdDir := paths.CommandDir()
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatalf("failed to create commands dir: %v", err)
		}

		commands, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if len(commands) != 0 {
			t.Errorf("List() returned %d commands, want 0", len(commands))
		}
	})

	t.Run("non-existent directory returns empty slice", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		commands, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if len(commands) != 0 {
			t.Errorf("List() returned %d commands, want 0", len(commands))
		}
	})

	t.Run("returns multiple commands", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		// Create commands directory with files
		cmdDir := paths.CommandDir()
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatalf("failed to create commands dir: %v", err)
		}

		// Write test commands
		testFiles := map[string]string{
			"build.md": "Build the project",
			"test.md":  "---\ndescription: Run tests\n---\n\nExecute test suite",
		}
		for name, content := range testFiles {
			if err := os.WriteFile(filepath.Join(cmdDir, name), []byte(content), 0o644); err != nil {
				t.Fatalf("failed to write %s: %v", name, err)
			}
		}

		commands, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if len(commands) != 2 {
			t.Errorf("List() returned %d commands, want 2", len(commands))
		}

		// Verify command names
		names := make(map[string]bool)
		for _, cmd := range commands {
			names[cmd.Name] = true
		}
		if !names["build"] || !names["test"] {
			t.Errorf("List() returned names %v, want [build, test]", names)
		}
	})

	t.Run("ignores non-md files and directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		cmdDir := paths.CommandDir()
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatalf("failed to create commands dir: %v", err)
		}

		// Create valid command
		if err := os.WriteFile(filepath.Join(cmdDir, "valid.md"), []byte("Valid command"), 0o644); err != nil {
			t.Fatalf("failed to write valid.md: %v", err)
		}

		// Create invalid files
		if err := os.WriteFile(filepath.Join(cmdDir, "readme.txt"), []byte("Not a command"), 0o644); err != nil {
			t.Fatalf("failed to write readme.txt: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(cmdDir, "subdir"), 0o755); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}

		commands, err := mgr.List()
		if err != nil {
			t.Fatalf("List() error = %v, want nil", err)
		}
		if len(commands) != 1 {
			t.Errorf("List() returned %d commands, want 1", len(commands))
		}
		if len(commands) > 0 && commands[0].Name != "valid" {
			t.Errorf("List() returned name %q, want %q", commands[0].Name, "valid")
		}
	})
}

func TestCommandManager_Get(t *testing.T) {
	t.Run("returns existing command", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		cmdDir := paths.CommandDir()
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatalf("failed to create commands dir: %v", err)
		}

		content := "---\ndescription: Build everything\n---\n\nRun the build process"
		if err := os.WriteFile(filepath.Join(cmdDir, "build.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write command file: %v", err)
		}

		cmd, err := mgr.Get("build")
		if err != nil {
			t.Fatalf("Get() error = %v, want nil", err)
		}
		if cmd.Name != "build" {
			t.Errorf("Get() name = %q, want %q", cmd.Name, "build")
		}
		if cmd.Description != "Build everything" {
			t.Errorf("Get() description = %q, want %q", cmd.Description, "Build everything")
		}
		if cmd.Instructions != "Run the build process" {
			t.Errorf("Get() instructions = %q, want %q", cmd.Instructions, "Run the build process")
		}
	})

	t.Run("returns error for non-existent command", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		_, err := mgr.Get("nonexistent")
		if !errors.Is(err, ErrCommandNotFound) {
			t.Errorf("Get() error = %v, want ErrCommandNotFound", err)
		}
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		_, err := mgr.Get("")
		if !errors.Is(err, ErrInvalidCommand) {
			t.Errorf("Get() error = %v, want ErrInvalidCommand", err)
		}
	})

	t.Run("command without frontmatter", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		cmdDir := paths.CommandDir()
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatalf("failed to create commands dir: %v", err)
		}

		content := "Just plain instructions\nNo frontmatter here"
		if err := os.WriteFile(filepath.Join(cmdDir, "simple.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write command file: %v", err)
		}

		cmd, err := mgr.Get("simple")
		if err != nil {
			t.Fatalf("Get() error = %v, want nil", err)
		}
		if cmd.Name != "simple" {
			t.Errorf("Get() name = %q, want %q", cmd.Name, "simple")
		}
		if cmd.Description != "" {
			t.Errorf("Get() description = %q, want empty", cmd.Description)
		}
		if cmd.Instructions != "Just plain instructions\nNo frontmatter here" {
			t.Errorf("Get() instructions = %q, want %q", cmd.Instructions, "Just plain instructions\nNo frontmatter here")
		}
	})

	t.Run("command with frontmatter", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		cmdDir := paths.CommandDir()
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatalf("failed to create commands dir: %v", err)
		}

		content := "---\ndescription: Test runner\n---\n\nRun all tests"
		if err := os.WriteFile(filepath.Join(cmdDir, "test.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write command file: %v", err)
		}

		cmd, err := mgr.Get("test")
		if err != nil {
			t.Fatalf("Get() error = %v, want nil", err)
		}
		if cmd.Description != "Test runner" {
			t.Errorf("Get() description = %q, want %q", cmd.Description, "Test runner")
		}
		if cmd.Instructions != "Run all tests" {
			t.Errorf("Get() instructions = %q, want %q", cmd.Instructions, "Run all tests")
		}
	})
}

func TestCommandManager_Install(t *testing.T) {
	t.Run("installs new command", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		cmd := &Command{
			Name:         "deploy",
			Description:  "Deploy to production",
			Instructions: "Deploy the application",
		}

		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v, want nil", err)
		}

		// Verify file was created
		cmdPath := paths.CommandPath("deploy")
		data, err := os.ReadFile(cmdPath)
		if err != nil {
			t.Fatalf("failed to read created file: %v", err)
		}

		content := string(data)
		if content == "" {
			t.Error("Install() created empty file")
		}

		// Verify we can read it back
		got, err := mgr.Get("deploy")
		if err != nil {
			t.Fatalf("Get() error = %v after Install()", err)
		}
		if got.Name != "deploy" {
			t.Errorf("Get() name = %q, want %q", got.Name, "deploy")
		}
		if got.Description != "Deploy to production" {
			t.Errorf("Get() description = %q, want %q", got.Description, "Deploy to production")
		}
		if got.Instructions != "Deploy the application" {
			t.Errorf("Get() instructions = %q, want %q", got.Instructions, "Deploy the application")
		}
	})

	t.Run("overwrites existing command", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		// Install initial version
		cmd := &Command{
			Name:         "build",
			Description:  "Version 1",
			Instructions: "Old instructions",
		}
		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Install updated version
		cmd.Description = "Version 2"
		cmd.Instructions = "New instructions"
		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v on overwrite", err)
		}

		// Verify update
		got, err := mgr.Get("build")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Description != "Version 2" {
			t.Errorf("Get() description = %q, want %q", got.Description, "Version 2")
		}
		if got.Instructions != "New instructions" {
			t.Errorf("Get() instructions = %q, want %q", got.Instructions, "New instructions")
		}
	})

	t.Run("creates commands directory if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		cmd := &Command{
			Name:         "init",
			Instructions: "Initialize project",
		}

		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Verify directory was created
		cmdDir := paths.CommandDir()
		info, err := os.Stat(cmdDir)
		if err != nil {
			t.Fatalf("commands directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("commands path is not a directory")
		}
	})

	t.Run("returns error for nil command", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		err := mgr.Install(nil)
		if !errors.Is(err, ErrInvalidCommand) {
			t.Errorf("Install(nil) error = %v, want ErrInvalidCommand", err)
		}
	})

	t.Run("returns error for command with empty name", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		cmd := &Command{
			Name:         "",
			Instructions: "Some instructions",
		}

		err := mgr.Install(cmd)
		if !errors.Is(err, ErrInvalidCommand) {
			t.Errorf("Install() error = %v, want ErrInvalidCommand", err)
		}
	})

	t.Run("command without description has no frontmatter", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		cmd := &Command{
			Name:         "simple",
			Instructions: "Just instructions",
		}

		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Read raw file content
		cmdPath := paths.CommandPath("simple")
		data, err := os.ReadFile(cmdPath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		content := string(data)
		if content == "" {
			t.Error("Install() created empty file")
		}

		// Should not contain frontmatter delimiter
		if content[:3] == "---" {
			t.Errorf("Install() added frontmatter when no description: %q", content)
		}
	})
}

func TestCommandManager_Uninstall(t *testing.T) {
	t.Run("removes existing command", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		// First install a command
		cmd := &Command{
			Name:         "remove-me",
			Instructions: "To be removed",
		}
		if err := mgr.Install(cmd); err != nil {
			t.Fatalf("Install() error = %v", err)
		}

		// Verify it exists
		if _, err := mgr.Get("remove-me"); err != nil {
			t.Fatalf("Get() error = %v, command should exist", err)
		}

		// Uninstall
		if err := mgr.Uninstall("remove-me"); err != nil {
			t.Fatalf("Uninstall() error = %v", err)
		}

		// Verify it's gone
		_, err := mgr.Get("remove-me")
		if !errors.Is(err, ErrCommandNotFound) {
			t.Errorf("Get() error = %v, want ErrCommandNotFound after Uninstall()", err)
		}
	})

	t.Run("idempotent for non-existent command", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		// Uninstall non-existent command should not error
		if err := mgr.Uninstall("never-existed"); err != nil {
			t.Errorf("Uninstall() error = %v, want nil for non-existent command", err)
		}
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		tmpDir := t.TempDir()
		paths := &ClaudePaths{scope: ScopeProject, projectRoot: tmpDir}
		mgr := NewCommandManager(paths)

		err := mgr.Uninstall("")
		if !errors.Is(err, ErrInvalidCommand) {
			t.Errorf("Uninstall(\"\") error = %v, want ErrInvalidCommand", err)
		}
	})
}

func TestParseCommandFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantDesc    string
		wantInstr   string
		wantErrLike string
	}{
		{
			name:      "plain markdown without frontmatter",
			content:   "Just some instructions",
			wantDesc:  "",
			wantInstr: "Just some instructions",
		},
		{
			name:      "with frontmatter",
			content:   "---\ndescription: My command\n---\n\nCommand body",
			wantDesc:  "My command",
			wantInstr: "Command body",
		},
		{
			name:      "frontmatter with no body",
			content:   "---\ndescription: Standalone\n---\n",
			wantDesc:  "Standalone",
			wantInstr: "",
		},
		{
			name:      "unclosed frontmatter treated as content",
			content:   "---\nno closing delimiter",
			wantDesc:  "",
			wantInstr: "---\nno closing delimiter",
		},
		{
			name:      "empty content",
			content:   "",
			wantDesc:  "",
			wantInstr: "",
		},
		{
			name:      "multiline instructions",
			content:   "Line 1\nLine 2\nLine 3",
			wantDesc:  "",
			wantInstr: "Line 1\nLine 2\nLine 3",
		},
		{
			name:      "frontmatter with multiline body",
			content:   "---\ndescription: Multi\n---\n\nPara 1\n\nPara 2",
			wantDesc:  "Multi",
			wantInstr: "Para 1\n\nPara 2",
		},
		{
			name:      "content starting with dashes but not frontmatter",
			content:   "--- not yaml ---\nJust text",
			wantDesc:  "",
			wantInstr: "--- not yaml ---\nJust text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parseCommandFile([]byte(tt.content))
			if tt.wantErrLike != "" {
				if err == nil || !containsSubstring(err.Error(), tt.wantErrLike) {
					t.Errorf("parseCommandFile() error = %v, want error containing %q", err, tt.wantErrLike)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCommandFile() error = %v, want nil", err)
			}
			if cmd.Description != tt.wantDesc {
				t.Errorf("parseCommandFile() description = %q, want %q", cmd.Description, tt.wantDesc)
			}
			if cmd.Instructions != tt.wantInstr {
				t.Errorf("parseCommandFile() instructions = %q, want %q", cmd.Instructions, tt.wantInstr)
			}
		})
	}
}

func TestFormatCommandFile(t *testing.T) {
	tests := []struct {
		name        string
		cmd         *Command
		wantContain []string
		wantNotHave []string
	}{
		{
			name: "command with description",
			cmd: &Command{
				Description:  "Build project",
				Instructions: "Run make build",
			},
			wantContain: []string{"---", "description: Build project", "Run make build"},
		},
		{
			name: "command without description",
			cmd: &Command{
				Instructions: "Just instructions",
			},
			wantContain: []string{"Just instructions"},
			wantNotHave: []string{"---"},
		},
		{
			name: "empty command",
			cmd:  &Command{},
			// Should produce minimal output, no frontmatter
			wantNotHave: []string{"---"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCommandFile(tt.cmd)

			for _, want := range tt.wantContain {
				if !containsSubstring(result, want) {
					t.Errorf("formatCommandFile() = %q, want to contain %q", result, want)
				}
			}
			for _, notWant := range tt.wantNotHave {
				if containsSubstring(result, notWant) {
					t.Errorf("formatCommandFile() = %q, should not contain %q", result, notWant)
				}
			}
		})
	}
}
