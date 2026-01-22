package commands

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
)

// mockAgentPlatform implements cli.Platform for testing agent operations.
type mockAgentPlatform struct {
	name          string
	displayName   string
	agents        map[string]any
	uninstallErr  error
	uninstallName string // records the name passed to UninstallAgent
}

func (m *mockAgentPlatform) Name() string        { return m.name }
func (m *mockAgentPlatform) DisplayName() string { return m.displayName }
func (m *mockAgentPlatform) IsAvailable() bool   { return true }

// Skill methods
func (m *mockAgentPlatform) SkillDir() string                     { return "/mock/skills" }
func (m *mockAgentPlatform) InstallSkill(_ any) error             { return nil }
func (m *mockAgentPlatform) UninstallSkill(_ string) error        { return nil }
func (m *mockAgentPlatform) ListSkills() ([]cli.SkillInfo, error) { return nil, nil }
func (m *mockAgentPlatform) GetSkill(_ string) (any, error)       { return nil, errors.New("not found") }

// Command methods
func (m *mockAgentPlatform) CommandDir() string                       { return "/mock/commands" }
func (m *mockAgentPlatform) InstallCommand(_ any) error               { return nil }
func (m *mockAgentPlatform) UninstallCommand(_ string) error          { return nil }
func (m *mockAgentPlatform) ListCommands() ([]cli.CommandInfo, error) { return nil, nil }
func (m *mockAgentPlatform) GetCommand(_ string) (any, error)         { return nil, errors.New("not found") }

// MCP methods
func (m *mockAgentPlatform) MCPConfigPath() string           { return "/mock/mcp.json" }
func (m *mockAgentPlatform) AddMCP(_ any) error              { return nil }
func (m *mockAgentPlatform) RemoveMCP(_ string) error        { return nil }
func (m *mockAgentPlatform) ListMCP() ([]cli.MCPInfo, error) { return nil, nil }
func (m *mockAgentPlatform) GetMCP(_ string) (any, error)    { return nil, errors.New("not found") }
func (m *mockAgentPlatform) EnableMCP(_ string) error        { return nil }
func (m *mockAgentPlatform) DisableMCP(_ string) error       { return nil }

// Agent methods
func (m *mockAgentPlatform) AgentDir() string { return "/mock/agents" }

func (m *mockAgentPlatform) InstallAgent(_ any) error { return nil }

func (m *mockAgentPlatform) UninstallAgent(name string) error {
	m.uninstallName = name
	if m.uninstallErr != nil {
		return m.uninstallErr
	}
	delete(m.agents, name)
	return nil
}

func (m *mockAgentPlatform) ListAgents() ([]cli.AgentInfo, error) {
	agents := make([]cli.AgentInfo, 0, len(m.agents))
	for name := range m.agents {
		agents = append(agents, cli.AgentInfo{Name: name})
	}
	return agents, nil
}

func (m *mockAgentPlatform) GetAgent(name string) (any, error) {
	agent, ok := m.agents[name]
	if !ok {
		return nil, errors.New("agent not found")
	}
	return agent, nil
}

func TestFindPlatformsWithAgent(t *testing.T) {
	tests := []struct {
		name      string
		platforms []cli.Platform
		agentName string
		wantCount int
	}{
		{
			name: "agent found on one platform",
			platforms: []cli.Platform{
				&mockAgentPlatform{name: "claude", agents: map[string]any{"my-agent": struct{}{}}},
				&mockAgentPlatform{name: "opencode", agents: map[string]any{}},
			},
			agentName: "my-agent",
			wantCount: 1,
		},
		{
			name: "agent found on all platforms",
			platforms: []cli.Platform{
				&mockAgentPlatform{name: "claude", agents: map[string]any{"my-agent": struct{}{}}},
				&mockAgentPlatform{name: "opencode", agents: map[string]any{"my-agent": struct{}{}}},
			},
			agentName: "my-agent",
			wantCount: 2,
		},
		{
			name: "agent not found on any platform",
			platforms: []cli.Platform{
				&mockAgentPlatform{name: "claude", agents: map[string]any{}},
				&mockAgentPlatform{name: "opencode", agents: map[string]any{}},
			},
			agentName: "my-agent",
			wantCount: 0,
		},
		{
			name:      "no platforms",
			platforms: []cli.Platform{},
			agentName: "my-agent",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findPlatformsWithAgent(tt.platforms, tt.agentName)
			if len(got) != tt.wantCount {
				t.Errorf("findPlatformsWithAgent() returned %d platforms, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestConfirmAgentRemoval(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		platforms []cli.Platform
		want      bool
	}{
		{
			name:  "yes confirms",
			input: "yes\n",
			platforms: []cli.Platform{
				&mockAgentPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "y confirms",
			input: "y\n",
			platforms: []cli.Platform{
				&mockAgentPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "Y confirms (case insensitive)",
			input: "Y\n",
			platforms: []cli.Platform{
				&mockAgentPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "YES confirms (case insensitive)",
			input: "YES\n",
			platforms: []cli.Platform{
				&mockAgentPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "no rejects",
			input: "no\n",
			platforms: []cli.Platform{
				&mockAgentPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "n rejects",
			input: "n\n",
			platforms: []cli.Platform{
				&mockAgentPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "empty input rejects",
			input: "\n",
			platforms: []cli.Platform{
				&mockAgentPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "random input rejects",
			input: "maybe\n",
			platforms: []cli.Platform{
				&mockAgentPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			in := strings.NewReader(tt.input)

			got := confirmAgentRemoval(&out, in, "test-agent", tt.platforms)
			if got != tt.want {
				t.Errorf("confirmAgentRemoval() = %v, want %v", got, tt.want)
			}

			// Verify prompt was written
			output := out.String()
			if !strings.Contains(output, "test-agent") {
				t.Error("prompt should contain agent name")
			}
			if !strings.Contains(output, "[y/N]") {
				t.Error("prompt should contain [y/N]")
			}
		})
	}
}

func TestConfirmAgentRemoval_ListsPlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&mockAgentPlatform{displayName: "Claude Code"},
		&mockAgentPlatform{displayName: "OpenCode"},
	}

	var out bytes.Buffer
	in := strings.NewReader("n\n")

	confirmAgentRemoval(&out, in, "my-agent", platforms)

	output := out.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should list Claude Code")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Error("output should list OpenCode")
	}
}

func TestAgentRemoveCommand_Metadata(t *testing.T) {
	if agentRemoveCmd.Use != "remove <name>" {
		t.Errorf("Use = %q, want %q", agentRemoveCmd.Use, "remove <name>")
	}

	if agentRemoveCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if agentRemoveCmd.Args == nil {
		t.Error("Args validator should be set")
	}

	// Verify aliases
	aliases := agentRemoveCmd.Aliases
	hasRm := false
	hasUninstall := false
	for _, alias := range aliases {
		if alias == "rm" {
			hasRm = true
		}
		if alias == "uninstall" {
			hasUninstall = true
		}
	}
	if !hasRm {
		t.Error("should have 'rm' alias")
	}
	if !hasUninstall {
		t.Error("should have 'uninstall' alias")
	}
}

func TestAgentRemoveCommand_ForceFlag(t *testing.T) {
	flag := agentRemoveCmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("--force flag should be defined")
	}
	if flag.Shorthand != "" {
		t.Error("--force should not have a shorthand") // Consistent with skill_remove
	}
	if flag.DefValue != "false" {
		t.Errorf("--force default value = %q, want %q", flag.DefValue, "false")
	}
}
