package gemini

import (
	"os"
	"strings"
	"testing"
)

func TestCommandManager(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewGeminiPaths(ScopeProject, tmpDir)
	mgr := NewCommandManager(paths)

	cmd := &Command{
		Name:         "test-cmd",
		Description:  "A test command",
		Instructions: "Run $ARGUMENTS",
	}

	// Test Install
	t.Run("Install", func(t *testing.T) {
		err := mgr.Install(cmd)
		if err != nil {
			t.Fatalf("Install failed: %v", err)
		}

		// Verify file exists and is TOML
		cmdPath := paths.CommandPath(cmd.Name)
		data, err := os.ReadFile(cmdPath)
		if err != nil {
			t.Fatalf("Failed to read command file: %v", err)
		}

		if !strings.Contains(string(data), "{{argument}}") {
			t.Errorf("Command content not translated: %s", string(data))
		}

		if !strings.Contains(string(data), "instructions =") {
			t.Errorf("Command not in TOML format: %s", string(data))
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		cmds, err := mgr.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(cmds) != 1 {
			t.Errorf("Expected 1 command, got %d", len(cmds))
		}

		if cmds[0].Name != "test-cmd" {
			t.Errorf("Expected test-cmd, got %s", cmds[0].Name)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		c, err := mgr.Get("test-cmd")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if c.Name != "test-cmd" {
			t.Errorf("Expected test-cmd, got %s", c.Name)
		}

		if c.Instructions != "Run $ARGUMENTS" {
			t.Errorf("Instructions not canonical: %s", c.Instructions)
		}
	})

	// Test Uninstall
	t.Run("Uninstall", func(t *testing.T) {
		err := mgr.Uninstall("test-cmd")
		if err != nil {
			t.Fatalf("Uninstall failed: %v", err)
		}

		cmdPath := paths.CommandPath("test-cmd")
		if _, err := os.Stat(cmdPath); !os.IsNotExist(err) {
			t.Errorf("Command file still exists after Uninstall")
		}
	})
}
