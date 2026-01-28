package cli

import (
	"errors"
	"testing"

	"github.com/thoreinstein/aix/internal/paths"
)

func TestNewPlatform(t *testing.T) {
	tests := []struct {
		name        string
		platformArg string
		wantName    string
		wantErr     error
	}{
		{
			name:        "claude platform",
			platformArg: "claude",
			wantName:    "claude",
			wantErr:     nil,
		},
		{
			name:        "opencode platform",
			platformArg: "opencode",
			wantName:    "opencode",
			wantErr:     nil,
		},
		{
			name:        "gemini platform",
			platformArg: "gemini",
			wantName:    "gemini",
			wantErr:     nil,
		},
		{
			name:        "unknown platform",
			platformArg: "unknown",
			wantName:    "",
			wantErr:     ErrUnknownPlatform,
		},
		{
			name:        "empty platform name",
			platformArg: "",
			wantName:    "",
			wantErr:     ErrUnknownPlatform,
		},
		{
			name:        "codex not supported yet",
			platformArg: "codex",
			wantName:    "",
			wantErr:     ErrUnknownPlatform,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPlatform(tt.platformArg)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewPlatform(%q) expected error, got nil", tt.platformArg)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewPlatform(%q) error = %v, want %v", tt.platformArg, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewPlatform(%q) unexpected error: %v", tt.platformArg, err)
				return
			}

			if p.Name() != tt.wantName {
				t.Errorf("NewPlatform(%q).Name() = %q, want %q", tt.platformArg, p.Name(), tt.wantName)
			}
		})
	}
}

func TestResolvePlatforms_EmptyNames(t *testing.T) {
	// When no names are provided, ResolvePlatforms should return detected platforms
	// or an error if no platforms are detected/available

	platforms, err := ResolvePlatforms(nil)

	// The result depends on the system where the test runs
	if err != nil {
		if !errors.Is(err, ErrNoPlatformsAvailable) {
			t.Errorf("ResolvePlatforms(nil) unexpected error: %v", err)
		}
		return
	}

	// Verify all returned platforms have adapters
	for _, p := range platforms {
		name := p.Name()
		if name != paths.PlatformClaude && name != paths.PlatformOpenCode && name != paths.PlatformGemini {
			t.Errorf("ResolvePlatforms(nil) returned unsupported platform: %q", name)
		}
	}
}

func TestResolvePlatforms_ValidNames(t *testing.T) {
	tests := []struct {
		name      string
		names     []string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "single valid platform",
			names:     []string{"claude"},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "multiple valid platforms",
			names:     []string{"claude", "opencode", "gemini"},
			wantCount: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platforms, err := ResolvePlatforms(tt.names)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolvePlatforms(%v) expected error, got nil", tt.names)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolvePlatforms(%v) unexpected error: %v", tt.names, err)
				return
			}

			if len(platforms) != tt.wantCount {
				t.Errorf("ResolvePlatforms(%v) returned %d platforms, want %d", tt.names, len(platforms), tt.wantCount)
			}
		})
	}
}

func TestResolvePlatforms_InvalidNames(t *testing.T) {
	tests := []struct {
		name  string
		names []string
	}{
		{
			name:  "single invalid platform",
			names: []string{"invalid"},
		},
		{
			name:  "mix of valid and invalid",
			names: []string{"claude", "invalid"},
		},
		{
			name:  "unsupported platform (codex)",
			names: []string{"codex"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResolvePlatforms(tt.names)

			if err == nil {
				t.Errorf("ResolvePlatforms(%v) expected error, got nil", tt.names)
				return
			}

			if !errors.Is(err, ErrUnknownPlatform) {
				t.Errorf("ResolvePlatforms(%v) error = %v, want ErrUnknownPlatform", tt.names, err)
			}
		})
	}
}

func TestPlatformInterface(t *testing.T) {
	// Verify that adapters properly implement the Platform interface
	platforms := []Platform{
		&claudeAdapter{},
		&opencodeAdapter{},
		&geminiAdapter{},
	}

	for _, p := range platforms {
		_ = p
	}
}

func TestSkillInfoFields(t *testing.T) {
	info := SkillInfo{
		Name:        "test-skill",
		Description: "A test skill",
		Source:      "local",
	}

	if info.Name != "test-skill" {
		t.Errorf("SkillInfo.Name = %q, want %q", info.Name, "test-skill")
	}
	if info.Description != "A test skill" {
		t.Errorf("SkillInfo.Description = %q, want %q", info.Description, "A test skill")
	}
	if info.Source != "local" {
		t.Errorf("SkillInfo.Source = %q, want %q", info.Source, "local")
	}
}

func TestCommandInfoFields(t *testing.T) {
	info := CommandInfo{
		Name:        "test-command",
		Description: "A test command",
		Source:      "installed",
	}

	if info.Name != "test-command" {
		t.Errorf("CommandInfo.Name = %q, want %q", info.Name, "test-command")
	}
	if info.Description != "A test command" {
		t.Errorf("CommandInfo.Description = %q, want %q", info.Description, "A test command")
	}
	if info.Source != "installed" {
		t.Errorf("CommandInfo.Source = %q, want %q", info.Source, "installed")
	}
}

func TestClaudeAdapter_InstallCommand_WrongType(t *testing.T) {
	p, err := NewPlatform("claude")
	if err != nil {
		t.Fatalf("NewPlatform(claude) unexpected error: %v", err)
	}

	err = p.InstallCommand("not a command", ScopeUser)
	if err == nil {
		t.Error("InstallCommand with wrong type expected error, got nil")
	}
}

func TestOpencodeAdapter_InstallCommand_WrongType(t *testing.T) {
	p, err := NewPlatform("opencode")
	if err != nil {
		t.Fatalf("NewPlatform(opencode) unexpected error: %v", err)
	}

	err = p.InstallCommand("not a command", ScopeUser)
	if err == nil {
		t.Error("InstallCommand with wrong type expected error, got nil")
	}
}

func TestGeminiAdapter_InstallCommand_WrongType(t *testing.T) {
	p, err := NewPlatform("gemini")
	if err != nil {
		t.Fatalf("NewPlatform(gemini) unexpected error: %v", err)
	}

	err = p.InstallCommand("not a command", ScopeUser)
	if err == nil {
		t.Error("InstallCommand with wrong type expected error, got nil")
	}
}
