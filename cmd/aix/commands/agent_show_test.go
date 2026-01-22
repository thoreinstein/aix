package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

func TestAgentShowCommand_Metadata(t *testing.T) {
	if agentShowCmd.Use != "show <name>" {
		t.Errorf("Use = %q, want %q", agentShowCmd.Use, "show <name>")
	}

	if agentShowCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if agentShowCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Check flags exist
	expectedFlags := []string{"json", "full"}
	for _, flagName := range expectedFlags {
		if agentShowCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("--%s flag should be defined", flagName)
		}
	}

	// Verify Args validator is set (ExactArgs(1))
	if agentShowCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestExtractAgentDetail_Claude(t *testing.T) {
	tests := []struct {
		name  string
		agent *claude.Agent
		want  *showAgentDetail
	}{
		{
			name: "full agent",
			agent: &claude.Agent{
				Name:         "code-reviewer",
				Description:  "Reviews code for quality",
				Instructions: "You are a code reviewer...",
			},
			want: &showAgentDetail{
				Name:         "code-reviewer",
				Description:  "Reviews code for quality",
				Instructions: "You are a code reviewer...",
			},
		},
		{
			name: "minimal agent",
			agent: &claude.Agent{
				Name: "simple-agent",
			},
			want: &showAgentDetail{
				Name: "simple-agent",
			},
		},
		{
			name: "agent with only instructions",
			agent: &claude.Agent{
				Name:         "instructions-only",
				Instructions: "Do the thing.",
			},
			want: &showAgentDetail{
				Name:         "instructions-only",
				Instructions: "Do the thing.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractClaudeAgent(tt.agent)
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.Instructions != tt.want.Instructions {
				t.Errorf("Instructions = %q, want %q", got.Instructions, tt.want.Instructions)
			}
			// Claude agents don't have Mode or Temperature
			if got.Mode != "" {
				t.Errorf("Mode = %q, want empty", got.Mode)
			}
			if got.Temperature != 0 {
				t.Errorf("Temperature = %f, want 0", got.Temperature)
			}
		})
	}
}

func TestExtractAgentDetail_OpenCode(t *testing.T) {
	tests := []struct {
		name  string
		agent *opencode.Agent
		want  *showAgentDetail
	}{
		{
			name: "full agent with all fields",
			agent: &opencode.Agent{
				Name:         "creative-writer",
				Description:  "Creative writing assistant",
				Mode:         "chat",
				Temperature:  0.8,
				Instructions: "You are a creative writer...",
			},
			want: &showAgentDetail{
				Name:         "creative-writer",
				Description:  "Creative writing assistant",
				Mode:         "chat",
				Temperature:  0.8,
				Instructions: "You are a creative writer...",
			},
		},
		{
			name: "agent without mode/temperature",
			agent: &opencode.Agent{
				Name:         "basic-agent",
				Description:  "A basic agent",
				Instructions: "Be helpful.",
			},
			want: &showAgentDetail{
				Name:         "basic-agent",
				Description:  "A basic agent",
				Instructions: "Be helpful.",
			},
		},
		{
			name: "agent with zero temperature",
			agent: &opencode.Agent{
				Name:        "precise-agent",
				Mode:        "edit",
				Temperature: 0.0,
			},
			want: &showAgentDetail{
				Name:        "precise-agent",
				Mode:        "edit",
				Temperature: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOpenCodeAgent(tt.agent)
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.Mode != tt.want.Mode {
				t.Errorf("Mode = %q, want %q", got.Mode, tt.want.Mode)
			}
			if got.Temperature != tt.want.Temperature {
				t.Errorf("Temperature = %f, want %f", got.Temperature, tt.want.Temperature)
			}
			if got.Instructions != tt.want.Instructions {
				t.Errorf("Instructions = %q, want %q", got.Instructions, tt.want.Instructions)
			}
		})
	}
}

func TestExtractAgentDetail_UnknownType(t *testing.T) {
	// Test that extractAgentDetail returns nil for unknown types
	got := extractAgentDetail("not an agent type")
	if got != nil {
		t.Errorf("extractAgentDetail() = %v, want nil for unknown type", got)
	}
}

func TestOutputAgentShowJSON(t *testing.T) {
	detail := &showAgentDetail{
		Name:         "test-agent",
		Description:  "A test agent",
		Mode:         "review",
		Temperature:  0.5,
		Instructions: "Test instructions.",
		Installations: []agentInstallLocation{
			{Platform: "Claude Code", Path: "/path/to/agents/test-agent.md"},
		},
	}

	var buf bytes.Buffer
	err := outputAgentShowJSON(&buf, detail)
	if err != nil {
		t.Fatalf("outputAgentShowJSON() error = %v", err)
	}

	// Verify valid JSON
	var unmarshaled showAgentDetail
	if err := json.Unmarshal(buf.Bytes(), &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v", err)
	}

	if unmarshaled.Name != detail.Name {
		t.Errorf("Name = %q, want %q", unmarshaled.Name, detail.Name)
	}
	if unmarshaled.Description != detail.Description {
		t.Errorf("Description = %q, want %q", unmarshaled.Description, detail.Description)
	}
	if unmarshaled.Mode != detail.Mode {
		t.Errorf("Mode = %q, want %q", unmarshaled.Mode, detail.Mode)
	}
	if unmarshaled.Temperature != detail.Temperature {
		t.Errorf("Temperature = %f, want %f", unmarshaled.Temperature, detail.Temperature)
	}
	if len(unmarshaled.Installations) != 1 {
		t.Errorf("Installations count = %d, want 1", len(unmarshaled.Installations))
	}
}

func TestOutputAgentShowText(t *testing.T) {
	tests := []struct {
		name           string
		detail         *showAgentDetail
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "full detail",
			detail: &showAgentDetail{
				Name:         "reviewer",
				Description:  "Reviews code",
				Mode:         "edit",
				Temperature:  0.3,
				Instructions: "Be thorough.",
				Installations: []agentInstallLocation{
					{Platform: "OpenCode", Path: "/agents/reviewer.md"},
				},
			},
			wantContains: []string{
				"Agent: reviewer",
				"Description: Reviews code",
				"Mode: edit",
				"Temperature: 0.30",
				"Installed On:",
				"OpenCode",
				"Instructions Preview:",
				"Be thorough.",
			},
		},
		{
			name: "minimal detail",
			detail: &showAgentDetail{
				Name: "minimal",
			},
			wantContains:   []string{"Agent: minimal"},
			wantNotContain: []string{"Description:", "Mode:", "Temperature:", "Instructions Preview:"},
		},
		{
			name: "claude agent (no mode/temperature)",
			detail: &showAgentDetail{
				Name:         "claude-agent",
				Description:  "A Claude agent",
				Instructions: "Do things.",
			},
			wantContains:   []string{"Agent: claude-agent", "Description:", "Instructions Preview:"},
			wantNotContain: []string{"Mode:", "Temperature:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := outputAgentShowText(&buf, tt.detail)
			if err != nil {
				t.Fatalf("outputAgentShowText() error = %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(output, notWant) {
					t.Errorf("output should not contain %q, got:\n%s", notWant, output)
				}
			}
		})
	}
}

func TestShowAgentDetailJSONTags(t *testing.T) {
	// Test that showAgentDetail has correct JSON tags and omitempty works
	detail := showAgentDetail{
		Name:         "test",
		Instructions: "test instructions",
		// Description, Mode, Temperature omitted
	}

	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("failed to marshal showAgentDetail: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Check that omitempty fields are not present when zero
	if _, ok := unmarshaled["description"]; ok {
		t.Error("description should be omitted when empty")
	}
	if _, ok := unmarshaled["mode"]; ok {
		t.Error("mode should be omitted when empty")
	}
	if _, ok := unmarshaled["temperature"]; ok {
		t.Error("temperature should be omitted when zero")
	}

	// Check required fields are present
	if _, ok := unmarshaled["name"]; !ok {
		t.Error("name should be present")
	}
	if _, ok := unmarshaled["instructions"]; !ok {
		t.Error("instructions should be present")
	}
}

func TestAgentShowFlags(t *testing.T) {
	// Test that the flag variables exist and have correct defaults
	jsonFlag := agentShowCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Fatal("--json flag not found")
	}
	if jsonFlag.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", jsonFlag.DefValue, "false")
	}

	fullFlag := agentShowCmd.Flags().Lookup("full")
	if fullFlag == nil {
		t.Fatal("--full flag not found")
	}
	if fullFlag.DefValue != "false" {
		t.Errorf("--full default = %q, want %q", fullFlag.DefValue, "false")
	}
}

func TestAgentInstallLocationJSONTags(t *testing.T) {
	loc := agentInstallLocation{
		Platform: "Claude Code",
		Path:     "/path/to/agent.md",
	}

	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatalf("failed to marshal agentInstallLocation: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled["platform"] != "Claude Code" {
		t.Errorf("platform = %v, want %q", unmarshaled["platform"], "Claude Code")
	}
	if unmarshaled["path"] != "/path/to/agent.md" {
		t.Errorf("path = %v, want %q", unmarshaled["path"], "/path/to/agent.md")
	}
}
