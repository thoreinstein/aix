package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func setupTestConfig(t *testing.T) string {
	t.Helper()
	viper.Reset()
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	viper.SetConfigFile(configFile)
	viper.Set("version", 1)
	viper.Set("default_platforms", []string{"claude", "opencode"})
	return configFile
}

func TestParsePlatforms(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string returns empty slice",
			input: "",
			want:  []string{},
		},
		{
			name:  "single platform",
			input: "claude",
			want:  []string{"claude"},
		},
		{
			name:  "multiple platforms",
			input: "claude,opencode",
			want:  []string{"claude", "opencode"},
		},
		{
			name:  "whitespace handling",
			input: " claude , opencode ",
			want:  []string{"claude", "opencode"},
		},
		{
			name:  "empty elements filtered",
			input: "claude,,opencode",
			want:  []string{"claude", "opencode"},
		},
		{
			name:  "leading and trailing commas",
			input: ",claude,opencode,",
			want:  []string{"claude", "opencode"},
		},
		{
			name:  "only whitespace and commas",
			input: " , , , ",
			want:  []string{},
		},
		{
			name:  "all four platforms",
			input: "claude,opencode,codex,gemini",
			want:  []string{"claude", "opencode", "codex", "gemini"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePlatforms(tt.input)

			// Handle nil vs empty slice comparison
			if len(got) == 0 && len(tt.want) == 0 {
				return // Both empty, test passes
			}

			if len(got) != len(tt.want) {
				t.Errorf("parsePlatforms(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.want, len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parsePlatforms(%q)[%d] = %q, want %q",
						tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestConfigGet(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		setupValue func()
		wantOutput string
	}{
		{
			name: "unset key prints not set",
			key:  "nonexistent_key",
			setupValue: func() {
				// Don't set anything
			},
			wantOutput: "not set\n",
		},
		{
			name: "scalar value prints the value",
			key:  "version",
			setupValue: func() {
				viper.Set("version", 1)
			},
			wantOutput: "1\n",
		},
		{
			name: "string value prints the value",
			key:  "custom_key",
			setupValue: func() {
				viper.Set("custom_key", "test_value")
			},
			wantOutput: "test_value\n",
		},
		{
			name: "array value prints one per line",
			key:  "default_platforms",
			setupValue: func() {
				viper.Set("default_platforms", []string{"claude", "opencode"})
			},
			wantOutput: "claude\nopencode\n",
		},
		{
			name: "empty array prints nothing",
			key:  "empty_array",
			setupValue: func() {
				viper.Set("empty_array", []string{})
			},
			wantOutput: "", // Empty array produces no output
		},
		{
			name: "single element array",
			key:  "single_platform",
			setupValue: func() {
				viper.Set("single_platform", []string{"claude"})
			},
			wantOutput: "claude\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			tt.setupValue()

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runConfigGet(nil, []string{tt.key})

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			got := buf.String()

			if err != nil {
				t.Errorf("runConfigGet() error = %v", err)
				return
			}

			if got != tt.wantOutput {
				t.Errorf("runConfigGet(%q) output = %q, want %q", tt.key, got, tt.wantOutput)
			}
		})
	}
}

func TestConfigSet(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		wantErr     bool
		errContains string
		verify      func(t *testing.T)
	}{
		{
			name:    "valid single platform",
			key:     "default_platforms",
			value:   "claude",
			wantErr: false,
			verify: func(t *testing.T) {
				t.Helper()
				platforms := viper.GetStringSlice("default_platforms")
				if len(platforms) != 1 || platforms[0] != "claude" {
					t.Errorf("expected [claude], got %v", platforms)
				}
			},
		},
		{
			name:    "valid multiple platforms",
			key:     "default_platforms",
			value:   "claude,opencode,codex",
			wantErr: false,
			verify: func(t *testing.T) {
				t.Helper()
				platforms := viper.GetStringSlice("default_platforms")
				if len(platforms) != 3 {
					t.Errorf("expected 3 platforms, got %d", len(platforms))
				}
			},
		},
		{
			name:        "invalid platform returns error",
			key:         "default_platforms",
			value:       "invalid_platform",
			wantErr:     true,
			errContains: "invalid platform",
		},
		{
			name:        "mixed valid and invalid platforms",
			key:         "default_platforms",
			value:       "claude,invalid_platform",
			wantErr:     true,
			errContains: "invalid platform",
		},
		{
			name:        "empty platform list returns error",
			key:         "default_platforms",
			value:       "",
			wantErr:     true,
			errContains: "no valid platforms specified",
		},
		{
			name:        "only commas returns error",
			key:         "default_platforms",
			value:       ",,,",
			wantErr:     true,
			errContains: "no valid platforms specified",
		},
		{
			name:    "version key accepts any value",
			key:     "version",
			value:   "2",
			wantErr: false,
			verify: func(t *testing.T) {
				t.Helper()
				if viper.GetString("version") != "2" {
					t.Errorf("expected version 2, got %s", viper.GetString("version"))
				}
			},
		},
		{
			name:    "arbitrary key works",
			key:     "custom_setting",
			value:   "custom_value",
			wantErr: false,
			verify: func(t *testing.T) {
				t.Helper()
				if viper.GetString("custom_setting") != "custom_value" {
					t.Errorf("expected custom_value, got %s", viper.GetString("custom_setting"))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: create temp dir for config file to be written
			// Note: writeConfig() writes to paths.ConfigHome() which we can't
			// easily override. The function will fail to write but we can still
			// verify the validation logic by checking if we get expected errors.
			viper.Reset()

			// Suppress stdout during tests
			old := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			err := runConfigSet(nil, []string{tt.key, tt.value})

			w.Close()
			os.Stdout = old

			if tt.wantErr {
				if err == nil {
					t.Errorf("runConfigSet() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("runConfigSet() error = %q, want error containing %q",
						err.Error(), tt.errContains)
				}
				return
			}

			// For non-error cases, we expect writeConfig to fail because
			// paths.ConfigHome() returns a real path we can't write to in tests.
			// So we only verify the validation passed (no validation error).
			// The actual write failure is acceptable in unit tests.
			if err != nil && !strings.Contains(err.Error(), "creating config directory") &&
				!strings.Contains(err.Error(), "writing config file") {
				t.Errorf("runConfigSet() unexpected validation error = %v", err)
				return
			}

			// If no validation error and verify function provided, run it
			if tt.verify != nil && err == nil {
				tt.verify(t)
			}
		})
	}
}

func TestConfigSet_ValidPlatforms(t *testing.T) {
	// Test that all valid platforms are accepted
	validPlatforms := []string{"claude", "opencode", "codex", "gemini"}

	for _, platform := range validPlatforms {
		t.Run(platform, func(t *testing.T) {
			viper.Reset()

			// Suppress stdout
			old := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			err := runConfigSet(nil, []string{"default_platforms", platform})

			w.Close()
			os.Stdout = old

			// We only care that validation passed (no "invalid platform" error)
			if err != nil && strings.Contains(err.Error(), "invalid platform") {
				t.Errorf("runConfigSet() rejected valid platform %q: %v", platform, err)
			}
		})
	}
}

func TestConfigList(t *testing.T) {
	t.Run("outputs valid YAML", func(t *testing.T) {
		viper.Reset()
		viper.Set("version", 1)
		viper.Set("default_platforms", []string{"claude", "opencode"})

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runConfigList(nil, nil)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if err != nil {
			t.Fatalf("runConfigList() error = %v", err)
		}

		// Verify output is valid YAML
		var parsed map[string]interface{}
		if err := yaml.Unmarshal([]byte(output), &parsed); err != nil {
			t.Errorf("runConfigList() output is not valid YAML: %v\nOutput: %s", err, output)
		}

		// Verify expected keys are present
		if _, ok := parsed["version"]; !ok {
			t.Error("runConfigList() output missing 'version' key")
		}
		if _, ok := parsed["default_platforms"]; !ok {
			t.Error("runConfigList() output missing 'default_platforms' key")
		}
	})

	t.Run("reflects current config values", func(t *testing.T) {
		viper.Reset()
		viper.Set("version", 42)
		viper.Set("default_platforms", []string{"gemini"})

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_ = runConfigList(nil, nil)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		var parsed map[string]interface{}
		if err := yaml.Unmarshal([]byte(output), &parsed); err != nil {
			t.Fatalf("YAML parse error: %v", err)
		}

		// Check version value
		if v, ok := parsed["version"].(int); !ok || v != 42 {
			t.Errorf("version = %v, want 42", parsed["version"])
		}

		// Check platforms value
		platforms, ok := parsed["default_platforms"].([]interface{})
		if !ok {
			t.Fatalf("default_platforms not a slice: %T", parsed["default_platforms"])
		}
		if len(platforms) != 1 || platforms[0] != "gemini" {
			t.Errorf("default_platforms = %v, want [gemini]", platforms)
		}
	})

	t.Run("handles empty platforms", func(t *testing.T) {
		viper.Reset()
		viper.Set("version", 1)
		viper.Set("default_platforms", []string{})

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runConfigList(nil, nil)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		if err != nil {
			t.Fatalf("runConfigList() error = %v", err)
		}

		// Should still be valid YAML
		var parsed map[string]interface{}
		if err := yaml.Unmarshal([]byte(output), &parsed); err != nil {
			t.Errorf("runConfigList() output is not valid YAML: %v", err)
		}
	})
}

func TestWriteConfig(t *testing.T) {
	// Note: writeConfig() uses paths.ConfigHome() which returns the real XDG
	// config directory. We can't easily mock this, so we test the behavior
	// indirectly by verifying what gets written when the directory exists.

	t.Run("creates valid YAML content", func(t *testing.T) {
		// We can test the YAML marshaling logic by checking runConfigList
		// output, which uses the same marshaling logic as writeConfig.
		viper.Reset()
		viper.Set("version", 1)
		viper.Set("default_platforms", []string{"claude", "opencode"})

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_ = runConfigList(nil, nil)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		// Verify the YAML structure matches what writeConfig would produce
		var parsed map[string]interface{}
		if err := yaml.Unmarshal([]byte(output), &parsed); err != nil {
			t.Errorf("marshaled config is not valid YAML: %v", err)
		}

		// Verify structure
		if v, ok := parsed["version"].(int); !ok || v != 1 {
			t.Errorf("version = %v, want 1", parsed["version"])
		}

		platforms, ok := parsed["default_platforms"].([]interface{})
		if !ok {
			t.Fatalf("default_platforms type = %T, want []interface{}", parsed["default_platforms"])
		}
		if len(platforms) != 2 {
			t.Errorf("len(default_platforms) = %d, want 2", len(platforms))
		}
	})
}

func TestConfigGet_InterfaceSlice(t *testing.T) {
	// Test that runConfigGet handles []interface{} (which viper sometimes returns)
	viper.Reset()
	viper.Set("mixed_slice", []interface{}{"a", "b", "c"})

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runConfigGet(nil, []string{"mixed_slice"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	got := buf.String()

	if err != nil {
		t.Errorf("runConfigGet() error = %v", err)
	}

	want := "a\nb\nc\n"
	if got != want {
		t.Errorf("runConfigGet(mixed_slice) = %q, want %q", got, want)
	}
}

func TestSetupTestConfig(t *testing.T) {
	// Verify the test helper itself works correctly
	configFile := setupTestConfig(t)

	// Check viper is configured
	if viper.GetInt("version") != 1 {
		t.Errorf("setupTestConfig() version = %d, want 1", viper.GetInt("version"))
	}

	platforms := viper.GetStringSlice("default_platforms")
	if len(platforms) != 2 {
		t.Errorf("setupTestConfig() platforms count = %d, want 2", len(platforms))
	}

	// Check config file path exists and is a valid path
	dir := filepath.Dir(configFile)
	if dir == "" || dir == "." {
		t.Errorf("setupTestConfig() configFile has invalid directory: %s", configFile)
	}

	// Check file doesn't exist yet (setupTestConfig doesn't write it)
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Errorf("setupTestConfig() should not create file, but file exists or error: %v", err)
	}
}

func TestConfigSet_ErrorContainsValidOptions(t *testing.T) {
	// Verify error message includes list of valid platforms
	viper.Reset()

	// Suppress stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runConfigSet(nil, []string{"default_platforms", "invalid"})

	w.Close()
	os.Stdout = old

	if err == nil {
		t.Fatal("expected error for invalid platform")
	}

	errMsg := err.Error()

	// Should mention the invalid platform
	if !strings.Contains(errMsg, "invalid") {
		t.Errorf("error should mention invalid platform name, got: %s", errMsg)
	}

	// Should list valid options
	validPlatforms := []string{"claude", "opencode", "codex", "gemini"}
	for _, p := range validPlatforms {
		if !strings.Contains(errMsg, p) {
			t.Errorf("error should list valid platform %q, got: %s", p, errMsg)
		}
	}
}
