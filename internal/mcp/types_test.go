package mcp

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestServer_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		server *Server
	}{
		{
			name: "minimal local server",
			server: &Server{
				Name:    "test",
				Command: "test-cmd",
			},
		},
		{
			name: "stdio server with args and env",
			server: &Server{
				Name:      "github",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-github"},
				Transport: TransportStdio,
				Env: map[string]string{
					"GITHUB_TOKEN": "${GITHUB_TOKEN}",
				},
			},
		},
		{
			name: "sse server with headers",
			server: &Server{
				Name:      "remote",
				URL:       "https://api.example.com/mcp",
				Transport: TransportSSE,
				Headers: map[string]string{
					"Authorization": "Bearer ${API_KEY}",
				},
			},
		},
		{
			name: "platform-specific disabled server",
			server: &Server{
				Name:      "linux-only",
				Command:   "linux-tool",
				Platforms: []string{"linux"},
				Disabled:  true,
			},
		},
		{
			name: "full server configuration",
			server: &Server{
				Name:      "full",
				Command:   "full-cmd",
				Args:      []string{"--verbose", "--config", "/etc/full.conf"},
				URL:       "http://localhost:8080",
				Transport: TransportStdio,
				Env: map[string]string{
					"KEY1": "value1",
					"KEY2": "value2",
				},
				Headers: map[string]string{
					"X-Custom": "header",
				},
				Platforms: []string{"darwin", "linux"},
				Disabled:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.server)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Unmarshal back
			var got Server
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Compare (excluding unknownFields which is unexported)
			if got.Name != tt.server.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.server.Name)
			}
			if got.Command != tt.server.Command {
				t.Errorf("Command = %q, want %q", got.Command, tt.server.Command)
			}
			if !reflect.DeepEqual(got.Args, tt.server.Args) {
				t.Errorf("Args = %v, want %v", got.Args, tt.server.Args)
			}
			if got.URL != tt.server.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.server.URL)
			}
			if got.Transport != tt.server.Transport {
				t.Errorf("Transport = %q, want %q", got.Transport, tt.server.Transport)
			}
			if !reflect.DeepEqual(got.Env, tt.server.Env) {
				t.Errorf("Env = %v, want %v", got.Env, tt.server.Env)
			}
			if !reflect.DeepEqual(got.Headers, tt.server.Headers) {
				t.Errorf("Headers = %v, want %v", got.Headers, tt.server.Headers)
			}
			if !reflect.DeepEqual(got.Platforms, tt.server.Platforms) {
				t.Errorf("Platforms = %v, want %v", got.Platforms, tt.server.Platforms)
			}
			if got.Disabled != tt.server.Disabled {
				t.Errorf("Disabled = %v, want %v", got.Disabled, tt.server.Disabled)
			}
		})
	}
}

func TestServer_PreservesUnknownFields(t *testing.T) {
	// JSON with unknown fields
	input := `{
		"name": "test",
		"command": "test-cmd",
		"futureField": "future value",
		"anotherField": {
			"nested": true
		}
	}`

	// Unmarshal
	var server Server
	if err := json.Unmarshal([]byte(input), &server); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify known fields parsed correctly
	if server.Name != "test" {
		t.Errorf("Name = %q, want %q", server.Name, "test")
	}
	if server.Command != "test-cmd" {
		t.Errorf("Command = %q, want %q", server.Command, "test-cmd")
	}

	// Marshal back
	data, err := json.Marshal(&server)
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

func TestConfig_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "empty config",
			config: NewConfig(),
		},
		{
			name: "single server",
			config: &Config{
				Servers: map[string]*Server{
					"test": {
						Name:    "test",
						Command: "test-cmd",
					},
				},
			},
		},
		{
			name: "multiple servers",
			config: &Config{
				Servers: map[string]*Server{
					"github": {
						Name:    "github",
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-github"},
					},
					"filesystem": {
						Name:    "filesystem",
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
					},
				},
			},
		},
		{
			name: "mixed local and remote servers",
			config: &Config{
				Servers: map[string]*Server{
					"local": {
						Name:      "local",
						Command:   "local-server",
						Transport: TransportStdio,
					},
					"remote": {
						Name:      "remote",
						URL:       "https://api.example.com/mcp",
						Transport: TransportSSE,
						Headers:   map[string]string{"Authorization": "Bearer token"},
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
			var got Config
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Compare server count
			if len(got.Servers) != len(tt.config.Servers) {
				t.Errorf("Servers count = %d, want %d", len(got.Servers), len(tt.config.Servers))
			}

			// Compare each server
			for name, wantServer := range tt.config.Servers {
				gotServer, ok := got.Servers[name]
				if !ok {
					t.Errorf("Server %q missing from result", name)
					continue
				}
				if gotServer.Name != wantServer.Name {
					t.Errorf("Server[%s].Name = %q, want %q", name, gotServer.Name, wantServer.Name)
				}
				if gotServer.Command != wantServer.Command {
					t.Errorf("Server[%s].Command = %q, want %q", name, gotServer.Command, wantServer.Command)
				}
			}
		})
	}
}

func TestConfig_PreservesUnknownFields(t *testing.T) {
	// JSON with unknown fields at root level
	input := `{
		"servers": {
			"test": {
				"name": "test",
				"command": "test-cmd"
			}
		},
		"futureField": "future value",
		"anotherField": {
			"nested": true
		}
	}`

	// Unmarshal
	var config Config
	if err := json.Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify known fields parsed correctly
	if len(config.Servers) != 1 {
		t.Errorf("Servers count = %d, want 1", len(config.Servers))
	}
	if config.Servers["test"].Command != "test-cmd" {
		t.Errorf("Servers[test].Command = %q, want %q", config.Servers["test"].Command, "test-cmd")
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

func TestConfig_PopulatesServerNameFromKey(t *testing.T) {
	// JSON where the server object lacks a "name" field
	input := `{
		"servers": {
			"github": {
				"command": "npx",
				"args": ["@modelcontextprotocol/server-github"]
			}
		}
	}`

	var config Config
	if err := json.Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(config.Servers) != 1 {
		t.Fatalf("Servers count = %d, want 1", len(config.Servers))
	}

	server, ok := config.Servers["github"]
	if !ok {
		t.Fatal("server 'github' not found in map")
	}

	// This is what we're testing: the Name field should be populated from the key "github"
	if server.Name != "github" {
		t.Errorf("server.Name = %q, want %q", server.Name, "github")
	}
}

func TestConfig_RoundTripWithUnknownFields(t *testing.T) {
	// Start with JSON that has unknown fields
	input := `{
		"servers": {"test": {"name": "test", "command": "cmd"}},
		"version": "2.0",
		"experimental": {"feature": true}
	}`

	// First unmarshal
	var config Config
	if err := json.Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("first Unmarshal() error = %v", err)
	}

	// Modify the known field
	config.Servers["new"] = &Server{
		Name:    "new",
		Command: "new-cmd",
	}

	// Marshal
	data, err := json.Marshal(&config)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal again
	var result Config
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("second Unmarshal() error = %v", err)
	}

	// Verify both servers exist
	if _, ok := result.Servers["test"]; !ok {
		t.Error("original server 'test' missing after round-trip")
	}
	if _, ok := result.Servers["new"]; !ok {
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

func TestServer_IsLocal(t *testing.T) {
	tests := []struct {
		name   string
		server *Server
		want   bool
	}{
		{
			name: "explicit stdio transport",
			server: &Server{
				Name:      "test",
				Transport: TransportStdio,
			},
			want: true,
		},
		{
			name: "command without transport",
			server: &Server{
				Name:    "test",
				Command: "some-cmd",
			},
			want: true,
		},
		{
			name: "command with stdio transport",
			server: &Server{
				Name:      "test",
				Command:   "some-cmd",
				Transport: TransportStdio,
			},
			want: true,
		},
		{
			name: "sse transport",
			server: &Server{
				Name:      "test",
				Transport: TransportSSE,
			},
			want: false,
		},
		{
			name: "url without transport",
			server: &Server{
				Name: "test",
				URL:  "https://example.com/mcp",
			},
			want: false,
		},
		{
			name: "empty server",
			server: &Server{
				Name: "test",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.server.IsLocal(); got != tt.want {
				t.Errorf("IsLocal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_IsRemote(t *testing.T) {
	tests := []struct {
		name   string
		server *Server
		want   bool
	}{
		{
			name: "explicit sse transport",
			server: &Server{
				Name:      "test",
				Transport: TransportSSE,
			},
			want: true,
		},
		{
			name: "url without transport",
			server: &Server{
				Name: "test",
				URL:  "https://example.com/mcp",
			},
			want: true,
		},
		{
			name: "url with sse transport",
			server: &Server{
				Name:      "test",
				URL:       "https://example.com/mcp",
				Transport: TransportSSE,
			},
			want: true,
		},
		{
			name: "stdio transport",
			server: &Server{
				Name:      "test",
				Transport: TransportStdio,
			},
			want: false,
		},
		{
			name: "command without transport",
			server: &Server{
				Name:    "test",
				Command: "some-cmd",
			},
			want: false,
		},
		{
			name: "url with command (command takes precedence)",
			server: &Server{
				Name:    "test",
				URL:     "https://example.com/mcp",
				Command: "some-cmd",
			},
			want: false,
		},
		{
			name: "empty server",
			server: &Server{
				Name: "test",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.server.IsRemote(); got != tt.want {
				t.Errorf("IsRemote() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config == nil {
		t.Fatal("NewConfig() returned nil")
	}

	if config.Servers == nil {
		t.Error("NewConfig().Servers is nil, want initialized map")
	}

	// Verify the map is usable
	config.Servers["test"] = &Server{Name: "test"}
	if len(config.Servers) != 1 {
		t.Errorf("Servers count = %d, want 1", len(config.Servers))
	}
}

func TestTransportConstants(t *testing.T) {
	// Verify transport constants have expected values
	if TransportStdio != "stdio" {
		t.Errorf("TransportStdio = %q, want %q", TransportStdio, "stdio")
	}
	if TransportSSE != "sse" {
		t.Errorf("TransportSSE = %q, want %q", TransportSSE, "sse")
	}
}
