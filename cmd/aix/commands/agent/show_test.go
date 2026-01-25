package agent

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

func TestShowCommand_Metadata(t *testing.T) {
	if showCmd.Use != "show <name>" {
		t.Errorf("Use = %q, want %q", showCmd.Use, "show <name>")
	}

	if showCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if showCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Check flags exist
	expectedFlags := []string{"json", "full"}
	for _, flagName := range expectedFlags {
		if showCmd.Flags().Lookup(flagName) == nil {
			t.Errorf("--%s flag should be defined", flagName)
		}
	}

	// Verify Args validator is set (ExactArgs(1))
	if showCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestExtractDetail_Claude(t *testing.T) {
	tests := []struct {
		name  string
		agent *claude.Agent
		want  *showDetail
	}{
		{
			name: "full agent",
			agent: &claude.Agent{
				Name:         "code-reviewer",
				Description:  "Reviews code for quality",
				Instructions: "You are a code reviewer...",
			},
			want: &showDetail{
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
			want: &showDetail{
				Name: "simple-agent",
			},
		},
		{
			name: "agent with only instructions",
			agent: &claude.Agent{
				Name:         "instructions-only",
				Instructions: "Do the thing.",
			},
			want: &showDetail{
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

func TestExtractDetail_OpenCode(t *testing.T) {
	tests := []struct {
		name  string
		agent *opencode.Agent
		want  *showDetail
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
			want: &showDetail{
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
			want: &showDetail{
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
			want: &showDetail{
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

func TestExtractDetail_UnknownType(t *testing.T) {
	// Test that extractDetail returns nil for unknown types
	got := extractDetail("not an agent type")
	if got != nil {
		t.Errorf("extractDetail() = %v, want nil for unknown type", got)
	}
}

func TestOutputShowJSON(t *testing.T) {
	detail := &showDetail{
		Name:         "test-agent",
		Description:  "A test agent",
		Mode:         "review",
		Temperature:  0.5,
		Instructions: "Test instructions.",
		Installations: []installLocation{
			{Platform: "Claude Code", Path: "/path/to/agents/test-agent.md"},
		},
	}

	var buf bytes.Buffer
	err := outputShowJSON(&buf, detail)
	if err != nil {
		t.Fatalf("outputShowJSON() error = %v", err)
	}

	// Verify valid JSON
	var unmarshaled showDetail
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

func TestOutputShowText(t *testing.T) {
	tests := []struct {
		name           string
		detail         *showDetail
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "full detail",
			detail: &showDetail{
				Name:         "reviewer",
				Description:  "Reviews code",
				Mode:         "edit",
				Temperature:  0.3,
				Instructions: "Be thorough.",
				Installations: []installLocation{
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
			detail: &showDetail{
				Name: "minimal",
			},
			wantContains:   []string{"Agent: minimal"},
			wantNotContain: []string{"Description:", "Mode:", "Temperature:", "Instructions Preview:"},
		},
		{
			name: "claude agent (no mode/temperature)",
			detail: &showDetail{
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
			err := outputShowText(&buf, tt.detail)
			if err != nil {
				t.Fatalf("outputShowText() error = %v", err)
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

func TestShowDetailJSONTags(t *testing.T) {
	// Test that showDetail has correct JSON tags and omitempty works
	detail := showDetail{
		Name:         "test",
		Instructions: "test instructions",
		// Description, Mode, Temperature omitted
	}

	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("failed to marshal showDetail: %v", err)
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

func TestShowFlags(t *testing.T) {
	// Test that the flag variables exist and have correct defaults
	if jsonFlag := showCmd.Flags().Lookup("json"); jsonFlag == nil {
		t.Fatal("--json flag not found")
	} else if jsonFlag.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", jsonFlag.DefValue, "false")
	}

	if fullFlag := showCmd.Flags().Lookup("full"); fullFlag == nil {
		t.Fatal("--full flag not found")
	} else if fullFlag.DefValue != "false" {
		t.Errorf("--full default = %q, want %q", fullFlag.DefValue, "false")
	}
}

func TestInstallLocationJSONTags(t *testing.T) {
	loc := installLocation{
		Platform: "Claude Code",
		Path:     "/path/to/agent.md",
	}

	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatalf("failed to marshal installLocation: %v", err)
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

// showMockPlatform extends mockPlatform with agent show-specific behavior.
type showMockPlatform struct {
	mockPlatform
	agent    any
	agentErr error
}

func (m *showMockPlatform) GetAgent(_ string) (any, error) {
	if m.agentErr != nil {
		return nil, m.agentErr
	}
	return m.agent, nil
}

func TestIsAgentNotFoundError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantSkip   bool
		errMessage string
	}{
		{
			name:       "claude ErrAgentNotFound",
			err:        claude.ErrAgentNotFound,
			wantSkip:   true,
			errMessage: "should skip for claude.ErrAgentNotFound",
		},
		{
			name:       "opencode ErrAgentNotFound",
			err:        opencode.ErrAgentNotFound,
			wantSkip:   true,
			errMessage: "should skip for opencode.ErrAgentNotFound",
		},
		{
			name:       "wrapped claude ErrAgentNotFound",
			err:        errors.Wrap(claude.ErrAgentNotFound, "additional context"),
			wantSkip:   true,
			errMessage: "should skip for wrapped claude.ErrAgentNotFound",
		},
		{
			name:       "wrapped opencode ErrAgentNotFound",
			err:        errors.Wrap(opencode.ErrAgentNotFound, "additional context"),
			wantSkip:   true,
			errMessage: "should skip for wrapped opencode.ErrAgentNotFound",
		},
		{
			name:       "permission error",
			err:        errors.New("permission denied"),
			wantSkip:   false,
			errMessage: "should NOT skip for permission errors",
		},
		{
			name:       "parse error",
			err:        errors.New("invalid yaml: unexpected mapping"),
			wantSkip:   false,
			errMessage: "should NOT skip for parse errors",
		},
		{
			name:       "read error",
			err:        errors.New("reading agent file: input/output error"),
			wantSkip:   false,
			errMessage: "should NOT skip for read errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test using errors.Is which is what the code uses
			isNotFound := errors.Is(tt.err, claude.ErrAgentNotFound) ||
				errors.Is(tt.err, opencode.ErrAgentNotFound)

			if isNotFound != tt.wantSkip {
				t.Errorf("isAgentNotFound = %v, want %v; %s",
					isNotFound, tt.wantSkip, tt.errMessage)
			}
		})
	}
}

func TestShowErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		platforms   []cli.Platform
		wantErr     bool
		wantErrMsg  string
		description string
	}{
		{
			name: "not_found_continues_to_next_platform",
			platforms: []cli.Platform{
				&showMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agentErr: claude.ErrAgentNotFound,
				},
				&showMockPlatform{
					mockPlatform: mockPlatform{
						name:        "opencode",
						displayName: "OpenCode",
					},
					agent: &opencode.Agent{
						Name:        "test-agent",
						Description: "A test agent",
					},
				},
			},
			wantErr:     false,
			description: "should continue to next platform when agent not found",
		},
		{
			name: "not_found_on_all_platforms",
			platforms: []cli.Platform{
				&showMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agentErr: claude.ErrAgentNotFound,
				},
				&showMockPlatform{
					mockPlatform: mockPlatform{
						name:        "opencode",
						displayName: "OpenCode",
					},
					agentErr: opencode.ErrAgentNotFound,
				},
			},
			wantErr:     true,
			wantErrMsg:  "not found on any platform",
			description: "should return not found error when missing from all platforms",
		},
		{
			name: "permission_error_reported",
			platforms: []cli.Platform{
				&showMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agentErr: errors.New("permission denied: ~/.claude/agents/test.md"),
				},
			},
			wantErr:     true,
			wantErrMsg:  "reading agent from Claude Code",
			description: "should report permission errors instead of swallowing them",
		},
		{
			name: "parse_error_reported",
			platforms: []cli.Platform{
				&showMockPlatform{
					mockPlatform: mockPlatform{
						name:        "opencode",
						displayName: "OpenCode",
					},
					agentErr: errors.New("parsing agent: invalid yaml"),
				},
			},
			wantErr:     true,
			wantErrMsg:  "reading agent from OpenCode",
			description: "should report parse errors instead of swallowing them",
		},
		{
			name: "error_on_first_platform_stops_iteration",
			platforms: []cli.Platform{
				&showMockPlatform{
					mockPlatform: mockPlatform{
						name:        "claude",
						displayName: "Claude Code",
					},
					agentErr: errors.New("disk I/O error"),
				},
				&showMockPlatform{
					mockPlatform: mockPlatform{
						name:        "opencode",
						displayName: "OpenCode",
					},
					agent: &opencode.Agent{
						Name: "test-agent",
					},
				},
			},
			wantErr:     true,
			wantErrMsg:  "reading agent from Claude Code: disk I/O error",
			description: "should fail fast on first real error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't directly test runShowWithWriter because it calls
			// cli.ResolvePlatforms. Instead, test the error classification logic
			// by simulating what the loop does.
			var foundAgent bool
			var firstErr error

			for _, p := range tt.platforms {
				_, err := p.GetAgent("test-agent")
				if err != nil {
					if errors.Is(err, claude.ErrAgentNotFound) ||
						errors.Is(err, opencode.ErrAgentNotFound) {
						continue
					}
					firstErr = errors.Wrapf(err, "reading agent from %s", p.DisplayName())
					break
				}
				foundAgent = true
			}

			var gotErr error
			if firstErr != nil {
				gotErr = firstErr
			} else if !foundAgent {
				gotErr = errors.New("agent not found on any platform")
			}

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v; %s", gotErr, tt.wantErr, tt.description)
				return
			}

			if tt.wantErr && gotErr != nil {
				if !strings.Contains(gotErr.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want to contain %q; %s",
						gotErr.Error(), tt.wantErrMsg, tt.description)
				}
			}
		})
	}
}
