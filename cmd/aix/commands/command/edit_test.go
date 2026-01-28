package command

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
)

func TestEditCommand_Metadata(t *testing.T) {
	if editCmd.Use != "edit <name>" {
		t.Errorf("Use = %q, want %q", editCmd.Use, "edit <name>")
	}

	if editCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if editCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestRunEditWithPlatforms(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		platforms   []cli.Platform
		wantPath    string
		wantErr     string
	}{
		{
			name:        "command found on first platform",
			commandName: "review",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", commands: map[string]any{"review": struct{}{}}},
				&mockPlatform{name: "opencode", commands: map[string]any{}},
			},
			wantPath: "/mock/commands/review.md",
		},
		{
			name:        "command found on second platform",
			commandName: "review",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", commands: map[string]any{}},
				&mockPlatform{name: "opencode", commands: map[string]any{"review": struct{}{}}},
			},
			wantPath: "/mock/commands/review.md",
		},
		{
			name:        "command found on both, opens first",
			commandName: "review",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", commands: map[string]any{"review": struct{}{}}},
				&mockPlatform{name: "opencode", commands: map[string]any{"review": struct{}{}}},
			},
			wantPath: "/mock/commands/review.md",
		},
		{
			name:        "command not found",
			commandName: "review",
			platforms: []cli.Platform{
				&mockPlatform{name: "claude", commands: map[string]any{}},
				&mockPlatform{name: "opencode", commands: map[string]any{}},
			},
			wantErr: `command "review" not found on any platform`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var openedPath string
			mockOpener := func(path string) error {
				openedPath = path
				return nil
			}

			err := runEditWithPlatforms(tt.commandName, tt.platforms, cli.ScopeUser, mockOpener)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// filepath.Clean used to handle platform differences if any,
			// though here we use forward slashes in mocks
			if filepath.Clean(openedPath) != filepath.Clean(tt.wantPath) {
				t.Errorf("openedPath = %q, want %q", openedPath, tt.wantPath)
			}
		})
	}
}
