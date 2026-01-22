package parser

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/thoreinstein/aix/internal/mcp"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		checkConfig func(t *testing.T, cfg *mcp.Config)
	}{
		{
			name:    "empty input returns empty config",
			input:   "",
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				if cfg == nil {
					t.Fatal("expected non-nil config")
				}
				if len(cfg.Servers) != 0 {
					t.Errorf("Servers len = %d, want 0", len(cfg.Servers))
				}
			},
		},
		{
			name:    "empty JSON object returns empty config",
			input:   "{}",
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				if cfg.Servers == nil {
					t.Error("Servers should be initialized, not nil")
				}
			},
		},
		{
			name:    "config with no servers",
			input:   `{"servers": {}}`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				if len(cfg.Servers) != 0 {
					t.Errorf("Servers len = %d, want 0", len(cfg.Servers))
				}
			},
		},
		{
			name: "config with single stdio server",
			input: `{
				"servers": {
					"github": {
						"name": "github",
						"command": "npx",
						"args": ["-y", "@modelcontextprotocol/server-github"],
						"env": {"GITHUB_TOKEN": "${GITHUB_TOKEN}"}
					}
				}
			}`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				if len(cfg.Servers) != 1 {
					t.Errorf("Servers len = %d, want 1", len(cfg.Servers))
				}
				server, ok := cfg.Servers["github"]
				if !ok {
					t.Fatal("expected server 'github'")
				}
				if server.Name != "github" {
					t.Errorf("Name = %q, want %q", server.Name, "github")
				}
				if server.Command != "npx" {
					t.Errorf("Command = %q, want %q", server.Command, "npx")
				}
				wantArgs := []string{"-y", "@modelcontextprotocol/server-github"}
				if !reflect.DeepEqual(server.Args, wantArgs) {
					t.Errorf("Args = %v, want %v", server.Args, wantArgs)
				}
				if server.Env["GITHUB_TOKEN"] != "${GITHUB_TOKEN}" {
					t.Errorf("Env[GITHUB_TOKEN] = %q, want %q", server.Env["GITHUB_TOKEN"], "${GITHUB_TOKEN}")
				}
			},
		},
		{
			name: "config with SSE server",
			input: `{
				"servers": {
					"remote": {
						"name": "remote",
						"url": "https://api.example.com/mcp",
						"transport": "sse",
						"headers": {"Authorization": "Bearer token"}
					}
				}
			}`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				server := cfg.Servers["remote"]
				if server == nil {
					t.Fatal("expected server 'remote'")
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
			name: "config with multiple servers",
			input: `{
				"servers": {
					"server1": {"name": "server1", "command": "cmd1"},
					"server2": {"name": "server2", "command": "cmd2"},
					"server3": {"name": "server3", "url": "http://localhost:8080", "transport": "sse"}
				}
			}`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				if len(cfg.Servers) != 3 {
					t.Errorf("Servers len = %d, want 3", len(cfg.Servers))
				}
			},
		},
		{
			name: "config with disabled server",
			input: `{
				"servers": {
					"disabled": {"name": "disabled", "command": "cmd", "disabled": true}
				}
			}`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				if !cfg.Servers["disabled"].Disabled {
					t.Error("Disabled should be true")
				}
			},
		},
		{
			name: "config with platforms",
			input: `{
				"servers": {
					"linux-only": {"name": "linux-only", "command": "cmd", "platforms": ["linux"]}
				}
			}`,
			wantErr: false,
			checkConfig: func(t *testing.T, cfg *mcp.Config) {
				t.Helper()
				wantPlatforms := []string{"linux"}
				if !reflect.DeepEqual(cfg.Servers["linux-only"].Platforms, wantPlatforms) {
					t.Errorf("Platforms = %v, want %v", cfg.Servers["linux-only"].Platforms, wantPlatforms)
				}
			},
		},
		{
			name:        "invalid JSON syntax",
			input:       `{"servers": {`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name:        "invalid JSON type for servers",
			input:       `{"servers": "not-an-object"}`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name:        "invalid JSON array instead of object",
			input:       `["not", "an", "object"]`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Fatal("Parse() expected error, got nil")
				}
				if tt.errContains != "" && !errors.Is(err, ErrInvalidJSON) {
					t.Errorf("expected ErrInvalidJSON, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if cfg == nil {
				t.Fatal("Parse() returned nil config")
			}
			if tt.checkConfig != nil {
				tt.checkConfig(t, cfg)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	t.Run("parses existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "mcp.json")

		content := `{
			"servers": {
				"test": {"name": "test", "command": "test-cmd"}
			}
		}`
		if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
			t.Fatalf("writing test file: %v", err)
		}

		cfg, err := ParseFile(cfgPath)
		if err != nil {
			t.Fatalf("ParseFile() error = %v", err)
		}
		if cfg.Servers["test"].Command != "test-cmd" {
			t.Errorf("Command = %q, want %q", cfg.Servers["test"].Command, "test-cmd")
		}
	})

	t.Run("returns empty config for missing file", func(t *testing.T) {
		cfg, err := ParseFile("/nonexistent/path/mcp.json")
		if err != nil {
			t.Fatalf("ParseFile() error = %v, expected nil for missing file", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if len(cfg.Servers) != 0 {
			t.Errorf("expected empty Servers, got %d", len(cfg.Servers))
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "mcp.json")

		if err := os.WriteFile(cfgPath, []byte(`{invalid`), 0o644); err != nil {
			t.Fatalf("writing test file: %v", err)
		}

		_, err := ParseFile(cfgPath)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("expected *ParseError, got %T", err)
		}
		if parseErr.Path != cfgPath {
			t.Errorf("ParseError.Path = %q, want %q", parseErr.Path, cfgPath)
		}
	})

	t.Run("returns error for permission denied", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "mcp.json")

		// Create file with no read permissions
		if err := os.WriteFile(cfgPath, []byte(`{}`), 0o000); err != nil {
			t.Fatalf("writing test file: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chmod(cfgPath, 0o644) // Restore permissions for cleanup
		})

		_, err := ParseFile(cfgPath)
		if err == nil {
			t.Fatal("expected error for permission denied")
		}
		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("expected *ParseError, got %T", err)
		}
	})
}

func TestWrite(t *testing.T) {
	t.Run("writes empty config", func(t *testing.T) {
		cfg := mcp.NewConfig()
		data, err := Write(cfg)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Verify it's valid JSON
		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
	})

	t.Run("writes nil config as empty", func(t *testing.T) {
		data, err := Write(nil)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
	})

	t.Run("produces indented output", func(t *testing.T) {
		cfg := &mcp.Config{
			Servers: map[string]*mcp.Server{
				"test": {Name: "test", Command: "cmd"},
			},
		}
		data, err := Write(cfg)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Check for indentation (should have newlines and spaces)
		output := string(data)
		if len(output) < 20 {
			t.Errorf("output seems too short to be indented: %q", output)
		}
		// Indented JSON should have multiple lines
		lines := 0
		for _, c := range output {
			if c == '\n' {
				lines++
			}
		}
		if lines < 3 {
			t.Errorf("expected multiple lines in indented output, got %d", lines)
		}
	})

	t.Run("ends with newline", func(t *testing.T) {
		cfg := mcp.NewConfig()
		data, err := Write(cfg)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		if len(data) == 0 || data[len(data)-1] != '\n' {
			t.Error("output should end with newline")
		}
	})

	t.Run("preserves all server fields", func(t *testing.T) {
		cfg := &mcp.Config{
			Servers: map[string]*mcp.Server{
				"full": {
					Name:      "full",
					Command:   "cmd",
					Args:      []string{"--flag", "value"},
					URL:       "http://localhost:8080",
					Transport: mcp.TransportStdio,
					Env:       map[string]string{"KEY": "value"},
					Headers:   map[string]string{"X-Custom": "header"},
					Platforms: []string{"darwin", "linux"},
					Disabled:  true,
				},
			},
		}

		data, err := Write(cfg)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Parse back and verify
		parsed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		server := parsed.Servers["full"]
		if server.Name != "full" {
			t.Errorf("Name = %q, want %q", server.Name, "full")
		}
		if server.Command != "cmd" {
			t.Errorf("Command = %q, want %q", server.Command, "cmd")
		}
		if !reflect.DeepEqual(server.Args, []string{"--flag", "value"}) {
			t.Errorf("Args = %v, want %v", server.Args, []string{"--flag", "value"})
		}
		if server.URL != "http://localhost:8080" {
			t.Errorf("URL = %q, want %q", server.URL, "http://localhost:8080")
		}
		if server.Transport != mcp.TransportStdio {
			t.Errorf("Transport = %q, want %q", server.Transport, mcp.TransportStdio)
		}
		if server.Env["KEY"] != "value" {
			t.Errorf("Env[KEY] = %q, want %q", server.Env["KEY"], "value")
		}
		if server.Headers["X-Custom"] != "header" {
			t.Errorf("Headers[X-Custom] = %q, want %q", server.Headers["X-Custom"], "header")
		}
		if !reflect.DeepEqual(server.Platforms, []string{"darwin", "linux"}) {
			t.Errorf("Platforms = %v, want %v", server.Platforms, []string{"darwin", "linux"})
		}
		if !server.Disabled {
			t.Error("Disabled should be true")
		}
	})
}

func TestWriteFile(t *testing.T) {
	t.Run("writes to new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "mcp.json")

		cfg := &mcp.Config{
			Servers: map[string]*mcp.Server{
				"test": {Name: "test", Command: "cmd"},
			},
		}

		if err := WriteFile(cfgPath, cfg); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		// Verify file exists and is readable
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatalf("reading written file: %v", err)
		}

		// Verify content is valid
		parsed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if parsed.Servers["test"].Command != "cmd" {
			t.Errorf("Command = %q, want %q", parsed.Servers["test"].Command, "cmd")
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "mcp.json")

		// Write initial content
		if err := os.WriteFile(cfgPath, []byte(`{"servers":{}}`), 0o644); err != nil {
			t.Fatalf("writing initial file: %v", err)
		}

		// Overwrite with new content
		cfg := &mcp.Config{
			Servers: map[string]*mcp.Server{
				"new": {Name: "new", Command: "new-cmd"},
			},
		}

		if err := WriteFile(cfgPath, cfg); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		// Verify new content
		parsed, err := ParseFile(cfgPath)
		if err != nil {
			t.Fatalf("ParseFile() error = %v", err)
		}
		if _, ok := parsed.Servers["new"]; !ok {
			t.Error("expected server 'new' after overwrite")
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "nested", "deep", "mcp.json")

		cfg := mcp.NewConfig()
		if err := WriteFile(cfgPath, cfg); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if _, err := os.Stat(cfgPath); err != nil {
			t.Errorf("file should exist: %v", err)
		}
	})

	t.Run("atomic write - no partial files on error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := filepath.Join(tmpDir, "mcp.json")

		// Write initial good content
		goodCfg := &mcp.Config{
			Servers: map[string]*mcp.Server{
				"good": {Name: "good", Command: "cmd"},
			},
		}
		if err := WriteFile(cfgPath, goodCfg); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		// Read original content
		original, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatalf("reading original: %v", err)
		}

		// Make directory read-only to prevent temp file creation
		// This should fail the write, but the original file should remain intact
		_ = os.Chmod(tmpDir, 0o555)
		t.Cleanup(func() {
			_ = os.Chmod(tmpDir, 0o755)
		})

		newCfg := &mcp.Config{
			Servers: map[string]*mcp.Server{
				"new": {Name: "new", Command: "new-cmd"},
			},
		}

		// This should fail
		err = WriteFile(cfgPath, newCfg)
		if err == nil {
			// If it didn't fail (maybe running as root), skip the rest
			t.Skip("write succeeded, skipping atomic check")
		}

		// Restore permissions to read the file
		_ = os.Chmod(tmpDir, 0o755)

		// Verify original content is still intact
		current, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatalf("reading after failed write: %v", err)
		}
		if string(current) != string(original) {
			t.Error("original file content was corrupted after failed write")
		}
	})

	t.Run("returns ParseError with path", func(t *testing.T) {
		// Try to write to a path that will fail
		err := WriteFile("/nonexistent/readonly/mcp.json", mcp.NewConfig())
		if err == nil {
			t.Fatal("expected error")
		}

		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("expected *ParseError, got %T", err)
		}
	})
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		cfg  *mcp.Config
	}{
		{
			name: "empty config",
			cfg:  mcp.NewConfig(),
		},
		{
			name: "single stdio server",
			cfg: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"github": {
						Name:      "github",
						Command:   "npx",
						Args:      []string{"-y", "@modelcontextprotocol/server-github"},
						Transport: mcp.TransportStdio,
						Env:       map[string]string{"GITHUB_TOKEN": "${GITHUB_TOKEN}"},
					},
				},
			},
		},
		{
			name: "single SSE server",
			cfg: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"remote": {
						Name:      "remote",
						URL:       "https://api.example.com/mcp",
						Transport: mcp.TransportSSE,
						Headers:   map[string]string{"Authorization": "Bearer ${TOKEN}"},
					},
				},
			},
		},
		{
			name: "multiple mixed servers",
			cfg: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"local1": {
						Name:    "local1",
						Command: "cmd1",
						Args:    []string{"--arg1"},
					},
					"local2": {
						Name:      "local2",
						Command:   "cmd2",
						Transport: mcp.TransportStdio,
						Env:       map[string]string{"VAR": "value"},
						Platforms: []string{"darwin"},
					},
					"remote1": {
						Name:      "remote1",
						URL:       "http://localhost:8080",
						Transport: mcp.TransportSSE,
						Disabled:  true,
					},
				},
			},
		},
		{
			name: "server with all fields",
			cfg: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"full": {
						Name:      "full",
						Command:   "full-cmd",
						Args:      []string{"--verbose", "--config=/etc/app.conf"},
						URL:       "http://backup.example.com",
						Transport: mcp.TransportStdio,
						Env: map[string]string{
							"KEY1": "value1",
							"KEY2": "value2",
						},
						Headers: map[string]string{
							"X-Custom-1": "header1",
							"X-Custom-2": "header2",
						},
						Platforms: []string{"darwin", "linux", "windows"},
						Disabled:  false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write to bytes
			data, err := Write(tt.cfg)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			// Parse back
			parsed, err := Parse(data)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Compare server count
			if len(parsed.Servers) != len(tt.cfg.Servers) {
				t.Errorf("Servers count = %d, want %d", len(parsed.Servers), len(tt.cfg.Servers))
			}

			// Compare each server
			for name, want := range tt.cfg.Servers {
				got, ok := parsed.Servers[name]
				if !ok {
					t.Errorf("server %q missing after round-trip", name)
					continue
				}

				if got.Name != want.Name {
					t.Errorf("server %q Name = %q, want %q", name, got.Name, want.Name)
				}
				if got.Command != want.Command {
					t.Errorf("server %q Command = %q, want %q", name, got.Command, want.Command)
				}
				if !reflect.DeepEqual(got.Args, want.Args) {
					t.Errorf("server %q Args = %v, want %v", name, got.Args, want.Args)
				}
				if got.URL != want.URL {
					t.Errorf("server %q URL = %q, want %q", name, got.URL, want.URL)
				}
				if got.Transport != want.Transport {
					t.Errorf("server %q Transport = %q, want %q", name, got.Transport, want.Transport)
				}
				if !reflect.DeepEqual(got.Env, want.Env) {
					t.Errorf("server %q Env = %v, want %v", name, got.Env, want.Env)
				}
				if !reflect.DeepEqual(got.Headers, want.Headers) {
					t.Errorf("server %q Headers = %v, want %v", name, got.Headers, want.Headers)
				}
				if !reflect.DeepEqual(got.Platforms, want.Platforms) {
					t.Errorf("server %q Platforms = %v, want %v", name, got.Platforms, want.Platforms)
				}
				if got.Disabled != want.Disabled {
					t.Errorf("server %q Disabled = %v, want %v", name, got.Disabled, want.Disabled)
				}
			}
		})
	}
}

func TestRoundTripFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "mcp.json")

	original := &mcp.Config{
		Servers: map[string]*mcp.Server{
			"github": {
				Name:      "github",
				Command:   "npx",
				Args:      []string{"-y", "@modelcontextprotocol/server-github"},
				Transport: mcp.TransportStdio,
				Env:       map[string]string{"GITHUB_TOKEN": "${GITHUB_TOKEN}"},
			},
			"remote": {
				Name:      "remote",
				URL:       "https://api.example.com/mcp",
				Transport: mcp.TransportSSE,
				Headers:   map[string]string{"Authorization": "Bearer ${TOKEN}"},
			},
		},
	}

	// Write to file
	if err := WriteFile(cfgPath, original); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Read back
	parsed, err := ParseFile(cfgPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	// Verify server count
	if len(parsed.Servers) != len(original.Servers) {
		t.Errorf("Servers count = %d, want %d", len(parsed.Servers), len(original.Servers))
	}

	// Verify specific servers
	if parsed.Servers["github"].Command != "npx" {
		t.Errorf("github.Command = %q, want %q", parsed.Servers["github"].Command, "npx")
	}
	if parsed.Servers["remote"].URL != "https://api.example.com/mcp" {
		t.Errorf("remote.URL = %q, want %q", parsed.Servers["remote"].URL, "https://api.example.com/mcp")
	}
}

func TestParseError(t *testing.T) {
	t.Run("formats with path", func(t *testing.T) {
		err := &ParseError{
			Path: "/some/path/mcp.json",
			Err:  errors.New("underlying error"),
		}
		expected := "parsing MCP config /some/path/mcp.json: underlying error"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("formats without path", func(t *testing.T) {
		err := &ParseError{
			Err: errors.New("underlying error"),
		}
		expected := "parsing MCP config: underlying error"
		if err.Error() != expected {
			t.Errorf("Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := &ParseError{
			Path: "/path.json",
			Err:  underlying,
		}
		if !errors.Is(err, underlying) {
			t.Error("Unwrap() should allow errors.Is to match underlying error")
		}
	})
}

func TestUnknownFieldsPreservation(t *testing.T) {
	input := `{
		"servers": {
			"test": {
				"name": "test",
				"command": "cmd",
				"futureServerField": "preserved"
			}
		},
		"futureConfigField": "also preserved",
		"version": "2.0"
	}`

	// Parse
	cfg, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Write back
	data, err := Write(cfg)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Check that unknown fields are preserved
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal to map error = %v", err)
	}

	// Config-level unknown fields
	if _, ok := result["futureConfigField"]; !ok {
		t.Error("futureConfigField not preserved")
	}
	if _, ok := result["version"]; !ok {
		t.Error("version not preserved")
	}

	// Server-level unknown fields
	servers, ok := result["servers"].(map[string]any)
	if !ok {
		t.Fatalf("servers is not a map: %T", result["servers"])
	}
	server, ok := servers["test"].(map[string]any)
	if !ok {
		t.Fatalf("test server is not a map: %T", servers["test"])
	}
	if _, ok := server["futureServerField"]; !ok {
		t.Error("futureServerField not preserved in server")
	}
}
