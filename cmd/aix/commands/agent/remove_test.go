package agent

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/cli/mocks"
	"github.com/thoreinstein/aix/internal/errors"
)

func TestFindPlatformsWithAgent(t *testing.T) {
	t.Run("agent found on one platform", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetAgent", "my-agent", cli.ScopeDefault).Return(struct{}{}, nil)

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetAgent", "my-agent", cli.ScopeDefault).Return(nil, errors.New("not found"))

		platforms := []cli.Platform{mockP1, mockP2}
		got := findPlatformsWithAgent(platforms, "my-agent")
		if len(got) != 1 {
			t.Errorf("findPlatformsWithAgent() returned %d platforms, want 1", len(got))
		}
	})

	t.Run("agent found on all platforms", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetAgent", "my-agent", cli.ScopeDefault).Return(struct{}{}, nil)

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetAgent", "my-agent", cli.ScopeDefault).Return(struct{}{}, nil)

		platforms := []cli.Platform{mockP1, mockP2}
		got := findPlatformsWithAgent(platforms, "my-agent")
		if len(got) != 2 {
			t.Errorf("findPlatformsWithAgent() returned %d platforms, want 2", len(got))
		}
	})

	t.Run("agent not found on any platform", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetAgent", "my-agent", cli.ScopeDefault).Return(nil, errors.New("not found"))

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetAgent", "my-agent", cli.ScopeDefault).Return(nil, errors.New("not found"))

		platforms := []cli.Platform{mockP1, mockP2}
		got := findPlatformsWithAgent(platforms, "my-agent")
		if len(got) != 0 {
			t.Errorf("findPlatformsWithAgent() returned %d platforms, want 0", len(got))
		}
	})

	t.Run("no platforms", func(t *testing.T) {
		platforms := []cli.Platform{}
		got := findPlatformsWithAgent(platforms, "my-agent")
		if len(got) != 0 {
			t.Errorf("findPlatformsWithAgent() returned %d platforms, want 0", len(got))
		}
	})
}

func TestConfirmRemoval(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "yes confirms",
			input: "yes\n",
			want:  true,
		},
		{
			name:  "y confirms",
			input: "y\n",
			want:  true,
		},
		{
			name:  "Y confirms (case insensitive)",
			input: "Y\n",
			want:  true,
		},
		{
			name:  "YES confirms (case insensitive)",
			input: "YES\n",
			want:  true,
		},
		{
			name:  "no rejects",
			input: "no\n",
			want:  false,
		},
		{
			name:  "n rejects",
			input: "n\n",
			want:  false,
		},
		{
			name:  "empty input rejects",
			input: "\n",
			want:  false,
		},
		{
			name:  "random input rejects",
			input: "maybe\n",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockP := mocks.NewMockPlatform(t)
			mockP.On("DisplayName").Return("Claude Code")

			var out bytes.Buffer
			in := strings.NewReader(tt.input)

			got := confirmRemoval(&out, in, "test-agent", []cli.Platform{mockP})
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
	mockP1 := mocks.NewMockPlatform(t)
	mockP1.On("DisplayName").Return("Claude Code")

	mockP2 := mocks.NewMockPlatform(t)
	mockP2.On("DisplayName").Return("OpenCode")

	platforms := []cli.Platform{mockP1, mockP2}

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
