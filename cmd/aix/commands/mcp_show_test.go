package commands

import (
	"encoding/json"
	"testing"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

// ptrBool returns a pointer to the given bool value.
func ptrBool(b bool) *bool {
	return &b
}

func TestMCPShowCommand_Metadata(t *testing.T) {
	if mcpShowCmd.Use != "show <name>" {
		t.Errorf("Use = %q, want %q", mcpShowCmd.Use, "show <name>")
	}

	if mcpShowCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if mcpShowCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Check flags exist
	expectedFlags := []string{"json", "show-secrets"}
	for _, flagName := range expectedFlags {
		if mcpShowCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("--%s flag should be defined", flagName)
		}
	}

	// Verify Args validator is set (ExactArgs(1))
	if mcpShowCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestFindDifferences(t *testing.T) {
	tests := []struct {
		name    string
		details map[string]*serverDetail
		want    []string
	}{
		{
			name: "single platform - no differences",
			details: map[string]*serverDetail{
				"claude": {Platform: "Claude Code", Transport: "stdio", Command: "npx"},
			},
			want: nil,
		},
		{
			name: "identical configs across platforms",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Transport: "stdio", Command: "npx", Args: []string{"-y", "pkg"}},
				"opencode": {Platform: "OpenCode", Transport: "stdio", Command: "npx", Args: []string{"-y", "pkg"}},
			},
			want: nil,
		},
		{
			name: "different transport",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Transport: "stdio"},
				"opencode": {Platform: "OpenCode", Transport: "sse"},
			},
			want: []string{"Transport differs: claude=stdio, opencode=sse"},
		},
		{
			name: "different command",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Transport: "stdio", Command: "npx"},
				"opencode": {Platform: "OpenCode", Transport: "stdio", Command: "node"},
			},
			want: []string{"Command differs: claude=\"npx\", opencode=\"node\""},
		},
		{
			name: "different args",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Args: []string{"-y", "pkg"}},
				"opencode": {Platform: "OpenCode", Args: []string{"-y", "other-pkg"}},
			},
			want: []string{"Args differ: claude=[-y pkg], opencode=[-y other-pkg]"},
		},
		{
			name: "different URL",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", URL: "http://a.com"},
				"opencode": {Platform: "OpenCode", URL: "http://b.com"},
			},
			want: []string{"URL differs: claude=\"http://a.com\", opencode=\"http://b.com\""},
		},
		{
			name: "different disabled status",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Disabled: false},
				"opencode": {Platform: "OpenCode", Disabled: true},
			},
			want: []string{"Status differs: claude=enabled, opencode=disabled"},
		},
		{
			name: "env key in one platform only",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Env: map[string]string{"TOKEN": "xxx"}},
				"opencode": {Platform: "OpenCode", Env: nil},
			},
			want: []string{"Env: claude has TOKEN, opencode does not"},
		},
		{
			name: "env value differs",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Env: map[string]string{"TOKEN": "aaa"}},
				"opencode": {Platform: "OpenCode", Env: map[string]string{"TOKEN": "bbb"}},
			},
			want: []string{"Env[TOKEN] value differs between platforms"},
		},
		{
			name: "header key in one platform only",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Headers: map[string]string{"Auth": "Bearer xxx"}},
				"opencode": {Platform: "OpenCode", Headers: nil},
			},
			want: []string{"Headers: claude has Auth, opencode does not"},
		},
		{
			name: "platforms (OS) differs",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Platforms: []string{"darwin"}},
				"opencode": {Platform: "OpenCode", Platforms: nil},
			},
			want: []string{"Platforms (OS) differs: claude=[darwin], opencode=[]"},
		},
		{
			name: "multiple differences",
			details: map[string]*serverDetail{
				"claude":   {Platform: "Claude Code", Transport: "stdio", Command: "npx", Disabled: false},
				"opencode": {Platform: "OpenCode", Transport: "sse", Command: "node", Disabled: true},
			},
			want: []string{
				"Transport differs: claude=stdio, opencode=sse",
				"Command differs: claude=\"npx\", opencode=\"node\"",
				"Status differs: claude=enabled, opencode=disabled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDifferences(tt.details)
			if tt.want == nil {
				if len(got) != 0 {
					t.Errorf("findDifferences() = %v, want nil/empty", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("findDifferences() returned %d differences, want %d: %v", len(got), len(tt.want), got)
				return
			}
			for i, wantDiff := range tt.want {
				if got[i] != wantDiff {
					t.Errorf("findDifferences()[%d] = %q, want %q", i, got[i], wantDiff)
				}
			}
		})
	}
}

func TestCompareMapKeys(t *testing.T) {
	tests := []struct {
		name      string
		m1        map[string]string
		m2        map[string]string
		name1     string
		name2     string
		fieldName string
		wantDiffs int
	}{
		{
			name:      "both nil",
			m1:        nil,
			m2:        nil,
			name1:     "a",
			name2:     "b",
			fieldName: "Env",
			wantDiffs: 0,
		},
		{
			name:      "both empty",
			m1:        map[string]string{},
			m2:        map[string]string{},
			name1:     "a",
			name2:     "b",
			fieldName: "Env",
			wantDiffs: 0,
		},
		{
			name:      "identical",
			m1:        map[string]string{"KEY": "value"},
			m2:        map[string]string{"KEY": "value"},
			name1:     "a",
			name2:     "b",
			fieldName: "Env",
			wantDiffs: 0,
		},
		{
			name:      "key in first only",
			m1:        map[string]string{"KEY": "value"},
			m2:        map[string]string{},
			name1:     "a",
			name2:     "b",
			fieldName: "Env",
			wantDiffs: 1,
		},
		{
			name:      "key in second only",
			m1:        map[string]string{},
			m2:        map[string]string{"KEY": "value"},
			name1:     "a",
			name2:     "b",
			fieldName: "Env",
			wantDiffs: 1,
		},
		{
			name:      "value differs",
			m1:        map[string]string{"KEY": "value1"},
			m2:        map[string]string{"KEY": "value2"},
			name1:     "a",
			name2:     "b",
			fieldName: "Env",
			wantDiffs: 1,
		},
		{
			name:      "multiple differences",
			m1:        map[string]string{"K1": "v1", "K2": "different"},
			m2:        map[string]string{"K2": "other", "K3": "v3"},
			name1:     "a",
			name2:     "b",
			fieldName: "Env",
			wantDiffs: 3, // K1 missing in m2, K3 missing in m1, K2 value differs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareMapKeys(tt.m1, tt.m2, tt.name1, tt.name2, tt.fieldName)
			if len(got) != tt.wantDiffs {
				t.Errorf("compareMapKeys() returned %d diffs, want %d: %v", len(got), tt.wantDiffs, got)
			}
		})
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		disabled bool
		want     string
	}{
		{disabled: false, want: "enabled"},
		{disabled: true, want: "disabled"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := statusString(tt.disabled)
			if got != tt.want {
				t.Errorf("statusString(%v) = %q, want %q", tt.disabled, got, tt.want)
			}
		})
	}
}

func TestExtractServerDetail_Claude(t *testing.T) {
	tests := []struct {
		name     string
		server   *claude.MCPServer
		platform string
		want     *serverDetail
	}{
		{
			name: "stdio server with command",
			server: &claude.MCPServer{
				Name:    "github",
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-github"},
				Type:    "stdio",
				Env:     map[string]string{"TOKEN": "xxx"},
			},
			platform: "Claude Code",
			want: &serverDetail{
				Platform:  "Claude Code",
				Transport: "stdio",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-github"},
				Env:       map[string]string{"TOKEN": "xxx"},
			},
		},
		{
			name: "http server with url",
			server: &claude.MCPServer{
				Name:    "api-gw",
				URL:     "https://api.example.com/mcp",
				Type:    "http",
				Headers: map[string]string{"Auth": "Bearer token"},
			},
			platform: "Claude Code",
			want: &serverDetail{
				Platform:  "Claude Code",
				Transport: "sse",
				URL:       "https://api.example.com/mcp",
				Headers:   map[string]string{"Auth": "Bearer token"},
			},
		},
		{
			name: "infer transport from url",
			server: &claude.MCPServer{
				Name: "api",
				URL:  "https://api.example.com/mcp",
			},
			platform: "Claude Code",
			want: &serverDetail{
				Platform:  "Claude Code",
				Transport: "sse",
				URL:       "https://api.example.com/mcp",
			},
		},
		{
			name: "infer transport from command",
			server: &claude.MCPServer{
				Name:    "local",
				Command: "/usr/bin/mcp-server",
			},
			platform: "Claude Code",
			want: &serverDetail{
				Platform:  "Claude Code",
				Transport: "stdio",
				Command:   "/usr/bin/mcp-server",
			},
		},
		{
			name: "disabled server",
			server: &claude.MCPServer{
				Name:     "disabled",
				Command:  "cmd",
				Disabled: true,
			},
			platform: "Claude Code",
			want: &serverDetail{
				Platform:  "Claude Code",
				Transport: "stdio",
				Command:   "cmd",
				Disabled:  true,
			},
		},
		{
			name: "with platforms (OS) restriction",
			server: &claude.MCPServer{
				Name:      "macos-only",
				Command:   "/usr/bin/mcp",
				Platforms: []string{"darwin"},
			},
			platform: "Claude Code",
			want: &serverDetail{
				Platform:  "Claude Code",
				Transport: "stdio",
				Command:   "/usr/bin/mcp",
				Platforms: []string{"darwin"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractClaudeMCPServer(tt.server, tt.platform)
			if got.Platform != tt.want.Platform {
				t.Errorf("Platform = %q, want %q", got.Platform, tt.want.Platform)
			}
			if got.Transport != tt.want.Transport {
				t.Errorf("Transport = %q, want %q", got.Transport, tt.want.Transport)
			}
			if got.Command != tt.want.Command {
				t.Errorf("Command = %q, want %q", got.Command, tt.want.Command)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
			if got.Disabled != tt.want.Disabled {
				t.Errorf("Disabled = %v, want %v", got.Disabled, tt.want.Disabled)
			}
		})
	}
}

func TestExtractServerDetail_OpenCode(t *testing.T) {
	tests := []struct {
		name     string
		server   *opencode.MCPServer
		platform string
		want     *serverDetail
	}{
		{
			name: "local server with command",
			server: &opencode.MCPServer{
				Name:        "github",
				Command:     []string{"npx", "-y", "@modelcontextprotocol/server-github"},
				Type:        "local",
				Environment: map[string]string{"TOKEN": "xxx"},
			},
			platform: "OpenCode",
			want: &serverDetail{
				Platform:  "OpenCode",
				Transport: "stdio",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-github"},
				Env:       map[string]string{"TOKEN": "xxx"},
			},
		},
		{
			name: "remote server with url",
			server: &opencode.MCPServer{
				Name:    "api-gw",
				URL:     "https://api.example.com/mcp",
				Type:    "remote",
				Headers: map[string]string{"Auth": "Bearer token"},
			},
			platform: "OpenCode",
			want: &serverDetail{
				Platform:  "OpenCode",
				Transport: "sse",
				URL:       "https://api.example.com/mcp",
				Headers:   map[string]string{"Auth": "Bearer token"},
			},
		},
		{
			name: "infer transport from type remote",
			server: &opencode.MCPServer{
				Name: "api",
				URL:  "https://api.example.com/mcp",
				Type: "remote",
			},
			platform: "OpenCode",
			want: &serverDetail{
				Platform:  "OpenCode",
				Transport: "sse",
				URL:       "https://api.example.com/mcp",
			},
		},
		{
			name: "infer transport from url alone",
			server: &opencode.MCPServer{
				Name: "api",
				URL:  "https://api.example.com/mcp",
			},
			platform: "OpenCode",
			want: &serverDetail{
				Platform:  "OpenCode",
				Transport: "sse",
				URL:       "https://api.example.com/mcp",
			},
		},
		{
			name: "disabled server",
			server: &opencode.MCPServer{
				Name:    "disabled",
				Command: []string{"cmd"},
				Enabled: ptrBool(false),
			},
			platform: "OpenCode",
			want: &serverDetail{
				Platform:  "OpenCode",
				Transport: "stdio",
				Command:   "cmd",
				Disabled:  true,
			},
		},
		{
			name: "empty command slice",
			server: &opencode.MCPServer{
				Name: "remote-only",
				URL:  "http://example.com",
			},
			platform: "OpenCode",
			want: &serverDetail{
				Platform:  "OpenCode",
				Transport: "sse",
				Command:   "",
				Args:      nil,
				URL:       "http://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOpenCodeMCPServer(tt.server, tt.platform)
			if got.Platform != tt.want.Platform {
				t.Errorf("Platform = %q, want %q", got.Platform, tt.want.Platform)
			}
			if got.Transport != tt.want.Transport {
				t.Errorf("Transport = %q, want %q", got.Transport, tt.want.Transport)
			}
			if got.Command != tt.want.Command {
				t.Errorf("Command = %q, want %q", got.Command, tt.want.Command)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
			if got.Disabled != tt.want.Disabled {
				t.Errorf("Disabled = %v, want %v", got.Disabled, tt.want.Disabled)
			}
		})
	}
}

func TestExtractServerDetail_UnknownType(t *testing.T) {
	// Test that extractServerDetail returns nil for unknown types
	got := extractServerDetail("not a server type", "Test")
	if got != nil {
		t.Errorf("extractServerDetail() = %v, want nil for unknown type", got)
	}
}

func TestOutputMCPShowJSON(t *testing.T) {
	details := map[string]*serverDetail{
		"claude": {
			Platform:  "Claude Code",
			Transport: "stdio",
			Command:   "npx",
			Args:      []string{"-y", "pkg"},
			Env:       map[string]string{"TOKEN": "****xxx"},
		},
	}
	differences := []string{"Transport differs: claude=stdio, opencode=sse"}

	// Test JSON output
	err := outputMCPShowJSON("github", details, differences)
	if err != nil {
		t.Fatalf("outputMCPShowJSON() error = %v", err)
	}

	// The function prints to stdout, so we just verify it doesn't error
	// A more comprehensive test would capture stdout
}

func TestMCPShowOutputStructure(t *testing.T) {
	// Test that mcpShowOutput can be marshaled correctly
	output := mcpShowOutput{
		Name: "github",
		Platforms: map[string]*serverDetail{
			"claude": {
				Platform:  "Claude Code",
				Transport: "stdio",
				Command:   "npx",
				Args:      []string{"-y", "pkg"},
				Disabled:  false,
				Env:       map[string]string{"TOKEN": "****xxx"},
				Headers:   map[string]string{"Auth": "****abc"},
			},
		},
		Differences: []string{"Transport differs"},
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal mcpShowOutput: %v", err)
	}

	// Verify we can unmarshal it back
	var unmarshaled mcpShowOutput
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal mcpShowOutput: %v", err)
	}

	if unmarshaled.Name != output.Name {
		t.Errorf("Name = %q, want %q", unmarshaled.Name, output.Name)
	}
	if len(unmarshaled.Platforms) != len(output.Platforms) {
		t.Errorf("Platforms count = %d, want %d", len(unmarshaled.Platforms), len(output.Platforms))
	}
	if len(unmarshaled.Differences) != len(output.Differences) {
		t.Errorf("Differences count = %d, want %d", len(unmarshaled.Differences), len(output.Differences))
	}
}

func TestServerDetailJSONTags(t *testing.T) {
	// Test that serverDetail has correct JSON tags and omitempty works
	detail := serverDetail{
		Platform:  "Test",
		Transport: "stdio",
		Command:   "cmd",
		// Args, URL, Env, Headers, Platforms all omitted
	}

	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("failed to marshal serverDetail: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Check that omitempty fields are not present
	if _, ok := unmarshaled["env"]; ok {
		t.Error("env should be omitted when nil")
	}
	if _, ok := unmarshaled["headers"]; ok {
		t.Error("headers should be omitted when nil")
	}
	if _, ok := unmarshaled["platforms"]; ok {
		t.Error("platforms should be omitted when nil")
	}

	// Check required fields are present
	if _, ok := unmarshaled["platform"]; !ok {
		t.Error("platform should be present")
	}
	if _, ok := unmarshaled["transport"]; !ok {
		t.Error("transport should be present")
	}
}

func TestFindDifferences_Deterministic(t *testing.T) {
	// Test that findDifferences produces deterministic output
	// by running it multiple times with the same input
	details := map[string]*serverDetail{
		"claude":   {Platform: "Claude Code", Transport: "stdio", Command: "npx"},
		"opencode": {Platform: "OpenCode", Transport: "sse", Command: "node"},
	}

	var firstResult []string
	for i := range 10 {
		result := findDifferences(details)
		if i == 0 {
			firstResult = result
		} else {
			if len(result) != len(firstResult) {
				t.Fatalf("iteration %d: got %d differences, want %d", i, len(result), len(firstResult))
			}
			for j, diff := range result {
				if diff != firstResult[j] {
					t.Errorf("iteration %d, diff %d: got %q, want %q", i, j, diff, firstResult[j])
				}
			}
		}
	}
}

func TestMCPShowSecretsFlags(t *testing.T) {
	// Test that the flag variables exist and have correct defaults
	// We need to check the command's flags, not the package variables
	jsonFlag := mcpShowCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Fatal("--json flag not found")
	}
	if jsonFlag.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", jsonFlag.DefValue, "false")
	}

	showSecretsFlag := mcpShowCmd.Flags().Lookup("show-secrets")
	if showSecretsFlag == nil {
		t.Fatal("--show-secrets flag not found")
	}
	if showSecretsFlag.DefValue != "false" {
		t.Errorf("--show-secrets default = %q, want %q", showSecretsFlag.DefValue, "false")
	}
}
