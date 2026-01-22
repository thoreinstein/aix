package mcp_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/thoreinstein/aix/internal/mcp"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

// Test fixtures representing realistic MCP server configurations.
var (
	// localStdioServer is a local stdio server with environment variables.
	localStdioServer = &mcp.Server{
		Name:      "github-mcp",
		Command:   "npx",
		Args:      []string{"-y", "@modelcontextprotocol/server-github"},
		Transport: mcp.TransportStdio,
		Env: map[string]string{
			"GITHUB_TOKEN":          "ghp_xxxxxxxxxxxx",
			"GITHUB_ENTERPRISE_URL": "https://github.example.com",
		},
	}

	// remoteSSEServer is a remote SSE server with authentication headers.
	remoteSSEServer = &mcp.Server{
		Name:      "api-gateway",
		URL:       "https://api.example.com/mcp/v1",
		Transport: mcp.TransportSSE,
		Headers: map[string]string{
			"Authorization": "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
			"X-API-Version": "2024-01",
		},
	}

	// platformRestrictedServer only runs on darwin.
	platformRestrictedServer = &mcp.Server{
		Name:      "macos-tools",
		Command:   "/usr/local/bin/macos-mcp-server",
		Args:      []string{"--verbose", "--port", "8080"},
		Transport: mcp.TransportStdio,
		Platforms: []string{"darwin"},
		Env: map[string]string{
			"HOME": "/Users/developer",
		},
	}

	// fullyPopulatedServer has all optional fields populated.
	fullyPopulatedServer = &mcp.Server{
		Name:      "full-featured",
		Command:   "python3",
		Args:      []string{"-m", "mcp_server", "--config", "/etc/mcp/config.yaml"},
		URL:       "", // Empty for local server
		Transport: mcp.TransportStdio,
		Env: map[string]string{
			"PYTHONPATH":  "/opt/mcp/lib",
			"LOG_LEVEL":   "debug",
			"CONFIG_PATH": "/etc/mcp",
		},
		Headers:   nil, // Not applicable for stdio
		Platforms: []string{"darwin", "linux"},
		Disabled:  false,
	}

	// minimalServer has only required fields.
	// Note: Transport is inferred as "stdio" during OpenCode round-trip because Command is set.
	minimalServer = &mcp.Server{
		Name:    "minimal",
		Command: "simple-server",
	}

	// minimalServerWithTransport is minimal but with explicit transport for OpenCode tests.
	minimalServerWithTransport = &mcp.Server{
		Name:      "minimal-explicit",
		Command:   "simple-server",
		Transport: mcp.TransportStdio,
	}

	// disabledServer is temporarily disabled.
	disabledServer = &mcp.Server{
		Name:      "maintenance",
		Command:   "maintenance-server",
		Transport: mcp.TransportStdio,
		Disabled:  true,
	}
)

// TestRoundTrip_Canonical_Claude_Canonical verifies that canonical configs
// survive a round-trip through Claude format without data loss.
func TestRoundTrip_Canonical_Claude_Canonical(t *testing.T) {
	translator := claude.NewMCPTranslator()

	tests := []struct {
		name   string
		server *mcp.Server
	}{
		{"local stdio server with env", localStdioServer},
		{"remote SSE server with headers", remoteSSEServer},
		{"platform restricted server", platformRestrictedServer},
		{"fully populated server", fullyPopulatedServer},
		{"minimal server", minimalServer},
		{"disabled server", disabledServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &mcp.Config{
				Servers: map[string]*mcp.Server{
					tt.server.Name: tt.server,
				},
			}

			// canonical -> Claude
			claudeJSON, err := translator.FromCanonical(original)
			if err != nil {
				t.Fatalf("FromCanonical() error = %v", err)
			}

			// Claude -> canonical
			result, err := translator.ToCanonical(claudeJSON)
			if err != nil {
				t.Fatalf("ToCanonical() error = %v", err)
			}

			// Verify server count
			if len(result.Servers) != 1 {
				t.Fatalf("len(Servers) = %d, want 1", len(result.Servers))
			}

			resultServer := result.Servers[tt.server.Name]
			if resultServer == nil {
				t.Fatalf("server %q not found in result", tt.server.Name)
			}

			// Verify all fields preserved
			assertServerEqual(t, tt.server, resultServer)
		})
	}
}

// TestRoundTrip_Canonical_OpenCode_Canonical verifies that canonical configs
// survive a round-trip through OpenCode format.
// NOTE: Platforms field is LOSSY - OpenCode does not support it.
func TestRoundTrip_Canonical_OpenCode_Canonical(t *testing.T) {
	translator := opencode.NewMCPTranslator()

	tests := []struct {
		name           string
		server         *mcp.Server
		lossyFields    []string // Fields that are expected to be lost
		inferTransport bool     // Whether Transport is inferred during round-trip
	}{
		{
			name:   "local stdio server with env",
			server: localStdioServer,
		},
		{
			name:   "remote SSE server with headers",
			server: remoteSSEServer,
		},
		{
			name:        "platform restricted server - platforms lost",
			server:      platformRestrictedServer,
			lossyFields: []string{"Platforms"},
		},
		{
			name:        "fully populated server - platforms lost",
			server:      fullyPopulatedServer,
			lossyFields: []string{"Platforms"},
		},
		{
			name:           "minimal server - transport inferred",
			server:         minimalServer,
			inferTransport: true, // OpenCode infers "local" -> "stdio" from Command
		},
		{
			name:   "minimal server with explicit transport",
			server: minimalServerWithTransport,
		},
		{
			name:   "disabled server",
			server: disabledServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &mcp.Config{
				Servers: map[string]*mcp.Server{
					tt.server.Name: tt.server,
				},
			}

			// canonical -> OpenCode
			openJSON, err := translator.FromCanonical(original)
			if err != nil {
				t.Fatalf("FromCanonical() error = %v", err)
			}

			// OpenCode -> canonical
			result, err := translator.ToCanonical(openJSON)
			if err != nil {
				t.Fatalf("ToCanonical() error = %v", err)
			}

			// Verify server count
			if len(result.Servers) != 1 {
				t.Fatalf("len(Servers) = %d, want 1", len(result.Servers))
			}

			resultServer := result.Servers[tt.server.Name]
			if resultServer == nil {
				t.Fatalf("server %q not found in result", tt.server.Name)
			}

			// Build expected server with lossy fields zeroed
			expected := copyServer(tt.server)
			for _, field := range tt.lossyFields {
				zeroField(expected, field)
			}

			// If transport is inferred, set expected transport to what OpenCode infers
			if tt.inferTransport && expected.Command != "" {
				expected.Transport = mcp.TransportStdio
			} else if tt.inferTransport && expected.URL != "" {
				expected.Transport = mcp.TransportSSE
			}

			assertServerEqual(t, expected, resultServer)
		})
	}
}

// TestRoundTrip_Claude_OpenCode_CrossPlatform verifies data flows correctly
// when translating Claude -> canonical -> OpenCode -> canonical.
func TestRoundTrip_Claude_OpenCode_CrossPlatform(t *testing.T) {
	claudeTranslator := claude.NewMCPTranslator()
	openCodeTranslator := opencode.NewMCPTranslator()

	// Start with Claude JSON format
	claudeJSON := []byte(`{
		"mcpServers": {
			"github": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-github"],
				"transport": "stdio",
				"env": {"GITHUB_TOKEN": "token123"}
			},
			"remote-api": {
				"url": "https://api.example.com/mcp",
				"transport": "sse",
				"headers": {"Authorization": "Bearer secret"}
			}
		}
	}`)

	// Claude -> canonical
	canonical1, err := claudeTranslator.ToCanonical(claudeJSON)
	if err != nil {
		t.Fatalf("Claude.ToCanonical() error = %v", err)
	}

	// canonical -> OpenCode
	openCodeJSON, err := openCodeTranslator.FromCanonical(canonical1)
	if err != nil {
		t.Fatalf("OpenCode.FromCanonical() error = %v", err)
	}

	// OpenCode -> canonical
	canonical2, err := openCodeTranslator.ToCanonical(openCodeJSON)
	if err != nil {
		t.Fatalf("OpenCode.ToCanonical() error = %v", err)
	}

	// Verify server count preserved
	if len(canonical2.Servers) != 2 {
		t.Fatalf("len(Servers) = %d, want 2", len(canonical2.Servers))
	}

	// Verify github server
	github := canonical2.Servers["github"]
	if github == nil {
		t.Fatal("github server not found")
	}
	if github.Command != "npx" {
		t.Errorf("github.Command = %q, want %q", github.Command, "npx")
	}
	if len(github.Args) != 2 || github.Args[0] != "-y" {
		t.Errorf("github.Args = %v, want [-y @modelcontextprotocol/server-github]", github.Args)
	}
	if github.Transport != mcp.TransportStdio {
		t.Errorf("github.Transport = %q, want %q", github.Transport, mcp.TransportStdio)
	}
	if github.Env["GITHUB_TOKEN"] != "token123" {
		t.Errorf("github.Env[GITHUB_TOKEN] = %q, want %q", github.Env["GITHUB_TOKEN"], "token123")
	}

	// Verify remote-api server
	remoteAPI := canonical2.Servers["remote-api"]
	if remoteAPI == nil {
		t.Fatal("remote-api server not found")
	}
	if remoteAPI.URL != "https://api.example.com/mcp" {
		t.Errorf("remote-api.URL = %q, want %q", remoteAPI.URL, "https://api.example.com/mcp")
	}
	if remoteAPI.Transport != mcp.TransportSSE {
		t.Errorf("remote-api.Transport = %q, want %q", remoteAPI.Transport, mcp.TransportSSE)
	}
	if remoteAPI.Headers["Authorization"] != "Bearer secret" {
		t.Errorf("remote-api.Headers[Authorization] = %q, want %q", remoteAPI.Headers["Authorization"], "Bearer secret")
	}
}

// TestRoundTrip_OpenCode_Claude_CrossPlatform verifies data flows correctly
// when translating OpenCode -> canonical -> Claude -> canonical.
func TestRoundTrip_OpenCode_Claude_CrossPlatform(t *testing.T) {
	claudeTranslator := claude.NewMCPTranslator()
	openCodeTranslator := opencode.NewMCPTranslator()

	// Start with OpenCode JSON format
	openCodeJSON := []byte(`{
		"mcp": {
			"filesystem": {
				"command": ["npx", "-y", "@modelcontextprotocol/server-filesystem", "/home/user"],
				"type": "local",
				"environment": {"HOME": "/home/user"}
			},
			"web-search": {
				"url": "https://search.example.com/mcp",
				"type": "remote",
				"headers": {"API-Key": "key123"}
			}
		}
	}`)

	// OpenCode -> canonical
	canonical1, err := openCodeTranslator.ToCanonical(openCodeJSON)
	if err != nil {
		t.Fatalf("OpenCode.ToCanonical() error = %v", err)
	}

	// canonical -> Claude
	claudeJSON, err := claudeTranslator.FromCanonical(canonical1)
	if err != nil {
		t.Fatalf("Claude.FromCanonical() error = %v", err)
	}

	// Claude -> canonical
	canonical2, err := claudeTranslator.ToCanonical(claudeJSON)
	if err != nil {
		t.Fatalf("Claude.ToCanonical() error = %v", err)
	}

	// Verify server count preserved
	if len(canonical2.Servers) != 2 {
		t.Fatalf("len(Servers) = %d, want 2", len(canonical2.Servers))
	}

	// Verify filesystem server
	fs := canonical2.Servers["filesystem"]
	if fs == nil {
		t.Fatal("filesystem server not found")
	}
	if fs.Command != "npx" {
		t.Errorf("filesystem.Command = %q, want %q", fs.Command, "npx")
	}
	if len(fs.Args) != 3 {
		t.Errorf("filesystem.len(Args) = %d, want 3", len(fs.Args))
	}
	if fs.Transport != mcp.TransportStdio {
		t.Errorf("filesystem.Transport = %q, want %q", fs.Transport, mcp.TransportStdio)
	}
	if fs.Env["HOME"] != "/home/user" {
		t.Errorf("filesystem.Env[HOME] = %q, want %q", fs.Env["HOME"], "/home/user")
	}

	// Verify web-search server
	webSearch := canonical2.Servers["web-search"]
	if webSearch == nil {
		t.Fatal("web-search server not found")
	}
	if webSearch.URL != "https://search.example.com/mcp" {
		t.Errorf("web-search.URL = %q, want %q", webSearch.URL, "https://search.example.com/mcp")
	}
	if webSearch.Transport != mcp.TransportSSE {
		t.Errorf("web-search.Transport = %q, want %q", webSearch.Transport, mcp.TransportSSE)
	}
	if webSearch.Headers["API-Key"] != "key123" {
		t.Errorf("web-search.Headers[API-Key] = %q, want %q", webSearch.Headers["API-Key"], "key123")
	}
}

// TestRoundTrip_UnknownFieldsPreservation_Canonical verifies that unknown fields
// in canonical Server survive JSON marshal/unmarshal round-trips.
func TestRoundTrip_UnknownFieldsPreservation_Canonical(t *testing.T) {
	// JSON with unknown fields at server level
	serverJSON := `{
		"name": "test-server",
		"command": "test-cmd",
		"args": ["--flag"],
		"transport": "stdio",
		"future_field": "should be preserved",
		"nested_config": {"key": "value", "num": 42}
	}`

	var server mcp.Server
	if err := json.Unmarshal([]byte(serverJSON), &server); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify known fields
	if server.Name != "test-server" {
		t.Errorf("Name = %q, want %q", server.Name, "test-server")
	}
	if server.Command != "test-cmd" {
		t.Errorf("Command = %q, want %q", server.Command, "test-cmd")
	}

	// Marshal back to JSON
	resultJSON, err := json.Marshal(&server)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal to generic map to verify unknown fields
	var result map[string]any
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("Unmarshal result error = %v", err)
	}

	// Check unknown string field preserved
	if result["future_field"] != "should be preserved" {
		t.Errorf("future_field = %v, want %q", result["future_field"], "should be preserved")
	}

	// Check unknown nested object preserved
	nested, ok := result["nested_config"].(map[string]any)
	if !ok {
		t.Fatalf("nested_config not a map: %T", result["nested_config"])
	}
	if nested["key"] != "value" {
		t.Errorf("nested_config.key = %v, want %q", nested["key"], "value")
	}
	// JSON numbers are float64
	if nested["num"] != float64(42) {
		t.Errorf("nested_config.num = %v, want 42", nested["num"])
	}
}

// TestRoundTrip_UnknownFieldsPreservation_Config verifies that unknown fields
// in canonical Config survive JSON marshal/unmarshal round-trips.
func TestRoundTrip_UnknownFieldsPreservation_Config(t *testing.T) {
	// JSON with unknown fields at config level
	configJSON := `{
		"servers": {
			"test": {
				"name": "test",
				"command": "cmd"
			}
		},
		"metadata": {"version": "1.0"},
		"global_setting": true
	}`

	var config mcp.Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify known fields
	if len(config.Servers) != 1 {
		t.Errorf("len(Servers) = %d, want 1", len(config.Servers))
	}

	// Marshal back to JSON
	resultJSON, err := json.Marshal(&config)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal to generic map to verify unknown fields
	var result map[string]any
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		t.Fatalf("Unmarshal result error = %v", err)
	}

	// Check unknown nested object preserved
	metadata, ok := result["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata not a map: %T", result["metadata"])
	}
	if metadata["version"] != "1.0" {
		t.Errorf("metadata.version = %v, want %q", metadata["version"], "1.0")
	}

	// Check unknown boolean field preserved
	if result["global_setting"] != true {
		t.Errorf("global_setting = %v, want true", result["global_setting"])
	}
}

// TestRoundTrip_PlatformsLossy_OpenCode explicitly documents that Platforms
// field is lost when round-tripping through OpenCode.
func TestRoundTrip_PlatformsLossy_OpenCode(t *testing.T) {
	openCodeTranslator := opencode.NewMCPTranslator()

	original := &mcp.Config{
		Servers: map[string]*mcp.Server{
			"darwin-only": {
				Name:      "darwin-only",
				Command:   "macos-specific-tool",
				Transport: mcp.TransportStdio,
				Platforms: []string{"darwin"}, // This WILL be lost
			},
			"multi-platform": {
				Name:      "multi-platform",
				Command:   "cross-platform-tool",
				Transport: mcp.TransportStdio,
				Platforms: []string{"darwin", "linux", "windows"}, // This WILL be lost
			},
		},
	}

	// canonical -> OpenCode
	openCodeJSON, err := openCodeTranslator.FromCanonical(original)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	// Verify platforms not in OpenCode JSON
	var openCodeConfig map[string]any
	if err := json.Unmarshal(openCodeJSON, &openCodeConfig); err != nil {
		t.Fatalf("Unmarshal OpenCode JSON error = %v", err)
	}

	mcpServers := openCodeConfig["mcp"].(map[string]any)
	for name, serverAny := range mcpServers {
		server := serverAny.(map[string]any)
		if _, exists := server["platforms"]; exists {
			t.Errorf("server %q: platforms field should not exist in OpenCode output", name)
		}
	}

	// OpenCode -> canonical
	result, err := openCodeTranslator.ToCanonical(openCodeJSON)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// Verify platforms is empty (lost) in all servers
	for name, server := range result.Servers {
		if len(server.Platforms) > 0 {
			t.Errorf("server %q: Platforms should be empty after OpenCode round-trip, got %v",
				name, server.Platforms)
		}
	}
}

// TestRoundTrip_MultipleServers verifies that configs with multiple servers
// survive round-trips correctly.
func TestRoundTrip_MultipleServers(t *testing.T) {
	claudeTranslator := claude.NewMCPTranslator()

	// Create servers using the same name for key and Name field
	github := copyServer(localStdioServer)
	github.Name = "github"

	api := copyServer(remoteSSEServer)
	api.Name = "api"

	macos := copyServer(platformRestrictedServer)
	macos.Name = "macos"

	full := copyServer(fullyPopulatedServer)
	full.Name = "full"

	minimal := copyServer(minimalServer)
	minimal.Name = "minimal"

	disabled := copyServer(disabledServer)
	disabled.Name = "disabled"

	original := &mcp.Config{
		Servers: map[string]*mcp.Server{
			"github":   github,
			"api":      api,
			"macos":    macos,
			"full":     full,
			"minimal":  minimal,
			"disabled": disabled,
		},
	}

	// canonical -> Claude -> canonical
	claudeJSON, err := claudeTranslator.FromCanonical(original)
	if err != nil {
		t.Fatalf("FromCanonical() error = %v", err)
	}

	result, err := claudeTranslator.ToCanonical(claudeJSON)
	if err != nil {
		t.Fatalf("ToCanonical() error = %v", err)
	}

	// Verify all servers present
	if len(result.Servers) != len(original.Servers) {
		t.Fatalf("len(Servers) = %d, want %d", len(result.Servers), len(original.Servers))
	}

	for name, origServer := range original.Servers {
		resultServer := result.Servers[name]
		if resultServer == nil {
			t.Errorf("server %q not found in result", name)
			continue
		}
		assertServerEqual(t, origServer, resultServer)
	}
}

// Helper functions

// assertServerEqual compares two servers and reports differences.
func assertServerEqual(t *testing.T, expected, actual *mcp.Server) {
	t.Helper()

	if actual.Name != expected.Name {
		t.Errorf("Name = %q, want %q", actual.Name, expected.Name)
	}
	if actual.Command != expected.Command {
		t.Errorf("Command = %q, want %q", actual.Command, expected.Command)
	}
	if !reflect.DeepEqual(actual.Args, expected.Args) {
		t.Errorf("Args = %v, want %v", actual.Args, expected.Args)
	}
	if actual.URL != expected.URL {
		t.Errorf("URL = %q, want %q", actual.URL, expected.URL)
	}
	if actual.Transport != expected.Transport {
		t.Errorf("Transport = %q, want %q", actual.Transport, expected.Transport)
	}
	if !reflect.DeepEqual(actual.Env, expected.Env) {
		t.Errorf("Env = %v, want %v", actual.Env, expected.Env)
	}
	if !reflect.DeepEqual(actual.Headers, expected.Headers) {
		t.Errorf("Headers = %v, want %v", actual.Headers, expected.Headers)
	}
	if !reflect.DeepEqual(actual.Platforms, expected.Platforms) {
		t.Errorf("Platforms = %v, want %v", actual.Platforms, expected.Platforms)
	}
	if actual.Disabled != expected.Disabled {
		t.Errorf("Disabled = %v, want %v", actual.Disabled, expected.Disabled)
	}
}

// copyServer creates a deep copy of a Server.
func copyServer(s *mcp.Server) *mcp.Server {
	cp := &mcp.Server{
		Name:      s.Name,
		Command:   s.Command,
		URL:       s.URL,
		Transport: s.Transport,
		Disabled:  s.Disabled,
	}

	if s.Args != nil {
		cp.Args = make([]string, len(s.Args))
		copy(cp.Args, s.Args)
	}
	if s.Env != nil {
		cp.Env = make(map[string]string, len(s.Env))
		for k, v := range s.Env {
			cp.Env[k] = v
		}
	}
	if s.Headers != nil {
		cp.Headers = make(map[string]string, len(s.Headers))
		for k, v := range s.Headers {
			cp.Headers[k] = v
		}
	}
	if s.Platforms != nil {
		cp.Platforms = make([]string, len(s.Platforms))
		copy(cp.Platforms, s.Platforms)
	}

	return cp
}

// zeroField sets a field to its zero value by name.
func zeroField(s *mcp.Server, field string) {
	switch field {
	case "Platforms":
		s.Platforms = nil
	case "Args":
		s.Args = nil
	case "Env":
		s.Env = nil
	case "Headers":
		s.Headers = nil
	case "Disabled":
		s.Disabled = false
	}
}
