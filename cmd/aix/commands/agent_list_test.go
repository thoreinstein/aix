package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
)

// agentListMockPlatform extends mockPlatform with agent list-specific behavior.
type agentListMockPlatform struct {
	mockPlatform
	agents   []cli.AgentInfo
	agentErr error
}

func (m *agentListMockPlatform) ListAgents() ([]cli.AgentInfo, error) {
	if m.agentErr != nil {
		return nil, m.agentErr
	}
	return m.agents, nil
}

func TestAgentListCommand_Metadata(t *testing.T) {
	if agentListCmd.Use != "list" {
		t.Errorf("Use = %q, want %q", agentListCmd.Use, "list")
	}

	if agentListCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Check flags exist
	if agentListCmd.Flags().Lookup("json") == nil {
		t.Error("--json flag should be defined")
	}
}

func TestOutputAgentsTabular_EmptyState(t *testing.T) {
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should contain platform name")
	}
	if !strings.Contains(output, "(no agents installed)") {
		t.Error("output should indicate no agents installed")
	}
}

func TestOutputAgentsTabular_WithAgents(t *testing.T) {
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{
				{
					Name:        "code-reviewer",
					Description: "Reviews code for quality and best practices",
				},
				{
					Name:        "test-generator",
					Description: "Generates unit tests for functions",
				},
			},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsTabular() error = %v", err)
	}

	output := buf.String()

	// Check headers
	if !strings.Contains(output, "NAME") {
		t.Error("output should contain NAME header")
	}
	if !strings.Contains(output, "DESCRIPTION") {
		t.Error("output should contain DESCRIPTION header")
	}

	// Check agents
	if !strings.Contains(output, "code-reviewer") {
		t.Error("output should contain code-reviewer agent")
	}
	if !strings.Contains(output, "test-generator") {
		t.Error("output should contain test-generator agent")
	}
	if !strings.Contains(output, "Reviews code") {
		t.Error("output should contain agent description")
	}
}

func TestOutputAgentsTabular_MultiplePlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{
				{Name: "code-reviewer", Description: "Reviews code"},
			},
		},
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
			},
			agents: []cli.AgentInfo{
				{Name: "test-generator", Description: "Generates tests"},
			},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should contain Claude Code")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Error("output should contain OpenCode")
	}
}

func TestOutputAgentsTabular_NoAgentsAcrossPlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{},
		},
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
			},
			agents: []cli.AgentInfo{},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No agents installed") {
		t.Error("output should indicate no agents installed across all platforms")
	}
}

func TestOutputAgentsTabular_TruncatesLongDescriptions(t *testing.T) {
	longDesc := strings.Repeat("a", 100)
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{
				{Name: "agent", Description: longDesc},
			},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsTabular(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsTabular() error = %v", err)
	}

	output := buf.String()
	// Should contain truncated description with "..."
	if !strings.Contains(output, "...") {
		t.Error("long description should be truncated with ...")
	}
	// Should not contain the full 100 character description
	if strings.Contains(output, longDesc) {
		t.Error("description should be truncated, not full length")
	}
}

func TestOutputAgentsJSON(t *testing.T) {
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{
				{
					Name:        "code-reviewer",
					Description: "Reviews code for quality",
				},
			},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsJSON(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsJSON() error = %v", err)
	}

	var result agentListOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	claudeAgents, ok := result["claude"]
	if !ok {
		t.Fatal("result should contain 'claude' key")
	}

	if len(claudeAgents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(claudeAgents))
	}

	agent := claudeAgents[0]
	if agent.Name != "code-reviewer" {
		t.Errorf("agent.Name = %q, want %q", agent.Name, "code-reviewer")
	}
	if agent.Description != "Reviews code for quality" {
		t.Errorf("agent.Description = %q, want %q", agent.Description, "Reviews code for quality")
	}
}

func TestOutputAgentsJSON_MultiplePlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{
				{Name: "agent1", Description: "Agent 1"},
			},
		},
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "opencode",
				displayName: "OpenCode",
			},
			agents: []cli.AgentInfo{
				{Name: "agent2", Description: "Agent 2"},
			},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsJSON(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsJSON() error = %v", err)
	}

	var result agentListOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if _, ok := result["claude"]; !ok {
		t.Error("result should contain 'claude' key")
	}
	if _, ok := result["opencode"]; !ok {
		t.Error("result should contain 'opencode' key")
	}
}

func TestOutputAgentsJSON_EmptyAgents(t *testing.T) {
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsJSON(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsJSON() error = %v", err)
	}

	var result agentListOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(result["claude"]) != 0 {
		t.Errorf("expected 0 agents, got %d", len(result["claude"]))
	}
}

func TestOutputAgentsJSON_FormattedOutput(t *testing.T) {
	platforms := []cli.Platform{
		&agentListMockPlatform{
			mockPlatform: mockPlatform{
				name:        "claude",
				displayName: "Claude Code",
			},
			agents: []cli.AgentInfo{
				{Name: "agent", Description: "Test agent"},
			},
		},
	}

	var buf bytes.Buffer
	err := outputAgentsJSON(&buf, platforms)
	if err != nil {
		t.Fatalf("outputAgentsJSON() error = %v", err)
	}

	// Check that output is indented (contains newlines and spaces for formatting)
	output := buf.String()
	if !strings.Contains(output, "\n") {
		t.Error("JSON output should be formatted with newlines")
	}
	if !strings.Contains(output, "  ") {
		t.Error("JSON output should be formatted with indentation")
	}
}

func TestOutputAgentsTabular_Error(t *testing.T) {
	tests := []struct {
		name        string
		platforms   []cli.Platform
		wantErrMsg  string
		description string
	}{
		{
			name: "permission_error",
			platforms: []cli.Platform{
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agentErr: errors.New("permission denied: ~/.claude/agents"),
				},
			},
			wantErrMsg:  "listing agents for claude: permission denied",
			description: "should wrap permission errors with platform context",
		},
		{
			name: "first_platform_error",
			platforms: []cli.Platform{
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agentErr: errors.New("directory not found"),
				},
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "opencode",
						displayName: "OpenCode",
					},
					agents: []cli.AgentInfo{
						{Name: "agent", Description: "Test"},
					},
				},
			},
			wantErrMsg:  "listing agents for claude",
			description: "should fail fast on first platform error",
		},
		{
			name: "second_platform_error",
			platforms: []cli.Platform{
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agents: []cli.AgentInfo{
						{Name: "agent", Description: "Test"},
					},
				},
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "opencode",
						displayName: "OpenCode",
					},
					agentErr: errors.New("read error"),
				},
			},
			wantErrMsg:  "listing agents for opencode: read error",
			description: "should propagate errors from subsequent platforms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := outputAgentsTabular(&buf, tt.platforms)

			if err == nil {
				t.Fatalf("expected error, got nil; %s", tt.description)
			}

			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("error = %q, want to contain %q; %s",
					err.Error(), tt.wantErrMsg, tt.description)
			}
		})
	}
}

func TestOutputAgentsJSON_Error(t *testing.T) {
	tests := []struct {
		name        string
		platforms   []cli.Platform
		wantErrMsg  string
		description string
	}{
		{
			name: "permission_error",
			platforms: []cli.Platform{
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agentErr: errors.New("permission denied: ~/.claude/agents"),
				},
			},
			wantErrMsg:  "listing agents for claude: permission denied",
			description: "should wrap permission errors with platform context",
		},
		{
			name: "first_platform_error",
			platforms: []cli.Platform{
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agentErr: errors.New("directory not found"),
				},
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "opencode",
						displayName: "OpenCode",
					},
					agents: []cli.AgentInfo{
						{Name: "agent", Description: "Test"},
					},
				},
			},
			wantErrMsg:  "listing agents for claude",
			description: "should fail fast on first platform error",
		},
		{
			name: "second_platform_error",
			platforms: []cli.Platform{
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agents: []cli.AgentInfo{
						{Name: "agent", Description: "Test"},
					},
				},
				&agentListMockPlatform{
					mockPlatform: mockPlatform{
						name:        "opencode",
						displayName: "OpenCode",
					},
					agentErr: errors.New("read error"),
				},
			},
			wantErrMsg:  "listing agents for opencode: read error",
			description: "should propagate errors from subsequent platforms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := outputAgentsJSON(&buf, tt.platforms)

			if err == nil {
				t.Fatalf("expected error, got nil; %s", tt.description)
			}

			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("error = %q, want to contain %q; %s",
					err.Error(), tt.wantErrMsg, tt.description)
			}
		})
	}
}
