package agent

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
)

// removeMockPlatform implements cli.Platform for testing agent remove operations.
type removeMockPlatform struct {
	name          string
	displayName   string
	agents        map[string]any
	uninstallErr  error
	uninstallName string // records the name passed to UninstallAgent
}

func (m *removeMockPlatform) Name() string        { return m.name }
func (m *removeMockPlatform) DisplayName() string { return m.displayName }
func (m *removeMockPlatform) IsAvailable() bool   { return true }

// Skill methods
func (m *removeMockPlatform) SkillDir() string                     { return "/mock/skills" }
func (m *removeMockPlatform) InstallSkill(_ any) error             { return nil }
func (m *removeMockPlatform) UninstallSkill(_ string) error        { return nil }
func (m *removeMockPlatform) ListSkills() ([]cli.SkillInfo, error) { return nil, nil }
func (m *removeMockPlatform) GetSkill(_ string) (any, error)       { return nil, errors.New("not found") }

// Command methods
func (m *removeMockPlatform) CommandDir() string                       { return "/mock/commands" }
func (m *removeMockPlatform) InstallCommand(_ any) error               { return nil }
func (m *removeMockPlatform) UninstallCommand(_ string) error          { return nil }
func (m *removeMockPlatform) ListCommands() ([]cli.CommandInfo, error) { return nil, nil }
func (m *removeMockPlatform) GetCommand(_ string) (any, error)         { return nil, errors.New("not found") }

// MCP methods
func (m *removeMockPlatform) MCPConfigPath() string           { return "/mock/mcp.json" }
func (m *removeMockPlatform) AddMCP(_ any) error              { return nil }
func (m *removeMockPlatform) RemoveMCP(_ string) error        { return nil }
func (m *removeMockPlatform) ListMCP() ([]cli.MCPInfo, error) { return nil, nil }
func (m *removeMockPlatform) GetMCP(_ string) (any, error)    { return nil, errors.New("not found") }
func (m *removeMockPlatform) EnableMCP(_ string) error        { return nil }
func (m *removeMockPlatform) DisableMCP(_ string) error       { return nil }

// Agent methods
func (m *removeMockPlatform) AgentDir() string { return "/mock/agents" }

func (m *removeMockPlatform) InstallAgent(_ any) error { return nil }

func (m *removeMockPlatform) UninstallAgent(name string) error {
	m.uninstallName = name
	if m.uninstallErr != nil {
		return m.uninstallErr
	}
	delete(m.agents, name)
	return nil
}

func (m *removeMockPlatform) ListAgents() ([]cli.AgentInfo, error) {
	agents := make([]cli.AgentInfo, 0, len(m.agents))
	for name := range m.agents {
		agents = append(agents, cli.AgentInfo{Name: name})
	}
	return agents, nil
}

func (m *removeMockPlatform) GetAgent(name string) (any, error) {
	agent, ok := m.agents[name]
	if !ok {
		return nil, errors.New("agent not found")
	}
	return agent, nil
}

// Backup methods for cli.Platform interface
func (m *removeMockPlatform) BackupPaths() []string { return []string{"/mock/backup"} }

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
				&removeMockPlatform{name: "claude", agents: map[string]any{"my-agent": struct{}{}}},
				&removeMockPlatform{name: "opencode", agents: map[string]any{}},
			},
			agentName: "my-agent",
			wantCount: 1,
		},
		{
			name: "agent found on all platforms",
			platforms: []cli.Platform{
				&removeMockPlatform{name: "claude", agents: map[string]any{"my-agent": struct{}{}}},
				&removeMockPlatform{name: "opencode", agents: map[string]any{"my-agent": struct{}{}}},
			},
			agentName: "my-agent",
			wantCount: 2,
		},
		{
			name: "agent not found on any platform",
			platforms: []cli.Platform{
				&removeMockPlatform{name: "claude", agents: map[string]any{}},
				&removeMockPlatform{name: "opencode", agents: map[string]any{}},
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

func TestConfirmRemoval(t *testing.T) {
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
				&removeMockPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "y confirms",
			input: "y\n",
			platforms: []cli.Platform{
				&removeMockPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "Y confirms (case insensitive)",
			input: "Y\n",
			platforms: []cli.Platform{
				&removeMockPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "YES confirms (case insensitive)",
			input: "YES\n",
			platforms: []cli.Platform{
				&removeMockPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "no rejects",
			input: "no\n",
			platforms: []cli.Platform{
				&removeMockPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "n rejects",
			input: "n\n",
			platforms: []cli.Platform{
				&removeMockPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "empty input rejects",
			input: "\n",
			platforms: []cli.Platform{
				&removeMockPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "random input rejects",
			input: "maybe\n",
			platforms: []cli.Platform{
				&removeMockPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			in := strings.NewReader(tt.input)

			got := confirmRemoval(&out, in, "test-agent", tt.platforms)
			if got != tt.want {
				t.Errorf("confirmRemoval() = %v, want %v", got, tt.want)
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

func TestConfirmRemoval_ListsPlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&removeMockPlatform{displayName: "Claude Code"},
		&removeMockPlatform{displayName: "OpenCode"},
	}

	var out bytes.Buffer
	in := strings.NewReader("n\n")

	confirmRemoval(&out, in, "my-agent", platforms)

	output := out.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should list Claude Code")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Error("output should list OpenCode")
	}
}

func TestRemoveCommand_Metadata(t *testing.T) {
	if removeCmd.Use != "remove <name>" {
		t.Errorf("Use = %q, want %q", removeCmd.Use, "remove <name>")
	}

	if removeCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if removeCmd.Args == nil {
		t.Error("Args validator should be set")
	}

	// Verify aliases
	aliases := removeCmd.Aliases
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

func TestRemoveCommand_ForceFlag(t *testing.T) {
	if flag := removeCmd.Flags().Lookup("force"); flag == nil {
		t.Fatal("--force flag should be defined")
	} else if flag.Shorthand != "" {
		t.Error("--force should not have a shorthand") // Consistent with skill_remove
	} else if flag.DefValue != "false" {
		t.Errorf("--force default value = %q, want %q", flag.DefValue, "false")
	}
}
