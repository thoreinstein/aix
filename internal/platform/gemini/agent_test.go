package gemini

import (
	"os"
	"strings"
	"testing"
)

func TestAgentManager(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewGeminiPaths(ScopeProject, tmpDir)
	mgr := NewAgentManager(paths)

	agent := &Agent{
		Name:         "test-agent",
		Description:  "A test agent",
		Instructions: "You are a test agent.",
	}

	// Test Install
	t.Run("Install", func(t *testing.T) {
		err := mgr.Install(agent)
		if err != nil {
			t.Fatalf("Install failed: %v", err)
		}

		// Verify file exists
		agentPath := paths.AgentPath(agent.Name)
		data, err := os.ReadFile(agentPath)
		if err != nil {
			t.Fatalf("Failed to read agent file: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "name: test-agent") {
			t.Errorf("Agent file missing name: %s", content)
		}
		if !strings.Contains(content, "description: A test agent") {
			t.Errorf("Agent file missing description: %s", content)
		}
		if !strings.Contains(content, "You are a test agent.") {
			t.Errorf("Agent file missing instructions: %s", content)
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		agents, err := mgr.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(agents) != 1 {
			t.Errorf("Expected 1 agent, got %d", len(agents))
		}

		if agents[0].Name != "test-agent" {
			t.Errorf("Expected test-agent, got %s", agents[0].Name)
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		a, err := mgr.Get("test-agent")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if a.Name != "test-agent" {
			t.Errorf("Expected test-agent, got %s", a.Name)
		}

		if a.Description != "A test agent" {
			t.Errorf("Expected description 'A test agent', got %q", a.Description)
		}

		if a.Instructions != "You are a test agent." {
			t.Errorf("Expected instructions, got %q", a.Instructions)
		}
	})

	// Test Uninstall
	t.Run("Uninstall", func(t *testing.T) {
		err := mgr.Uninstall("test-agent")
		if err != nil {
			t.Fatalf("Uninstall failed: %v", err)
		}

		// Verify file is gone
		agentPath := paths.AgentPath("test-agent")
		if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
			t.Errorf("Agent file still exists after Uninstall")
		}
	})
}

func TestGeminiPlatform_InstallAgent_EnablesExperimental(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewGeminiPlatform(WithScope(ScopeProject), WithProjectRoot(tmpDir))

	agent := &Agent{
		Name:         "test-agent",
		Instructions: "Instructions",
	}

	err := p.InstallAgent(agent)
	if err != nil {
		t.Fatalf("InstallAgent failed: %v", err)
	}

	// Verify settings.toml has enableAgents = true
	settingsPath := p.MCPConfigPath()
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "enableAgents = true") {
		t.Errorf("settings.toml does not enable agents: %s", content)
	}
}
