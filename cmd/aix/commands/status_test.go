package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
)

// statusMockPlatform implements cli.Platform for status command testing.
// It extends the base mockPlatform with status-specific fields.
type statusMockPlatform struct {
	name        string
	displayName string
	available   bool
	skills      []cli.SkillInfo
	skillsErr   error
	commands    []cli.CommandInfo
	commandsErr error
	mcp         []cli.MCPInfo
	mcpErr      error
}

func (m *statusMockPlatform) Name() string        { return m.name }
func (m *statusMockPlatform) DisplayName() string { return m.displayName }
func (m *statusMockPlatform) IsAvailable() bool   { return m.available }
func (m *statusMockPlatform) SkillDir() string    { return "/mock/skills" }

func (m *statusMockPlatform) InstallSkill(_ any) error { return nil }
func (m *statusMockPlatform) UninstallSkill(_ string) error {
	return errors.New("not implemented")
}
func (m *statusMockPlatform) ListSkills() ([]cli.SkillInfo, error) {
	return m.skills, m.skillsErr
}
func (m *statusMockPlatform) GetSkill(_ string) (any, error) {
	return nil, errors.New("not implemented")
}

func (m *statusMockPlatform) CommandDir() string         { return "/mock/commands" }
func (m *statusMockPlatform) InstallCommand(_ any) error { return nil }
func (m *statusMockPlatform) UninstallCommand(_ string) error {
	return errors.New("not implemented")
}
func (m *statusMockPlatform) ListCommands() ([]cli.CommandInfo, error) {
	return m.commands, m.commandsErr
}
func (m *statusMockPlatform) GetCommand(_ string) (any, error) {
	return nil, errors.New("not implemented")
}

func (m *statusMockPlatform) MCPConfigPath() string           { return "/mock/mcp.json" }
func (m *statusMockPlatform) AddMCP(_ any) error              { return nil }
func (m *statusMockPlatform) RemoveMCP(_ string) error        { return nil }
func (m *statusMockPlatform) ListMCP() ([]cli.MCPInfo, error) { return m.mcp, m.mcpErr }
func (m *statusMockPlatform) GetMCP(_ string) (any, error) {
	return nil, errors.New("not implemented")
}
func (m *statusMockPlatform) EnableMCP(_ string) error  { return nil }
func (m *statusMockPlatform) DisableMCP(_ string) error { return nil }

func TestValidateStatusFlags(t *testing.T) {
	tests := []struct {
		name        string
		jsonFlag    bool
		quietFlag   bool
		verboseFlag bool
		wantErr     bool
	}{
		{
			name:        "no flags set",
			jsonFlag:    false,
			quietFlag:   false,
			verboseFlag: false,
			wantErr:     false,
		},
		{
			name:        "only json flag",
			jsonFlag:    true,
			quietFlag:   false,
			verboseFlag: false,
			wantErr:     false,
		},
		{
			name:        "only quiet flag",
			jsonFlag:    false,
			quietFlag:   true,
			verboseFlag: false,
			wantErr:     false,
		},
		{
			name:        "only verbose flag",
			jsonFlag:    false,
			quietFlag:   false,
			verboseFlag: true,
			wantErr:     false,
		},
		{
			name:        "json and quiet flags",
			jsonFlag:    true,
			quietFlag:   true,
			verboseFlag: false,
			wantErr:     true,
		},
		{
			name:        "json and verbose flags",
			jsonFlag:    true,
			quietFlag:   false,
			verboseFlag: true,
			wantErr:     true,
		},
		{
			name:        "quiet and verbose flags",
			jsonFlag:    false,
			quietFlag:   true,
			verboseFlag: true,
			wantErr:     true,
		},
		{
			name:        "all three flags",
			jsonFlag:    true,
			quietFlag:   true,
			verboseFlag: true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global flags
			oldJSON := statusJSON
			oldQuiet := statusQuiet
			oldVerbose := statusVerbose
			defer func() {
				statusJSON = oldJSON
				statusQuiet = oldQuiet
				statusVerbose = oldVerbose
			}()

			statusJSON = tt.jsonFlag
			statusQuiet = tt.quietFlag
			statusVerbose = tt.verboseFlag

			err := validateStatusFlags(nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStatusFlags() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && !strings.Contains(err.Error(), "mutually exclusive") {
				t.Errorf("error should mention 'mutually exclusive', got: %v", err)
			}
		})
	}
}

func TestMcpCounts(t *testing.T) {
	tests := []struct {
		name         string
		servers      []cli.MCPInfo
		wantTotal    int
		wantEnabled  int
		wantDisabled int
	}{
		{
			name:         "empty slice",
			servers:      []cli.MCPInfo{},
			wantTotal:    0,
			wantEnabled:  0,
			wantDisabled: 0,
		},
		{
			name:         "nil slice",
			servers:      nil,
			wantTotal:    0,
			wantEnabled:  0,
			wantDisabled: 0,
		},
		{
			name: "all enabled",
			servers: []cli.MCPInfo{
				{Name: "server1", Disabled: false},
				{Name: "server2", Disabled: false},
				{Name: "server3", Disabled: false},
			},
			wantTotal:    3,
			wantEnabled:  3,
			wantDisabled: 0,
		},
		{
			name: "all disabled",
			servers: []cli.MCPInfo{
				{Name: "server1", Disabled: true},
				{Name: "server2", Disabled: true},
			},
			wantTotal:    2,
			wantEnabled:  0,
			wantDisabled: 2,
		},
		{
			name: "mixed enabled and disabled",
			servers: []cli.MCPInfo{
				{Name: "server1", Disabled: false},
				{Name: "server2", Disabled: true},
				{Name: "server3", Disabled: false},
				{Name: "server4", Disabled: true},
				{Name: "server5", Disabled: false},
			},
			wantTotal:    5,
			wantEnabled:  3,
			wantDisabled: 2,
		},
		{
			name: "single enabled",
			servers: []cli.MCPInfo{
				{Name: "server1", Disabled: false},
			},
			wantTotal:    1,
			wantEnabled:  1,
			wantDisabled: 0,
		},
		{
			name: "single disabled",
			servers: []cli.MCPInfo{
				{Name: "server1", Disabled: true},
			},
			wantTotal:    1,
			wantEnabled:  0,
			wantDisabled: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, gotEnabled, gotDisabled := mcpCounts(tt.servers)
			if gotTotal != tt.wantTotal {
				t.Errorf("mcpCounts() total = %d, want %d", gotTotal, tt.wantTotal)
			}
			if gotEnabled != tt.wantEnabled {
				t.Errorf("mcpCounts() enabled = %d, want %d", gotEnabled, tt.wantEnabled)
			}
			if gotDisabled != tt.wantDisabled {
				t.Errorf("mcpCounts() disabled = %d, want %d", gotDisabled, tt.wantDisabled)
			}
		})
	}
}

func TestCollectPlatformStatus(t *testing.T) {
	t.Run("unavailable platform returns early", func(t *testing.T) {
		mock := &statusMockPlatform{
			name:        "test",
			displayName: "Test Platform",
			available:   false,
		}

		status := collectPlatformStatus(mock)

		if status.Available {
			t.Error("status.Available should be false for unavailable platform")
		}
		if status.Skills != nil {
			t.Error("status.Skills should be nil for unavailable platform")
		}
		if status.Commands != nil {
			t.Error("status.Commands should be nil for unavailable platform")
		}
		if status.MCP != nil {
			t.Error("status.MCP should be nil for unavailable platform")
		}
	})

	t.Run("available platform collects all data", func(t *testing.T) {
		mock := &statusMockPlatform{
			name:        "test",
			displayName: "Test Platform",
			available:   true,
			skills: []cli.SkillInfo{
				{Name: "skill1", Description: "Skill 1"},
			},
			commands: []cli.CommandInfo{
				{Name: "cmd1", Description: "Command 1"},
			},
			mcp: []cli.MCPInfo{
				{Name: "mcp1", Transport: "stdio"},
			},
		}

		status := collectPlatformStatus(mock)

		if !status.Available {
			t.Error("status.Available should be true")
		}
		if len(status.Skills) != 1 {
			t.Errorf("expected 1 skill, got %d", len(status.Skills))
		}
		if len(status.Commands) != 1 {
			t.Errorf("expected 1 command, got %d", len(status.Commands))
		}
		if len(status.MCP) != 1 {
			t.Errorf("expected 1 mcp, got %d", len(status.MCP))
		}
	})

	t.Run("captures errors from list operations", func(t *testing.T) {
		expectedErr := errors.New("test error")
		mock := &statusMockPlatform{
			name:        "test",
			displayName: "Test Platform",
			available:   true,
			skillsErr:   expectedErr,
			commandsErr: expectedErr,
			mcpErr:      expectedErr,
		}

		status := collectPlatformStatus(mock)

		if status.SkillsErr == nil {
			t.Error("status.SkillsErr should not be nil")
		}
		if status.CommandsErr == nil {
			t.Error("status.CommandsErr should not be nil")
		}
		if status.MCPErr == nil {
			t.Error("status.MCPErr should not be nil")
		}
	})
}

func TestOutputStatusJSON(t *testing.T) {
	t.Run("basic structure", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				skills: []cli.SkillInfo{
					{Name: "debug", Description: "Debug skill"},
				},
				commands: []cli.CommandInfo{
					{Name: "test", Description: "Test command"},
				},
				mcp: []cli.MCPInfo{
					{Name: "github", Transport: "stdio", Command: "npx", Disabled: false},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusJSON(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusJSON() error = %v", err)
		}

		var result statusJSONOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if result.Version != Version {
			t.Errorf("version = %q, want %q", result.Version, Version)
		}

		platformEntry, ok := result.Platforms["claude"]
		if !ok {
			t.Fatal("expected 'claude' in platforms")
		}

		if !platformEntry.Available {
			t.Error("platform should be available")
		}
		if platformEntry.Skills == nil || platformEntry.Skills.Count != 1 {
			t.Errorf("expected 1 skill, got %v", platformEntry.Skills)
		}
		if platformEntry.Commands == nil || platformEntry.Commands.Count != 1 {
			t.Errorf("expected 1 command, got %v", platformEntry.Commands)
		}
		if platformEntry.MCP == nil || platformEntry.MCP.Count != 1 {
			t.Errorf("expected 1 mcp, got %v", platformEntry.MCP)
		}
	})

	t.Run("unavailable platform", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
				available:   false,
			},
		}

		var buf bytes.Buffer
		err := outputStatusJSON(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusJSON() error = %v", err)
		}

		var result statusJSONOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		platformEntry, ok := result.Platforms["opencode"]
		if !ok {
			t.Fatal("expected 'opencode' in platforms")
		}

		if platformEntry.Available {
			t.Error("platform should not be available")
		}
		if platformEntry.Skills != nil {
			t.Error("skills should be nil for unavailable platform")
		}
	})

	t.Run("mcp counts enabled and disabled", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				mcp: []cli.MCPInfo{
					{Name: "server1", Disabled: false},
					{Name: "server2", Disabled: true},
					{Name: "server3", Disabled: false},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusJSON(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusJSON() error = %v", err)
		}

		var result statusJSONOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		mcp := result.Platforms["claude"].MCP
		if mcp.Count != 3 {
			t.Errorf("mcp.Count = %d, want 3", mcp.Count)
		}
		if mcp.Enabled != 2 {
			t.Errorf("mcp.Enabled = %d, want 2", mcp.Enabled)
		}
		if mcp.Disabled != 1 {
			t.Errorf("mcp.Disabled = %d, want 1", mcp.Disabled)
		}
	})

	t.Run("credential redaction", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				mcp: []cli.MCPInfo{
					{
						Name:      "github",
						Transport: "stdio",
						Command:   "npx",
						Env: map[string]string{
							"GITHUB_TOKEN": "ghp_xxxxxxxxxxxx1234",
							"DEBUG":        "true",
							"API_KEY":      "sk-secret-key-value",
						},
					},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusJSON(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusJSON() error = %v", err)
		}

		var result statusJSONOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		env := result.Platforms["claude"].MCP.Items[0].Env

		// GITHUB_TOKEN should be masked (contains TOKEN)
		if env["GITHUB_TOKEN"] == "ghp_xxxxxxxxxxxx1234" {
			t.Error("GITHUB_TOKEN should be masked")
		}
		if !strings.HasPrefix(env["GITHUB_TOKEN"], "****") {
			t.Errorf("GITHUB_TOKEN should start with ****, got %q", env["GITHUB_TOKEN"])
		}

		// DEBUG should NOT be masked
		if env["DEBUG"] != "true" {
			t.Errorf("DEBUG should not be masked, got %q", env["DEBUG"])
		}

		// API_KEY should be masked (contains KEY)
		if env["API_KEY"] == "sk-secret-key-value" {
			t.Error("API_KEY should be masked")
		}
	})

	t.Run("handles errors in platform data", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				skillsErr:   errors.New("skills error"),
				commandsErr: errors.New("commands error"),
				mcpErr:      errors.New("mcp error"),
			},
		}

		var buf bytes.Buffer
		err := outputStatusJSON(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusJSON() error = %v", err)
		}

		var result statusJSONOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		entry := result.Platforms["claude"]
		if entry.Skills == nil || entry.Skills.Error != "skills error" {
			t.Errorf("skills error not captured: %+v", entry.Skills)
		}
		if entry.Commands == nil || entry.Commands.Error != "commands error" {
			t.Errorf("commands error not captured: %+v", entry.Commands)
		}
		if entry.MCP == nil || entry.MCP.Error != "mcp error" {
			t.Errorf("mcp error not captured: %+v", entry.MCP)
		}
	})
}

func TestOutputStatusQuiet(t *testing.T) {
	t.Run("available platform with data", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				skills: []cli.SkillInfo{
					{Name: "skill1"},
					{Name: "skill2"},
				},
				commands: []cli.CommandInfo{
					{Name: "cmd1"},
				},
				mcp: []cli.MCPInfo{
					{Name: "mcp1"},
					{Name: "mcp2"},
					{Name: "mcp3"},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusQuiet(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusQuiet() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "claude:") {
			t.Error("output should contain platform name")
		}
		if !strings.Contains(output, "2 skills") {
			t.Error("output should contain skill count")
		}
		if !strings.Contains(output, "1 commands") {
			t.Error("output should contain command count")
		}
		if !strings.Contains(output, "3 mcp") {
			t.Error("output should contain mcp count")
		}
	})

	t.Run("unavailable platform", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
				available:   false,
			},
		}

		var buf bytes.Buffer
		err := outputStatusQuiet(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusQuiet() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "opencode: (not installed)") {
			t.Errorf("output should indicate not installed, got: %q", output)
		}
	})

	t.Run("handles errors gracefully", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				skillsErr:   errors.New("skills error"),
				commandsErr: errors.New("commands error"),
				mcpErr:      errors.New("mcp error"),
			},
		}

		var buf bytes.Buffer
		err := outputStatusQuiet(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusQuiet() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "skills: error") {
			t.Error("output should indicate skills error")
		}
		if !strings.Contains(output, "commands: error") {
			t.Error("output should indicate commands error")
		}
		if !strings.Contains(output, "mcp: error") {
			t.Error("output should indicate mcp error")
		}
	})
}

func TestOutputStatusCompact(t *testing.T) {
	t.Run("includes version header", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
			},
		}

		var buf bytes.Buffer
		err := outputStatusCompact(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusCompact() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "aix version") {
			t.Error("output should contain version header")
		}
	})

	t.Run("shows platform display name", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
			},
		}

		var buf bytes.Buffer
		err := outputStatusCompact(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusCompact() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Claude Code") {
			t.Error("output should contain platform display name")
		}
	})

	t.Run("unavailable platform shows not installed", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
				available:   false,
			},
		}

		var buf bytes.Buffer
		err := outputStatusCompact(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusCompact() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "(not installed)") {
			t.Error("output should indicate not installed")
		}
	})

	t.Run("shows disabled mcp count", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				mcp: []cli.MCPInfo{
					{Name: "server1", Disabled: false},
					{Name: "server2", Disabled: true},
					{Name: "server3", Disabled: true},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusCompact(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusCompact() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "2 disabled") {
			t.Errorf("output should show disabled count, got: %s", output)
		}
	})

	t.Run("handles errors", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				skillsErr:   errors.New("skills error"),
			},
		}

		var buf bytes.Buffer
		err := outputStatusCompact(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusCompact() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "error") {
			t.Error("output should indicate error")
		}
	})
}

func TestOutputStatusVerbose(t *testing.T) {
	t.Run("includes version header", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "aix version") {
			t.Error("output should contain version header")
		}
	})

	t.Run("shows skill details", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				skills: []cli.SkillInfo{
					{Name: "debug", Description: "Debug skill for testing"},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "debug") {
			t.Error("output should contain skill name")
		}
		if !strings.Contains(output, "Debug skill") {
			t.Error("output should contain skill description")
		}
	})

	t.Run("shows command details with slash prefix", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				commands: []cli.CommandInfo{
					{Name: "test", Description: "Test command"},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "/test") {
			t.Error("output should contain command name with slash prefix")
		}
	})

	t.Run("shows mcp server details", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				mcp: []cli.MCPInfo{
					{
						Name:      "github",
						Transport: "stdio",
						Command:   "npx -y @modelcontextprotocol/server-github",
						Env: map[string]string{
							"DEBUG": "true",
						},
					},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "github") {
			t.Error("output should contain server name")
		}
		if !strings.Contains(output, "Transport: stdio") {
			t.Error("output should contain transport")
		}
		if !strings.Contains(output, "Command:") {
			t.Error("output should contain command")
		}
	})

	t.Run("shows url for sse transport", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				mcp: []cli.MCPInfo{
					{
						Name:      "api-gateway",
						Transport: "sse",
						URL:       "https://api.example.com/mcp",
					},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "URL: https://api.example.com/mcp") {
			t.Errorf("output should contain URL, got: %s", output)
		}
	})

	t.Run("credential redaction in env vars", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				mcp: []cli.MCPInfo{
					{
						Name:      "github",
						Transport: "stdio",
						Command:   "npx",
						Env: map[string]string{
							"GITHUB_TOKEN": "ghp_secret_token_value",
							"DEBUG":        "true",
						},
					},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()

		// Token should be masked
		if strings.Contains(output, "ghp_secret_token_value") {
			t.Error("GITHUB_TOKEN value should be masked")
		}
		if !strings.Contains(output, "GITHUB_TOKEN=****") {
			t.Errorf("GITHUB_TOKEN should show masked value, got: %s", output)
		}

		// DEBUG should NOT be masked
		if !strings.Contains(output, "DEBUG=true") {
			t.Errorf("DEBUG should not be masked, got: %s", output)
		}
	})

	t.Run("shows enabled/disabled status for mcp servers", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				mcp: []cli.MCPInfo{
					{Name: "enabled-server", Transport: "stdio", Disabled: false},
					{Name: "disabled-server", Transport: "stdio", Disabled: true},
				},
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		// Status text appears inside brackets with ANSI codes, so check for the status words
		if !strings.Contains(output, "enabled") {
			t.Error("output should show enabled status")
		}
		if !strings.Contains(output, "disabled") {
			t.Error("output should show disabled status")
		}
	})

	t.Run("empty sections show (none)", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "claude",
				displayName: "Claude Code",
				available:   true,
				skills:      []cli.SkillInfo{},
				commands:    []cli.CommandInfo{},
				mcp:         []cli.MCPInfo{},
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		// Count occurrences of "(none)" - should appear for skills, commands, and mcp
		count := strings.Count(output, "(none)")
		if count != 3 {
			t.Errorf("expected 3 '(none)' markers, got %d in: %s", count, output)
		}
	})

	t.Run("unavailable platform shows not installed", func(t *testing.T) {
		platforms := []cli.Platform{
			&statusMockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
				available:   false,
			},
		}

		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "(not installed)") {
			t.Error("output should indicate not installed")
		}
	})
}

func TestOutputStatusVerbose_LongDescriptionTruncation(t *testing.T) {
	longDesc := strings.Repeat("a", 100)
	platforms := []cli.Platform{
		&statusMockPlatform{
			name:        "claude",
			displayName: "Claude Code",
			available:   true,
			skills: []cli.SkillInfo{
				{Name: "test", Description: longDesc},
			},
		},
	}

	var buf bytes.Buffer
	err := outputStatusVerbose(&buf, platforms)
	if err != nil {
		t.Fatalf("outputStatusVerbose() error = %v", err)
	}

	output := buf.String()
	// Description should be truncated (truncate function limits to 60 chars)
	if strings.Contains(output, longDesc) {
		t.Error("long description should be truncated")
	}
}

func TestStatusCommand_Metadata(t *testing.T) {
	if statusCmd.Use != "status" {
		t.Errorf("Use = %q, want %q", statusCmd.Use, "status")
	}

	if statusCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Check flags exist
	if statusCmd.Flags().Lookup("json") == nil {
		t.Error("--json flag should be defined")
	}
	if statusCmd.Flags().Lookup("quiet") == nil {
		t.Error("--quiet flag should be defined")
	}
	if statusCmd.Flags().Lookup("verbose") == nil {
		t.Error("--verbose flag should be defined")
	}
}

func TestMultiplePlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&statusMockPlatform{
			name:        "claude",
			displayName: "Claude Code",
			available:   true,
			skills: []cli.SkillInfo{
				{Name: "skill1"},
			},
		},
		&statusMockPlatform{
			name:        "opencode",
			displayName: "OpenCode",
			available:   true,
			skills: []cli.SkillInfo{
				{Name: "skill1"},
				{Name: "skill2"},
			},
		},
	}

	t.Run("JSON output includes all platforms", func(t *testing.T) {
		var buf bytes.Buffer
		err := outputStatusJSON(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusJSON() error = %v", err)
		}

		var result statusJSONOutput
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if len(result.Platforms) != 2 {
			t.Errorf("expected 2 platforms, got %d", len(result.Platforms))
		}
		if _, ok := result.Platforms["claude"]; !ok {
			t.Error("expected claude in platforms")
		}
		if _, ok := result.Platforms["opencode"]; !ok {
			t.Error("expected opencode in platforms")
		}
	})

	t.Run("quiet output includes all platforms", func(t *testing.T) {
		var buf bytes.Buffer
		err := outputStatusQuiet(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusQuiet() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "claude:") {
			t.Error("output should contain claude")
		}
		if !strings.Contains(output, "opencode:") {
			t.Error("output should contain opencode")
		}
	})

	t.Run("compact output includes all platforms", func(t *testing.T) {
		var buf bytes.Buffer
		err := outputStatusCompact(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusCompact() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Claude Code") {
			t.Error("output should contain Claude Code")
		}
		if !strings.Contains(output, "OpenCode") {
			t.Error("output should contain OpenCode")
		}
	})

	t.Run("verbose output includes all platforms", func(t *testing.T) {
		var buf bytes.Buffer
		err := outputStatusVerbose(&buf, platforms)
		if err != nil {
			t.Fatalf("outputStatusVerbose() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Claude Code") {
			t.Error("output should contain Claude Code")
		}
		if !strings.Contains(output, "OpenCode") {
			t.Error("output should contain OpenCode")
		}
	})
}
