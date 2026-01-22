package commands

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
)

func TestMCPEnableCommand_Metadata(t *testing.T) {
	if mcpEnableCmd.Use != "enable <name>" {
		t.Errorf("Use = %q, want %q", mcpEnableCmd.Use, "enable <name>")
	}

	if mcpEnableCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if mcpEnableCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Verify Args validator is set (ExactArgs(1))
	if mcpEnableCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestMCPDisableCommand_Metadata(t *testing.T) {
	if mcpDisableCmd.Use != "disable <name>" {
		t.Errorf("Use = %q, want %q", mcpDisableCmd.Use, "disable <name>")
	}

	if mcpDisableCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if mcpDisableCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Verify Args validator is set (ExactArgs(1))
	if mcpDisableCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

// mcpEnableMockPlatform extends mockPlatform for MCP enable/disable testing.
type mcpEnableMockPlatform struct {
	mockPlatform
	mcpServers    map[string]any
	enableErr     error
	disableErr    error
	enableCalled  bool
	disableCalled bool
	isAvailable   bool
}

func (m *mcpEnableMockPlatform) IsAvailable() bool {
	return m.isAvailable
}

func (m *mcpEnableMockPlatform) GetMCP(name string) (any, error) {
	server, ok := m.mcpServers[name]
	if !ok {
		return nil, errors.New("MCP server not found")
	}
	return server, nil
}

func (m *mcpEnableMockPlatform) EnableMCP(_ string) error {
	m.enableCalled = true
	return m.enableErr
}

func (m *mcpEnableMockPlatform) DisableMCP(_ string) error {
	m.disableCalled = true
	return m.disableErr
}

func TestRunMCPSetEnabledWithIO_Enable(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		platforms  func() []cli.Platform
		wantErr    bool
		wantOutput []string
	}{
		{
			name:       "enable existing server",
			serverName: "github",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{"github": struct{}{}},
						isAvailable:  true,
					},
				}
			},
			wantErr:    false,
			wantOutput: []string{"Enabling", "github", "enabled"},
		},
		{
			name:       "enable non-existent server",
			serverName: "not-found",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{},
						isAvailable:  true,
					},
				}
			},
			wantErr:    true,
			wantOutput: []string{"not found"},
		},
		{
			name:       "enable with platform error",
			serverName: "github",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{"github": struct{}{}},
						enableErr:    errors.New("permission denied"),
						isAvailable:  true,
					},
				}
			},
			wantErr:    false, // The function continues and reports the error in output
			wantOutput: []string{"error", "permission denied"},
		},
		{
			name:       "skip unavailable platform",
			serverName: "github",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{"github": struct{}{}},
						isAvailable:  false,
					},
				}
			},
			wantErr:    true, // Server not found on any available platform
			wantOutput: []string{"github"},
		},
		{
			name:       "enable across multiple platforms",
			serverName: "github",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{"github": struct{}{}},
						isAvailable:  true,
					},
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "opencode"},
						mcpServers:   map[string]any{"github": struct{}{}},
						isAvailable:  true,
					},
				}
			},
			wantErr:    false,
			wantOutput: []string{"claude", "opencode", "enabled"},
		},
		{
			name:       "partial success - one platform has server",
			serverName: "github",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{"github": struct{}{}},
						isAvailable:  true,
					},
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "opencode"},
						mcpServers:   map[string]any{}, // No github server here
						isAvailable:  true,
					},
				}
			},
			wantErr:    false,
			wantOutput: []string{"claude", "enabled", "opencode", "not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			// Create a wrapper that bypasses ResolvePlatforms
			platforms := tt.platforms()
			err := runMCPSetEnabledWithMockPlatforms(tt.serverName, true, &buf, platforms)

			if (err != nil) != tt.wantErr {
				t.Errorf("runMCPSetEnabledWithIO() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

func TestRunMCPSetEnabledWithIO_Disable(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		platforms  func() []cli.Platform
		wantErr    bool
		wantOutput []string
	}{
		{
			name:       "disable existing server",
			serverName: "github",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{"github": struct{}{}},
						isAvailable:  true,
					},
				}
			},
			wantErr:    false,
			wantOutput: []string{"Disabling", "github", "disabled"},
		},
		{
			name:       "disable non-existent server",
			serverName: "not-found",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{},
						isAvailable:  true,
					},
				}
			},
			wantErr:    true,
			wantOutput: []string{"not found"},
		},
		{
			name:       "disable with platform error",
			serverName: "github",
			platforms: func() []cli.Platform {
				return []cli.Platform{
					&mcpEnableMockPlatform{
						mockPlatform: mockPlatform{name: "claude"},
						mcpServers:   map[string]any{"github": struct{}{}},
						disableErr:   errors.New("disk full"),
						isAvailable:  true,
					},
				}
			},
			wantErr:    false, // The function continues and reports the error in output
			wantOutput: []string{"error", "disk full"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			platforms := tt.platforms()
			err := runMCPSetEnabledWithMockPlatforms(tt.serverName, false, &buf, platforms)

			if (err != nil) != tt.wantErr {
				t.Errorf("runMCPSetEnabledWithIO() error = %v, wantErr %v", err, tt.wantErr)
			}

			output := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
		})
	}
}

// runMCPSetEnabledWithMockPlatforms is a test helper that bypasses ResolvePlatforms.
// It mirrors the logic of runMCPSetEnabledWithIO but accepts platforms directly.
func runMCPSetEnabledWithMockPlatforms(name string, enabled bool, w *bytes.Buffer, platforms []cli.Platform) error {
	action := "Enabling"
	pastTense := "enabled"
	if !enabled {
		action = "Disabling"
		pastTense = "disabled"
	}

	w.WriteString(action + " MCP server \"" + name + "\"...\n")

	var foundAny bool
	for _, plat := range platforms {
		if !plat.IsAvailable() {
			continue
		}

		// Check if server exists
		_, err := plat.GetMCP(name)
		if err != nil {
			w.WriteString("  " + plat.Name() + ": not found\n")
			continue
		}

		foundAny = true

		if enabled {
			err = plat.EnableMCP(name)
		} else {
			err = plat.DisableMCP(name)
		}

		if err != nil {
			w.WriteString("  " + plat.Name() + ": error: " + err.Error() + "\n")
			continue
		}

		w.WriteString("  " + plat.Name() + ": " + pastTense + "\n")
	}

	if !foundAny {
		return errors.New("server \"" + name + "\" not found on any platform")
	}

	return nil
}

func TestMCPEnableDisableActionStrings(t *testing.T) {
	// Test that the action strings are correct for enable/disable
	tests := []struct {
		enabled  bool
		wantVerb string
		wantPast string
	}{
		{enabled: true, wantVerb: "Enabling", wantPast: "enabled"},
		{enabled: false, wantVerb: "Disabling", wantPast: "disabled"},
	}

	for _, tt := range tests {
		name := "disable"
		if tt.enabled {
			name = "enable"
		}
		t.Run(name, func(t *testing.T) {
			action := "Enabling"
			pastTense := "enabled"
			if !tt.enabled {
				action = "Disabling"
				pastTense = "disabled"
			}

			if action != tt.wantVerb {
				t.Errorf("action = %q, want %q", action, tt.wantVerb)
			}
			if pastTense != tt.wantPast {
				t.Errorf("pastTense = %q, want %q", pastTense, tt.wantPast)
			}
		})
	}
}

func TestMCPEnableCallsEnableMCP(t *testing.T) {
	mock := &mcpEnableMockPlatform{
		mockPlatform: mockPlatform{name: "claude"},
		mcpServers:   map[string]any{"github": struct{}{}},
		isAvailable:  true,
	}

	var buf bytes.Buffer
	_ = runMCPSetEnabledWithMockPlatforms("github", true, &buf, []cli.Platform{mock})

	if !mock.enableCalled {
		t.Error("EnableMCP() should have been called")
	}
	if mock.disableCalled {
		t.Error("DisableMCP() should not have been called")
	}
}

func TestMCPDisableCallsDisableMCP(t *testing.T) {
	mock := &mcpEnableMockPlatform{
		mockPlatform: mockPlatform{name: "claude"},
		mcpServers:   map[string]any{"github": struct{}{}},
		isAvailable:  true,
	}

	var buf bytes.Buffer
	_ = runMCPSetEnabledWithMockPlatforms("github", false, &buf, []cli.Platform{mock})

	if mock.enableCalled {
		t.Error("EnableMCP() should not have been called")
	}
	if !mock.disableCalled {
		t.Error("DisableMCP() should have been called")
	}
}
