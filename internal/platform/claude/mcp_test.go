package claude

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMCPManager_List_NonExistentConfig(t *testing.T) {
	dir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	servers, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v, want nil", err)
	}
	if len(servers) != 0 {
		t.Errorf("List() returned %d servers, want 0", len(servers))
	}
}

func TestMCPManager_List_MultipleServers(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := `{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"]
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    },
    "api-server": {
      "url": "https://api.example.com/mcp",
      "type": "http"
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	servers, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(servers) != 3 {
		t.Fatalf("List() returned %d servers, want 3", len(servers))
	}

	// Verify sorted order (api-server, filesystem, github)
	expectedNames := []string{"api-server", "filesystem", "github"}
	for i, name := range expectedNames {
		if servers[i].Name != name {
			t.Errorf("servers[%d].Name = %q, want %q", i, servers[i].Name, name)
		}
	}

	// Verify server details
	var githubServer *MCPServer
	for _, s := range servers {
		if s.Name == "github" {
			githubServer = s
			break
		}
	}
	if githubServer == nil {
		t.Fatal("github server not found")
	}
	if githubServer.Command != "npx" {
		t.Errorf("github.Command = %q, want %q", githubServer.Command, "npx")
	}
	if !reflect.DeepEqual(githubServer.Args, []string{"-y", "@modelcontextprotocol/server-github"}) {
		t.Errorf("github.Args = %v, want %v", githubServer.Args, []string{"-y", "@modelcontextprotocol/server-github"})
	}
}

func TestMCPManager_Get_ExistingServer(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := `{
  "mcpServers": {
    "test-server": {
      "command": "test-cmd",
      "args": ["--verbose"],
      "env": {"KEY": "value"}
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	server, err := mgr.Get("test-server")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if server.Name != "test-server" {
		t.Errorf("Name = %q, want %q", server.Name, "test-server")
	}
	if server.Command != "test-cmd" {
		t.Errorf("Command = %q, want %q", server.Command, "test-cmd")
	}
	if !reflect.DeepEqual(server.Args, []string{"--verbose"}) {
		t.Errorf("Args = %v, want %v", server.Args, []string{"--verbose"})
	}
	if server.Env["KEY"] != "value" {
		t.Errorf("Env[KEY] = %q, want %q", server.Env["KEY"], "value")
	}
}

func TestMCPManager_Get_NonExistentServer(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := `{"mcpServers": {}}`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	_, err := mgr.Get("nonexistent")
	if !errors.Is(err, ErrMCPServerNotFound) {
		t.Errorf("Get() error = %v, want %v", err, ErrMCPServerNotFound)
	}
}

func TestMCPManager_Add_NewServer(t *testing.T) {
	dir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	server := &MCPServer{
		Name:    "new-server",
		Command: "npx",
		Args:    []string{"-y", "some-package"},
		Env: map[string]string{
			"TOKEN": "secret",
		},
	}

	if err := mgr.Add(server); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify it was written
	got, err := mgr.Get("new-server")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Command != "npx" {
		t.Errorf("Command = %q, want %q", got.Command, "npx")
	}
	if got.Env["TOKEN"] != "secret" {
		t.Errorf("Env[TOKEN] = %q, want %q", got.Env["TOKEN"], "secret")
	}

	// Verify file exists
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestMCPManager_Add_OverwriteExisting(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := `{
  "mcpServers": {
    "existing": {
      "command": "old-cmd"
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	// Overwrite with new config
	server := &MCPServer{
		Name:    "existing",
		Command: "new-cmd",
		Args:    []string{"--new-arg"},
	}

	if err := mgr.Add(server); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, err := mgr.Get("existing")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Command != "new-cmd" {
		t.Errorf("Command = %q, want %q", got.Command, "new-cmd")
	}
	if !reflect.DeepEqual(got.Args, []string{"--new-arg"}) {
		t.Errorf("Args = %v, want %v", got.Args, []string{"--new-arg"})
	}
}

func TestMCPManager_Add_InvalidServer(t *testing.T) {
	dir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	tests := []struct {
		name   string
		server *MCPServer
	}{
		{
			name:   "nil server",
			server: nil,
		},
		{
			name:   "empty name",
			server: &MCPServer{Name: "", Command: "cmd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.Add(tt.server)
			if !errors.Is(err, ErrInvalidMCPServer) {
				t.Errorf("Add() error = %v, want %v", err, ErrInvalidMCPServer)
			}
		})
	}
}

func TestMCPManager_Remove_ExistingServer(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := `{
  "mcpServers": {
    "to-remove": {"command": "cmd1"},
    "to-keep": {"command": "cmd2"}
  }
}`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	if err := mgr.Remove("to-remove"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// Verify removal
	_, err := mgr.Get("to-remove")
	if !errors.Is(err, ErrMCPServerNotFound) {
		t.Errorf("Get(to-remove) error = %v, want %v", err, ErrMCPServerNotFound)
	}

	// Verify other server still exists
	_, err = mgr.Get("to-keep")
	if err != nil {
		t.Errorf("Get(to-keep) error = %v, want nil", err)
	}
}

func TestMCPManager_Remove_NonExistentServer(t *testing.T) {
	dir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	// Should be idempotent - no error for removing non-existent
	if err := mgr.Remove("nonexistent"); err != nil {
		t.Errorf("Remove() error = %v, want nil", err)
	}
}

func TestMCPManager_Enable(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := `{
  "mcpServers": {
    "disabled-server": {
      "command": "cmd",
      "disabled": true
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	if err := mgr.Enable("disabled-server"); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}

	got, err := mgr.Get("disabled-server")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Disabled {
		t.Error("Disabled = true, want false")
	}
}

func TestMCPManager_Disable(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configData := `{
  "mcpServers": {
    "enabled-server": {
      "command": "cmd",
      "disabled": false
    }
  }
}`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	if err := mgr.Disable("enabled-server"); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}

	got, err := mgr.Get("enabled-server")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !got.Disabled {
		t.Error("Disabled = false, want true")
	}
}

func TestMCPManager_Enable_NonExistentServer(t *testing.T) {
	dir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	err := mgr.Enable("nonexistent")
	if !errors.Is(err, ErrMCPServerNotFound) {
		t.Errorf("Enable() error = %v, want %v", err, ErrMCPServerNotFound)
	}
}

func TestMCPManager_Disable_NonExistentServer(t *testing.T) {
	dir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	err := mgr.Disable("nonexistent")
	if !errors.Is(err, ErrMCPServerNotFound) {
		t.Errorf("Disable() error = %v, want %v", err, ErrMCPServerNotFound)
	}
}

func TestMCPManager_PreservesUnknownFields(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Config with unknown fields at root level
	configData := `{
  "mcpServers": {
    "existing": {
      "command": "cmd"
    }
  },
  "futureField": "future value",
  "anotherUnknown": {
    "nested": true,
    "value": 42
  }
}`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	// Add a new server
	if err := mgr.Add(&MCPServer{Name: "new", Command: "new-cmd"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Read the file back and verify unknown fields are preserved
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Check unknown fields are preserved
	if result["futureField"] != "future value" {
		t.Errorf("futureField = %v, want %q", result["futureField"], "future value")
	}

	nested, ok := result["anotherUnknown"].(map[string]any)
	if !ok {
		t.Fatalf("anotherUnknown is not a map: %T", result["anotherUnknown"])
	}
	if nested["nested"] != true {
		t.Errorf("anotherUnknown.nested = %v, want true", nested["nested"])
	}
	// JSON numbers are float64
	if nested["value"] != float64(42) {
		t.Errorf("anotherUnknown.value = %v, want 42", nested["value"])
	}

	// Verify the new server was added
	servers, ok := result["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers is not a map: %T", result["mcpServers"])
	}
	if _, ok := servers["new"]; !ok {
		t.Error("new server was not added")
	}
	if _, ok := servers["existing"]; !ok {
		t.Error("existing server was removed")
	}
}

func TestMCPManager_CreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	// Parent directory doesn't exist yet
	configDir := filepath.Join(dir, ".claude")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf(".claude directory should not exist yet")
	}

	// Add a server - should create directory
	if err := mgr.Add(&MCPServer{Name: "test", Command: "cmd"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error(".claude directory was not created")
	}

	// Verify config file exists
	configPath := filepath.Join(configDir, ".mcp.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error(".mcp.json was not created")
	}
}

func TestMCPManager_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write initial content
	initialData := `{"mcpServers": {"initial": {"command": "cmd"}}}`
	if err := os.WriteFile(configPath, []byte(initialData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	// Add a new server
	if err := mgr.Add(&MCPServer{Name: "new", Command: "new-cmd"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify no temp files left behind
	entries, err := os.ReadDir(filepath.Dir(configPath))
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	for _, entry := range entries {
		if entry.Name() != ".mcp.json" {
			t.Errorf("unexpected file left behind: %s", entry.Name())
		}
	}
}

func TestMCPManager_JSONFormatting(t *testing.T) {
	dir := t.TempDir()
	paths := NewClaudePaths(ScopeProject, dir)
	mgr := NewMCPManager(paths)

	// Add a server
	if err := mgr.Add(&MCPServer{
		Name:    "test",
		Command: "npx",
		Args:    []string{"-y", "package"},
	}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Read the file and verify formatting
	configPath := filepath.Join(dir, ".claude", ".mcp.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	// Should have 2-space indentation
	content := string(data)
	if content[:2] != "{\n" {
		t.Error("JSON should start with {\\n")
	}
	// Check for 2-space indentation (first indented line)
	if len(content) > 4 && content[2:4] != "  " {
		t.Errorf("JSON should use 2-space indentation, got: %q", content[2:4])
	}
	// Should end with newline
	if content[len(content)-1] != '\n' {
		t.Error("JSON should end with newline")
	}
}

func TestMCPManager_ScopeLocal(t *testing.T) {
	tmpHome := t.TempDir()
	tmpProject := t.TempDir()

	// Mock HOME for user/local scope resolution
	t.Setenv("HOME", tmpHome)

	// Ensure .claude.json doesn't exist
	configPath := filepath.Join(tmpHome, ".claude.json")

	// Create manager with ScopeLocal
	paths := NewClaudePaths(ScopeLocal, tmpProject)
	mgr := NewMCPManager(paths)

	server := &MCPServer{
		Name:    "local-server",
		Command: "local-cmd",
	}

	// Add server in Local scope
	if err := mgr.Add(server); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify it was written nested under project path
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read global config: %v", err)
	}

	var fullConfig map[string]json.RawMessage
	if err := json.Unmarshal(data, &fullConfig); err != nil {
		t.Fatalf("failed to parse global config: %v", err)
	}

	absProj, _ := filepath.Abs(tmpProject)
	projectData, ok := fullConfig[absProj]
	if !ok {
		t.Fatalf("project path key %q not found in config", absProj)
	}

	var projectConfig MCPConfig
	if err := json.Unmarshal(projectData, &projectConfig); err != nil {
		t.Fatalf("failed to parse project config: %v", err)
	}

	if _, ok := projectConfig.MCPServers["local-server"]; !ok {
		t.Error("local-server not found in project config")
	}

	// Verify List() in Local scope only returns local servers
	servers, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("List() returned %d servers, want 1", len(servers))
	}
	if servers[0].Name != "local-server" {
		t.Errorf("List()[0].Name = %q, want %q", servers[0].Name, "local-server")
	}

	// Verify Get() in Local scope
	got, err := mgr.Get("local-server")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name != "local-server" {
		t.Errorf("Get().Name = %q, want %q", got.Name, "local-server")
	}

	// Now add a User scope server and verify separation
	userPaths := NewClaudePaths(ScopeUser, tmpProject)
	userMgr := NewMCPManager(userPaths)

	userServer := &MCPServer{
		Name:    "user-server",
		Command: "user-cmd",
	}
	if err := userMgr.Add(userServer); err != nil {
		t.Fatalf("userMgr.Add() error = %v", err)
	}

	// List in User scope should only show user server
	userServers, _ := userMgr.List()
	if len(userServers) != 1 || userServers[0].Name != "user-server" {
		t.Errorf("userMgr.List() = %v, want [user-server]", userServers)
	}

	// List in Local scope should still only show local server
	localServers, _ := mgr.List()
	if len(localServers) != 1 || localServers[0].Name != "local-server" {
		t.Errorf("mgr.List() = %v, want [local-server]", localServers)
	}
}
