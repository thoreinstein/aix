package opencode

import (
	"encoding/json"
	"testing"

	"github.com/thoreinstein/aix/internal/mcp"
)

func TestMCPTranslator_Platform(t *testing.T) {
	translator := NewMCPTranslator()
	if got := translator.Platform(); got != "opencode" {
		t.Errorf("Platform() = %q, want %q", got, "opencode")
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
			name: "valid config with mcp wrapper",
			input: `{
				"mcp": {
					"github": {
						"command": ["npx", "-y", "@modelcontextprotocol/server-github"],
						"type": "local",
						"environment": {"GITHUB_TOKEN": "token123"}
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
				// Command[0] → Command
				if server.Command != "npx" {
					t.Errorf("Command = %q, want %q", server.Command, "npx")
				}
				// Command[1:] → Args
				if len(server.Args) != 2 {
					t.Errorf("len(Args) = %d, want 2", len(server.Args))
				} else if server.Args[0] != "-y" {
					t.Errorf("Args[0] = %q, want %q", server.Args[0], "-y")
				}
				// Environment → Env
				if server.Env["GITHUB_TOKEN"] != "token123" {
					t.Errorf("Env[GITHUB_TOKEN] = %q, want %q", server.Env["GITHUB_TOKEN"], "token123")
				}
				// Type "local" → Transport "stdio"
				if server.Transport != mcp.TransportStdio {
					t.Errorf("Transport = %q, want %q", server.Transport, mcp.TransportStdio)
				}
			},
		},
		{
			name: "type local maps to transport stdio",
			input: `{
				"mcp": {
					"local-server": {
						"command": ["cmd"],
						"type": "local"
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["local-server"]
				if server == nil {
					t.Fatal("local-server not found")
				}
				if server.Transport != mcp.TransportStdio {
					t.Errorf("Transport = %q, want %q", server.Transport, mcp.TransportStdio)
				}
			},
		},
		{
			name: "type remote maps to transport sse",
			input: `{
				"mcp": {
					"remote-server": {
						"url": "https://api.example.com/mcp",
						"type": "remote",
						"headers": {"Authorization": "Bearer token"}
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["remote-server"]
				if server == nil {
					t.Fatal("remote-server not found")
				}
				if server.URL != "https://api.example.com/mcp" {
					t.Errorf("URL = %q, want %q", server.URL, "https://api.example.com/mcp")
				}
				if server.Transport != mcp.TransportSSE {
					t.Errorf("Transport = %q, want %q", server.Transport, mcp.TransportSSE)
				}
				if server.Headers["Authorization"] != "Bearer token" {
					t.Errorf("Headers[Authorization] = %q, want %q", server.Headers["Authorization"], "Bearer token")
				}
			},
		},
		{
			name: "environment maps to env",
			input: `{
				"mcp": {
					"env-server": {
						"command": ["cmd"],
						"environment": {"KEY1": "val1", "KEY2": "val2"}
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["env-server"]
				if server == nil {
					t.Fatal("env-server not found")
				}
				if server.Env["KEY1"] != "val1" {
					t.Errorf("Env[KEY1] = %q, want %q", server.Env["KEY1"], "val1")
				}
				if server.Env["KEY2"] != "val2" {
					t.Errorf("Env[KEY2] = %q, want %q", server.Env["KEY2"], "val2")
				}
			},
		},
		{
			name: "infers transport from url when type missing",
			input: `{
				"mcp": {
					"inferred-remote": {
						"url": "https://example.com"
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["inferred-remote"]
				if server == nil {
					t.Fatal("inferred-remote not found")
				}
				if server.Transport != mcp.TransportSSE {
					t.Errorf("Transport = %q, want %q (inferred from URL)", server.Transport, mcp.TransportSSE)
				}
			},
		},
		{
			name: "infers transport from command when type missing",
			input: `{
				"mcp": {
					"inferred-local": {
						"command": ["run-server"]
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["inferred-local"]
				if server == nil {
					t.Fatal("inferred-local not found")
				}
				if server.Transport != mcp.TransportStdio {
					t.Errorf("Transport = %q, want %q (inferred from Command)", server.Transport, mcp.TransportStdio)
				}
			},
		},
		{
			name: "bare servers map (no mcp wrapper)",
			input: `{
				"test-server": {
					"command": ["test-cmd"]
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
			name: "disabled server",
			input: `{
				"mcp": {
					"disabled-server": {
						"command": ["cmd"],
						"disabled": true
					}
				}
			}`,
			wantErr:   false,
			wantCount: 1,
			checkFirst: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["disabled-server"]
				if server == nil {
					t.Fatal("disabled-server not found")
				}
				if !server.Disabled {
					t.Error("Disabled = false, want true")
				}
			},
		},
		{
			name:      "empty config",
			input:     `{"mcp": {}}`,
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
			name: "stdio server - command and args joined",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Command:   "npx",
						Args:      []string{"-y", "server"},
						Transport: mcp.TransportStdio,
						Env:       map[string]string{"KEY": "value"},
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

				servers, ok := result["mcp"].(map[string]any)
				if !ok {
					t.Fatal("mcp not found or not a map")
				}

				server, ok := servers["test"].(map[string]any)
				if !ok {
					t.Fatal("test server not found")
				}

				// Command + Args → Command []string
				cmd, ok := server["command"].([]any)
				if !ok {
					t.Fatal("command not found or not an array")
				}
				if len(cmd) != 3 {
					t.Errorf("len(command) = %d, want 3", len(cmd))
				} else {
					if cmd[0] != "npx" {
						t.Errorf("command[0] = %v, want %q", cmd[0], "npx")
					}
					if cmd[1] != "-y" {
						t.Errorf("command[1] = %v, want %q", cmd[1], "-y")
					}
					if cmd[2] != "server" {
						t.Errorf("command[2] = %v, want %q", cmd[2], "server")
					}
				}

				// Transport "stdio" → Type "local"
				if server["type"] != "local" {
					t.Errorf("type = %v, want %q", server["type"], "local")
				}

				// Env → Environment
				env, ok := server["environment"].(map[string]any)
				if !ok {
					t.Fatal("environment not found")
				}
				if env["KEY"] != "value" {
					t.Errorf("environment[KEY] = %v, want %q", env["KEY"], "value")
				}
			},
		},
		{
			name: "sse server - transport maps to type remote",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"remote": {
						Name:      "remote",
						URL:       "https://api.example.com",
						Transport: mcp.TransportSSE,
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

				servers := result["mcp"].(map[string]any)
				server := servers["remote"].(map[string]any)

				if server["url"] != "https://api.example.com" {
					t.Errorf("url = %v, want %q", server["url"], "https://api.example.com")
				}
				if server["type"] != "remote" {
					t.Errorf("type = %v, want %q", server["type"], "remote")
				}
			},
		},
		{
			name: "platforms field is NOT in output (lossy)",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"limited": {
						Name:      "limited",
						Command:   "cmd",
						Transport: mcp.TransportStdio,
						Platforms: []string{"darwin", "linux"}, // Should be lost
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

				servers := result["mcp"].(map[string]any)
				server := servers["limited"].(map[string]any)

				// Platforms should NOT be present in output
				if _, exists := server["platforms"]; exists {
					t.Error("platforms field should not be present in OpenCode output")
				}
			},
		},
		{
			name: "command without args",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"simple": {
						Name:      "simple",
						Command:   "simple-cmd",
						Transport: mcp.TransportStdio,
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

				servers := result["mcp"].(map[string]any)
				server := servers["simple"].(map[string]any)

				cmd, ok := server["command"].([]any)
				if !ok {
					t.Fatal("command not found or not an array")
				}
				if len(cmd) != 1 {
					t.Errorf("len(command) = %d, want 1", len(cmd))
				}
				if cmd[0] != "simple-cmd" {
					t.Errorf("command[0] = %v, want %q", cmd[0], "simple-cmd")
				}
			},
		},
		{
			name: "disabled server",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"disabled": {
						Name:      "disabled",
						Command:   "cmd",
						Transport: mcp.TransportStdio,
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

				servers := result["mcp"].(map[string]any)
				server := servers["disabled"].(map[string]any)

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

				servers, ok := result["mcp"].(map[string]any)
				if !ok {
					t.Fatal("mcp not found")
				}
				if len(servers) != 0 {
					t.Errorf("len(servers) = %d, want 0", len(servers))
				}
			},
		},
		{
			name: "infers type from url when transport empty",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"inferred": {
						Name: "inferred",
						URL:  "https://example.com",
						// Transport is empty
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

				servers := result["mcp"].(map[string]any)
				server := servers["inferred"].(map[string]any)

				if server["type"] != "remote" {
					t.Errorf("type = %v, want %q (inferred from URL)", server["type"], "remote")
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

func TestMCPTranslator_RoundTrip_CanonicalToOpenCode(t *testing.T) {
	translator := NewMCPTranslator()

	original := &mcp.Config{
		Servers: map[string]*mcp.Server{
			"local-server": {
				Name:      "local-server",
				Command:   "npx",
				Args:      []string{"-y", "@mcp/server"},
				Transport: mcp.TransportStdio,
				Env:       map[string]string{"TOKEN": "secret"},
				// NOTE: Platforms intentionally omitted - it would be lost
				Disabled: false,
			},
			"remote-server": {
				Name:      "remote-server",
				URL:       "https://api.example.com/mcp",
				Transport: mcp.TransportSSE,
				Headers:   map[string]string{"Authorization": "Bearer token"},
				Disabled:  true,
			},
		},
	}

	// canonical → opencode
	openData, err := translator.FromCanonical(original)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	// opencode → canonical
	result, err := translator.ToCanonical(openData)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// Verify all fields preserved (except Platforms which is lossy)
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
	}
}

func TestMCPTranslator_RoundTrip_OpenCodeToCanonical(t *testing.T) {
	translator := NewMCPTranslator()

	originalJSON := `{
		"mcp": {
			"github": {
				"command": ["npx", "-y", "@modelcontextprotocol/server-github"],
				"type": "local",
				"environment": {"GITHUB_TOKEN": "token123"}
			},
			"api-server": {
				"url": "https://api.example.com/mcp",
				"type": "remote",
				"headers": {"Authorization": "Bearer secret"},
				"disabled": true
			}
		}
	}`

	// opencode → canonical
	canonical, err := translator.ToCanonical([]byte(originalJSON))
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// canonical → opencode
	openData, err := translator.FromCanonical(canonical)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	// opencode → canonical (again)
	result, err := translator.ToCanonical(openData)
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
	if github.Transport != mcp.TransportStdio {
		t.Errorf("github.Transport = %q, want %q", github.Transport, mcp.TransportStdio)
	}

	// Check api-server
	apiServer := result.Servers["api-server"]
	if apiServer == nil {
		t.Fatal("api-server not found")
	}
	if apiServer.URL != "https://api.example.com/mcp" {
		t.Errorf("api-server.URL = %q, want %q", apiServer.URL, "https://api.example.com/mcp")
	}
	if apiServer.Transport != mcp.TransportSSE {
		t.Errorf("api-server.Transport = %q, want %q", apiServer.Transport, mcp.TransportSSE)
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
				Name:      "test",
				Command:   "cmd",
				Transport: mcp.TransportStdio,
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

func TestMCPTranslator_PlatformsFieldLossy(t *testing.T) {
	translator := NewMCPTranslator()

	// Create a canonical config WITH platforms
	original := &mcp.Config{
		Servers: map[string]*mcp.Server{
			"darwin-only": {
				Name:      "darwin-only",
				Command:   "mac-cmd",
				Transport: mcp.TransportStdio,
				Platforms: []string{"darwin"}, // This will be lost
			},
		},
	}

	// canonical → opencode
	openData, err := translator.FromCanonical(original)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	// Verify platforms is not in the JSON
	var result map[string]any
	if err := json.Unmarshal(openData, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	servers := result["mcp"].(map[string]any)
	server := servers["darwin-only"].(map[string]any)

	if _, exists := server["platforms"]; exists {
		t.Error("platforms field should NOT exist in OpenCode output")
	}

	// opencode → canonical
	roundTripped, err := translator.ToCanonical(openData)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// Verify platforms is now empty (lost)
	resultServer := roundTripped.Servers["darwin-only"]
	if len(resultServer.Platforms) != 0 {
		t.Errorf("Platforms should be lost after round-trip, got: %v", resultServer.Platforms)
	}
}
