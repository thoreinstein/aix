package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigSemanticCheck_Name(t *testing.T) {
	c := NewConfigSemanticCheck()
	if got := c.Name(); got != "config-semantics" {
		t.Errorf("Name() = %q, want %q", got, "config-semantics")
	}
}

func TestConfigSemanticCheck_Category(t *testing.T) {
	c := NewConfigSemanticCheck()
	if got := c.Category(); got != "config" {
		t.Errorf("Category() = %q, want %q", got, "config")
	}
}

func TestConfigSemanticCheck_ImplementsCheck(t *testing.T) {
	var _ Check = (*ConfigSemanticCheck)(nil)
}

func TestConfigSemanticCheck_parseClaudeServers(t *testing.T) {
	c := NewConfigSemanticCheck()

	tests := []struct {
		name        string
		input       string
		wantServers int
		wantErr     bool
	}{
		{
			name:        "valid config with servers",
			input:       `{"mcpServers": {"github": {"command": "npx", "args": ["-y", "@mcp/server-github"]}}}`,
			wantServers: 1,
			wantErr:     false,
		},
		{
			name:        "empty mcpServers",
			input:       `{"mcpServers": {}}`,
			wantServers: 0,
			wantErr:     false,
		},
		{
			name:        "no mcpServers key",
			input:       `{"otherKey": "value"}`,
			wantServers: 0,
			wantErr:     false,
		},
		{
			name:        "remote server",
			input:       `{"mcpServers": {"api": {"url": "https://api.example.com", "type": "http"}}}`,
			wantServers: 1,
			wantErr:     false,
		},
		{
			name:        "invalid json",
			input:       `{invalid}`,
			wantServers: 0,
			wantErr:     true,
		},
		{
			name:        "multiple servers",
			input:       `{"mcpServers": {"a": {"command": "a"}, "b": {"command": "b"}, "c": {"url": "http://c"}}}`,
			wantServers: 3,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servers, err := c.parseClaudeServers([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseClaudeServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(servers) != tt.wantServers {
				t.Errorf("parseClaudeServers() got %d servers, want %d", len(servers), tt.wantServers)
			}
		})
	}
}

func TestConfigSemanticCheck_parseOpenCodeServers(t *testing.T) {
	c := NewConfigSemanticCheck()

	tests := []struct {
		name        string
		input       string
		wantServers int
		wantCmd     string
		wantArgs    []string
	}{
		{
			name:        "valid config with command array",
			input:       `{"mcp": {"github": {"command": ["npx", "-y", "@mcp/server-github"]}}}`,
			wantServers: 1,
			wantCmd:     "npx",
			wantArgs:    []string{"-y", "@mcp/server-github"},
		},
		{
			name:        "remote server",
			input:       `{"mcp": {"api": {"url": "https://api.example.com", "type": "remote"}}}`,
			wantServers: 1,
			wantCmd:     "",
			wantArgs:    nil,
		},
		{
			name:        "empty mcp",
			input:       `{"mcp": {}}`,
			wantServers: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servers, err := c.parseOpenCodeServers([]byte(tt.input))
			if err != nil {
				t.Errorf("parseOpenCodeServers() error = %v", err)
				return
			}
			if len(servers) != tt.wantServers {
				t.Errorf("parseOpenCodeServers() got %d servers, want %d", len(servers), tt.wantServers)
				return
			}
			if tt.wantCmd != "" {
				for _, s := range servers {
					if s.Command != tt.wantCmd {
						t.Errorf("parseOpenCodeServers() command = %q, want %q", s.Command, tt.wantCmd)
					}
					if len(s.Args) != len(tt.wantArgs) {
						t.Errorf("parseOpenCodeServers() args = %v, want %v", s.Args, tt.wantArgs)
					}
				}
			}
		})
	}
}

func TestConfigSemanticCheck_isLocalServer(t *testing.T) {
	c := NewConfigSemanticCheck()

	tests := []struct {
		name   string
		server *mcpServerInfo
		want   bool
	}{
		{
			name:   "explicit stdio transport",
			server: &mcpServerInfo{Command: "npx", Transport: "stdio"},
			want:   true,
		},
		{
			name:   "explicit local transport (opencode)",
			server: &mcpServerInfo{Command: "npx", Transport: "local"},
			want:   true,
		},
		{
			name:   "implicit local via command",
			server: &mcpServerInfo{Command: "npx"},
			want:   true,
		},
		{
			name:   "remote server with url",
			server: &mcpServerInfo{URL: "https://api.example.com", Transport: "sse"},
			want:   false,
		},
		{
			name:   "empty server",
			server: &mcpServerInfo{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.isLocalServer(tt.server); got != tt.want {
				t.Errorf("isLocalServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigSemanticCheck_isRemoteServer(t *testing.T) {
	c := NewConfigSemanticCheck()

	tests := []struct {
		name   string
		server *mcpServerInfo
		want   bool
	}{
		{
			name:   "explicit sse transport",
			server: &mcpServerInfo{URL: "https://api.example.com", Transport: "sse"},
			want:   true,
		},
		{
			name:   "explicit http transport (claude)",
			server: &mcpServerInfo{URL: "https://api.example.com", Transport: "http"},
			want:   true,
		},
		{
			name:   "explicit remote transport (opencode)",
			server: &mcpServerInfo{URL: "https://api.example.com", Transport: "remote"},
			want:   true,
		},
		{
			name:   "implicit remote via url",
			server: &mcpServerInfo{URL: "https://api.example.com"},
			want:   true,
		},
		{
			name:   "local server",
			server: &mcpServerInfo{Command: "npx", Transport: "stdio"},
			want:   false,
		},
		{
			name:   "empty server",
			server: &mcpServerInfo{},
			want:   false,
		},
		{
			name:   "both command and url with no transport",
			server: &mcpServerInfo{Command: "npx", URL: "https://api.example.com"},
			want:   false, // command takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.isRemoteServer(tt.server); got != tt.want {
				t.Errorf("isRemoteServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigSemanticCheck_validateServer(t *testing.T) {
	c := NewConfigSemanticCheck()

	tests := []struct {
		name       string
		server     *mcpServerInfo
		wantIssues int
		wantTypes  []string
	}{
		{
			name:       "valid local server",
			server:     &mcpServerInfo{Command: "ls"}, // ls should exist
			wantIssues: 0,
		},
		{
			name:       "valid remote server",
			server:     &mcpServerInfo{URL: "https://api.example.com"},
			wantIssues: 0,
		},
		{
			name:       "empty server (no command or url)",
			server:     &mcpServerInfo{},
			wantIssues: 1,
			wantTypes:  []string{"transport_mismatch"},
		},
		{
			name:       "both command and url",
			server:     &mcpServerInfo{Command: "ls", URL: "https://api.example.com"},
			wantIssues: 1,
			wantTypes:  []string{"transport_mismatch"},
		},
		{
			name:       "nonexistent command",
			server:     &mcpServerInfo{Command: "this-command-definitely-does-not-exist-xyz123"},
			wantIssues: 1,
			wantTypes:  []string{"missing_command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := c.validateServer("/test/path", "test", "test-server", tt.server)
			if len(issues) != tt.wantIssues {
				t.Errorf("validateServer() got %d issues, want %d", len(issues), tt.wantIssues)
				for _, i := range issues {
					t.Logf("  issue: %s - %s", i.Type, i.Problem)
				}
				return
			}
			if tt.wantTypes != nil {
				for i, wantType := range tt.wantTypes {
					if i < len(issues) && issues[i].Type != wantType {
						t.Errorf("validateServer() issue[%d].Type = %q, want %q", i, issues[i].Type, wantType)
					}
				}
			}
		})
	}
}

func TestConfigSemanticCheck_validateCommand(t *testing.T) {
	c := NewConfigSemanticCheck()
	tempDir := t.TempDir()

	// Create a test executable
	testExec := filepath.Join(tempDir, "test-exec")
	if err := os.WriteFile(testExec, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		command    string
		wantIssues int
	}{
		{
			name:       "command in PATH",
			command:    "ls",
			wantIssues: 0,
		},
		{
			name:       "absolute path exists",
			command:    testExec,
			wantIssues: 0,
		},
		{
			name:       "absolute path not found",
			command:    "/nonexistent/path/to/command",
			wantIssues: 1,
		},
		{
			name:       "command not in PATH",
			command:    "nonexistent-command-xyz123",
			wantIssues: 1,
		},
		{
			name:       "empty command",
			command:    "",
			wantIssues: 0, // empty is handled elsewhere
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := c.validateCommand("/test/path", "test", "test-server", tt.command)
			if len(issues) != tt.wantIssues {
				t.Errorf("validateCommand(%q) got %d issues, want %d", tt.command, len(issues), tt.wantIssues)
			}
		})
	}
}

func TestConfigSemanticCheck_lookPath(t *testing.T) {
	c := NewConfigSemanticCheck()

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "ls should exist",
			command: "ls",
			wantErr: false,
		},
		{
			name:    "nonexistent command",
			command: "definitely-not-a-real-command-xyz123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.lookPath(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("lookPath(%q) error = %v, wantErr %v", tt.command, err, tt.wantErr)
			}
		})
	}
}

func TestConfigSemanticCheck_buildSemanticResult(t *testing.T) {
	c := NewConfigSemanticCheck()

	t.Run("no issues with servers", func(t *testing.T) {
		result := c.buildSemanticResult(nil, 3)
		if result.Status != SeverityPass {
			t.Errorf("buildSemanticResult() status = %v, want %v", result.Status, SeverityPass)
		}
		if result.Details["checked_servers"] != 3 {
			t.Errorf("buildSemanticResult() checked_servers = %v, want 3", result.Details["checked_servers"])
		}
	})

	t.Run("no issues no servers", func(t *testing.T) {
		result := c.buildSemanticResult(nil, 0)
		if result.Status != SeverityInfo {
			t.Errorf("buildSemanticResult() status = %v, want %v", result.Status, SeverityInfo)
		}
	})

	t.Run("with warnings", func(t *testing.T) {
		issues := []semanticIssue{
			{
				Path:     "/test/path",
				Platform: "test",
				Server:   "test-server",
				Type:     "missing_command",
				Problem:  "command not found",
				Severity: SeverityWarning,
			},
		}
		result := c.buildSemanticResult(issues, 1)
		if result.Status != SeverityWarning {
			t.Errorf("buildSemanticResult() status = %v, want %v", result.Status, SeverityWarning)
		}
	})

	t.Run("with errors", func(t *testing.T) {
		issues := []semanticIssue{
			{
				Path:     "/test/path",
				Platform: "test",
				Server:   "test-server",
				Type:     "transport_mismatch",
				Problem:  "no transport configured",
				Severity: SeverityError,
			},
		}
		result := c.buildSemanticResult(issues, 1)
		if result.Status != SeverityError {
			t.Errorf("buildSemanticResult() status = %v, want %v", result.Status, SeverityError)
		}
	})

	t.Run("errors take precedence over warnings", func(t *testing.T) {
		issues := []semanticIssue{
			{
				Path:     "/test/path",
				Platform: "test",
				Type:     "missing_command",
				Problem:  "command not found",
				Severity: SeverityWarning,
			},
			{
				Path:     "/test/path",
				Platform: "test",
				Type:     "transport_mismatch",
				Problem:  "no transport configured",
				Severity: SeverityError,
			},
		}
		result := c.buildSemanticResult(issues, 2)
		if result.Status != SeverityError {
			t.Errorf("buildSemanticResult() status = %v, want %v (errors should take precedence)", result.Status, SeverityError)
		}
	})
}

func TestConfigSemanticCheck_validateMCPConfig(t *testing.T) {
	c := NewConfigSemanticCheck()
	tempDir := t.TempDir()

	t.Run("nonexistent file", func(t *testing.T) {
		issues, count := c.validateMCPConfig(filepath.Join(tempDir, "nonexistent.json"), "test")
		if len(issues) != 0 {
			t.Errorf("validateMCPConfig() got %d issues, want 0 for nonexistent file", len(issues))
		}
		if count != 0 {
			t.Errorf("validateMCPConfig() count = %d, want 0", count)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.json")
		if err := os.WriteFile(emptyFile, []byte(""), 0600); err != nil {
			t.Fatal(err)
		}
		issues, count := c.validateMCPConfig(emptyFile, "test")
		if len(issues) != 0 {
			t.Errorf("validateMCPConfig() got %d issues, want 0 for empty file", len(issues))
		}
		if count != 0 {
			t.Errorf("validateMCPConfig() count = %d, want 0", count)
		}
	})

	t.Run("valid claude config", func(t *testing.T) {
		configFile := filepath.Join(tempDir, "claude.json")
		content := `{"mcpServers": {"test": {"command": "ls"}}}`
		if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		issues, count := c.validateMCPConfig(configFile, "claude")
		if count != 1 {
			t.Errorf("validateMCPConfig() count = %d, want 1", count)
		}
		// Should have no issues since "ls" exists
		for _, issue := range issues {
			if issue.Severity == SeverityError {
				t.Errorf("validateMCPConfig() unexpected error: %s", issue.Problem)
			}
		}
	})

	t.Run("invalid server config", func(t *testing.T) {
		configFile := filepath.Join(tempDir, "invalid.json")
		content := `{"mcpServers": {"broken": {}}}`
		if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		issues, _ := c.validateMCPConfig(configFile, "claude")
		hasTransportError := false
		for _, issue := range issues {
			if issue.Type == "transport_mismatch" && issue.Severity == SeverityError {
				hasTransportError = true
			}
		}
		if !hasTransportError {
			t.Error("validateMCPConfig() expected transport_mismatch error for empty server config")
		}
	})
}
