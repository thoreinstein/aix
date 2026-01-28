package skill

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
)

// mockPlatform implements cli.Platform for testing.
type mockPlatform struct {
	name         string
	displayName  string
	skills       map[string]any
	commands     map[string]any
	uninstallErr error
}

func (m *mockPlatform) Name() string        { return m.name }
func (m *mockPlatform) DisplayName() string { return m.displayName }
func (m *mockPlatform) IsAvailable() bool   { return true }
func (m *mockPlatform) SkillDir() string    { return "/mock/skills" }

func (m *mockPlatform) InstallSkill(_ any, _ cli.Scope) error { return nil }

func (m *mockPlatform) UninstallSkill(name string, _ cli.Scope) error {
	if m.uninstallErr != nil {
		return m.uninstallErr
	}
	delete(m.skills, name)
	return nil
}

func (m *mockPlatform) ListSkills() ([]cli.SkillInfo, error) {
	skills := make([]cli.SkillInfo, 0, len(m.skills))
	for name := range m.skills {
		skills = append(skills, cli.SkillInfo{Name: name})
	}
	return skills, nil
}

func (m *mockPlatform) GetSkill(name string) (any, error) {
	skill, ok := m.skills[name]
	if !ok {
		return nil, errors.New("skill not found")
	}
	return skill, nil
}

func (m *mockPlatform) CommandDir() string { return "/mock/commands" }

func (m *mockPlatform) InstallCommand(_ any, _ cli.Scope) error { return nil }

func (m *mockPlatform) UninstallCommand(name string, _ cli.Scope) error {
	delete(m.commands, name)
	return nil
}

func (m *mockPlatform) ListCommands() ([]cli.CommandInfo, error) {
	commands := make([]cli.CommandInfo, 0, len(m.commands))
	for name := range m.commands {
		commands = append(commands, cli.CommandInfo{Name: name})
	}
	return commands, nil
}

func (m *mockPlatform) GetCommand(name string) (any, error) {
	cmd, ok := m.commands[name]
	if !ok {
		return nil, errors.New("command not found")
	}
	return cmd, nil
}

// MCP methods for cli.Platform interface
func (m *mockPlatform) MCPConfigPath() string                 { return "/mock/mcp.json" }
func (m *mockPlatform) AddMCP(_ any, _ cli.Scope) error       { return nil }
func (m *mockPlatform) RemoveMCP(_ string, _ cli.Scope) error { return nil }
func (m *mockPlatform) ListMCP() ([]cli.MCPInfo, error)       { return nil, nil }
func (m *mockPlatform) GetMCP(_ string) (any, error)          { return nil, errors.New("not found") }
func (m *mockPlatform) EnableMCP(_ string) error              { return nil }
func (m *mockPlatform) DisableMCP(_ string) error             { return nil }

// Agent methods for cli.Platform interface
func (m *mockPlatform) AgentDir() string                           { return "/mock/agents" }
func (m *mockPlatform) InstallAgent(_ any, _ cli.Scope) error      { return nil }
func (m *mockPlatform) UninstallAgent(_ string, _ cli.Scope) error { return nil }
func (m *mockPlatform) ListAgents() ([]cli.AgentInfo, error)       { return nil, nil }
func (m *mockPlatform) GetAgent(_ string) (any, error)             { return nil, errors.New("not found") }

// Backup methods for cli.Platform interface
func (m *mockPlatform) BackupPaths() []string { return []string{"/mock/backup"} }

func TestFindPlatformsWithSkill(t *testing.T) {
	tests := []struct {
		name      string
		platforms []cli.Platform
		skillName string
		wantCount int
	}{
		{
			name: "skill found on one platform",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", skills: map[string]any{"debug": struct{}{}}},
				&mockPlatform{name: "opencode", skills: map[string]any{}},
			},
			skillName: "debug",
			wantCount: 1,
		},
		{
			name: "skill found on all platforms",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", skills: map[string]any{"debug": struct{}{}}},
				&mockPlatform{name: "opencode", skills: map[string]any{"debug": struct{}{}}},
			},
			skillName: "debug",
			wantCount: 2,
		},
		{
			name: "skill not found on any platform",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", skills: map[string]any{}},
				&mockPlatform{name: "opencode", skills: map[string]any{}},
			},
			skillName: "debug",
			wantCount: 0,
		},
		{
			name:      "no platforms",
			platforms: []cli.Platform{},
			skillName: "debug",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findPlatformsWithSkill(tt.platforms, tt.skillName)
			if len(got) != tt.wantCount {
				t.Errorf("findPlatformsWithSkill() returned %d platforms, want %d", len(got), tt.wantCount)
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
				&mockPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "y confirms",
			input: "y\n",
			platforms: []cli.Platform{
				&mockPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "Y confirms (case insensitive)",
			input: "Y\n",
			platforms: []cli.Platform{
				&mockPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "YES confirms (case insensitive)",
			input: "YES\n",
			platforms: []cli.Platform{
				&mockPlatform{displayName: "Claude Code"},
			},
			want: true,
		},
		{
			name:  "no rejects",
			input: "no\n",
			platforms: []cli.Platform{
				&mockPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "n rejects",
			input: "n\n",
			platforms: []cli.Platform{
				&mockPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "empty input rejects",
			input: "\n",
			platforms: []cli.Platform{
				&mockPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
		{
			name:  "random input rejects",
			input: "maybe\n",
			platforms: []cli.Platform{
				&mockPlatform{displayName: "Claude Code"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			in := strings.NewReader(tt.input)

			got := confirmRemoval(&out, in, "test-skill", tt.platforms)
			if got != tt.want {
				t.Errorf("confirmRemoval() = %v, want %v", got, tt.want)
			}

			// Verify prompt was written
			output := out.String()
			if !strings.Contains(output, "test-skill") {
				t.Error("prompt should contain skill name")
			}
			if !strings.Contains(output, "[y/N]") {
				t.Error("prompt should contain [y/N]")
			}
		})
	}
}

func TestConfirmRemoval_ListsPlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&mockPlatform{displayName: "Claude Code"},
		&mockPlatform{displayName: "OpenCode"},
	}

	var out bytes.Buffer
	in := strings.NewReader("n\n")

	confirmRemoval(&out, in, "debug", platforms)

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
}
