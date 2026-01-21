package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
)

func TestFindPlatformsWithCommand(t *testing.T) {
	tests := []struct {
		name        string
		platforms   []cli.Platform
		commandName string
		wantCount   int
	}{
		{
			name: "command found on one platform",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", commands: map[string]any{"review": struct{}{}}},
				&mockPlatform{name: "opencode", commands: map[string]any{}},
			},
			commandName: "review",
			wantCount:   1,
		},
		{
			name: "command found on all platforms",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", commands: map[string]any{"review": struct{}{}}},
				&mockPlatform{name: "opencode", commands: map[string]any{"review": struct{}{}}},
			},
			commandName: "review",
			wantCount:   2,
		},
		{
			name: "command not found on any platform",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", commands: map[string]any{}},
				&mockPlatform{name: "opencode", commands: map[string]any{}},
			},
			commandName: "review",
			wantCount:   0,
		},
		{
			name:        "no platforms",
			platforms:   []cli.Platform{},
			commandName: "review",
			wantCount:   0,
		},
		{
			name: "different commands on different platforms",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", commands: map[string]any{"review": struct{}{}}},
				&mockPlatform{name: "opencode", commands: map[string]any{"build": struct{}{}}},
			},
			commandName: "deploy",
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findPlatformsWithCommand(tt.platforms, tt.commandName)
			if len(got) != tt.wantCount {
				t.Errorf("findPlatformsWithCommand() returned %d platforms, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestConfirmCommandRemoval(t *testing.T) {
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

			got := confirmCommandRemoval(&out, in, "test-command", tt.platforms)
			if got != tt.want {
				t.Errorf("confirmCommandRemoval() = %v, want %v", got, tt.want)
			}

			// Verify prompt was written
			output := out.String()
			if !strings.Contains(output, "test-command") {
				t.Error("prompt should contain command name")
			}
			if !strings.Contains(output, "[y/N]") {
				t.Error("prompt should contain [y/N]")
			}
		})
	}
}

func TestConfirmCommandRemoval_ListsPlatforms(t *testing.T) {
	platforms := []cli.Platform{
		&mockPlatform{displayName: "Claude Code"},
		&mockPlatform{displayName: "OpenCode"},
	}

	var out bytes.Buffer
	in := strings.NewReader("n\n")

	confirmCommandRemoval(&out, in, "review", platforms)

	output := out.String()
	if !strings.Contains(output, "Claude Code") {
		t.Error("output should list Claude Code")
	}
	if !strings.Contains(output, "OpenCode") {
		t.Error("output should list OpenCode")
	}
}

func TestCommandRemoveCommand_Metadata(t *testing.T) {
	if commandRemoveCmd.Use != "remove <name>" {
		t.Errorf("Use = %q, want %q", commandRemoveCmd.Use, "remove <name>")
	}

	if commandRemoveCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if commandRemoveCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestRunCommandRemoveWithIO_CommandNotFound(t *testing.T) {
	// This test would require mocking cli.ResolvePlatforms
	// which is difficult without dependency injection.
	// For now, we skip this as it would require broader refactoring.
	t.Skip("requires mocking cli.ResolvePlatforms")
}
