package claude

import (
	"encoding/json"
	"testing"

	"github.com/thoreinstein/aix/internal/mcp"
)

func TestMCPTranslator_Platform(t *testing.T) {
	translator := NewMCPTranslator()
	if got := translator.Platform(); got != "claude" {
		t.Errorf("Platform() = %q, want %q", got, "claude")
	}
}

func TestMCPTranslator_ToCanonical(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		wantCount  int
		checkFirst func(t *testing.T, cfg *mcp.Config)
	}{
		{
			name: "valid config with mcpServers wrapper",
			input: `{
				"mcpServers": {
					"github": {
						"command": "npx",
						"args": ["-y", "@modelcontextprotocol/server-github"],
						"env": {"GITHUB_TOKEN": "token123"}
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["github"]
				if server == nil {
					t.Fatal("github server not found")
				}
				if server.Command != "npx" {
					t.Errorf("Command = %q, want %q", server.Command, "npx")
				}
				if len(server.Args) != 2 {
					t.Errorf("len(Args) = %d, want 2", len(server.Args))
				}
				if server.Env["GITHUB_TOKEN"] != "token123" {
					t.Errorf("Env[GITHUB_TOKEN] = %q, want %q", server.Env["GITHUB_TOKEN"], "token123")
				}
			},
		},
		{
			name: "valid config with sse transport",
			input: `{
				"mcpServers": {
					"remote": {
						"url": "https://api.example.com/mcp",
						"transport": "sse",
						"headers": {"Authorization": "Bearer token"}
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["remote"]
				if server == nil {
					t.Fatal("remote server not found")
				}
				if server.URL != "https://api.example.com/mcp" {
					t.Errorf("URL = %q, want %q", server.URL, "https://api.example.com/mcp")
				}
				if server.Transport != "sse" {
					t.Errorf("Transport = %q, want %q", server.Transport, "sse")
				}
				if server.Headers["Authorization"] != "Bearer token" {
					t.Errorf("Headers[Authorization] = %q, want %q", server.Headers["Authorization"], "Bearer token")
				}
			},
		},
		{
			name: "config with platforms and disabled",
			input: `{
				"mcpServers": {
					"darwin-only": {
						"command": "mac-cmd",
						"platforms": ["darwin"],
						"disabled": true
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["darwin-only"]
				if server == nil {
					t.Fatal("darwin-only server not found")
				}
				if len(server.Platforms) != 1 || server.Platforms[0] != "darwin" {
					t.Errorf("Platforms = %v, want [darwin]", server.Platforms)
				}
				if !server.Disabled {
					t.Error("Disabled = false, want true")
				}
			},
		},
		{
			name: "bare servers map (no mcpServers wrapper)",
			input: `{
				"test-server": {
					"command": "test-cmd"
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["test-server"]
				if server == nil {
					t.Fatal("test-server not found")
				}
				if server.Command != "test-cmd" {
					t.Errorf("Command = %q, want %q", server.Command, "test-cmd")
				}
			},
		},
		{
			name:      "empty config",
			input:     `{"mcpServers": {}}`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewMCPTranslator()
			cfg, err := translator.ToCanonical([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Error("ToCanonical() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("ToCanonical() error = %v", err)
			}

			if len(cfg.Servers) != tt.wantCount {
				t.Errorf("len(Servers) = %d, want %d", len(cfg.Servers), tt.wantCount)
			}

			if tt.checkFirst != nil {
				tt.checkFirst(t, cfg)
			}
		})
	}
}

func TestMCPTranslator_FromCanonical(t *testing.T) {
	tests := []struct {
		name    string
		config  *mcp.Config
		wantErr bool
		check   func(t *testing.T, data []byte)
	}{
		{
			name: "valid stdio server",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:    "test",
						Command: "npx",
						Args:    []string{"-y", "server"},
						Env:     map[string]string{"KEY": "value"},
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				t.Helper()
				var result map[string]any
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				servers, ok := result["mcpServers"].(map[string]any)
				if !ok {
					t.Fatal("mcpServers not found or not a map")
				}

				server, ok := servers["test"].(map[string]any)
				if !ok {
					t.Fatal("test server not found")
				}

				if server["command"] != "npx" {
					t.Errorf("command = %v, want %q", server["command"], "npx")
				}
			},
		},
		{
			name: "valid sse server",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"remote": {
						Name:      "remote",
						URL:       "https://api.example.com",
						Transport: "sse",
						Headers:   map[string]string{"Auth": "token"},
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				t.Helper()
				var result map[string]any
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				servers := result["mcpServers"].(map[string]any)
				server := servers["remote"].(map[string]any)

				if server["url"] != "https://api.example.com" {
					t.Errorf("url = %v, want %q", server["url"], "https://api.example.com")
				}
				if server["transport"] != "sse" {
					t.Errorf("transport = %v, want %q", server["transport"], "sse")
				}
			},
		},
		{
			name: "server with platforms and disabled",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"limited": {
						Name:      "limited",
						Command:   "cmd",
						Platforms: []string{"darwin", "linux"},
						Disabled:  true,
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				t.Helper()
				var result map[string]any
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				servers := result["mcpServers"].(map[string]any)
				server := servers["limited"].(map[string]any)

				platforms, ok := server["platforms"].([]any)
				if !ok {
					t.Fatal("platforms not found")
				}
				if len(platforms) != 2 {
					t.Errorf("len(platforms) = %d, want 2", len(platforms))
				}

				if server["disabled"] != true {
					t.Errorf("disabled = %v, want true", server["disabled"])
				}
			},
		},
		{
			name:    "nil config returns error",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty servers",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{},
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				t.Helper()
				var result map[string]any
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				servers, ok := result["mcpServers"].(map[string]any)
				if !ok {
					t.Fatal("mcpServers not found")
				}
				if len(servers) != 0 {
					t.Errorf("len(servers) = %d, want 0", len(servers))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := NewMCPTranslator()
			data, err := translator.FromCanonical(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("FromCanonical() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("FromCanonical() error = %v", err)
			}

			if tt.check != nil {
				tt.check(t, data)
			}
		})
	}
}

func TestMCPTranslator_RoundTrip_CanonicalToClaude(t *testing.T) {
	translator := NewMCPTranslator()

	original := &mcp.Config{
		Servers: map[string]*mcp.Server{
			"full-server": {
				Name:      "full-server",
				Command:   "npx",
				Args:      []string{"-y", "@mcp/server"},
				Transport: "stdio",
				Env:       map[string]string{"TOKEN": "secret"},
				Platforms: []string{"darwin", "linux"},
				Disabled:  false,
			},
			"remote-server": {
				Name:      "remote-server",
				URL:       "https://api.example.com/mcp",
				Transport: "sse",
				Headers:   map[string]string{"Authorization": "Bearer token"},
				Disabled:  true,
			},
		},
	}

	// canonical → claude
	claudeData, err := translator.FromCanonical(original)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	// claude → canonical
	result, err := translator.ToCanonical(claudeData)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// Verify all fields preserved
	if len(result.Servers) != len(original.Servers) {
		t.Fatalf("len(Servers) = %d, want %d", len(result.Servers), len(original.Servers))
	}

	for name, origServer := range original.Servers {
		resultServer := result.Servers[name]
		if resultServer == nil {
			t.Errorf("server %q not found in result", name)
			continue
		}

		if resultServer.Command != origServer.Command {
			t.Errorf("%s.Command = %q, want %q", name, resultServer.Command, origServer.Command)
		}
		if len(resultServer.Args) != len(origServer.Args) {
			t.Errorf("%s.len(Args) = %d, want %d", name, len(resultServer.Args), len(origServer.Args))
		}
		if resultServer.URL != origServer.URL {
			t.Errorf("%s.URL = %q, want %q", name, resultServer.URL, origServer.URL)
		}
		if resultServer.Transport != origServer.Transport {
			t.Errorf("%s.Transport = %q, want %q", name, resultServer.Transport, origServer.Transport)
		}
		if resultServer.Disabled != origServer.Disabled {
			t.Errorf("%s.Disabled = %v, want %v", name, resultServer.Disabled, origServer.Disabled)
		}
		if len(resultServer.Platforms) != len(origServer.Platforms) {
			t.Errorf("%s.len(Platforms) = %d, want %d", name, len(resultServer.Platforms), len(origServer.Platforms))
		}
	}
}

func TestMCPTranslator_RoundTrip_ClaudeToCanonical(t *testing.T) {
	translator := NewMCPTranslator()

	originalJSON := `{
		"mcpServers": {
			"github": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-github"],
				"env": {"GITHUB_TOKEN": "token123"},
				"transport": "stdio",
				"platforms": ["darwin", "linux", "windows"]
			},
			"api-server": {
				"url": "https://api.example.com/mcp",
				"transport": "sse",
				"headers": {"Authorization": "Bearer secret"},
				"disabled": true
			}
		}
	}`

	// claude → canonical
	canonical, err := translator.ToCanonical([]byte(originalJSON))
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// canonical → claude
	claudeData, err := translator.FromCanonical(canonical)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	// claude → canonical (again)
	result, err := translator.ToCanonical(claudeData)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// Verify fields preserved through round-trip
	if len(result.Servers) != 2 {
		t.Fatalf("len(Servers) = %d, want 2", len(result.Servers))
	}

	// Check github server
	github := result.Servers["github"]
	if github == nil {
		t.Fatal("github server not found")
	}
	if github.Command != "npx" {
		t.Errorf("github.Command = %q, want %q", github.Command, "npx")
	}
	if len(github.Args) != 2 {
		t.Errorf("github.len(Args) = %d, want 2", len(github.Args))
	}
	if github.Env["GITHUB_TOKEN"] != "token123" {
		t.Errorf("github.Env[GITHUB_TOKEN] = %q, want %q", github.Env["GITHUB_TOKEN"], "token123")
	}
	if len(github.Platforms) != 3 {
		t.Errorf("github.len(Platforms) = %d, want 3", len(github.Platforms))
	}

	// Check api-server
	apiServer := result.Servers["api-server"]
	if apiServer == nil {
		t.Fatal("api-server not found")
	}
	if apiServer.URL != "https://api.example.com/mcp" {
		t.Errorf("api-server.URL = %q, want %q", apiServer.URL, "https://api.example.com/mcp")
	}
	if apiServer.Transport != "sse" {
		t.Errorf("api-server.Transport = %q, want %q", apiServer.Transport, "sse")
	}
	if !apiServer.Disabled {
		t.Error("api-server.Disabled = false, want true")
	}
}

func TestMCPTranslator_OutputFormat(t *testing.T) {
	translator := NewMCPTranslator()

	config := &mcp.Config{
		Servers: map[string]*mcp.Server{
			"test": {
				Name:    "test",
				Command: "cmd",
			},
		},
	}

	data, err := translator.FromCanonical(config)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	// Verify JSON is formatted with 2-space indentation
	content := string(data)
	if len(content) < 3 || content[0:2] != "{\n" {
		t.Error("JSON should start with {\\n")
	}
	// Check for 2-space indentation
	if len(content) > 4 && content[2:4] != "  " {
		t.Errorf("JSON should use 2-space indentation, got first indent: %q", content[2:4])
	}
}
