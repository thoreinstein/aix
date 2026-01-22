package opencode

import (
	"encoding/json"
	"reflect"
	"testing"
)

// ptrBool returns a pointer to the given bool value.
func ptrBool(b bool) *bool {
	return &b
}

func TestMCPConfig_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		config  *MCPConfig
		wantErr bool
	}{
		{
			name: "empty config",
			config: &MCPConfig{
				MCP: map[string]*MCPServer{},
			},
			wantErr: false,
		},
		{
			name: "single server",
			config: &MCPConfig{
				MCP: map[string]*MCPServer{
					"test": {
						Name:    "test",
						Command: []string{"test-cmd"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "server with all fields",
			config: &MCPConfig{
				MCP: map[string]*MCPServer{
					"full": {
						Name:    "full",
						Command: []string{"npx", "-y", "server"},
						Type:    "local",
						URL:     "http://localhost:8080",
						Environment: map[string]string{
							"KEY": "value",
						},
						Headers: map[string]string{
							"Authorization": "Bearer token",
						},
						Enabled: ptrBool(false),
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			// Verify it's valid JSON
			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("output is not valid JSON: %v", err)
			}

			// Verify mcp key exists
			if _, ok := result["mcp"]; !ok {
				t.Error("output missing 'mcp' key")
			}
		})
	}
}

func TestMCPConfig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    bool
		checkExtra bool // whether to check for unknown fields
		check      func(t *testing.T, cfg *MCPConfig)
	}{
		{
			name:    "empty mcp",
			input:   `{"mcp": {}}`,
			wantErr: false,
			check: func(t *testing.T, cfg *MCPConfig) {
				if len(cfg.MCP) != 0 {
					t.Errorf("MCP count = %d, want 0", len(cfg.MCP))
				}
			},
		},
		{
			name:    "single server",
			input:   `{"mcp": {"test": {"command": ["test-cmd"]}}}`,
			wantErr: false,
			check: func(t *testing.T, cfg *MCPConfig) {
				if len(cfg.MCP) != 1 {
					t.Errorf("MCP count = %d, want 1", len(cfg.MCP))
				}
				server := cfg.MCP["test"]
				if server == nil {
					t.Fatal("test server not found")
				}
				// Name is not populated from JSON (json:"-"), it's populated from map key elsewhere
				if len(server.Command) != 1 || server.Command[0] != "test-cmd" {
					t.Errorf("Command = %v, want [test-cmd]", server.Command)
				}
			},
		},
		{
			name:    "server with all fields",
			input:   `{"mcp": {"full": {"command": ["cmd", "arg"], "type": "local", "url": "http://localhost", "environment": {"KEY": "val"}, "headers": {"Auth": "token"}, "enabled": false}}}`,
			wantErr: false,
			check: func(t *testing.T, cfg *MCPConfig) {
				server := cfg.MCP["full"]
				if server == nil {
					t.Fatal("full server not found")
				}
				if len(server.Command) != 2 {
					t.Errorf("len(Command) = %d, want 2", len(server.Command))
				}
				if server.Type != "local" {
					t.Errorf("Type = %q, want %q", server.Type, "local")
				}
				if server.URL != "http://localhost" {
					t.Errorf("URL = %q, want %q", server.URL, "http://localhost")
				}
				if server.Environment["KEY"] != "val" {
					t.Errorf("Environment[KEY] = %q, want %q", server.Environment["KEY"], "val")
				}
				if server.Headers["Auth"] != "token" {
					t.Errorf("Headers[Auth] = %q, want %q", server.Headers["Auth"], "token")
				}
				if server.Enabled == nil || *server.Enabled != false {
					t.Errorf("Enabled = %v, want false", server.Enabled)
				}
			},
		},
		{
			name:       "preserves unknown fields",
			input:      `{"mcp": {"test": {}}, "futureField": "value"}`,
			wantErr:    false,
			checkExtra: true,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config MCPConfig
			err := json.Unmarshal([]byte(tt.input), &config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if tt.check != nil {
				tt.check(t, &config)
			}

			if tt.checkExtra && config.unknownFields == nil {
				t.Error("unknownFields should be populated for unknown fields")
			}
		})
	}
}

func TestMCPConfig_PreservesUnknownFields(t *testing.T) {
	// JSON with unknown fields at root level
	input := `{
		"mcp": {
			"test": {
				"name": "test",
				"command": ["test-cmd"]
			}
		},
		"futureField": "future value",
		"anotherField": {
			"nested": true
		}
	}`

	// Unmarshal
	var config MCPConfig
	if err := json.Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify known fields parsed correctly
	if len(config.MCP) != 1 {
		t.Errorf("MCP count = %d, want 1", len(config.MCP))
	}
	if config.MCP["test"] == nil {
		t.Fatal("MCP[test] is nil")
	}
	if len(config.MCP["test"].Command) != 1 || config.MCP["test"].Command[0] != "test-cmd" {
		t.Errorf("MCP[test].Command = %v, want [test-cmd]", config.MCP["test"].Command)
	}

	// Marshal back
	data, err := json.Marshal(&config)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal to generic map to verify unknown fields preserved
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal to map error = %v", err)
	}

	// Check unknown fields are present
	if _, ok := result["futureField"]; !ok {
		t.Error("futureField not preserved in output")
	}
	if _, ok := result["anotherField"]; !ok {
		t.Error("anotherField not preserved in output")
	}

	// Verify futureField value
	if result["futureField"] != "future value" {
		t.Errorf("futureField = %v, want %q", result["futureField"], "future value")
	}

	// Verify nested unknown field
	nested, ok := result["anotherField"].(map[string]any)
	if !ok {
		t.Fatalf("anotherField is not a map: %T", result["anotherField"])
	}
	if nested["nested"] != true {
		t.Errorf("anotherField.nested = %v, want true", nested["nested"])
	}
}

func TestMCPConfig_RoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		config *MCPConfig
	}{
		{
			name: "empty config",
			config: &MCPConfig{
				MCP: map[string]*MCPServer{},
			},
		},
		{
			name: "single server",
			config: &MCPConfig{
				MCP: map[string]*MCPServer{
					"test": {
						Name:    "test",
						Command: []string{"test-cmd"},
					},
				},
			},
		},
		{
			name: "multiple servers",
			config: &MCPConfig{
				MCP: map[string]*MCPServer{
					"github": {
						Name:    "github",
						Command: []string{"npx", "-y", "@modelcontextprotocol/server-github"},
						Type:    "local",
						Environment: map[string]string{
							"GITHUB_TOKEN": "${GITHUB_TOKEN}",
						},
					},
					"remote": {
						Name:    "remote",
						Type:    "remote",
						URL:     "https://api.example.com/mcp",
						Headers: map[string]string{"Authorization": "Bearer ${API_KEY}"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Unmarshal back
			var got MCPConfig
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Compare MCP servers by checking individual fields
			// (can't use reflect.DeepEqual due to pointer comparison issues)
			if len(got.MCP) != len(tt.config.MCP) {
				t.Errorf("MCP server count mismatch: got %d, want %d", len(got.MCP), len(tt.config.MCP))
				return
			}

			for name, want := range tt.config.MCP {
				gotServer, ok := got.MCP[name]
				if !ok {
					t.Errorf("missing server %q", name)
					continue
				}
				// Name is set from map key during unmarshal
				if gotServer.Name != name {
					t.Errorf("server %q: Name = %q, want %q", name, gotServer.Name, name)
				}
				if !reflect.DeepEqual(gotServer.Command, want.Command) {
					t.Errorf("server %q: Command = %v, want %v", name, gotServer.Command, want.Command)
				}
				if gotServer.Type != want.Type {
					t.Errorf("server %q: Type = %q, want %q", name, gotServer.Type, want.Type)
				}
				if gotServer.URL != want.URL {
					t.Errorf("server %q: URL = %q, want %q", name, gotServer.URL, want.URL)
				}
				if !reflect.DeepEqual(gotServer.Environment, want.Environment) {
					t.Errorf("server %q: Environment = %v, want %v", name, gotServer.Environment, want.Environment)
				}
				if !reflect.DeepEqual(gotServer.Headers, want.Headers) {
					t.Errorf("server %q: Headers = %v, want %v", name, gotServer.Headers, want.Headers)
				}
			}
		})
	}
}

func TestMCPConfig_RoundTripWithUnknownFields(t *testing.T) {
	// Start with JSON that has unknown fields
	input := `{
		"mcp": {"test": {"name": "test", "command": ["cmd"]}},
		"version": "2.0",
		"experimental": {"feature": true}
	}`

	// First unmarshal
	var config MCPConfig
	if err := json.Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("first Unmarshal() error = %v", err)
	}

	// Modify the known field
	config.MCP["new"] = &MCPServer{
		Name:    "new",
		Command: []string{"new-cmd"},
	}

	// Marshal
	data, err := json.Marshal(&config)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal again
	var result MCPConfig
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("second Unmarshal() error = %v", err)
	}

	// Verify both servers exist
	if _, ok := result.MCP["test"]; !ok {
		t.Error("original server 'test' missing after round-trip")
	}
	if _, ok := result.MCP["new"]; !ok {
		t.Error("new server 'new' missing after round-trip")
	}

	// Verify unknown fields still exist by marshaling to generic map
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map error = %v", err)
	}
	if _, ok := raw["version"]; !ok {
		t.Error("unknown field 'version' not preserved after modification")
	}
	if _, ok := raw["experimental"]; !ok {
		t.Error("unknown field 'experimental' not preserved after modification")
	}
}
