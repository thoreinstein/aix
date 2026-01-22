package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
)

// mcpListMockPlatform extends mockPlatform with MCP list-specific behavior.
type mcpListMockPlatform struct {
	mockPlatform
	mcpServers []cli.MCPInfo
	mcpErr     error
}

func (m *mcpListMockPlatform) ListMCP() ([]cli.MCPInfo, error) {
	if m.mcpErr != nil {
		return nil, m.mcpErr
	}
	return m.mcpServers, nil
}

// Note: MaskSecrets unit tests are in internal/doctor/redact_test.go.
// The integration tests below verify the command behavior including masking.

func TestMCPListCommand_Metadata(t *testing.T) {
	if mcpListCmd.Use != "list" {
		t.Errorf("Use = %q, want %q", mcpListCmd.Use, "list")
	}

	if mcpListCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Check flags exist
	if mcpListCmd.Flags().Lookup("json") == nil {
		t.Error("--json flag should be defined")
	}
	if mcpListCmd.Flags().Lookup("show-secrets") == nil {
		t.Error("--show-secrets flag should be defined")
	}
}

func TestOutputMCPTabular_EmptyState(t *testing.T) {
	platforms := []cli.Platform{
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			mcpServers: []cli.MCPInfo{},
		},
	}

	var buf bytes.Buffer
	err := outputMCPTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputMCPTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should contain platform name")
	}
	if !strings.Contains(output, "(no MCP servers configured)") {
		t.Error("output should indicate no servers configured")
	}
}

func TestOutputMCPTabular_WithServers(t *testing.T) {
	platforms := []cli.Platform{
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			mcpServers: []cli.MCPInfo{
				{
					Name:      "github",
					Transport: "stdio",
					Command:   "npx",
					Disabled:  false,
				},
				{
					Name:      "api-gw",
					Transport: "sse",
					URL:       "https://api.example.com/mcp",
					Disabled:  false,
				},
				{
					Name:      "disabled-server",
					Transport: "stdio",
					Command:   "/usr/bin/disabled-server",
					Disabled:  true,
				},
			},
		},
	}

	var buf bytes.Buffer
	err := outputMCPTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputMCPTabular() error = %v", err)
	}

	output := buf.String()

	// Check headers
	if !strings.Contains(output, "NAME") {
		t.Error("output should contain NAME header")
	}
	if !strings.Contains(output, "TRANSPORT") {
		t.Error("output should contain TRANSPORT header")
	}
	if !strings.Contains(output, "COMMAND/URL") {
		t.Error("output should contain COMMAND/URL header")
	}
	if !strings.Contains(output, "STATUS") {
		t.Error("output should contain STATUS header")
	}

	// Check servers
	if !strings.Contains(output, "github") {
		t.Error("output should contain github server")
	}
	if !strings.Contains(output, "api-gw") {
		t.Error("output should contain api-gw server")
	}
	if !strings.Contains(output, "disabled-server") {
		t.Error("output should contain disabled-server")
	}
	if !strings.Contains(output, "enabled") {
		t.Error("output should contain enabled status")
	}
	if !strings.Contains(output, "disabled") {
		t.Error("output should contain disabled status")
	}
}

func TestOutputMCPTabular_MultiplePlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			mcpServers: []cli.MCPInfo{
				{Name: "github", Transport: "stdio", Command: "npx"},
			},
		},
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
			},
			mcpServers: []cli.MCPInfo{
				{Name: "github", Transport: "stdio", Command: "npx"},
			},
		},
	}

	var buf bytes.Buffer
	err := outputMCPTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputMCPTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should contain Claude Code")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Error("output should contain OpenCode")
	}
}

func TestOutputMCPTabular_NoServersAcrossPlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			mcpServers: []cli.MCPInfo{},
		},
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
			},
			mcpServers: []cli.MCPInfo{},
		},
	}

	var buf bytes.Buffer
	err := outputMCPTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputMCPTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No MCP servers configured") {
		t.Error("output should indicate no servers configured across all platforms")
	}
}

func TestOutputMCPJSON(t *testing.T) {
	// Save and restore global flag
	oldShowSecrets := mcpListShowSecrets
	defer func() { mcpListShowSecrets = oldShowSecrets }()
	mcpListShowSecrets = false

	platforms := []cli.Platform{
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			mcpServers: []cli.MCPInfo{
				{
					Name:      "github",
					Transport: "stdio",
					Command:   "npx",
					Disabled:  false,
					Env: map[string]string{
						"GITHUB_TOKEN": "ghp_xxxxxxxxxxxx1234",
						"DEBUG":        "true",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := outputMCPJSON(&buf, platforms)
	if err != nil {
		t.Fatalf("outputMCPJSON() error = %v", err)
	}

	var result []mcpListPlatformOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(result))
	}
	if result[0].Platform != "claude" {
		t.Errorf("platform = %q, want %q", result[0].Platform, "claude")
	}
	if len(result[0].Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(result[0].Servers))
	}

	server := result[0].Servers[0]
	if server.Name != "github" {
		t.Errorf("server.Name = %q, want %q", server.Name, "github")
	}
	if server.Transport != "stdio" {
		t.Errorf("server.Transport = %q, want %q", server.Transport, "stdio")
	}
	if server.Command != "npx" {
		t.Errorf("server.Command = %q, want %q", server.Command, "npx")
	}
	if server.Disabled {
		t.Error("server.Disabled should be false")
	}

	// Check secret masking
	if server.Env["GITHUB_TOKEN"] != "****1234" {
		t.Errorf("GITHUB_TOKEN should be masked, got %q", server.Env["GITHUB_TOKEN"])
	}
	if server.Env["DEBUG"] != "true" {
		t.Errorf("DEBUG should not be masked, got %q", server.Env["DEBUG"])
	}
}

func TestOutputMCPJSON_ShowSecrets(t *testing.T) {
	// Save and restore global flag
	oldShowSecrets := mcpListShowSecrets
	defer func() { mcpListShowSecrets = oldShowSecrets }()
	mcpListShowSecrets = true

	platforms := []cli.Platform{
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			mcpServers: []cli.MCPInfo{
				{
					Name:      "github",
					Transport: "stdio",
					Command:   "npx",
					Env: map[string]string{
						"GITHUB_TOKEN": "ghp_xxxxxxxxxxxx1234",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := outputMCPJSON(&buf, platforms)
	if err != nil {
		t.Fatalf("outputMCPJSON() error = %v", err)
	}

	var result []mcpListPlatformOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Check secret is NOT masked
	if result[0].Servers[0].Env["GITHUB_TOKEN"] != "ghp_xxxxxxxxxxxx1234" {
		t.Errorf("GITHUB_TOKEN should not be masked with --show-secrets, got %q",
			result[0].Servers[0].Env["GITHUB_TOKEN"])
	}
}

func TestOutputMCPJSON_EmptyServers(t *testing.T) {
	platforms := []cli.Platform{
		&mcpListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			mcpServers: []cli.MCPInfo{},
		},
	}

	var buf bytes.Buffer
	err := outputMCPJSON(&buf, platforms)
	if err != nil {
		t.Fatalf("outputMCPJSON() error = %v", err)
	}

	var result []mcpListPlatformOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(result))
	}
	if len(result[0].Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(result[0].Servers))
	}
}
