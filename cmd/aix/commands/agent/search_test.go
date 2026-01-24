package agent

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/pkg/fileutil"
)

// setupSearchTest creates a temporary config directory and returns the path.
// The caller should use t.Setenv("AIX_CONFIG_DIR", configDir) before calling
// runSearchWithWriter, and reset the searchRepo flag.
func setupSearchTest(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Write an empty config file (no repos)
	cfg := &config.Config{
		Version:          1,
		DefaultPlatforms: []string{"claude", "opencode"},
		Repos:            nil,
	}
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := fileutil.AtomicWriteYAML(configPath, cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return tmpDir
}

func TestSearchCommand_NoReposConfigured(t *testing.T) {
	configDir := setupSearchTest(t)
	t.Setenv("AIX_CONFIG_DIR", configDir)

	// Reset flags
	searchRepo = ""
	searchJSON = false

	var buf bytes.Buffer
	err := runSearchWithWriter(&buf, []string{"test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No repositories configured") {
		t.Errorf("expected 'No repositories configured' message, got:\n%s", output)
	}
	if !strings.Contains(output, "aix repo add <url>") {
		t.Errorf("expected 'aix repo add <url>' hint, got:\n%s", output)
	}
}

func TestSearchCommand_InvalidRepoFilter(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIX_CONFIG_DIR", tmpDir)

	// Create a fake repo directory (needed for scanner)
	repoDir := filepath.Join(tmpDir, "repos", "test-repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	// Write config with one repo
	cfg := &config.Config{
		Version:          1,
		DefaultPlatforms: []string{"claude", "opencode"},
		Repos: map[string]config.RepoConfig{
			"test-repo": {
				Name: "test-repo",
				URL:  "https://github.com/test/test-repo.git",
				Path: repoDir,
			},
		},
	}
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := fileutil.AtomicWriteYAML(configPath, cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Set invalid repo filter
	searchRepo = "nonexistent-repo"
	searchJSON = false

	var buf bytes.Buffer
	err := runSearchWithWriter(&buf, []string{"test"})
	if err == nil {
		t.Fatal("expected error for invalid repo filter, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, `repository "nonexistent-repo" not found`) {
		t.Errorf("expected error to mention 'nonexistent-repo', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "aix repo list") {
		t.Errorf("expected error to suggest 'aix repo list', got: %s", errMsg)
	}
}

func TestSearchCommand_NoResults(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIX_CONFIG_DIR", tmpDir)

	// Create a repo directory with agents dir (but no agents)
	repoDir := filepath.Join(tmpDir, "repos", "empty-repo")
	agentsDir := filepath.Join(repoDir, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Write config with one repo
	cfg := &config.Config{
		Version:          1,
		DefaultPlatforms: []string{"claude", "opencode"},
		Repos: map[string]config.RepoConfig{
			"empty-repo": {
				Name: "empty-repo",
				URL:  "https://github.com/test/empty-repo.git",
				Path: repoDir,
			},
		},
	}
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := fileutil.AtomicWriteYAML(configPath, cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Reset flags
	searchRepo = ""
	searchJSON = false

	var buf bytes.Buffer
	err := runSearchWithWriter(&buf, []string{"nonexistent-agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `No agents found matching "nonexistent-agent"`) {
		t.Errorf("expected 'No agents found matching' message, got:\n%s", output)
	}
}

func TestSearchCommand_Metadata(t *testing.T) {
	if searchCmd.Use != "search <query>" {
		t.Errorf("Use = %q, want %q", searchCmd.Use, "search <query>")
	}

	if searchCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if searchCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Verify flags are registered
	repoFlag := searchCmd.Flags().Lookup("repo")
	if repoFlag == nil {
		t.Error("--repo flag not registered")
	}

	jsonFlag := searchCmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("--json flag not registered")
	}
}

func TestSearchCommand_ValidRepoFilter(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIX_CONFIG_DIR", tmpDir)

	// Create a repo directory with an agent
	repoDir := filepath.Join(tmpDir, "repos", "my-repo")
	agentDir := filepath.Join(repoDir, "agents", "test-agent")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agent dir: %v", err)
	}

	// Write a valid AGENT.md
	agentContent := `---
name: test-agent
description: A test agent for searching
---
This is a test agent.
`
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(agentContent), 0o644); err != nil {
		t.Fatalf("failed to write AGENT.md: %v", err)
	}

	// Write config with one repo
	cfg := &config.Config{
		Version:          1,
		DefaultPlatforms: []string{"claude", "opencode"},
		Repos: map[string]config.RepoConfig{
			"my-repo": {
				Name: "my-repo",
				URL:  "https://github.com/test/my-repo.git",
				Path: repoDir,
			},
		},
	}
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := fileutil.AtomicWriteYAML(configPath, cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Set valid repo filter
	searchRepo = "my-repo"
	searchJSON = false

	var buf bytes.Buffer
	err := runSearchWithWriter(&buf, []string{"test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should find the agent - check for agent name in output
	if !strings.Contains(output, "test-agent") {
		t.Errorf("expected to find 'test-agent' in output, got:\n%s", output)
	}
}

func TestSearchCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("AIX_CONFIG_DIR", tmpDir)

	// Create a repo directory with an agent
	repoDir := filepath.Join(tmpDir, "repos", "json-repo")
	agentDir := filepath.Join(repoDir, "agents", "json-agent")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agent dir: %v", err)
	}

	// Write a valid AGENT.md
	agentContent := `---
name: json-agent
description: An agent for JSON output testing
---
This is a test agent.
`
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(agentContent), 0o644); err != nil {
		t.Fatalf("failed to write AGENT.md: %v", err)
	}

	// Write config
	cfg := &config.Config{
		Version:          1,
		DefaultPlatforms: []string{"claude", "opencode"},
		Repos: map[string]config.RepoConfig{
			"json-repo": {
				Name: "json-repo",
				URL:  "https://github.com/test/json-repo.git",
				Path: repoDir,
			},
		},
	}
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := fileutil.AtomicWriteYAML(configPath, cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Set JSON output flag
	searchRepo = ""
	searchJSON = true

	var buf bytes.Buffer
	err := runSearchWithWriter(&buf, []string{"json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should be valid JSON with expected fields
	if !strings.Contains(output, `"name": "json-agent"`) {
		t.Errorf("expected JSON output with agent name, got:\n%s", output)
	}
	if !strings.Contains(output, `"repository": "json-repo"`) {
		t.Errorf("expected JSON output with repository, got:\n%s", output)
	}
}
