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
	if err := os.WriteFile(configPath, content, 0644); err != nil {
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
