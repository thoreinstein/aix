package gemini

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"

	"github.com/thoreinstein/aix/internal/mcp"
)

func TestMCPManager(t *testing.T) {
	tmpDir := t.TempDir()
	paths := NewGeminiPaths(ScopeProject, tmpDir)
	mgr := NewMCPManager(paths)

	// Test settings.toml preservation
	configPath := paths.MCPConfigPath()
	initialSettings := `other = "value"
[mcp.servers]
`
	err := os.MkdirAll(filepath.Dir(configPath), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(configPath, []byte(initialSettings), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	server := &MCPServer{
		Name:    "test-server",
		Command: "node",
		Args:    []string{"server.js"},
		Enabled: true,
	}

	// Test Add
	t.Run("Add", func(t *testing.T) {
		err := mgr.Add(server)
		if err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Verify settings.toml preservation
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}

		var raw map[string]any
		if err := toml.Unmarshal(data, &raw); err != nil {
			t.Fatalf("Failed to unmarshal settings: %v", err)
		}
		if raw["other"] != "value" {
			t.Errorf("Initial setting 'other' not preserved: %v", raw["other"])
		}

		if _, ok := raw["mcp"]; !ok {
			t.Errorf("mcp section missing")
		}
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		servers, err := mgr.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(servers) != 1 {
			t.Errorf("Expected 1 server, got %d", len(servers))
		}

		if servers[0].Name != "test-server" {
			t.Errorf("Expected test-server, got %s", servers[0].Name)
		}
	})

	// Test Remove
	t.Run("Remove", func(t *testing.T) {
		err := mgr.Remove("test-server")
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		servers, _ := mgr.List()
		if len(servers) != 0 {
			t.Errorf("Server still exists after Remove")
		}
	})
}

func TestMCPTranslator(t *testing.T) {
	translator := NewMCPTranslator()

	t.Run("ToCanonical", func(t *testing.T) {
		geminiTOML := `
[servers]
  [servers.myserver]
    command = "node"
    enabled = true
`
		config, err := translator.ToCanonical([]byte(geminiTOML))
		if err != nil {
			t.Fatalf("ToCanonical failed: %v", err)
		}

		server, ok := config.Servers["myserver"]
		if !ok {
			t.Fatal("server myserver not found")
		}

		if server.Command != "node" {
			t.Errorf("Expected command node, got %s", server.Command)
		}

		if server.Disabled {
			t.Errorf("Expected enabled (disabled=false), got disabled=true")
		}
	})

	t.Run("FromCanonical", func(t *testing.T) {
		config := mcp.NewConfig()
		config.Servers["myserver"] = &mcp.Server{
			Name:     "myserver",
			Command:  "node",
			Args:     []string{"server.js"},
			Disabled: false,
		}

		data, err := translator.FromCanonical(config)
		if err != nil {
			t.Fatalf("FromCanonical failed: %v", err)
		}

		var geminiConfig MCPConfig
		err = toml.Unmarshal(data, &geminiConfig)
		if err != nil {
			t.Fatal(err)
		}

		server, ok := geminiConfig.Servers["myserver"]
		if !ok {
			t.Fatal("server myserver not found in output")
		}

		if !server.Enabled {
			t.Errorf("Expected enabled=true, got enabled=false")
		}
	})
}
