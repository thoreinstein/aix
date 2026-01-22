package validator

import (
	"errors"
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
		wantErr        error
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
			wantErr:        ErrEmptyConfig,
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
			wantErr:        ErrMissingServerName,
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
			wantErr:        ErrInvalidTransport,
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
			wantErr:        ErrMissingCommand,
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
			wantErr:        ErrMissingURL,
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
			wantErr:        ErrInvalidPlatform,
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
			wantErr:        ErrEmptyEnvKey,
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
			wantErr:        ErrEmptyHeaderKey,
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
			errs := v.Validate(tt.config)

			errCount := len(Errors(errs))
			warnCount := len(Warnings(errs))

			if errCount != tt.wantErrCount {
				t.Errorf("Validate() got %d errors, want %d; errors: %v", errCount, tt.wantErrCount, errs)
			}
			if tt.wantWarnCount > 0 && warnCount != tt.wantWarnCount {
				t.Errorf("Validate() got %d warnings, want %d; errors: %v", warnCount, tt.wantWarnCount, errs)
			}

			// Check for specific error details
			if len(errs) > 0 && (tt.wantServerName != "" || tt.wantField != "" || tt.wantMsgContain != "") {
				found := false
				for _, err := range errs {
					if tt.wantServerName != "" && err.ServerName != tt.wantServerName {
						continue
					}
					if tt.wantField != "" && err.Field != tt.wantField {
						continue
					}
					if tt.wantMsgContain != "" && !strings.Contains(err.Message, tt.wantMsgContain) {
						continue
					}
					found = true
					break
				}
				if !found {
					t.Errorf("expected error with server=%q field=%q msg containing %q, got: %v",
						tt.wantServerName, tt.wantField, tt.wantMsgContain, errs)
				}
			}

			// Check for specific sentinel error
			if tt.wantErr != nil && len(errs) > 0 {
				found := false
				for _, err := range errs {
					if errors.Is(err, tt.wantErr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error %v, got: %v", tt.wantErr, errs)
				}
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ValidationError
		want string
	}{
		{
			name: "error with server and field",
			err: &ValidationError{
				ServerName: "myserver",
				Field:      "command",
				Message:    "is required",
				Severity:   SeverityError,
			},
			want: `error: server "myserver" field "command": is required`,
		},
		{
			name: "error with server only",
			err: &ValidationError{
				ServerName: "myserver",
				Message:    "is misconfigured",
				Severity:   SeverityError,
			},
			want: `error: server "myserver": is misconfigured`,
		},
		{
			name: "error with field only",
			err: &ValidationError{
				Field:    "servers",
				Message:  "must not be empty",
				Severity: SeverityError,
			},
			want: `error: field "servers": must not be empty`,
		},
		{
			name: "error with message only",
			err: &ValidationError{
				Message:  "config is invalid",
				Severity: SeverityError,
			},
			want: "error: config is invalid",
		},
		{
			name: "warning with server and field",
			err: &ValidationError{
				ServerName: "myserver",
				Field:      "transport",
				Message:    "should be explicit",
				Severity:   SeverityWarning,
			},
			want: `warning: server "myserver" field "transport": should be explicit`,
		},
		{
			name: "warning with message only",
			err: &ValidationError{
				Message:  "config may have issues",
				Severity: SeverityWarning,
			},
			want: "warning: config may have issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	underlying := ErrMissingCommand
	err := &ValidationError{
		ServerName: "test",
		Field:      "command",
		Message:    "stdio transport requires command",
		Severity:   SeverityError,
		Err:        underlying,
	}

	if !errors.Is(err, underlying) {
		t.Error("Unwrap() should allow errors.Is to match underlying error")
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name string
		errs []*ValidationError
		want bool
	}{
		{
			name: "nil slice",
			errs: nil,
			want: false,
		},
		{
			name: "empty slice",
			errs: []*ValidationError{},
			want: false,
		},
		{
			name: "only warnings",
			errs: []*ValidationError{
				{Severity: SeverityWarning, Message: "warn1"},
				{Severity: SeverityWarning, Message: "warn2"},
			},
			want: false,
		},
		{
			name: "only errors",
			errs: []*ValidationError{
				{Severity: SeverityError, Message: "err1"},
			},
			want: true,
		},
		{
			name: "mixed",
			errs: []*ValidationError{
				{Severity: SeverityWarning, Message: "warn1"},
				{Severity: SeverityError, Message: "err1"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasErrors(tt.errs); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasWarnings(t *testing.T) {
	tests := []struct {
		name string
		errs []*ValidationError
		want bool
	}{
		{
			name: "nil slice",
			errs: nil,
			want: false,
		},
		{
			name: "only errors",
			errs: []*ValidationError{
				{Severity: SeverityError, Message: "err1"},
			},
			want: false,
		},
		{
			name: "only warnings",
			errs: []*ValidationError{
				{Severity: SeverityWarning, Message: "warn1"},
			},
			want: true,
		},
		{
			name: "mixed",
			errs: []*ValidationError{
				{Severity: SeverityError, Message: "err1"},
				{Severity: SeverityWarning, Message: "warn1"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasWarnings(tt.errs); got != tt.want {
				t.Errorf("HasWarnings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	errs := []*ValidationError{
		{Severity: SeverityError, Message: "err1"},
		{Severity: SeverityWarning, Message: "warn1"},
		{Severity: SeverityError, Message: "err2"},
		{Severity: SeverityWarning, Message: "warn2"},
	}

	result := Errors(errs)
	if len(result) != 2 {
		t.Errorf("Errors() returned %d items, want 2", len(result))
	}
	for _, e := range result {
		if e.Severity != SeverityError {
			t.Errorf("Errors() returned non-error: %v", e)
		}
	}
}

func TestWarnings(t *testing.T) {
	errs := []*ValidationError{
		{Severity: SeverityError, Message: "err1"},
		{Severity: SeverityWarning, Message: "warn1"},
		{Severity: SeverityError, Message: "err2"},
		{Severity: SeverityWarning, Message: "warn2"},
	}

	result := Warnings(errs)
	if len(result) != 2 {
		t.Errorf("Warnings() returned %d items, want 2", len(result))
	}
	for _, w := range result {
		if w.Severity != SeverityWarning {
			t.Errorf("Warnings() returned non-warning: %v", w)
		}
	}
}

func TestNew_Options(t *testing.T) {
	t.Run("default does not allow empty", func(t *testing.T) {
		v := New()
		errs := v.Validate(&mcp.Config{Servers: map[string]*mcp.Server{}})
		if len(errs) == 0 {
			t.Error("default validator should not allow empty config")
		}
	})

	t.Run("WithAllowEmpty(true) allows empty", func(t *testing.T) {
		v := New(WithAllowEmpty(true))
		errs := v.Validate(&mcp.Config{Servers: map[string]*mcp.Server{}})
		if len(errs) != 0 {
			t.Errorf("validator with allowEmpty should allow empty config, got errors: %v", errs)
		}
	})

	t.Run("WithAllowEmpty(false) does not allow empty", func(t *testing.T) {
		v := New(WithAllowEmpty(false))
		errs := v.Validate(&mcp.Config{Servers: map[string]*mcp.Server{}})
		if len(errs) == 0 {
			t.Error("validator with !allowEmpty should not allow empty config")
		}
	})
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{Severity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("Severity.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
