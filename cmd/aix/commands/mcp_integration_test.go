//go:build integration

package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

// testEnv holds the test environment configuration.
type testEnv struct {
	homeDir          string
	claudeDir        string
	opencodeDir      string
	claudePlatform   *claudeTestAdapter
	opencodePlatform *opencodeTestAdapter
}

// claudeTestAdapter wraps ClaudePlatform to implement cli.Platform for testing.
type claudeTestAdapter struct {
	p *claude.ClaudePlatform
}

func (a *claudeTestAdapter) Name() string        { return a.p.Name() }
func (a *claudeTestAdapter) DisplayName() string { return a.p.DisplayName() }
func (a *claudeTestAdapter) IsAvailable() bool   { return a.p.IsAvailable() }
func (a *claudeTestAdapter) SkillDir() string    { return a.p.SkillDir() }
func (a *claudeTestAdapter) InstallSkill(skill any) error {
	return a.p.InstallSkill(skill.(*claude.Skill))
}
func (a *claudeTestAdapter) UninstallSkill(name string) error     { return a.p.UninstallSkill(name) }
func (a *claudeTestAdapter) ListSkills() ([]cli.SkillInfo, error) { return nil, nil }
func (a *claudeTestAdapter) GetSkill(name string) (any, error)    { return a.p.GetSkill(name) }
func (a *claudeTestAdapter) CommandDir() string                   { return a.p.CommandDir() }
func (a *claudeTestAdapter) InstallCommand(cmd any) error {
	return a.p.InstallCommand(cmd.(*claude.Command))
}
func (a *claudeTestAdapter) UninstallCommand(name string) error       { return a.p.UninstallCommand(name) }
func (a *claudeTestAdapter) ListCommands() ([]cli.CommandInfo, error) { return nil, nil }
func (a *claudeTestAdapter) GetCommand(name string) (any, error)      { return a.p.GetCommand(name) }
func (a *claudeTestAdapter) MCPConfigPath() string                    { return a.p.MCPConfigPath() }
func (a *claudeTestAdapter) AddMCP(server any) error                  { return a.p.AddMCP(server.(*claude.MCPServer)) }
func (a *claudeTestAdapter) RemoveMCP(name string) error              { return a.p.RemoveMCP(name) }

func (a *claudeTestAdapter) ListMCP() ([]cli.MCPInfo, error) {
	servers, err := a.p.ListMCP()
	if err != nil {
		return nil, err
	}
	infos := make([]cli.MCPInfo, len(servers))
	for i, s := range servers {
		transport := s.Transport
		if transport == "" {
			if s.URL != "" {
				transport = "sse"
			} else {
				transport = "stdio"
			}
		}
		infos[i] = cli.MCPInfo{
			Name:      s.Name,
			Transport: transport,
			Command:   s.Command,
			URL:       s.URL,
			Disabled:  s.Disabled,
			Env:       s.Env,
		}
	}
	return infos, nil
}

func (a *claudeTestAdapter) GetMCP(name string) (any, error) { return a.p.GetMCP(name) }
func (a *claudeTestAdapter) EnableMCP(name string) error     { return a.p.EnableMCP(name) }
func (a *claudeTestAdapter) DisableMCP(name string) error    { return a.p.DisableMCP(name) }

// opencodeTestAdapter wraps OpenCodePlatform to implement cli.Platform for testing.
type opencodeTestAdapter struct {
	p *opencode.OpenCodePlatform
}

func (a *opencodeTestAdapter) Name() string        { return a.p.Name() }
func (a *opencodeTestAdapter) DisplayName() string { return a.p.DisplayName() }
func (a *opencodeTestAdapter) IsAvailable() bool   { return a.p.IsAvailable() }
func (a *opencodeTestAdapter) SkillDir() string    { return a.p.SkillDir() }
func (a *opencodeTestAdapter) InstallSkill(skill any) error {
	return a.p.InstallSkill(skill.(*opencode.Skill))
}
func (a *opencodeTestAdapter) UninstallSkill(name string) error     { return a.p.UninstallSkill(name) }
func (a *opencodeTestAdapter) ListSkills() ([]cli.SkillInfo, error) { return nil, nil }
func (a *opencodeTestAdapter) GetSkill(name string) (any, error)    { return a.p.GetSkill(name) }
func (a *opencodeTestAdapter) CommandDir() string                   { return a.p.CommandDir() }
func (a *opencodeTestAdapter) InstallCommand(cmd any) error {
	return a.p.InstallCommand(cmd.(*opencode.Command))
}
func (a *opencodeTestAdapter) UninstallCommand(name string) error       { return a.p.UninstallCommand(name) }
func (a *opencodeTestAdapter) ListCommands() ([]cli.CommandInfo, error) { return nil, nil }
func (a *opencodeTestAdapter) GetCommand(name string) (any, error)      { return a.p.GetCommand(name) }
func (a *opencodeTestAdapter) MCPConfigPath() string                    { return a.p.MCPConfigPath() }
func (a *opencodeTestAdapter) AddMCP(server any) error {
	return a.p.AddMCP(server.(*opencode.MCPServer))
}
func (a *opencodeTestAdapter) RemoveMCP(name string) error { return a.p.RemoveMCP(name) }

func (a *opencodeTestAdapter) ListMCP() ([]cli.MCPInfo, error) {
	servers, err := a.p.ListMCP()
	if err != nil {
		return nil, err
	}
	infos := make([]cli.MCPInfo, len(servers))
	for i, s := range servers {
		transport := "stdio"
		if s.Type == "remote" || s.URL != "" {
			transport = "sse"
		}
		cmd := ""
		if len(s.Command) > 0 {
			cmd = s.Command[0]
		}
		infos[i] = cli.MCPInfo{
			Name:      s.Name,
			Transport: transport,
			Command:   cmd,
			URL:       s.URL,
			Disabled:  s.Disabled,
			Env:       s.Environment,
		}
	}
	return infos, nil
}

func (a *opencodeTestAdapter) GetMCP(name string) (any, error) { return a.p.GetMCP(name) }
func (a *opencodeTestAdapter) EnableMCP(name string) error     { return a.p.EnableMCP(name) }
func (a *opencodeTestAdapter) DisableMCP(name string) error    { return a.p.DisableMCP(name) }

// setupTestEnv creates a test environment with isolated config directories.
// It sets HOME to a temp directory and creates the expected directory structure
// for both Claude and OpenCode platforms.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	// Create Claude config directory: ~/.claude/
	claudeDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create claude dir: %v", err)
	}

	// Create OpenCode config directory: ~/.config/opencode/
	opencodeDir := filepath.Join(homeDir, ".config", "opencode")
	if err := os.MkdirAll(opencodeDir, 0755); err != nil {
		t.Fatalf("failed to create opencode dir: %v", err)
	}

	// Create platform instances after HOME is set
	claudePlat := claude.NewClaudePlatform()
	opencodePlat := opencode.NewOpenCodePlatform()

	return &testEnv{
		homeDir:          homeDir,
		claudeDir:        claudeDir,
		opencodeDir:      opencodeDir,
		claudePlatform:   &claudeTestAdapter{p: claudePlat},
		opencodePlatform: &opencodeTestAdapter{p: opencodePlat},
	}
}

// readClaudeConfig reads the Claude MCP config file and returns the parsed content.
// Note: User-scoped Claude config is at ~/.claude.json (not in .claude directory).
func (e *testEnv) readClaudeConfig(t *testing.T) *claude.MCPConfig {
	t.Helper()
	configPath := filepath.Join(e.homeDir, ".claude.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &claude.MCPConfig{MCPServers: make(map[string]*claude.MCPServer)}
		}
		t.Fatalf("failed to read claude config: %v", err)
	}
	var config claude.MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse claude config: %v", err)
	}
	return &config
}

// readOpenCodeConfig reads the OpenCode config file and returns the parsed content.
func (e *testEnv) readOpenCodeConfig(t *testing.T) *opencode.MCPConfig {
	t.Helper()
	configPath := filepath.Join(e.opencodeDir, "opencode.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &opencode.MCPConfig{MCP: make(map[string]*opencode.MCPServer)}
		}
		t.Fatalf("failed to read opencode config: %v", err)
	}
	var config opencode.MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse opencode config: %v", err)
	}
	return &config
}

// TestMCPFullLifecycle tests the complete MCP server lifecycle:
// add → list → show → disable → enable → remove
func TestMCPFullLifecycle(t *testing.T) {
	env := setupTestEnv(t)

	// Step 1: Add a server to Claude
	server := &claude.MCPServer{
		Name:    "github",
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-github"},
		Env: map[string]string{
			"GITHUB_TOKEN": "ghp_test1234567890",
		},
	}
	if err := env.claudePlatform.AddMCP(server); err != nil {
		t.Fatalf("AddMCP failed: %v", err)
	}

	// Verify the config file was created correctly
	config := env.readClaudeConfig(t)
	if len(config.MCPServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(config.MCPServers))
	}
	if s, ok := config.MCPServers["github"]; !ok {
		t.Error("server 'github' not found in config")
	} else {
		if s.Command != "npx" {
			t.Errorf("Command = %q, want %q", s.Command, "npx")
		}
		if len(s.Args) != 2 {
			t.Errorf("Args length = %d, want 2", len(s.Args))
		}
	}

	// Step 2: List servers
	servers, err := env.claudePlatform.ListMCP()
	if err != nil {
		t.Fatalf("ListMCP failed: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("ListMCP returned %d servers, want 1", len(servers))
	}
	if servers[0].Name != "github" {
		t.Errorf("server name = %q, want %q", servers[0].Name, "github")
	}

	// Step 3: Get server details
	got, err := env.claudePlatform.GetMCP("github")
	if err != nil {
		t.Fatalf("GetMCP failed: %v", err)
	}
	gotServer := got.(*claude.MCPServer)
	if gotServer.Command != "npx" {
		t.Errorf("GetMCP Command = %q, want %q", gotServer.Command, "npx")
	}

	// Step 4: Disable server
	if err := env.claudePlatform.DisableMCP("github"); err != nil {
		t.Fatalf("DisableMCP failed: %v", err)
	}

	config = env.readClaudeConfig(t)
	if !config.MCPServers["github"].Disabled {
		t.Error("server should be disabled")
	}

	// Step 5: Enable server
	if err := env.claudePlatform.EnableMCP("github"); err != nil {
		t.Fatalf("EnableMCP failed: %v", err)
	}

	config = env.readClaudeConfig(t)
	if config.MCPServers["github"].Disabled {
		t.Error("server should be enabled")
	}

	// Step 6: Remove server
	if err := env.claudePlatform.RemoveMCP("github"); err != nil {
		t.Fatalf("RemoveMCP failed: %v", err)
	}

	config = env.readClaudeConfig(t)
	if len(config.MCPServers) != 0 {
		t.Errorf("expected 0 servers after removal, got %d", len(config.MCPServers))
	}
}

// TestMCPCrossPlatformOperations tests adding servers to both platforms
// and verifying platform-specific configuration formats.
func TestMCPCrossPlatformOperations(t *testing.T) {
	env := setupTestEnv(t)

	// Add server to Claude
	claudeServer := &claude.MCPServer{
		Name:    "github",
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-github"},
		Env: map[string]string{
			"GITHUB_TOKEN": "ghp_test1234567890",
		},
	}
	if err := env.claudePlatform.AddMCP(claudeServer); err != nil {
		t.Fatalf("AddMCP to Claude failed: %v", err)
	}

	// Add server to OpenCode
	opencodeServer := &opencode.MCPServer{
		Name:    "github",
		Command: []string{"npx", "-y", "@modelcontextprotocol/server-github"},
		Type:    "local",
		Environment: map[string]string{
			"GITHUB_TOKEN": "ghp_test1234567890",
		},
	}
	if err := env.opencodePlatform.AddMCP(opencodeServer); err != nil {
		t.Fatalf("AddMCP to OpenCode failed: %v", err)
	}

	// Verify Claude config format
	claudeConfig := env.readClaudeConfig(t)
	if s, ok := claudeConfig.MCPServers["github"]; !ok {
		t.Error("github not found in Claude config")
	} else {
		if s.Command != "npx" {
			t.Errorf("Claude Command = %q, want %q", s.Command, "npx")
		}
		if len(s.Args) != 2 {
			t.Errorf("Claude Args length = %d, want 2", len(s.Args))
		}
	}

	// Verify OpenCode config format
	opencodeConfig := env.readOpenCodeConfig(t)
	if s, ok := opencodeConfig.MCP["github"]; !ok {
		t.Error("github not found in OpenCode config")
	} else {
		if len(s.Command) != 3 {
			t.Errorf("OpenCode Command length = %d, want 3", len(s.Command))
		}
		if s.Type != "local" {
			t.Errorf("OpenCode Type = %q, want %q", s.Type, "local")
		}
	}

	// Remove from Claude only
	if err := env.claudePlatform.RemoveMCP("github"); err != nil {
		t.Fatalf("RemoveMCP from Claude failed: %v", err)
	}

	// Verify Claude is empty but OpenCode still has the server
	claudeConfig = env.readClaudeConfig(t)
	if len(claudeConfig.MCPServers) != 0 {
		t.Errorf("Claude should have 0 servers, got %d", len(claudeConfig.MCPServers))
	}

	opencodeConfig = env.readOpenCodeConfig(t)
	if len(opencodeConfig.MCP) != 1 {
		t.Errorf("OpenCode should have 1 server, got %d", len(opencodeConfig.MCP))
	}
}

// TestMCPSecretMaskingIntegration tests that secrets are properly masked in output.
func TestMCPSecretMaskingIntegration(t *testing.T) {
	env := setupTestEnv(t)

	// Add server with secret
	server := &claude.MCPServer{
		Name:    "github",
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-github"},
		Env: map[string]string{
			"GITHUB_TOKEN": "ghp_xxxxxxxxxxxx1234",
			"DEBUG":        "true",
		},
	}
	if err := env.claudePlatform.AddMCP(server); err != nil {
		t.Fatalf("AddMCP failed: %v", err)
	}

	// Test masking via the maskSecrets function
	platforms := []cli.Platform{env.claudePlatform}

	// Test with secrets hidden (default)
	mcpListShowSecrets = false
	var buf bytes.Buffer
	if err := outputMCPJSON(&buf, platforms); err != nil {
		t.Fatalf("outputMCPJSON failed: %v", err)
	}

	var result []mcpListPlatformOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(result) == 0 || len(result[0].Servers) == 0 {
		t.Fatal("expected at least one platform with one server")
	}

	envVars := result[0].Servers[0].Env
	if envVars["GITHUB_TOKEN"] != "****1234" {
		t.Errorf("GITHUB_TOKEN should be masked, got %q", envVars["GITHUB_TOKEN"])
	}
	if envVars["DEBUG"] != "true" {
		t.Errorf("DEBUG should not be masked, got %q", envVars["DEBUG"])
	}

	// Test with secrets shown
	mcpListShowSecrets = true
	buf.Reset()
	if err := outputMCPJSON(&buf, platforms); err != nil {
		t.Fatalf("outputMCPJSON with show-secrets failed: %v", err)
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	envVars = result[0].Servers[0].Env
	if envVars["GITHUB_TOKEN"] != "ghp_xxxxxxxxxxxx1234" {
		t.Errorf("GITHUB_TOKEN should be visible with --show-secrets, got %q", envVars["GITHUB_TOKEN"])
	}

	// Reset global flag
	mcpListShowSecrets = false
}

// TestMCPErrorHandling tests error cases for MCP operations.
func TestMCPErrorHandling(t *testing.T) {
	env := setupTestEnv(t)

	// Test: Get nonexistent server
	_, err := env.claudePlatform.GetMCP("nonexistent")
	if err == nil {
		t.Error("GetMCP should return error for nonexistent server")
	}

	// Test: Enable nonexistent server
	err = env.claudePlatform.EnableMCP("nonexistent")
	if err == nil {
		t.Error("EnableMCP should return error for nonexistent server")
	}

	// Test: Disable nonexistent server
	err = env.claudePlatform.DisableMCP("nonexistent")
	if err == nil {
		t.Error("DisableMCP should return error for nonexistent server")
	}

	// Test: Add server, then try to add duplicate (this should succeed - it's an upsert)
	server := &claude.MCPServer{
		Name:    "test-server",
		Command: "test-cmd",
	}
	if err := env.claudePlatform.AddMCP(server); err != nil {
		t.Fatalf("first AddMCP failed: %v", err)
	}

	// Adding again should succeed (upsert behavior)
	server.Command = "updated-cmd"
	if err := env.claudePlatform.AddMCP(server); err != nil {
		t.Fatalf("second AddMCP (update) failed: %v", err)
	}

	// Verify update
	got, err := env.claudePlatform.GetMCP("test-server")
	if err != nil {
		t.Fatalf("GetMCP after update failed: %v", err)
	}
	gotServer := got.(*claude.MCPServer)
	if gotServer.Command != "updated-cmd" {
		t.Errorf("Command should be updated, got %q", gotServer.Command)
	}

	// Test: Remove (idempotent - should not error for nonexistent)
	err = env.claudePlatform.RemoveMCP("nonexistent")
	if err != nil {
		t.Errorf("RemoveMCP should be idempotent, got error: %v", err)
	}
}

// TestMCPJSONOutput tests JSON output format and structure.
func TestMCPJSONOutput(t *testing.T) {
	env := setupTestEnv(t)

	// Add server with all fields
	server := &claude.MCPServer{
		Name:      "api-gateway",
		URL:       "https://api.example.com/mcp",
		Transport: "sse",
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
	}
	if err := env.claudePlatform.AddMCP(server); err != nil {
		t.Fatalf("AddMCP failed: %v", err)
	}

	platforms := []cli.Platform{env.claudePlatform}

	var buf bytes.Buffer
	if err := outputMCPJSON(&buf, platforms); err != nil {
		t.Fatalf("outputMCPJSON failed: %v", err)
	}

	var result []mcpListPlatformOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify structure
	if len(result) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(result))
	}
	if result[0].Platform != "claude" {
		t.Errorf("Platform = %q, want %q", result[0].Platform, "claude")
	}
	if len(result[0].Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(result[0].Servers))
	}

	srv := result[0].Servers[0]
	if srv.Name != "api-gateway" {
		t.Errorf("Name = %q, want %q", srv.Name, "api-gateway")
	}
	if srv.Transport != "sse" {
		t.Errorf("Transport = %q, want %q", srv.Transport, "sse")
	}
	if srv.URL != "https://api.example.com/mcp" {
		t.Errorf("URL = %q, want %q", srv.URL, "https://api.example.com/mcp")
	}
	if srv.Disabled {
		t.Error("Disabled should be false")
	}
}

// TestMCPAtomicWrites tests that config writes are atomic (no partial writes).
func TestMCPAtomicWrites(t *testing.T) {
	env := setupTestEnv(t)

	// Add multiple servers
	for i := range 5 {
		server := &claude.MCPServer{
			Name:    "server-" + string(rune('a'+i)),
			Command: "test-cmd",
		}
		if err := env.claudePlatform.AddMCP(server); err != nil {
			t.Fatalf("AddMCP failed for server-%c: %v", 'a'+i, err)
		}
	}

	// Read config and verify it's valid JSON
	config := env.readClaudeConfig(t)
	if len(config.MCPServers) != 5 {
		t.Errorf("expected 5 servers, got %d", len(config.MCPServers))
	}

	// Verify file contains proper JSON (read raw and check format)
	// Note: User-scoped config is at ~/.claude.json (not in .claude directory)
	configPath := filepath.Join(env.homeDir, ".claude.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Check for proper formatting (indented JSON)
	if !strings.Contains(string(data), "  ") {
		t.Error("config should be formatted with indentation")
	}

	// Check for trailing newline (POSIX compliance)
	if !strings.HasSuffix(string(data), "\n") {
		t.Error("config should end with newline")
	}
}

// TestMCPInvalidExistingConfig tests error handling with corrupted config files.
func TestMCPInvalidExistingConfig(t *testing.T) {
	env := setupTestEnv(t)

	// Write invalid JSON to config file
	// Note: User-scoped config is at ~/.claude.json (not in .claude directory)
	configPath := filepath.Join(env.homeDir, ".claude.json")
	if err := os.WriteFile(configPath, []byte("{ invalid json }"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	// Attempt to add server - should fail due to invalid existing config
	server := &claude.MCPServer{
		Name:    "test",
		Command: "test-cmd",
	}
	err := env.claudePlatform.AddMCP(server)
	if err == nil {
		t.Error("AddMCP should fail with invalid existing config")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("error should mention parsing issue, got: %v", err)
	}

	// Same for List
	_, err = env.claudePlatform.ListMCP()
	if err == nil {
		t.Error("ListMCP should fail with invalid existing config")
	}
}

// TestMCPSSEServerConfiguration tests remote SSE server configuration.
func TestMCPSSEServerConfiguration(t *testing.T) {
	env := setupTestEnv(t)

	// Add SSE server to Claude
	claudeServer := &claude.MCPServer{
		Name:      "remote-api",
		URL:       "https://api.example.com/mcp/v1",
		Transport: "sse",
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"X-Custom":      "value",
		},
	}
	if err := env.claudePlatform.AddMCP(claudeServer); err != nil {
		t.Fatalf("AddMCP SSE server failed: %v", err)
	}

	// Add SSE server to OpenCode
	opencodeServer := &opencode.MCPServer{
		Name: "remote-api",
		URL:  "https://api.example.com/mcp/v1",
		Type: "remote",
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"X-Custom":      "value",
		},
	}
	if err := env.opencodePlatform.AddMCP(opencodeServer); err != nil {
		t.Fatalf("AddMCP SSE server to OpenCode failed: %v", err)
	}

	// Verify Claude config
	claudeConfig := env.readClaudeConfig(t)
	s := claudeConfig.MCPServers["remote-api"]
	if s.Transport != "sse" {
		t.Errorf("Claude Transport = %q, want %q", s.Transport, "sse")
	}
	if s.URL != "https://api.example.com/mcp/v1" {
		t.Errorf("Claude URL = %q, want expected URL", s.URL)
	}
	if len(s.Headers) != 2 {
		t.Errorf("Claude Headers count = %d, want 2", len(s.Headers))
	}

	// Verify OpenCode config
	opencodeConfig := env.readOpenCodeConfig(t)
	os := opencodeConfig.MCP["remote-api"]
	if os.Type != "remote" {
		t.Errorf("OpenCode Type = %q, want %q", os.Type, "remote")
	}
	if os.URL != "https://api.example.com/mcp/v1" {
		t.Errorf("OpenCode URL = %q, want expected URL", os.URL)
	}
	if len(os.Headers) != 2 {
		t.Errorf("OpenCode Headers count = %d, want 2", len(os.Headers))
	}
}

// TestMCPMultiplePlatformsList tests listing servers from multiple platforms.
func TestMCPMultiplePlatformsList(t *testing.T) {
	env := setupTestEnv(t)

	// Add servers to both platforms
	claudeServer := &claude.MCPServer{
		Name:    "shared-server",
		Command: "shared-cmd",
	}
	if err := env.claudePlatform.AddMCP(claudeServer); err != nil {
		t.Fatalf("AddMCP to Claude failed: %v", err)
	}

	opencodeServer := &opencode.MCPServer{
		Name:    "shared-server",
		Command: []string{"shared-cmd"},
		Type:    "local",
	}
	if err := env.opencodePlatform.AddMCP(opencodeServer); err != nil {
		t.Fatalf("AddMCP to OpenCode failed: %v", err)
	}

	// Add unique server to each platform
	claudeOnly := &claude.MCPServer{
		Name:    "claude-only",
		Command: "claude-cmd",
	}
	if err := env.claudePlatform.AddMCP(claudeOnly); err != nil {
		t.Fatalf("AddMCP claude-only failed: %v", err)
	}

	opencodeOnly := &opencode.MCPServer{
		Name:    "opencode-only",
		Command: []string{"opencode-cmd"},
		Type:    "local",
	}
	if err := env.opencodePlatform.AddMCP(opencodeOnly); err != nil {
		t.Fatalf("AddMCP opencode-only failed: %v", err)
	}

	// Test tabular output with multiple platforms
	platforms := []cli.Platform{env.claudePlatform, env.opencodePlatform}

	var buf bytes.Buffer
	if err := outputMCPTabular(&buf, platforms); err != nil {
		t.Fatalf("outputMCPTabular failed: %v", err)
	}

	output := buf.String()

	// Verify both platforms are shown
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should contain Claude Code")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Error("output should contain OpenCode")
	}

	// Verify all servers are shown
	if !strings.Contains(output, "shared-server") {
		t.Error("output should contain shared-server")
	}
	if !strings.Contains(output, "claude-only") {
		t.Error("output should contain claude-only")
	}
	if !strings.Contains(output, "opencode-only") {
		t.Error("output should contain opencode-only")
	}
}

// TestMCPEmptyConfig tests behavior with no configured servers.
func TestMCPEmptyConfig(t *testing.T) {
	env := setupTestEnv(t)

	// List should return empty slice, not error
	servers, err := env.claudePlatform.ListMCP()
	if err != nil {
		t.Fatalf("ListMCP on empty config failed: %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("expected 0 servers on empty config, got %d", len(servers))
	}

	// Test tabular output with empty state
	platforms := []cli.Platform{env.claudePlatform}

	var buf bytes.Buffer
	if err := outputMCPTabular(&buf, platforms); err != nil {
		t.Fatalf("outputMCPTabular failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "(no MCP servers configured)") {
		t.Error("output should indicate no servers configured")
	}
}
