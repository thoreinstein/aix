package validator

import (
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/mcp"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name           string
		config         *mcp.Config
		allowEmpty     bool
		wantErrCount   int
		wantWarnCount  int
		wantServerName string
		wantField      string
		wantMsgContain string
	}{
		// Valid configs
		{
			name: "valid stdio server",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:    "test",
						Command: "test-cmd",
					},
				},
			},
			wantErrCount:  0,
			wantWarnCount: 0,
		},
		{
			name: "valid stdio server with explicit transport",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Command:   "test-cmd",
						Transport: mcp.TransportStdio,
					},
				},
			},
			wantErrCount:  0,
			wantWarnCount: 0,
		},
		{
			name: "valid sse server",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"remote": {
						Name:      "remote",
						URL:       "https://api.example.com/mcp",
						Transport: mcp.TransportSSE,
					},
				},
			},
			wantErrCount:  0,
			wantWarnCount: 0,
		},
		{
			name: "valid sse server with URL only (no explicit transport)",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"remote": {
						Name: "remote",
						URL:  "https://api.example.com/mcp",
					},
				},
			},
			wantErrCount:  0,
			wantWarnCount: 0,
		},
		{
			name: "valid server with all fields",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"full": {
						Name:      "full",
						Command:   "cmd",
						Args:      []string{"--flag"},
						Transport: mcp.TransportStdio,
						Env:       map[string]string{"KEY": "value"},
						Platforms: []string{"darwin", "linux"},
					},
				},
			},
			wantErrCount:  0,
			wantWarnCount: 0,
		},
		{
			name: "valid multiple servers",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"local":  {Name: "local", Command: "cmd1"},
					"remote": {Name: "remote", URL: "http://localhost:8080", Transport: mcp.TransportSSE},
				},
			},
			wantErrCount:  0,
			wantWarnCount: 0,
		},

		// Empty config
		{
			name: "empty config not allowed",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{},
			},
			allowEmpty:     false,
			wantErrCount:   1,
			wantMsgContain: "no servers",
		},
		{
			name: "empty config allowed with option",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{},
			},
			allowEmpty:   true,
			wantErrCount: 0,
		},
		{
			name:           "nil config",
			config:         nil,
			wantErrCount:   1,
			wantMsgContain: "nil",
		},

		// Name validation
		{
			name: "missing server name",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Command: "cmd",
					},
				},
			},
			wantErrCount:   1,
			wantServerName: "test",
			wantField:      "name",
			wantMsgContain: "required",
		},

		// Transport validation
		{
			name: "invalid transport value",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Command:   "cmd",
						Transport: "invalid",
					},
				},
			},
			wantErrCount:   1,
			wantServerName: "test",
			wantField:      "transport",
			wantMsgContain: "stdio",
		},

		// Command/URL validation
		{
			name: "stdio transport missing command",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Transport: mcp.TransportStdio,
					},
				},
			},
			wantErrCount:   1,
			wantServerName: "test",
			wantField:      "command",
			wantMsgContain: "requires command",
		},
		{
			name: "sse transport missing URL",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Transport: mcp.TransportSSE,
					},
				},
			},
			wantErrCount:   1,
			wantServerName: "test",
			wantField:      "url",
			wantMsgContain: "requires URL",
		},
		{
			name: "no command or URL with no transport",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name: "test",
					},
				},
			},
			wantErrCount:   1,
			wantServerName: "test",
			wantMsgContain: "command",
		},
		{
			name: "both command and URL generates warning",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:    "test",
						Command: "cmd",
						URL:     "http://localhost:8080",
					},
				},
			},
			wantErrCount:   0,
			wantWarnCount:  1,
			wantServerName: "test",
			wantMsgContain: "both command and URL",
		},
		{
			name: "both command and URL with explicit stdio transport",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Command:   "cmd",
						URL:       "http://localhost:8080",
						Transport: mcp.TransportStdio,
					},
				},
			},
			wantErrCount:   0,
			wantWarnCount:  1,
			wantMsgContain: "command will be used",
		},
		{
			name: "both command and URL with explicit sse transport",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Command:   "cmd",
						URL:       "http://localhost:8080",
						Transport: mcp.TransportSSE,
					},
				},
			},
			wantErrCount:   0,
			wantWarnCount:  1,
			wantMsgContain: "URL will be used",
		},

		// Platform validation
		{
			name: "valid platforms",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Command:   "cmd",
						Platforms: []string{"darwin", "linux", "windows"},
					},
				},
			},
			wantErrCount: 0,
		},
		{
			name: "invalid platform",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						Command:   "cmd",
						Platforms: []string{"darwin", "invalid"},
					},
				},
			},
			wantErrCount:   1,
			wantServerName: "test",
			wantField:      "platforms",
			wantMsgContain: "invalid platform",
		},

		// Env validation
		{
			name: "valid env",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:    "test",
						Command: "cmd",
						Env:     map[string]string{"KEY": "value", "KEY2": ""},
					},
				},
			},
			wantErrCount: 0,
		},
		{
			name: "empty env key",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:    "test",
						Command: "cmd",
						Env:     map[string]string{"": "value"},
					},
				},
			},
			wantErrCount:   1,
			wantServerName: "test",
			wantField:      "env",
			wantMsgContain: "empty",
		},

		// Headers validation
		{
			name: "valid headers",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						URL:       "http://localhost:8080",
						Transport: mcp.TransportSSE,
						Headers:   map[string]string{"Authorization": "Bearer token"},
					},
				},
			},
			wantErrCount: 0,
		},
		{
			name: "empty header key",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"test": {
						Name:      "test",
						URL:       "http://localhost:8080",
						Transport: mcp.TransportSSE,
						Headers:   map[string]string{"": "value"},
					},
				},
			},
			wantErrCount:   1,
			wantServerName: "test",
			wantField:      "headers",
			wantMsgContain: "empty",
		},

		// Multiple errors
		{
			name: "multiple errors collected",
			config: &mcp.Config{
				Servers: map[string]*mcp.Server{
					"bad1": {
						Command: "cmd", // missing name
					},
					"bad2": {
						Name: "bad2", // missing command/url
					},
				},
			},
			wantErrCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []Option{}
			if tt.allowEmpty {
				opts = append(opts, WithAllowEmpty(true))
			}
			v := New(opts...)
			result := v.Validate(tt.config)

			errCount := len(result.Errors())
			warnCount := len(result.Warnings())

			if errCount != tt.wantErrCount {
				t.Errorf("Validate() got %d errors, want %d; issues: %v", errCount, tt.wantErrCount, result.Issues)
			}
			if tt.wantWarnCount > 0 && warnCount != tt.wantWarnCount {
				t.Errorf("Validate() got %d warnings, want %d; issues: %v", warnCount, tt.wantWarnCount, result.Issues)
			}

			// Check for specific error details
			if len(result.Issues) > 0 && (tt.wantServerName != "" || tt.wantField != "" || tt.wantMsgContain != "") {
				found := false
				for _, issue := range result.Issues {
					if tt.wantServerName != "" && issue.Context["server"] != tt.wantServerName {
						continue
					}
					if tt.wantField != "" && issue.Field != tt.wantField {
						continue
					}
					if tt.wantMsgContain != "" && !strings.Contains(issue.Message, tt.wantMsgContain) {
						continue
					}
					found = true
					break
				}
				if !found {
					t.Errorf("expected error with server=%q field=%q msg containing %q, got: %v",
						tt.wantServerName, tt.wantField, tt.wantMsgContain, result.Issues)
				}
			}
		})
	}
}

func TestNew_Options(t *testing.T) {
	t.Run("default does not allow empty", func(t *testing.T) {
		v := New()
		result := v.Validate(&mcp.Config{Servers: map[string]*mcp.Server{}})
		if !result.HasErrors() {
			t.Error("default validator should not allow empty config")
		}
	})

	t.Run("WithAllowEmpty(true) allows empty", func(t *testing.T) {
		v := New(WithAllowEmpty(true))
		result := v.Validate(&mcp.Config{Servers: map[string]*mcp.Server{}})
		if result.HasErrors() {
			t.Errorf("validator with allowEmpty should allow empty config, got errors: %v", result.Issues)
		}
	})

	t.Run("WithAllowEmpty(false) does not allow empty", func(t *testing.T) {
		v := New(WithAllowEmpty(false))
		result := v.Validate(&mcp.Config{Servers: map[string]*mcp.Server{}})
		if !result.HasErrors() {
			t.Error("validator with !allowEmpty should not allow empty config")
		}
	})
}
