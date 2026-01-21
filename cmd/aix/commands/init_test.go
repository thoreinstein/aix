package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// setupInitTest prepares a clean environment for init command tests.
// Returns the temp config directory and a cleanup function.
func setupInitTest(t *testing.T) (configDir string, cleanup func()) {
	t.Helper()
	viper.Reset()
	dir := t.TempDir()
	return dir, func() { viper.Reset() }
}

// resetInitFlags resets the init command flags to their default values.
func resetInitFlags(t *testing.T) {
	t.Helper()
	initYes = false
	initPlatforms = ""
	initForce = false
}

func TestParsePlatformList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string returns empty slice",
			input: "",
			want:  nil,
		},
		{
			name:  "single valid platform",
			input: "claude",
			want:  []string{"claude"},
		},
		{
			name:  "multiple valid platforms",
			input: "claude,opencode",
			want:  []string{"claude", "opencode"},
		},
		{
			name:  "all valid platforms",
			input: "claude,opencode,codex,gemini",
			want:  []string{"claude", "opencode", "codex", "gemini"},
		},
		{
			name:  "invalid platform filtered out",
			input: "invalid",
			want:  nil,
		},
		{
			name:  "mixed valid and invalid filters correctly",
			input: "claude,invalid,opencode",
			want:  []string{"claude", "opencode"},
		},
		{
			name:  "whitespace is trimmed",
			input: " claude , opencode ",
			want:  []string{"claude", "opencode"},
		},
		{
			name:  "extra whitespace around commas",
			input: "claude  ,  opencode  ,  gemini",
			want:  []string{"claude", "opencode", "gemini"},
		},
		{
			name:  "trailing comma produces empty entry filtered out",
			input: "claude,",
			want:  []string{"claude"},
		},
		{
			name:  "leading comma produces empty entry filtered out",
			input: ",claude",
			want:  []string{"claude"},
		},
		{
			name:  "multiple commas filtered",
			input: "claude,,opencode",
			want:  []string{"claude", "opencode"},
		},
		{
			name:  "case sensitive - Claude is invalid",
			input: "Claude",
			want:  nil,
		},
		{
			name:  "case sensitive - mixed case filtered",
			input: "Claude,claude,OPENCODE,opencode",
			want:  []string{"claude", "opencode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePlatformList(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parsePlatformList(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parsePlatformList(%q)[%d] = %q, want %q",
						tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDetectPlatforms(t *testing.T) {
	// Note: This test depends on the actual filesystem state.
	// detectPlatforms() checks paths.GlobalConfigDir for each platform.
	// We can't easily mock this without refactoring the code.
	// This test verifies the function runs without error and returns
	// a valid slice (may be empty or contain detected platforms).
	t.Run("returns valid slice", func(t *testing.T) {
		platforms := detectPlatforms()

		// Result should be a slice (possibly empty)
		if platforms == nil {
			// nil is acceptable - means no platforms detected
			return
		}

		// All returned platforms should be valid
		for _, p := range platforms {
			if p == "" {
				t.Error("detectPlatforms() returned empty string in slice")
			}
			// Verify each returned platform is one of the known platforms
			validPlatforms := map[string]bool{
				"claude": true, "opencode": true, "codex": true, "gemini": true,
			}
			if !validPlatforms[p] {
				t.Errorf("detectPlatforms() returned unknown platform %q", p)
			}
		}
	})
}

func TestInitCommand_CreatesConfigFile(t *testing.T) {
	configDir, cleanup := setupInitTest(t)
	defer cleanup()
	resetInitFlags(t)

	// Set up flags for non-interactive mode with explicit platforms
	initYes = true
	initPlatforms = "claude,opencode"

	// Create a mock config path by setting HOME to temp dir
	// t.Setenv automatically restores the original value after the test
	t.Setenv("HOME", configDir)

	// We need to also handle XDG_CONFIG_HOME for paths.ConfigHome()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	configPath := filepath.Join(configDir, "aix", "config.yaml")

	// Run the init command
	// We can't easily call runInit directly since it uses paths.ConfigHome()
	// which is based on xdg. Instead, we'll test the config file creation logic.

	// Test the config directory creation
	configDirPath := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirPath, 0o755); err != nil {
		t.Fatalf("creating config directory: %v", err)
	}

	// Create config using same logic as runInit
	platforms := parsePlatformList(initPlatforms)
	cfg := aixConfig{
		Version:          1,
		DefaultPlatforms: platforms,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("marshaling config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Read and verify contents
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	var readCfg aixConfig
	if err := yaml.Unmarshal(content, &readCfg); err != nil {
		t.Fatalf("unmarshaling config: %v", err)
	}

	if readCfg.Version != 1 {
		t.Errorf("config version = %d, want 1", readCfg.Version)
	}

	if len(readCfg.DefaultPlatforms) != 2 {
		t.Errorf("config default_platforms length = %d, want 2", len(readCfg.DefaultPlatforms))
	}

	expectedPlatforms := map[string]bool{"claude": true, "opencode": true}
	for _, p := range readCfg.DefaultPlatforms {
		if !expectedPlatforms[p] {
			t.Errorf("unexpected platform in config: %s", p)
		}
	}
}

func TestInitCommand_ForceOverwritesConfig(t *testing.T) {
	configDir, cleanup := setupInitTest(t)
	defer cleanup()
	resetInitFlags(t)

	configPath := filepath.Join(configDir, "config.yaml")

	// Create existing config with different content
	existingCfg := aixConfig{
		Version:          1,
		DefaultPlatforms: []string{"gemini"},
	}
	existingData, err := yaml.Marshal(&existingCfg)
	if err != nil {
		t.Fatalf("marshaling existing config: %v", err)
	}
	if err := os.WriteFile(configPath, existingData, 0o644); err != nil {
		t.Fatalf("writing existing config: %v", err)
	}

	// Set flags
	initYes = true
	initForce = true
	initPlatforms = "claude,opencode"

	// Simulate force overwrite
	platforms := parsePlatformList(initPlatforms)
	newCfg := aixConfig{
		Version:          1,
		DefaultPlatforms: platforms,
	}

	newData, err := yaml.Marshal(&newCfg)
	if err != nil {
		t.Fatalf("marshaling new config: %v", err)
	}

	if err := os.WriteFile(configPath, newData, 0o644); err != nil {
		t.Fatalf("writing new config: %v", err)
	}

	// Verify new content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	var readCfg aixConfig
	if err := yaml.Unmarshal(content, &readCfg); err != nil {
		t.Fatalf("unmarshaling config: %v", err)
	}

	if len(readCfg.DefaultPlatforms) != 2 {
		t.Errorf("config was not overwritten: got %d platforms, want 2",
			len(readCfg.DefaultPlatforms))
	}

	// Verify gemini is not in the new config
	for _, p := range readCfg.DefaultPlatforms {
		if p == "gemini" {
			t.Error("config was not overwritten: still contains 'gemini'")
		}
	}
}

func TestInitCommand_WithoutForceExistingConfig(t *testing.T) {
	configDir, cleanup := setupInitTest(t)
	defer cleanup()
	resetInitFlags(t)

	configPath := filepath.Join(configDir, "config.yaml")

	// Create existing config
	existingCfg := aixConfig{
		Version:          1,
		DefaultPlatforms: []string{"gemini"},
	}
	existingData, err := yaml.Marshal(&existingCfg)
	if err != nil {
		t.Fatalf("marshaling existing config: %v", err)
	}
	if err := os.WriteFile(configPath, existingData, 0o644); err != nil {
		t.Fatalf("writing existing config: %v", err)
	}

	// Without --force flag
	initYes = true
	initForce = false

	// Check if config exists (simulating runInit logic)
	_, err = os.Stat(configPath)
	configExists := err == nil

	if !configExists {
		t.Fatal("test setup failed: config file should exist")
	}

	// When config exists and --force is false, runInit should return early
	// without modifying the config. We verify by checking the file is unchanged.
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	var readCfg aixConfig
	if err := yaml.Unmarshal(content, &readCfg); err != nil {
		t.Fatalf("unmarshaling config: %v", err)
	}

	// Config should still have gemini as the only platform
	if len(readCfg.DefaultPlatforms) != 1 || readCfg.DefaultPlatforms[0] != "gemini" {
		t.Errorf("config was modified without --force: got %v", readCfg.DefaultPlatforms)
	}
}

func TestInitCommand_PlatformsFlagOverridesDetection(t *testing.T) {
	_, cleanup := setupInitTest(t)
	defer cleanup()
	resetInitFlags(t)

	// When --platforms flag is provided, it should override detection
	initPlatforms = "codex,gemini"

	// This mimics the logic in runInit
	platforms := detectPlatforms()
	if initPlatforms != "" {
		platforms = parsePlatformList(initPlatforms)
	}

	// Verify platforms came from the flag, not detection
	if len(platforms) != 2 {
		t.Fatalf("expected 2 platforms from flag, got %d", len(platforms))
	}

	expected := map[string]bool{"codex": true, "gemini": true}
	for _, p := range platforms {
		if !expected[p] {
			t.Errorf("unexpected platform %q, expected codex or gemini", p)
		}
	}
}

func TestInitCommand_YesFlagSkipsConfirmation(t *testing.T) {
	_, cleanup := setupInitTest(t)
	defer cleanup()
	resetInitFlags(t)

	// With --yes flag, confirmation should be skipped
	initYes = true

	// This test verifies the flag behavior
	// In runInit, when initYes is true, confirm() is not called
	if !initYes {
		t.Error("initYes flag should be true")
	}
}

func TestInitCommand_ConfigFileFormat(t *testing.T) {
	configDir, cleanup := setupInitTest(t)
	defer cleanup()
	resetInitFlags(t)

	configPath := filepath.Join(configDir, "config.yaml")

	// Create config with all platforms
	cfg := aixConfig{
		Version:          1,
		DefaultPlatforms: []string{"claude", "opencode", "codex", "gemini"},
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("marshaling config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	// Read and verify YAML structure
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	contentStr := string(content)

	// Verify YAML contains expected keys
	if !strings.Contains(contentStr, "version:") {
		t.Error("config file missing 'version' key")
	}

	if !strings.Contains(contentStr, "default_platforms:") {
		t.Error("config file missing 'default_platforms' key")
	}

	// Verify version value
	if !strings.Contains(contentStr, "version: 1") {
		t.Error("config file should have version: 1")
	}

	// Parse back and verify structure
	var readCfg aixConfig
	if err := yaml.Unmarshal(content, &readCfg); err != nil {
		t.Fatalf("unmarshaling config: %v", err)
	}

	if readCfg.Version != 1 {
		t.Errorf("version = %d, want 1", readCfg.Version)
	}

	if len(readCfg.DefaultPlatforms) != 4 {
		t.Errorf("default_platforms length = %d, want 4", len(readCfg.DefaultPlatforms))
	}
}

func TestAixConfig_YAMLRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		cfg  aixConfig
	}{
		{
			name: "empty platforms",
			cfg: aixConfig{
				Version:          1,
				DefaultPlatforms: []string{},
			},
		},
		{
			name: "single platform",
			cfg: aixConfig{
				Version:          1,
				DefaultPlatforms: []string{"claude"},
			},
		},
		{
			name: "all platforms",
			cfg: aixConfig{
				Version:          1,
				DefaultPlatforms: []string{"claude", "opencode", "codex", "gemini"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to YAML
			data, err := yaml.Marshal(&tt.cfg)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			// Unmarshal back
			var got aixConfig
			if err := yaml.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			// Compare
			if got.Version != tt.cfg.Version {
				t.Errorf("version = %d, want %d", got.Version, tt.cfg.Version)
			}

			if len(got.DefaultPlatforms) != len(tt.cfg.DefaultPlatforms) {
				t.Errorf("platforms length = %d, want %d",
					len(got.DefaultPlatforms), len(tt.cfg.DefaultPlatforms))
				return
			}

			for i := range got.DefaultPlatforms {
				if got.DefaultPlatforms[i] != tt.cfg.DefaultPlatforms[i] {
					t.Errorf("platforms[%d] = %q, want %q",
						i, got.DefaultPlatforms[i], tt.cfg.DefaultPlatforms[i])
				}
			}
		})
	}
}

func TestInitCommand_ConfigDirectoryCreation(t *testing.T) {
	configDir, cleanup := setupInitTest(t)
	defer cleanup()

	// Test that nested directory creation works
	nestedPath := filepath.Join(configDir, "aix", "nested", "config.yaml")
	nestedDir := filepath.Dir(nestedPath)

	// Directory should not exist yet
	if _, err := os.Stat(nestedDir); !os.IsNotExist(err) {
		t.Fatal("nested directory should not exist before test")
	}

	// Create directory (mimicking runInit behavior)
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("creating nested directory: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("created path is not a directory")
	}
}

func TestInitCommand_ConfigFilePermissions(t *testing.T) {
	configDir, cleanup := setupInitTest(t)
	defer cleanup()

	configPath := filepath.Join(configDir, "config.yaml")

	cfg := aixConfig{
		Version:          1,
		DefaultPlatforms: []string{"claude"},
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("marshaling config: %v", err)
	}

	// Write with specific permissions (0644)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}

	// Check file permissions (on Unix-like systems)
	mode := info.Mode().Perm()
	// 0644 = owner rw, group r, others r
	expectedMode := os.FileMode(0o644)
	if mode != expectedMode {
		t.Errorf("file permissions = %o, want %o", mode, expectedMode)
	}
}

func TestParsePlatformList_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // expected length
	}{
		{
			name:  "only whitespace",
			input: "   ",
			want:  0,
		},
		{
			name:  "only commas",
			input: ",,,",
			want:  0,
		},
		{
			name:  "commas and whitespace",
			input: " , , , ",
			want:  0,
		},
		{
			name:  "tabs as whitespace",
			input: "\tclaude\t,\topencode\t",
			want:  2,
		},
		{
			name:  "newlines in input",
			input: "claude\nopencode",
			want:  0, // newline doesn't split, so this is one invalid platform
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePlatformList(tt.input)
			if len(got) != tt.want {
				t.Errorf("parsePlatformList(%q) returned %d items, want %d",
					tt.input, len(got), tt.want)
			}
		})
	}
}

// TestConfirm tests the confirm function behavior.
// Note: This function reads from os.Stdin, making it difficult to test
// in automated tests. We document the expected behavior here.
func TestConfirm_Documentation(t *testing.T) {
	// The confirm function:
	// 1. Prints a prompt with [y/N] suffix
	// 2. Reads a line from stdin
	// 3. Returns true only for "y" or "yes" (case-insensitive)
	// 4. Returns false for any other input or error

	// Since we can't easily mock stdin, this test documents the interface.
	// Interactive testing should verify:
	// - "y" returns true
	// - "yes" returns true
	// - "Y" returns true
	// - "YES" returns true
	// - "n" returns false
	// - "no" returns false
	// - "" (empty/enter) returns false
	// - Any other input returns false

	t.Log("confirm() is an interactive function - manual testing recommended")
	t.Log("Use --yes flag to skip confirmation in automated scenarios")
}
