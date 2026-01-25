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

		if !strings.Contains(string(data), "prompt =") {
			t.Errorf("Command not in TOML format (missing 'prompt ='): %s", string(data))
		}

		if strings.Contains(string(data), "name =") {
			t.Errorf("Command TOML should not contain 'name =': %s", string(data))
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

func TestCommandManager_Multiline(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewGeminiPaths(ScopeProject, tmpDir)
	mgr := NewCommandManager(paths)

	cmd := &Command{
		Name:        "multiline-cmd",
		Description: "A command with multiline instructions.",
		Instructions: `This is the first line.
This is the second line.
This is the third line.`,
	}

	// Test Install with multiline
	if err := mgr.Install(cmd); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify file exists and contains multiline TOML
	cmdPath := paths.CommandPath(cmd.Name)
	data, err := os.ReadFile(cmdPath)
	if err != nil {
		t.Fatalf("Failed to read command file: %v", err)
	}
	content := string(data)

	// Check for TOML multiline syntax
	if !strings.Contains(content, `prompt = """`) {
		t.Errorf("Expected multiline TOML syntax ('\"\"\"'), but not found in:\n%s", content)
	}

	// Check if newlines are preserved
	expectedLines := strings.Split(cmd.Instructions, "\n")
	actualLines := strings.Split(strings.Trim(content, "\n"), "\n")

	// Quick and dirty check, assumes 'instructions' is the last field
	lastThreeLines := actualLines[len(actualLines)-3:]
	for i, line := range expectedLines {
		if !strings.Contains(strings.Join(lastThreeLines, "\n"), line) {
			t.Errorf("Expected line %d (%q) not found in the output TOML", i+1, line)
		}
	}

	// Clean up
	if err := mgr.Uninstall(cmd.Name); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}
}
