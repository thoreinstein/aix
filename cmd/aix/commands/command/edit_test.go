package command

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/cli/mocks"
	"github.com/thoreinstein/aix/internal/errors"
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
	t.Run("command found on first platform", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetCommand", "review", cli.ScopeUser).Return(struct{}{}, nil)
		mockP1.On("CommandDir").Return("/mock/commands")

		mockP2 := mocks.NewMockPlatform(t)

		platforms := []cli.Platform{mockP1, mockP2}
		wantPath := "/mock/commands/review.md"

		var openedPath string
		mockOpener := func(path string) error {
			openedPath = path
			return nil
		}

		err := runEditWithPlatforms("review", platforms, cli.ScopeUser, mockOpener)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if filepath.Clean(openedPath) != filepath.Clean(wantPath) {
			t.Errorf("openedPath = %q, want %q", openedPath, wantPath)
		}
	})

	t.Run("command found on second platform", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetCommand", "review", cli.ScopeUser).Return(nil, errors.New("not found"))

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetCommand", "review", cli.ScopeUser).Return(struct{}{}, nil)
		mockP2.On("CommandDir").Return("/mock/commands")

		platforms := []cli.Platform{mockP1, mockP2}
		wantPath := "/mock/commands/review.md"

		var openedPath string
		mockOpener := func(path string) error {
			openedPath = path
			return nil
		}

		err := runEditWithPlatforms("review", platforms, cli.ScopeUser, mockOpener)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if filepath.Clean(openedPath) != filepath.Clean(wantPath) {
			t.Errorf("openedPath = %q, want %q", openedPath, wantPath)
		}
	})

	t.Run("command found on both, opens first", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetCommand", "review", cli.ScopeUser).Return(struct{}{}, nil)
		mockP1.On("CommandDir").Return("/mock/commands")

		mockP2 := mocks.NewMockPlatform(t)
		// mockP2 won't be called

		platforms := []cli.Platform{mockP1, mockP2}
		wantPath := "/mock/commands/review.md"

		var openedPath string
		mockOpener := func(path string) error {
			openedPath = path
			return nil
		}

		err := runEditWithPlatforms("review", platforms, cli.ScopeUser, mockOpener)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if filepath.Clean(openedPath) != filepath.Clean(wantPath) {
			t.Errorf("openedPath = %q, want %q", openedPath, wantPath)
		}
	})

	t.Run("command not found", func(t *testing.T) {
		mockP1 := mocks.NewMockPlatform(t)
		mockP1.On("GetCommand", "review", cli.ScopeUser).Return(nil, errors.New("not found"))

		mockP2 := mocks.NewMockPlatform(t)
		mockP2.On("GetCommand", "review", cli.ScopeUser).Return(nil, errors.New("not found"))

		platforms := []cli.Platform{mockP1, mockP2}
		wantErr := `command "review" not found on any platform`

		err := runEditWithPlatforms("review", platforms, cli.ScopeUser, func(s string) error { return nil })
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), wantErr) {
			t.Errorf("error = %q, want to contain %q", err.Error(), wantErr)
		}
	})
}
