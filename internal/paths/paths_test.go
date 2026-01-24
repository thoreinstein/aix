package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/errors"
)

func TestHome(t *testing.T) {
	got := Home()
	want, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}
	if got != want {
		t.Errorf("Home() = %q, want %q", got, want)
	}
}

func TestResolveHome(t *testing.T) {
	got, err := ResolveHome()
	want, _ := os.UserHomeDir()

	if err != nil {
		// This might happen in some restricted environments,
		// but normally should succeed.
		if !errors.Is(err, ErrHomeDirNotFound) {
			t.Errorf("unexpected error type: %v", err)
		}
	} else if got != want {
		t.Errorf("ResolveHome() = %q, want %q", got, want)
	}
}

func TestConfigHome(t *testing.T) {
	got := ConfigHome()
	if got == "" {
		t.Error("ConfigHome() returned empty string")
	}
	// Verify it's an absolute path
	if !filepath.IsAbs(got) {
		t.Errorf("ConfigHome() = %q, want absolute path", got)
	}
}

func TestDataHome(t *testing.T) {
	got := DataHome()
	if got == "" {
		t.Error("DataHome() returned empty string")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("DataHome() = %q, want absolute path", got)
	}
}

func TestCacheHome(t *testing.T) {
	got := CacheHome()
	if got == "" {
		t.Error("CacheHome() returned empty string")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("CacheHome() = %q, want absolute path", got)
	}
}

func TestReposCacheDir(t *testing.T) {
	got := ReposCacheDir()
	if got == "" {
		t.Error("ReposCacheDir() returned empty string")
	}
	if !filepath.IsAbs(got) {
		t.Errorf("ReposCacheDir() = %q, want absolute path", got)
	}

	// Verify path ends with aix/repos
	wantSuffix := filepath.Join("aix", "repos")
	if !strings.HasSuffix(got, wantSuffix) {
		t.Errorf("ReposCacheDir() = %q, want path ending with %q", got, wantSuffix)
	}

	// Verify it's under CacheHome
	cacheHome := CacheHome()
	if !strings.HasPrefix(got, cacheHome) {
		t.Errorf("ReposCacheDir() = %q, want path under CacheHome %q", got, cacheHome)
	}
}

func TestValidPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     bool
	}{
		{
			name:     "claude is valid",
			platform: PlatformClaude,
			want:     true,
		},
		{
			name:     "opencode is valid",
			platform: PlatformOpenCode,
			want:     true,
		},
		{
			name:     "codex is valid",
			platform: PlatformCodex,
			want:     true,
		},
		{
			name:     "gemini is valid",
			platform: PlatformGemini,
			want:     true,
		},
		{
			name:     "unknown platform is invalid",
			platform: "unknown",
			want:     false,
		},
		{
			name:     "empty string is invalid",
			platform: "",
			want:     false,
		},
		{
			name:     "case sensitive - Claude is invalid",
			platform: "Claude",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidPlatform(tt.platform)
			if got != tt.want {
				t.Errorf("ValidPlatform(%q) = %v, want %v", tt.platform, got, tt.want)
			}
		})
	}
}

func TestPlatforms(t *testing.T) {
	platforms := Platforms()

	if len(platforms) != 4 {
		t.Errorf("Platforms() returned %d platforms, want 4", len(platforms))
	}

	// Verify all expected platforms are present
	expected := map[string]bool{
		PlatformClaude:   false,
		PlatformOpenCode: false,
		PlatformCodex:    false,
		PlatformGemini:   false,
	}

	for _, p := range platforms {
		if _, ok := expected[p]; !ok {
			t.Errorf("Platforms() contains unexpected platform %q", p)
		}
		expected[p] = true
	}

	for p, found := range expected {
		if !found {
			t.Errorf("Platforms() missing expected platform %q", p)
		}
	}
}

func TestGlobalConfigDir(t *testing.T) {
	home := Home()
	if home == "" {
		t.Skip("Could not determine home directory")
	}

	tests := []struct {
		name      string
		platform  string
		xdgConfig string
		want      string
	}{
		{
			name:     "claude global config",
			platform: PlatformClaude,
			want:     filepath.Join(home, ".claude"),
		},
		{
			name:     "opencode global config",
			platform: PlatformOpenCode,
			want:     filepath.Join(home, ".config", "opencode"),
		},
		{
			name:     "codex global config",
			platform: PlatformCodex,
			want:     filepath.Join(home, ".codex"),
		},
		{
			name:     "gemini global config",
			platform: PlatformGemini,
			want:     filepath.Join(home, ".gemini"),
		},
		{
			name:      "gemini global config with XDG_CONFIG_HOME",
			platform:  PlatformGemini,
			xdgConfig: "/custom/config",
			want:      filepath.Join("/custom/config", "gemini"),
		},
		{
			name:     "unknown platform returns empty",
			platform: "unknown",
			want:     "",
		},
		{
			name:     "empty platform returns empty",
			platform: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.xdgConfig != "" {
				t.Setenv("XDG_CONFIG_HOME", tt.xdgConfig)
			} else {
				t.Setenv("XDG_CONFIG_HOME", "")
			}

			got := GlobalConfigDir(tt.platform)
			if got != tt.want {
				t.Errorf("GlobalConfigDir(%q) = %q, want %q", tt.platform, got, tt.want)
			}
		})
	}
}

func TestProjectConfigDir(t *testing.T) {
	projectRoot := "/home/user/myproject"
	if runtime.GOOS == "windows" {
		projectRoot = `C:\Users\user\myproject`
	}

	tests := []struct {
		name        string
		platform    string
		projectRoot string
		want        string
	}{
		{
			name:        "claude project config",
			platform:    PlatformClaude,
			projectRoot: projectRoot,
			want:        filepath.Join(projectRoot, ".claude"),
		},
		{
			name:        "opencode uses project root",
			platform:    PlatformOpenCode,
			projectRoot: projectRoot,
			want:        projectRoot, // OpenCode uses root directly
		},
		{
			name:        "codex project config",
			platform:    PlatformCodex,
			projectRoot: projectRoot,
			want:        filepath.Join(projectRoot, ".codex"),
		},
		{
			name:        "gemini project config",
			platform:    PlatformGemini,
			projectRoot: projectRoot,
			want:        filepath.Join(projectRoot, ".gemini"),
		},
		{
			name:        "unknown platform returns empty",
			platform:    "unknown",
			projectRoot: projectRoot,
			want:        "",
		},
		{
			name:        "empty project root returns empty",
			platform:    PlatformClaude,
			projectRoot: "",
			want:        "",
		},
		{
			name:        "empty platform and root returns empty",
			platform:    "",
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProjectConfigDir(tt.platform, tt.projectRoot)
			if got != tt.want {
				t.Errorf("ProjectConfigDir(%q, %q) = %q, want %q", tt.platform, tt.projectRoot, got, tt.want)
			}
		})
	}
}

func TestInstructionsPath(t *testing.T) {
	projectRoot := "/home/user/myproject"
	if runtime.GOOS == "windows" {
		projectRoot = `C:\Users\user\myproject`
	}

	tests := []struct {
		name        string
		platform    string
		projectRoot string
		want        string
	}{
		{
			name:        "claude instructions",
			platform:    PlatformClaude,
			projectRoot: projectRoot,
			want:        filepath.Join(projectRoot, "CLAUDE.md"),
		},
		{
			name:        "opencode instructions",
			platform:    PlatformOpenCode,
			projectRoot: projectRoot,
			want:        filepath.Join(projectRoot, "AGENTS.md"),
		},
		{
			name:        "codex instructions",
			platform:    PlatformCodex,
			projectRoot: projectRoot,
			want:        filepath.Join(projectRoot, "AGENTS.md"),
		},
		{
			name:        "gemini instructions",
			platform:    PlatformGemini,
			projectRoot: projectRoot,
			want:        filepath.Join(projectRoot, "GEMINI.md"),
		},
		{
			name:        "unknown platform returns empty",
			platform:    "unknown",
			projectRoot: projectRoot,
			want:        "",
		},
		{
			name:        "empty project root returns empty",
			platform:    PlatformClaude,
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InstructionsPath(tt.platform, tt.projectRoot)
			if got != tt.want {
				t.Errorf("InstructionsPath(%q, %q) = %q, want %q", tt.platform, tt.projectRoot, got, tt.want)
			}
		})
	}
}

func TestSkillDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home := Home()
	if home == "" {
		t.Skip("Could not determine home directory")
	}

	tests := []struct {
		name     string
		platform string
		want     string
	}{
		{
			name:     "claude skills",
			platform: PlatformClaude,
			want:     filepath.Join(home, ".claude", "skills"),
		},
		{
			name:     "opencode skills",
			platform: PlatformOpenCode,
			want:     filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			name:     "codex skills",
			platform: PlatformCodex,
			want:     filepath.Join(home, ".codex", "skills"),
		},
		{
			name:     "gemini skills",
			platform: PlatformGemini,
			want:     filepath.Join(home, ".gemini", "skills"),
		},
		{
			name:     "unknown platform returns empty",
			platform: "unknown",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SkillDir(tt.platform)
			if got != tt.want {
				t.Errorf("SkillDir(%q) = %q, want %q", tt.platform, got, tt.want)
			}
		})
	}
}

func TestCommandDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home := Home()
	if home == "" {
		t.Skip("Could not determine home directory")
	}

	tests := []struct {
		name     string
		platform string
		want     string
	}{
		{
			name:     "claude commands",
			platform: PlatformClaude,
			want:     filepath.Join(home, ".claude", "commands"),
		},
		{
			name:     "opencode commands",
			platform: PlatformOpenCode,
			want:     filepath.Join(home, ".config", "opencode", "commands"),
		},
		{
			name:     "codex commands",
			platform: PlatformCodex,
			want:     filepath.Join(home, ".codex", "commands"),
		},
		{
			name:     "gemini commands",
			platform: PlatformGemini,
			want:     filepath.Join(home, ".gemini", "commands"),
		},
		{
			name:     "unknown platform returns empty",
			platform: "unknown",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CommandDir(tt.platform)
			if got != tt.want {
				t.Errorf("CommandDir(%q) = %q, want %q", tt.platform, got, tt.want)
			}
		})
	}
}

func TestMCPConfigPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home := Home()
	if home == "" {
		t.Skip("Could not determine home directory")
	}

	tests := []struct {
		name     string
		platform string
		want     string
	}{
		{
			name:     "claude MCP config is in home directory",
			platform: PlatformClaude,
			want:     filepath.Join(home, ".claude.json"),
		},
		{
			name:     "opencode MCP config",
			platform: PlatformOpenCode,
			want:     filepath.Join(home, ".config", "opencode", "opencode.json"),
		},
		{
			name:     "codex MCP config",
			platform: PlatformCodex,
			want:     filepath.Join(home, ".codex", "mcp.json"),
		},
		{
			name:     "gemini MCP config",
			platform: PlatformGemini,
			want:     filepath.Join(home, ".gemini", "settings.json"),
		},
		{
			name:     "unknown platform returns empty",
			platform: "unknown",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MCPConfigPath(tt.platform)
			if got != tt.want {
				t.Errorf("MCPConfigPath(%q) = %q, want %q", tt.platform, got, tt.want)
			}
		})
	}
}

func TestInstructionFilename(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     string
	}{
		{
			name:     "claude filename",
			platform: PlatformClaude,
			want:     "CLAUDE.md",
		},
		{
			name:     "opencode filename",
			platform: PlatformOpenCode,
			want:     "AGENTS.md",
		},
		{
			name:     "codex filename",
			platform: PlatformCodex,
			want:     "AGENTS.md",
		},
		{
			name:     "gemini filename",
			platform: PlatformGemini,
			want:     "GEMINI.md",
		},
		{
			name:     "unknown platform returns empty",
			platform: "unknown",
			want:     "",
		},
		{
			name:     "empty platform returns empty",
			platform: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InstructionFilename(tt.platform)
			if got != tt.want {
				t.Errorf("InstructionFilename(%q) = %q, want %q", tt.platform, got, tt.want)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("creates new directory with default perms", func(t *testing.T) {
		path := filepath.Join(tmpDir, "new-dir")
		err := EnsureDir(path, 0)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat failed: %v", err)
		}
		if !info.IsDir() {
			t.Errorf("expected directory, got file")
		}
		// On some systems (like macOS), the mode might have extra bits (like 0700 or 0755)
		// but we want to check the permission bits.
		if info.Mode().Perm() != DefaultDirPerm {
			t.Errorf("expected perm %o, got %o", DefaultDirPerm, info.Mode().Perm())
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		path := filepath.Join(tmpDir, "parent", "child", "grandchild")
		err := EnsureDir(path, 0o755)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat failed: %v", err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Errorf("expected perm 0755, got %o", info.Mode().Perm())
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		path := filepath.Join(tmpDir, "existing")
		err := os.Mkdir(path, 0o755)
		if err != nil {
			t.Fatal(err)
		}

		err = EnsureDir(path, 0o700)
		if err != nil {
			t.Errorf("EnsureDir failed on existing directory: %v", err)
		}

		// Note: MkdirAll (and thus EnsureDir) does NOT change permissions of existing directories.
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Errorf("expected original perm 0755 to be preserved, got %o", info.Mode().Perm())
		}
	})
}

// TestXDGHomeConsistency verifies XDG functions return consistent results
// across multiple calls.
func TestXDGHomeConsistency(t *testing.T) {
	// Call each function twice and verify consistency
	configHome1 := ConfigHome()
	configHome2 := ConfigHome()
	if configHome1 != configHome2 {
		t.Errorf("ConfigHome() not consistent: %q != %q", configHome1, configHome2)
	}

	dataHome1 := DataHome()
	dataHome2 := DataHome()
	if dataHome1 != dataHome2 {
		t.Errorf("DataHome() not consistent: %q != %q", dataHome1, dataHome2)
	}

	cacheHome1 := CacheHome()
	cacheHome2 := CacheHome()
	if cacheHome1 != cacheHome2 {
		t.Errorf("CacheHome() not consistent: %q != %q", cacheHome1, cacheHome2)
	}
}

// TestPlatformConstantsMatchMaps verifies that the platform constants
// are properly registered in all lookup maps.
func TestPlatformConstantsMatchMaps(t *testing.T) {
	platforms := []string{PlatformClaude, PlatformOpenCode, PlatformCodex, PlatformGemini}

	for _, p := range platforms {
		t.Run(p, func(t *testing.T) {
			if !ValidPlatform(p) {
				t.Errorf("Platform constant %q is not valid", p)
			}

			if GlobalConfigDir(p) == "" {
				t.Errorf("GlobalConfigDir(%q) returned empty string", p)
			}

			if InstructionFilename(p) == "" {
				t.Errorf("InstructionFilename(%q) returned empty string", p)
			}

			if MCPConfigPath(p) == "" {
				t.Errorf("MCPConfigPath(%q) returned empty string", p)
			}

			if SkillDir(p) == "" {
				t.Errorf("SkillDir(%q) returned empty string", p)
			}

			if CommandDir(p) == "" {
				t.Errorf("CommandDir(%q) returned empty string", p)
			}

			// ProjectConfigDir requires a projectRoot
			projectRoot := "/test/project"
			if runtime.GOOS == "windows" {
				projectRoot = `C:\test\project`
			}
			if ProjectConfigDir(p, projectRoot) == "" {
				t.Errorf("ProjectConfigDir(%q, %q) returned empty string", p, projectRoot)
			}

			if InstructionsPath(p, projectRoot) == "" {
				t.Errorf("InstructionsPath(%q, %q) returned empty string", p, projectRoot)
			}
		})
	}
}
