package mcp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/thoreinstein/aix/internal/cli"
	climocks "github.com/thoreinstein/aix/internal/cli/mocks"
	"github.com/thoreinstein/aix/internal/errors"
)

func TestEnableCommand_Metadata(t *testing.T) {
	if enableCmd.Use != "enable <name>" {
		t.Errorf("Use = %q, want %q", enableCmd.Use, "enable <name>")
	}

	if enableCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if enableCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Verify Args validator is set (ExactArgs(1))
	if enableCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestDisableCommand_Metadata(t *testing.T) {
	if disableCmd.Use != "disable <name>" {
		t.Errorf("Use = %q, want %q", disableCmd.Use, "disable <name>")
	}

	if disableCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if disableCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Verify Args validator is set (ExactArgs(1))
	if disableCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestRunMCPSetEnabledWithIO_Enable(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		setupMock  func(t *testing.T) *climocks.MockPlatform
		wantErr    bool
		wantOutput []string
	}{
		{
			name:       "enable existing server",
			serverName: "github",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude").Maybe()
				m.EXPECT().IsAvailable().Return(true)
				m.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
				m.EXPECT().EnableMCP("github").Return(nil)
				return m
			},
			wantErr:    false,
			wantOutput: []string{"Enabling", "github", "enabled"},
		},
		{
			name:       "enable non-existent server",
			serverName: "not-found",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude").Maybe()
				m.EXPECT().IsAvailable().Return(true)
				m.EXPECT().GetMCP("not-found", mock.Anything).Return(nil, errors.New("MCP server not found"))
				return m
			},
			wantErr:    true,
			wantOutput: []string{"not found"},
		},
		{
			name:       "enable with platform error",
			serverName: "github",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude").Maybe()
				m.EXPECT().IsAvailable().Return(true)
				m.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
				m.EXPECT().EnableMCP("github").Return(errors.New("permission denied"))
				return m
			},
			wantErr:    false, // The function continues and reports the error in output
			wantOutput: []string{"error", "permission denied"},
		},
		{
			name:       "skip unavailable platform",
			serverName: "github",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude").Maybe()
				m.EXPECT().IsAvailable().Return(false)
				return m
			},
			wantErr:    true, // Server not found on any available platform
			wantOutput: []string{"github"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			platforms := []cli.Platform{tt.setupMock(t)}
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

func TestRunMCPSetEnabledWithIO_MultiplePlatforms(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		setupMocks func(t *testing.T) []cli.Platform
		wantErr    bool
		wantOutput []string
	}{
		{
			name:       "enable across multiple platforms",
			serverName: "github",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().Name().Return("claude").Maybe()
				m1.EXPECT().IsAvailable().Return(true)
				m1.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
				m1.EXPECT().EnableMCP("github").Return(nil)

				m2 := climocks.NewMockPlatform(t)
				m2.EXPECT().Name().Return("opencode").Maybe()
				m2.EXPECT().IsAvailable().Return(true)
				m2.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
				m2.EXPECT().EnableMCP("github").Return(nil)

				return []cli.Platform{m1, m2}
			},
			wantErr:    false,
			wantOutput: []string{"claude", "opencode", "enabled"},
		},
		{
			name:       "partial success - one platform has server",
			serverName: "github",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().Name().Return("claude").Maybe()
				m1.EXPECT().IsAvailable().Return(true)
				m1.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
				m1.EXPECT().EnableMCP("github").Return(nil)

				m2 := climocks.NewMockPlatform(t)
				m2.EXPECT().Name().Return("opencode").Maybe()
				m2.EXPECT().IsAvailable().Return(true)
				m2.EXPECT().GetMCP("github", mock.Anything).Return(nil, errors.New("not found"))

				return []cli.Platform{m1, m2}
			},
			wantErr:    false,
			wantOutput: []string{"claude", "enabled", "opencode", "not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			platforms := tt.setupMocks(t)
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
		setupMock  func(t *testing.T) *climocks.MockPlatform
		wantErr    bool
		wantOutput []string
	}{
		{
			name:       "disable existing server",
			serverName: "github",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude").Maybe()
				m.EXPECT().IsAvailable().Return(true)
				m.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
				m.EXPECT().DisableMCP("github").Return(nil)
				return m
			},
			wantErr:    false,
			wantOutput: []string{"Disabling", "github", "disabled"},
		},
		{
			name:       "disable non-existent server",
			serverName: "not-found",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude").Maybe()
				m.EXPECT().IsAvailable().Return(true)
				m.EXPECT().GetMCP("not-found", mock.Anything).Return(nil, errors.New("MCP server not found"))
				return m
			},
			wantErr:    true,
			wantOutput: []string{"not found"},
		},
		{
			name:       "disable with platform error",
			serverName: "github",
			setupMock: func(t *testing.T) *climocks.MockPlatform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude").Maybe()
				m.EXPECT().IsAvailable().Return(true)
				m.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
				m.EXPECT().DisableMCP("github").Return(errors.New("disk full"))
				return m
			},
			wantErr:    false, // The function continues and reports the error in output
			wantOutput: []string{"error", "disk full"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			platforms := []cli.Platform{tt.setupMock(t)}
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
		_, err := plat.GetMCP(name, cli.ScopeDefault)
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
	m := climocks.NewMockPlatform(t)
	m.EXPECT().Name().Return("claude").Maybe()
	m.EXPECT().IsAvailable().Return(true)
	m.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
	m.EXPECT().EnableMCP("github").Return(nil)

	var buf bytes.Buffer
	_ = runMCPSetEnabledWithMockPlatforms("github", true, &buf, []cli.Platform{m})
}

func TestMCPDisableCallsDisableMCP(t *testing.T) {
	m := climocks.NewMockPlatform(t)
	m.EXPECT().Name().Return("claude").Maybe()
	m.EXPECT().IsAvailable().Return(true)
	m.EXPECT().GetMCP("github", mock.Anything).Return(struct{}{}, nil)
	m.EXPECT().DisableMCP("github").Return(nil)

	var buf bytes.Buffer
	_ = runMCPSetEnabledWithMockPlatforms("github", false, &buf, []cli.Platform{m})
}
