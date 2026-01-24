package mcp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
)

// removeMockPlatform extends mockPlatform for MCP remove testing.
type removeMockPlatform struct {
	mockPlatform
	mcpServers   map[string]any
	removeErr    error
	removeCalled bool
}

func (m *removeMockPlatform) GetMCP(name string) (any, error) {
	server, ok := m.mcpServers[name]
	if !ok {
		return nil, errors.New("MCP server not found")
	}
	return server, nil
}

func (m *removeMockPlatform) RemoveMCP(_ string) error {
	m.removeCalled = true
	return m.removeErr
}

func TestFindPlatformsWithMCP(t *testing.T) {
	tests := []struct {
		name       string
		platforms  []cli.Platform
		serverName string
		wantCount  int
	}{
		{
			name: "server found on one platform",
			platforms: []cli.Platform{
				&removeMockPlatform{
					mockPlatform: mockPlatform{name: "claude"},
					mcpServers:   map[string]any{"github": struct{}{}},
				},
				&removeMockPlatform{
					mockPlatform: mockPlatform{name: "opencode"},
					mcpServers:   map[string]any{},
				},
			},
			serverName: "github",
			wantCount:  1,
		},
		{
			name: "server found on all platforms",
			platforms: []cli.Platform{
				&removeMockPlatform{
					mockPlatform: mockPlatform{name: "claude"},
					mcpServers:   map[string]any{"github": struct{}{}},
				},
				&removeMockPlatform{
					mockPlatform: mockPlatform{name: "opencode"},
					mcpServers:   map[string]any{"github": struct{}{}},
				},
			},
			serverName: "github",
			wantCount:  2,
		},
		{
			name: "server not found on any platform",
			platforms: []cli.Platform{
				&removeMockPlatform{
					mockPlatform: mockPlatform{name: "claude"},
					mcpServers:   map[string]any{},
				},
				&removeMockPlatform{
					mockPlatform: mockPlatform{name: "opencode"},
					mcpServers:   map[string]any{},
				},
			},
			serverName: "github",
			wantCount:  0,
		},
		{
			name:       "no platforms",
			platforms:  []cli.Platform{},
			serverName: "github",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findPlatformsWithMCP(tt.platforms, tt.serverName)
			if len(got) != tt.wantCount {
				t.Errorf("findPlatformsWithMCP() returned %d platforms, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestConfirmMCPRemoval(t *testing.T) {
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
				&removeMockPlatform{mockPlatform: mockPlatform{displayName: "Claude Code"}},
			},
			want: true,
		},
		{
			name:  "y confirms",
			input: "y\n",
			platforms: []cli.Platform{
				&removeMockPlatform{mockPlatform: mockPlatform{displayName: "Claude Code"}},
			},
			want: true,
		},
		{
			name:  "Y confirms (case insensitive)",
			input: "Y\n",
			platforms: []cli.Platform{
				&removeMockPlatform{mockPlatform: mockPlatform{displayName: "Claude Code"}},
			},
			want: true,
		},
		{
			name:  "no rejects",
			input: "no\n",
			platforms: []cli.Platform{
				&removeMockPlatform{mockPlatform: mockPlatform{displayName: "Claude Code"}},
			},
			want: false,
		},
		{
			name:  "empty input rejects (default N)",
			input: "\n",
			platforms: []cli.Platform{
				&removeMockPlatform{mockPlatform: mockPlatform{displayName: "Claude Code"}},
			},
			want: false,
		},
		{
			name:  "random input rejects",
			input: "maybe\n",
			platforms: []cli.Platform{
				&removeMockPlatform{mockPlatform: mockPlatform{displayName: "Claude Code"}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			in := strings.NewReader(tt.input)

			got := confirmRemoval(&out, in, "github", tt.platforms)
			if got != tt.want {
				t.Errorf("confirmMCPRemoval() = %v, want %v", got, tt.want)
			}

			// Verify prompt was written
			output := out.String()
			if !strings.Contains(output, "github") {
				t.Error("prompt should contain server name")
			}
			if !strings.Contains(output, "[y/N]") {
				t.Error("prompt should contain [y/N]")
			}
		})
	}
}

func TestConfirmMCPRemoval_ListsPlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&removeMockPlatform{mockPlatform: mockPlatform{displayName: "Claude Code"}},
		&removeMockPlatform{mockPlatform: mockPlatform{displayName: "OpenCode"}},
	}

	var out bytes.Buffer
	in := strings.NewReader("n\n")

	confirmRemoval(&out, in, "github", platforms)

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
