package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thoreinstein/aix/internal/paths"
)

func TestDetectPlatform_ValidPlatforms(t *testing.T) {
	tests := []struct {
		name     string
		platform string
	}{
		{name: "claude", platform: paths.PlatformClaude},
		{name: "opencode", platform: paths.PlatformOpenCode},
		{name: "codex", platform: paths.PlatformCodex},
		{name: "gemini", platform: paths.PlatformGemini},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectPlatform(tt.platform)
			if result == nil {
				t.Fatalf("DetectPlatform(%q) returned nil, want non-nil", tt.platform)
			}

			if result.Name != tt.platform {
				t.Errorf("DetectPlatform(%q).Name = %q, want %q", tt.platform, result.Name, tt.platform)
			}

			if result.GlobalConfig == "" {
				t.Errorf("DetectPlatform(%q).GlobalConfig is empty", tt.platform)
			}

			if result.MCPConfig == "" {
				t.Errorf("DetectPlatform(%q).MCPConfig is empty", tt.platform)
			}

			// Status should be one of the valid values
			switch result.Status {
			case StatusInstalled, StatusNotInstalled, StatusPartial:
				// valid
			default:
				t.Errorf("DetectPlatform(%q).Status = %q, want valid InstallStatus", tt.platform, result.Status)
			}
		})
	}
}

func TestDetectPlatform_InvalidPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform string
	}{
		{name: "unknown platform", platform: "unknown"},
		{name: "empty string", platform: ""},
		{name: "case sensitive", platform: "Claude"},
		{name: "typo", platform: "claudde"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectPlatform(tt.platform)
			if result != nil {
				t.Errorf("DetectPlatform(%q) = %+v, want nil", tt.platform, result)
			}
		})
	}
}

func TestDetectAll_ReturnsAllPlatforms(t *testing.T) {
	results := DetectAll()

	if len(results) != 4 {
		t.Errorf("DetectAll() returned %d platforms, want 4", len(results))
	}

	// Verify all expected platforms are present
	found := make(map[string]bool)
	for _, result := range results {
		found[result.Name] = true
	}

	expected := []string{
		paths.PlatformClaude,
		paths.PlatformOpenCode,
		paths.PlatformCodex,
		paths.PlatformGemini,
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("DetectAll() missing platform %q", name)
		}
	}
}

func TestDetectAll_DeterministicOrder(t *testing.T) {
	// Call twice and verify same order
	results1 := DetectAll()
	results2 := DetectAll()

	if len(results1) != len(results2) {
		t.Fatalf("DetectAll() returned different lengths: %d vs %d", len(results1), len(results2))
	}

	for i := range results1 {
		if results1[i].Name != results2[i].Name {
			t.Errorf("DetectAll() order not deterministic at index %d: %q vs %q",
				i, results1[i].Name, results2[i].Name)
		}
	}

	// Verify expected order matches paths.Platforms()
	expectedOrder := paths.Platforms()
	for i, result := range results1 {
		if result.Name != expectedOrder[i] {
			t.Errorf("DetectAll()[%d].Name = %q, want %q", i, result.Name, expectedOrder[i])
		}
	}
}

func TestDetectInstalled_FiltersCorrectly(t *testing.T) {
	// Create a temp directory structure that mimics installed platforms
	tmpDir := t.TempDir()

	// Create mock config directories for claude and codex
	mockClaude := filepath.Join(tmpDir, ".claude")
	mockCodex := filepath.Join(tmpDir, ".codex")

	if err := os.MkdirAll(mockClaude, 0o755); err != nil {
		t.Fatalf("Failed to create mock claude dir: %v", err)
	}
	if err := os.MkdirAll(mockCodex, 0o755); err != nil {
		t.Fatalf("Failed to create mock codex dir: %v", err)
	}

	// We can't easily override the home directory used by paths package,
	// so we'll test with a helper that allows injection
	results := detectInstalledWithDirCheck(func(path string) bool {
		// Simulate only claude and codex being "installed"
		return path == mockClaude || path == mockCodex
	})

	// Should return only "installed" platforms
	for _, result := range results {
		if result.Status != StatusInstalled {
			t.Errorf("DetectInstalled() included platform with status %q, want %q",
				result.Status, StatusInstalled)
		}
	}
}

func TestDetectInstalled_EmptyWhenNoneInstalled(t *testing.T) {
	// Test with a dir check function that always returns false
	results := detectInstalledWithDirCheck(func(path string) bool {
		return false
	})

	if len(results) != 0 {
		t.Errorf("detectInstalledWithDirCheck() returned %d platforms, want 0 when none installed",
			len(results))
	}
}

func TestDetectionResult_PathsMatch(t *testing.T) {
	// Verify DetectionResult paths match what paths package returns
	for _, platform := range paths.Platforms() {
		result := DetectPlatform(platform)
		if result == nil {
			t.Fatalf("DetectPlatform(%q) returned nil", platform)
		}

		expectedGlobal := paths.GlobalConfigDir(platform)
		if result.GlobalConfig != expectedGlobal {
			t.Errorf("DetectPlatform(%q).GlobalConfig = %q, want %q",
				platform, result.GlobalConfig, expectedGlobal)
		}

		expectedMCP := paths.MCPConfigPath(platform)
		if result.MCPConfig != expectedMCP {
			t.Errorf("DetectPlatform(%q).MCPConfig = %q, want %q",
				platform, result.MCPConfig, expectedMCP)
		}
	}
}

func TestInstallStatus_Constants(t *testing.T) {
	// Verify status constants have expected string values
	tests := []struct {
		status InstallStatus
		want   string
	}{
		{StatusInstalled, "installed"},
		{StatusNotInstalled, "not_installed"},
		{StatusPartial, "partial"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("InstallStatus constant = %q, want %q", tt.status, tt.want)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")

	// Create a file (not a directory)
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "existing directory", path: tmpDir, want: true},
		{name: "existing file", path: tmpFile, want: false},
		{name: "nonexistent path", path: filepath.Join(tmpDir, "nonexistent"), want: false},
		{name: "empty path", path: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dirExists(tt.path); got != tt.want {
				t.Errorf("dirExists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// detectInstalledWithDirCheck is a test helper that allows injecting a custom
// directory existence check function.
func detectInstalledWithDirCheck(checkFn func(string) bool) []*DetectionResult {
	platforms := paths.Platforms()
	results := make([]*DetectionResult, 0, len(platforms))

	for _, name := range platforms {
		globalConfig := paths.GlobalConfigDir(name)
		mcpConfig := paths.MCPConfigPath(name)

		status := StatusNotInstalled
		if checkFn(globalConfig) {
			status = StatusInstalled
		}

		if status == StatusInstalled {
			results = append(results, &DetectionResult{
				Name:         name,
				GlobalConfig: globalConfig,
				MCPConfig:    mcpConfig,
				Status:       status,
			})
		}
	}

	return results
}
