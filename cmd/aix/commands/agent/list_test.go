package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/thoreinstein/aix/internal/cli"
	climocks "github.com/thoreinstein/aix/internal/cli/mocks"
)

func TestListCommand_Metadata(t *testing.T) {
	if listCmd.Use != "list" {
		t.Errorf("Use = %q, want %q", listCmd.Use, "list")
	}

	if listCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Check flags exist
	if listCmd.Flags().Lookup("json") == nil {
		t.Error("--json flag should be defined")
	}
}

func TestOutputListTabular_EmptyState(t *testing.T) {
	m := climocks.NewMockPlatform(t)
	m.EXPECT().Name().Return("claude").Maybe()
	m.EXPECT().DisplayName().Return("Claude Code")
	m.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{}, nil)

	platforms := []cli.Platform{m}

	var buf bytes.Buffer
	err := outputListTabular(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should contain platform name")
	}
	if !strings.Contains(output, "(no agents installed)") {
		t.Error("output should indicate no agents installed")
	}
}

func TestOutputListTabular_WithAgents(t *testing.T) {
	m := climocks.NewMockPlatform(t)
	m.EXPECT().Name().Return("claude").Maybe()
	m.EXPECT().DisplayName().Return("Claude Code")
	m.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
		{
			Name:        "code-reviewer",
			Description: "Reviews code for quality and best practices",
		},
		{
			Name:        "test-generator",
			Description: "Generates unit tests for functions",
		},
	}, nil)

	platforms := []cli.Platform{m}

	var buf bytes.Buffer
	err := outputListTabular(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListTabular() error = %v", err)
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

func TestOutputListTabular_MultiplePlatforms(t *testing.T) {
	m1 := climocks.NewMockPlatform(t)
	m1.EXPECT().Name().Return("claude").Maybe()
	m1.EXPECT().DisplayName().Return("Claude Code")
	m1.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
		{Name: "code-reviewer", Description: "Reviews code"},
	}, nil)

	m2 := climocks.NewMockPlatform(t)
	m2.EXPECT().Name().Return("opencode").Maybe()
	m2.EXPECT().DisplayName().Return("OpenCode")
	m2.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
		{Name: "test-generator", Description: "Generates tests"},
	}, nil)

	platforms := []cli.Platform{m1, m2}

	var buf bytes.Buffer
	err := outputListTabular(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should contain Claude Code")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Error("output should contain OpenCode")
	}
}

func TestOutputListTabular_NoAgentsAcrossPlatforms(t *testing.T) {
	m1 := climocks.NewMockPlatform(t)
	m1.EXPECT().Name().Return("claude").Maybe()
	m1.EXPECT().DisplayName().Return("Claude Code")
	m1.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{}, nil)

	m2 := climocks.NewMockPlatform(t)
	m2.EXPECT().Name().Return("opencode").Maybe()
	m2.EXPECT().DisplayName().Return("OpenCode")
	m2.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{}, nil)

	platforms := []cli.Platform{m1, m2}

	var buf bytes.Buffer
	err := outputListTabular(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListTabular() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No agents installed") {
		t.Error("output should indicate no agents installed across all platforms")
	}
}

func TestOutputListTabular_TruncatesLongDescriptions(t *testing.T) {
	longDesc := strings.Repeat("a", 100)
	m := climocks.NewMockPlatform(t)
	m.EXPECT().Name().Return("claude").Maybe()
	m.EXPECT().DisplayName().Return("Claude Code")
	m.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
		{Name: "agent", Description: longDesc},
	}, nil)

	platforms := []cli.Platform{m}

	var buf bytes.Buffer
	err := outputListTabular(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListTabular() error = %v", err)
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

func TestOutputListJSON(t *testing.T) {
	m := climocks.NewMockPlatform(t)
	m.EXPECT().Name().Return("claude")
	m.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
		{
			Name:        "code-reviewer",
			Description: "Reviews code for quality",
		},
	}, nil)

	platforms := []cli.Platform{m}

	var buf bytes.Buffer
	err := outputListJSON(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListJSON() error = %v", err)
	}

	var result listOutput
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

func TestOutputListJSON_MultiplePlatforms(t *testing.T) {
	m1 := climocks.NewMockPlatform(t)
	m1.EXPECT().Name().Return("claude")
	m1.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
		{Name: "agent1", Description: "Agent 1"},
	}, nil)

	m2 := climocks.NewMockPlatform(t)
	m2.EXPECT().Name().Return("opencode")
	m2.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
		{Name: "agent2", Description: "Agent 2"},
	}, nil)

	platforms := []cli.Platform{m1, m2}

	var buf bytes.Buffer
	err := outputListJSON(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListJSON() error = %v", err)
	}

	var result listOutput
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

func TestOutputListJSON_EmptyAgents(t *testing.T) {
	m := climocks.NewMockPlatform(t)
	m.EXPECT().Name().Return("claude")
	m.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{}, nil)

	platforms := []cli.Platform{m}

	var buf bytes.Buffer
	err := outputListJSON(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListJSON() error = %v", err)
	}

	var result listOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if len(result["claude"]) != 0 {
		t.Errorf("expected 0 agents, got %d", len(result["claude"]))
	}
}

func TestOutputListJSON_FormattedOutput(t *testing.T) {
	m := climocks.NewMockPlatform(t)
	m.EXPECT().Name().Return("claude")
	m.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
		{Name: "agent", Description: "Test agent"},
	}, nil)

	platforms := []cli.Platform{m}

	var buf bytes.Buffer
	err := outputListJSON(&buf, platforms, cli.ScopeUser)
	if err != nil {
		t.Fatalf("outputListJSON() error = %v", err)
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

func TestOutputListTabular_Error(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(t *testing.T) []cli.Platform
		wantErrMsg  string
		description string
	}{
		{
			name: "permission_error",
			setupMocks: func(t *testing.T) []cli.Platform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude")
				m.EXPECT().ListAgents(mock.Anything).Return(nil, errors.New("permission denied: ~/.claude/agents"))
				return []cli.Platform{m}
			},
			wantErrMsg:  "listing agents for claude: permission denied",
			description: "should wrap permission errors with platform context",
		},
		{
			name: "first_platform_error",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().Name().Return("claude")
				m1.EXPECT().ListAgents(mock.Anything).Return(nil, errors.New("directory not found"))

				m2 := climocks.NewMockPlatform(t)
				// m2 won't be called because we fail fast on first error
				return []cli.Platform{m1, m2}
			},
			wantErrMsg:  "listing agents for claude",
			description: "should fail fast on first platform error",
		},
		{
			name: "second_platform_error",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().Name().Return("claude").Maybe()
				m1.EXPECT().DisplayName().Return("Claude Code").Maybe()
				m1.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
					{Name: "agent", Description: "Test"},
				}, nil)

				m2 := climocks.NewMockPlatform(t)
				m2.EXPECT().Name().Return("opencode")
				m2.EXPECT().ListAgents(mock.Anything).Return(nil, errors.New("read error"))

				return []cli.Platform{m1, m2}
			},
			wantErrMsg:  "listing agents for opencode: read error",
			description: "should propagate errors from subsequent platforms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := outputListTabular(&buf, tt.setupMocks(t), cli.ScopeUser)

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

func TestOutputListJSON_Error(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(t *testing.T) []cli.Platform
		wantErrMsg  string
		description string
	}{
		{
			name: "permission_error",
			setupMocks: func(t *testing.T) []cli.Platform {
				m := climocks.NewMockPlatform(t)
				m.EXPECT().Name().Return("claude")
				m.EXPECT().ListAgents(mock.Anything).Return(nil, errors.New("permission denied: ~/.claude/agents"))
				return []cli.Platform{m}
			},
			wantErrMsg:  "listing agents for claude: permission denied",
			description: "should wrap permission errors with platform context",
		},
		{
			name: "first_platform_error",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().Name().Return("claude")
				m1.EXPECT().ListAgents(mock.Anything).Return(nil, errors.New("directory not found"))

				m2 := climocks.NewMockPlatform(t)
				// m2 won't be called because we fail fast on first error
				return []cli.Platform{m1, m2}
			},
			wantErrMsg:  "listing agents for claude",
			description: "should fail fast on first platform error",
		},
		{
			name: "second_platform_error",
			setupMocks: func(t *testing.T) []cli.Platform {
				m1 := climocks.NewMockPlatform(t)
				m1.EXPECT().Name().Return("claude")
				m1.EXPECT().ListAgents(mock.Anything).Return([]cli.AgentInfo{
					{Name: "agent", Description: "Test"},
				}, nil)

				m2 := climocks.NewMockPlatform(t)
				m2.EXPECT().Name().Return("opencode")
				m2.EXPECT().ListAgents(mock.Anything).Return(nil, errors.New("read error"))

				return []cli.Platform{m1, m2}
			},
			wantErrMsg:  "listing agents for opencode: read error",
			description: "should propagate errors from subsequent platforms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := outputListJSON(&buf, tt.setupMocks(t), cli.ScopeUser)

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
