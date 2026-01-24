package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	// Reset viper state
	viper.Reset()

	Init()

	// Check defaults are set
	if viper.GetInt("version") != 1 {
		t.Errorf("expected version default 1, got %d", viper.GetInt("version"))
	}

	platforms := viper.GetStringSlice("default_platforms")
	if len(platforms) == 0 {
		t.Error("expected default_platforms to have values")
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	viper.Reset()

	// Set AIX_CONFIG_DIR to a temp dir to avoid loading system config
	tempDir := t.TempDir()
	t.Setenv("AIX_CONFIG_DIR", tempDir)

	Init()

	// Load with no config file should not error
	cfg, err := Load("")
	if err != nil {
		t.Errorf("Load() with no config file should not error: %v", err)
	}
	if cfg == nil {
		t.Error("expected config to be returned")
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	viper.Reset()

	// Create temp config file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	content := []byte("default_platforms:\n  - claude\n  - opencode\n")
	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatal(err)
	}

	Init()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.DefaultPlatforms) != 2 {
		t.Errorf("expected 2 platforms, got %d", len(cfg.DefaultPlatforms))
	}
}

func TestLoad_ExplicitPathNotFound(t *testing.T) {
	viper.Reset()
	Init()

	// Load with non-existent config file should error
	_, err := Load("/non/existent/path/config.yaml")
	if err == nil {
		t.Error("Load() with non-existent explicit path should error")
	}
}

func TestLoad_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "invalid version",
			content: "version: 2\n",
			wantErr: "unsupported config version: 2",
		},
		{
			name:    "invalid default platform",
			content: "default_platforms:\n  - invalid_platform\n",
			wantErr: "invalid default platform: invalid_platform",
		},
		{
			name:    "invalid platform override",
			content: "platforms:\n  invalid_platform:\n    config_dir: /tmp\n",
			wantErr: "invalid platform override key: invalid_platform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			Init()

			dir := t.TempDir()
			configPath := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}

			_, err := Load(configPath)
			if err == nil {
				t.Error("Load() expected error, got nil")
			} else if err.Error() != "validating config: "+tt.wantErr {
				t.Errorf("Load() error = %v, want %v", err, "validating config: "+tt.wantErr)
			}
		})
	}
}
