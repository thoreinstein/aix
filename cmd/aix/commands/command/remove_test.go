package command

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/cli/mocks"
	"github.com/thoreinstein/aix/internal/errors"
)

func TestFindPlatformsWithCommand(t *testing.T) {
	t.Run("command found on one platform", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetCommand", "review", cli.ScopeDefault).Return(struct{}{}, nil)

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetCommand", "review", cli.ScopeDefault).Return(nil, errors.New("not found"))

		platforms := []cli.Platform{mockP1, mockP2}
		got := findPlatformsWithCommand(platforms, "review")
		if len(got) != 1 {
			t.Errorf("findPlatformsWithCommand() returned %d platforms, want 1", len(got))
		}
	})

	t.Run("command found on all platforms", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetCommand", "review", cli.ScopeDefault).Return(struct{}{}, nil)

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetCommand", "review", cli.ScopeDefault).Return(struct{}{}, nil)

		platforms := []cli.Platform{mockP1, mockP2}
		got := findPlatformsWithCommand(platforms, "review")
		if len(got) != 2 {
			t.Errorf("findPlatformsWithCommand() returned %d platforms, want 2", len(got))
		}
	})

	t.Run("command not found on any platform", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetCommand", "review", cli.ScopeDefault).Return(nil, errors.New("not found"))

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetCommand", "review", cli.ScopeDefault).Return(nil, errors.New("not found"))

		platforms := []cli.Platform{mockP1, mockP2}
		got := findPlatformsWithCommand(platforms, "review")
		if len(got) != 0 {
			t.Errorf("findPlatformsWithCommand() returned %d platforms, want 0", len(got))
		}
	})

	t.Run("no platforms", func(t *testing.T) {
		platforms := []cli.Platform{}
		got := findPlatformsWithCommand(platforms, "review")
		if len(got) != 0 {
			t.Errorf("findPlatformsWithCommand() returned %d platforms, want 0", len(got))
		}
	})

	t.Run("different commands on different platforms", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetCommand", "deploy", cli.ScopeDefault).Return(nil, errors.New("not found"))

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetCommand", "deploy", cli.ScopeDefault).Return(nil, errors.New("not found"))

		platforms := []cli.Platform{mockP1, mockP2}
		got := findPlatformsWithCommand(platforms, "deploy")
		if len(got) != 0 {
			t.Errorf("findPlatformsWithCommand() returned %d platforms, want 0", len(got))
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

			got := confirmRemoval(&out, in, "test-command", []cli.Platform{mockP})
			if got != tt.want {
				t.Errorf("confirmRemoval() = %v, want %v", got, tt.want)
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

func TestConfirmRemoval_ListsPlatforms(t *testing.T) {
	mockP1 := mocks.NewMockPlatform(t)
	mockP1.On("DisplayName").Return("Claude Code")

	mockP2 := mocks.NewMockPlatform(t)
	mockP2.On("DisplayName").Return("OpenCode")

	platforms := []cli.Platform{mockP1, mockP2}

	var out bytes.Buffer
	in := strings.NewReader("n\n")

	confirmRemoval(&out, in, "review", platforms)

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

func TestRunRemoveWithIO_CommandNotFound(t *testing.T) {
	// This test would require mocking cli.ResolvePlatforms
	// which is difficult without dependency injection.
	// For now, we skip this as it would require broader refactoring.
	t.Skip("requires mocking cli.ResolvePlatforms")
}
