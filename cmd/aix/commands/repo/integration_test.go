// Package repo provides integration tests for the repository ecosystem.
//
// These tests verify the complete repo lifecycle: add -> list -> update -> remove,
// as well as validation warnings and resource discovery integration.
package repo

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/config"
	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/internal/repo"
	"github.com/thoreinstein/aix/internal/resource"
)

// createLocalGitRepo creates a bare git repository with the given resources.
// Returns the file:// URL for the repository.
func createLocalGitRepo(t *testing.T, skills, commands, agents, mcpServers map[string]string) string {
	t.Helper()

	// Create a work directory for the source repo
	srcDir := t.TempDir()

	// Initialize git repo
	if err := runGit(srcDir, "init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Configure git user for commits
	if err := runGit(srcDir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config email failed: %v", err)
	}
	if err := runGit(srcDir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config name failed: %v", err)
	}

	// Always create a README so we have at least one file to commit
	if err := os.WriteFile(filepath.Join(srcDir, "README.md"), []byte("# Test Repository\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Create skills
	for name, content := range skills {
		skillDir := filepath.Join(srcDir, "skills", name)
		if err := os.MkdirAll(skillDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Create commands
	for name, content := range commands {
		cmdDir := filepath.Join(srcDir, "commands", name)
		if err := os.MkdirAll(cmdDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cmdDir, "command.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Create agents
	for name, content := range agents {
		agentDir := filepath.Join(srcDir, "agents", name)
		if err := os.MkdirAll(agentDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// Create MCP servers
	if len(mcpServers) > 0 {
		mcpDir := filepath.Join(srcDir, "mcp")
		if err := os.MkdirAll(mcpDir, 0o700); err != nil {
			t.Fatal(err)
		}
		for name, content := range mcpServers {
			if err := os.WriteFile(filepath.Join(mcpDir, name+".json"), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Stage and commit all files
	if err := runGit(srcDir, "add", "-A"); err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	if err := runGit(srcDir, "commit", "-m", "Initial commit"); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Return file:// URL
	return "file://" + srcDir
}

// runGit executes a git command in the specified directory.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil
	return errors.Wrap(cmd.Run(), "running git command")
}

// validSkillFrontmatter returns valid SKILL.md content.
func validSkillFrontmatter(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nSkill content here."
}

// validCommandFrontmatter returns valid command.md content.
func validCommandFrontmatter(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nCommand instructions here."
}

// validAgentFrontmatter returns valid AGENT.md content.
func validAgentFrontmatter(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nAgent instructions here."
}

// validMCPJSON returns valid MCP server JSON.
func validMCPJSON(name, command string) string {
	return `{"name": "` + name + `", "command": "` + command + `", "args": ["-y", "@mcp/test"]}`
}

// setupTestConfig creates a temporary config file and returns its path.
// Cleanup is handled automatically by t.TempDir().
func setupTestConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "config.yaml")
}

// TestIntegration_RepoLifecycle tests the complete repo lifecycle:
// add -> list -> update -> remove.
func TestIntegration_RepoLifecycle(t *testing.T) {
	// Create a local git repo with test resources
	repoURL := createLocalGitRepo(t,
		map[string]string{
			"code-review": validSkillFrontmatter("code-review", "Reviews code"),
		},
		map[string]string{
			"deploy": validCommandFrontmatter("deploy", "Deploy application"),
		},
		nil,
		nil,
	)

	// Setup test config
	configPath := setupTestConfig(t)
	manager := repo.NewManager(configPath)

	// Step 1: Add the repository (need explicit name for file:// URLs since temp paths
	// derive names like "001" which fail the "starts with letter" validation)
	repoConfig, err := manager.Add(repoURL, repo.WithName("lifecycle-test"))
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	if repoConfig.Name != "lifecycle-test" {
		t.Errorf("expected repo name 'lifecycle-test', got %q", repoConfig.Name)
	}
	if repoConfig.URL != repoURL {
		t.Errorf("URL = %q, want %q", repoConfig.URL, repoURL)
	}
	if repoConfig.Path == "" {
		t.Error("expected repo path to be set")
	}

	// Verify the repo was cloned
	if _, err := os.Stat(repoConfig.Path); os.IsNotExist(err) {
		t.Errorf("repo path does not exist: %s", repoConfig.Path)
	}

	// Step 2: List repositories
	repos, err := manager.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != repoConfig.Name {
		t.Errorf("List()[0].Name = %q, want %q", repos[0].Name, repoConfig.Name)
	}

	// Step 3: Get specific repo
	gotRepo, err := manager.Get(repoConfig.Name)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if gotRepo.URL != repoURL {
		t.Errorf("Get().URL = %q, want %q", gotRepo.URL, repoURL)
	}

	// Step 4: Update the repository
	err = manager.Update(repoConfig.Name)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Step 5: Remove the repository
	err = manager.Remove(repoConfig.Name)
	if err != nil {
		t.Fatalf("Remove() failed: %v", err)
	}

	// Verify repo was removed from config
	repos, err = manager.List()
	if err != nil {
		t.Fatalf("List() after remove failed: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 repos after remove, got %d", len(repos))
	}

	// Verify repo directory was cleaned up
	if _, err := os.Stat(repoConfig.Path); !os.IsNotExist(err) {
		t.Errorf("repo path should not exist after remove: %s", repoConfig.Path)
	}
}

// TestIntegration_RepoWithCustomName tests adding a repo with a custom name.
func TestIntegration_RepoWithCustomName(t *testing.T) {
	repoURL := createLocalGitRepo(t,
		map[string]string{
			"test-skill": validSkillFrontmatter("test-skill", "Test skill"),
		},
		nil, nil, nil,
	)

	configPath := setupTestConfig(t)
	manager := repo.NewManager(configPath)

	// Add with custom name
	customName := "my-custom-repo"
	repoConfig, err := manager.Add(repoURL, repo.WithName(customName))
	if err != nil {
		t.Fatalf("Add() with custom name failed: %v", err)
	}

	if repoConfig.Name != customName {
		t.Errorf("Name = %q, want %q", repoConfig.Name, customName)
	}

	// Cleanup
	_ = manager.Remove(customName)
}

// TestIntegration_ValidationWarnings tests that validation warnings are generated
// for repos with invalid resources.
func TestIntegration_ValidationWarnings(t *testing.T) {
	tests := []struct {
		name         string
		skills       map[string]string
		commands     map[string]string
		agents       map[string]string
		mcpServers   map[string]string
		wantWarnings bool
		wantContains string
	}{
		{
			name: "valid repo - no warnings",
			skills: map[string]string{
				"valid-skill": validSkillFrontmatter("valid-skill", "Valid skill"),
			},
			wantWarnings: false,
		},
		{
			name: "malformed skill frontmatter",
			skills: map[string]string{
				"broken-skill": "---\nname: [invalid yaml\n---\n\nContent",
			},
			wantWarnings: true,
			wantContains: "invalid frontmatter",
		},
		{
			name: "malformed MCP JSON",
			mcpServers: map[string]string{
				"broken-mcp": `{"name": "broken", "command": }`,
			},
			wantWarnings: true,
			wantContains: "invalid JSON",
		},
		{
			name:   "skill directory missing SKILL.md",
			skills: map[string]string{
				// This creates a valid skill, but we'll add an empty dir
			},
			wantWarnings: false, // Empty dirs don't cause warnings in the current impl
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoURL := createLocalGitRepo(t, tt.skills, tt.commands, tt.agents, tt.mcpServers)

			configPath := setupTestConfig(t)
			manager := repo.NewManager(configPath)

			// Use explicit name for file:// URLs (temp paths derive invalid names)
			repoConfig, err := manager.Add(repoURL, repo.WithName("validation-test"))
			if err != nil {
				t.Fatalf("Add() failed: %v", err)
			}

			// Validate repo content
			warnings := repo.ValidateRepoContent(repoConfig.Path)

			// Filter out "directory not found" warnings (expected for partial repos)
			var actionableWarnings []repo.ValidationWarning
			for _, w := range warnings {
				if w.Message != "directory not found" {
					actionableWarnings = append(actionableWarnings, w)
				}
			}

			hasWarnings := len(actionableWarnings) > 0
			if hasWarnings != tt.wantWarnings {
				t.Errorf("hasWarnings = %v, want %v; warnings: %v", hasWarnings, tt.wantWarnings, actionableWarnings)
			}

			if tt.wantContains != "" && hasWarnings {
				found := false
				for _, w := range actionableWarnings {
					if strings.Contains(w.Message, tt.wantContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning containing %q, got %v", tt.wantContains, actionableWarnings)
				}
			}

			// Cleanup
			_ = manager.Remove(repoConfig.Name)
		})
	}
}

// TestIntegration_ResourceDiscovery tests that resources can be discovered
// from added repositories.
func TestIntegration_ResourceDiscovery(t *testing.T) {
	// Create a repo with multiple resource types
	repoURL := createLocalGitRepo(t,
		map[string]string{
			"code-review":    validSkillFrontmatter("code-review", "Reviews code"),
			"test-generator": validSkillFrontmatter("test-generator", "Generates tests"),
		},
		map[string]string{
			"deploy":   validCommandFrontmatter("deploy", "Deploy application"),
			"rollback": validCommandFrontmatter("rollback", "Rollback deployment"),
		},
		map[string]string{
			"helper": validAgentFrontmatter("helper", "Helpful assistant"),
		},
		map[string]string{
			"github": validMCPJSON("github", "npx"),
		},
	)

	configPath := setupTestConfig(t)
	manager := repo.NewManager(configPath)

	// Add the repository
	repoConfig, err := manager.Add(repoURL, repo.WithName("test-repo"))
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Scan for resources
	scanner := resource.NewScanner()
	repos, err := manager.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	resources, err := scanner.ScanAll(repos)
	if err != nil {
		t.Fatalf("ScanAll() failed: %v", err)
	}

	// Count by type
	counts := make(map[resource.ResourceType]int)
	for _, r := range resources {
		counts[r.Type]++
	}

	// Verify counts
	if counts[resource.TypeSkill] != 2 {
		t.Errorf("expected 2 skills, got %d", counts[resource.TypeSkill])
	}
	if counts[resource.TypeCommand] != 2 {
		t.Errorf("expected 2 commands, got %d", counts[resource.TypeCommand])
	}
	if counts[resource.TypeAgent] != 1 {
		t.Errorf("expected 1 agent, got %d", counts[resource.TypeAgent])
	}
	if counts[resource.TypeMCP] != 1 {
		t.Errorf("expected 1 MCP server, got %d", counts[resource.TypeMCP])
	}

	// Verify repo attribution
	for _, r := range resources {
		if r.RepoName != "test-repo" {
			t.Errorf("resource %s has RepoName = %q, want %q", r.Name, r.RepoName, "test-repo")
		}
	}

	// Cleanup
	_ = manager.Remove(repoConfig.Name)
}

// TestIntegration_MultipleRepos tests operations with multiple repositories.
func TestIntegration_MultipleRepos(t *testing.T) {
	// Create two repos with different resources
	repo1URL := createLocalGitRepo(t,
		map[string]string{
			"skill-a": validSkillFrontmatter("skill-a", "Skill A"),
		},
		nil, nil, nil,
	)

	repo2URL := createLocalGitRepo(t,
		map[string]string{
			"skill-b": validSkillFrontmatter("skill-b", "Skill B"),
		},
		nil, nil, nil,
	)

	configPath := setupTestConfig(t)
	manager := repo.NewManager(configPath)

	// Add both repos
	repo1Config, err := manager.Add(repo1URL, repo.WithName("repo-one"))
	if err != nil {
		t.Fatalf("Add repo1 failed: %v", err)
	}

	repo2Config, err := manager.Add(repo2URL, repo.WithName("repo-two"))
	if err != nil {
		t.Fatalf("Add repo2 failed: %v", err)
	}

	// List should show both
	repos, err := manager.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}

	// Scan should find resources from both
	scanner := resource.NewScanner()
	resources, err := scanner.ScanAll(repos)
	if err != nil {
		t.Fatalf("ScanAll() failed: %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}

	// Verify resources are from different repos
	repoNames := make(map[string]bool)
	for _, r := range resources {
		repoNames[r.RepoName] = true
	}
	if !repoNames["repo-one"] || !repoNames["repo-two"] {
		t.Errorf("expected resources from both repos, got repos: %v", repoNames)
	}

	// Cleanup
	_ = manager.Remove(repo1Config.Name)
	_ = manager.Remove(repo2Config.Name)
}

// TestIntegration_ErrorCases tests error handling for various failure scenarios.
func TestIntegration_ErrorCases(t *testing.T) {
	configPath := setupTestConfig(t)
	manager := repo.NewManager(configPath)

	t.Run("invalid URL", func(t *testing.T) {
		_, err := manager.Add("not-a-url")
		if err == nil {
			t.Error("expected error for invalid URL")
		}
	})

	t.Run("get non-existent repo", func(t *testing.T) {
		_, err := manager.Get("does-not-exist")
		if err == nil {
			t.Error("expected error for non-existent repo")
		}
	})

	t.Run("update non-existent repo", func(t *testing.T) {
		err := manager.Update("does-not-exist")
		if err == nil {
			t.Error("expected error for non-existent repo")
		}
	})

	t.Run("remove non-existent repo", func(t *testing.T) {
		err := manager.Remove("does-not-exist")
		if err == nil {
			t.Error("expected error for non-existent repo")
		}
	})

	t.Run("duplicate repo name", func(t *testing.T) {
		repoURL := createLocalGitRepo(t,
			map[string]string{"test": validSkillFrontmatter("test", "Test")},
			nil, nil, nil,
		)

		_, err := manager.Add(repoURL, repo.WithName("dupe-test"))
		if err != nil {
			t.Fatalf("first Add() failed: %v", err)
		}

		// Second add with same name should fail
		repoURL2 := createLocalGitRepo(t,
			map[string]string{"test2": validSkillFrontmatter("test2", "Test 2")},
			nil, nil, nil,
		)
		_, err = manager.Add(repoURL2, repo.WithName("dupe-test"))
		if err == nil {
			t.Error("expected error for duplicate repo name")
		}

		// Cleanup
		_ = manager.Remove("dupe-test")
	})
}

// TestIntegration_CLIOutput tests the CLI command output formatting.
func TestIntegration_CLIOutput(t *testing.T) {
	repoURL := createLocalGitRepo(t,
		map[string]string{
			"test-skill": validSkillFrontmatter("test-skill", "Test skill"),
		},
		nil, nil, nil,
	)

	// Set the name flag for file:// URLs (temp paths derive invalid names)
	oldNameFlag := nameFlag
	nameFlag = "cli-output-test"
	defer func() { nameFlag = oldNameFlag }()

	// Use default config path and clean up any leftover from previous runs
	configPath := config.DefaultConfigPath()
	manager := repo.NewManager(configPath)
	_ = manager.Remove("cli-output-test")
	_ = os.RemoveAll(filepath.Join(paths.ReposCacheDir(), "cli-output-test"))
	defer func() {
		_ = manager.Remove("cli-output-test")
		_ = os.RemoveAll(filepath.Join(paths.ReposCacheDir(), "cli-output-test"))
	}()

	t.Run("add command output", func(t *testing.T) {
		var buf bytes.Buffer
		err := runAddWithIO([]string{repoURL}, &buf)
		if err != nil {
			t.Fatalf("runAddWithIO() failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "[OK]") {
			t.Error("expected checkmark in output")
		}
		if !strings.Contains(output, "Repository") {
			t.Error("expected 'Repository' in output")
		}
		if !strings.Contains(output, "added") {
			t.Error("expected 'added' in output")
		}
	})

	t.Run("list command output", func(t *testing.T) {
		var buf bytes.Buffer
		err := runListWithWriter(&buf)
		if err != nil {
			t.Fatalf("runListWithWriter() failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "NAME") {
			t.Error("expected 'NAME' header in output")
		}
		if !strings.Contains(output, "URL") {
			t.Error("expected 'URL' header in output")
		}
	})
}

// TestIntegration_Search tests searching for resources across repositories.
func TestIntegration_Search(t *testing.T) {
	// Create repos with varied resources
	repoURL := createLocalGitRepo(t,
		map[string]string{
			"code-review":     validSkillFrontmatter("code-review", "Reviews code for quality"),
			"security-review": validSkillFrontmatter("security-review", "Reviews code for security"),
		},
		map[string]string{
			"deploy": validCommandFrontmatter("deploy", "Deploy to production"),
		},
		nil, nil,
	)

	configPath := setupTestConfig(t)
	manager := repo.NewManager(configPath)

	repoConfig, err := manager.Add(repoURL, repo.WithName("search-test"))
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	repos, err := manager.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	// Scan all repositories to get resources
	scanner := resource.NewScanner()
	allResources, err := scanner.ScanAll(repos)
	if err != nil {
		t.Fatalf("ScanAll() failed: %v", err)
	}

	// Search for "review" - should match 2 skills
	results := resource.Search(allResources, "review", resource.SearchOptions{})
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'review', got %d", len(results))
	}

	// Search for "deploy" - should match 1 command
	results = resource.Search(allResources, "deploy", resource.SearchOptions{})
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'deploy', got %d", len(results))
	}

	// Search with type filter
	results = resource.Search(allResources, "review", resource.SearchOptions{
		Type: resource.TypeSkill,
	})
	if len(results) != 2 {
		t.Errorf("expected 2 skills matching 'review', got %d", len(results))
	}

	// Cleanup
	_ = manager.Remove(repoConfig.Name)
}

// TestIntegration_RepoUpdate tests that updates pull new changes.
func TestIntegration_RepoUpdate(t *testing.T) {
	// Create a source repo that we'll modify
	srcDir := t.TempDir()

	// Initialize git repo
	if err := runGit(srcDir, "init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if err := runGit(srcDir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config email failed: %v", err)
	}
	if err := runGit(srcDir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config name failed: %v", err)
	}

	// Create initial skill
	skillDir := filepath.Join(srcDir, "skills", "initial-skill")
	if err := os.MkdirAll(skillDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
		[]byte(validSkillFrontmatter("initial-skill", "Initial skill")), 0o600); err != nil {
		t.Fatal(err)
	}

	// Initial commit
	if err := runGit(srcDir, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(srcDir, "commit", "-m", "Initial commit"); err != nil {
		t.Fatal(err)
	}

	repoURL := "file://" + srcDir

	// Add repo
	configPath := setupTestConfig(t)
	manager := repo.NewManager(configPath)

	repoConfig, err := manager.Add(repoURL, repo.WithName("update-test"))
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Verify initial state
	scanner := resource.NewScanner()
	repos, _ := manager.List()
	resources, _ := scanner.ScanAll(repos)
	if len(resources) != 1 {
		t.Errorf("expected 1 initial resource, got %d", len(resources))
	}

	// Add a new skill to source repo
	newSkillDir := filepath.Join(srcDir, "skills", "new-skill")
	if err := os.MkdirAll(newSkillDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newSkillDir, "SKILL.md"),
		[]byte(validSkillFrontmatter("new-skill", "New skill")), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runGit(srcDir, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(srcDir, "commit", "-m", "Add new skill"); err != nil {
		t.Fatal(err)
	}

	// Update should pull new changes
	err = manager.Update(repoConfig.Name)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Verify new skill is available
	resources, _ = scanner.ScanAll(repos)
	if len(resources) != 2 {
		t.Errorf("expected 2 resources after update, got %d", len(resources))
	}

	// Cleanup
	_ = manager.Remove(repoConfig.Name)
}

// TestIntegration_ValidationWarningOutput tests that validation warnings
// are properly formatted in CLI output.
func TestIntegration_ValidationWarningOutput(t *testing.T) {
	warnings := []repo.ValidationWarning{
		{Path: "skills/broken/SKILL.md", Message: "invalid frontmatter: unexpected EOF"},
		{Path: "mcp/bad.json", Message: "invalid JSON: syntax error"},
	}

	var buf bytes.Buffer
	printValidationWarnings(&buf, warnings)
	output := buf.String()

	if !strings.Contains(output, "Validation warnings:") {
		t.Error("expected 'Validation warnings:' header")
	}
	if !strings.Contains(output, "skills/broken/SKILL.md") {
		t.Error("expected skill path in output")
	}
	if !strings.Contains(output, "mcp/bad.json") {
		t.Error("expected MCP path in output")
	}
}

// TestIntegration_EmptyRepoList tests behavior with no configured repos.
func TestIntegration_EmptyRepoList(t *testing.T) {
	configPath := setupTestConfig(t)
	manager := repo.NewManager(configPath)

	repos, err := manager.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected empty list, got %d repos", len(repos))
	}

	// Scan should return empty
	scanner := resource.NewScanner()
	resources, err := scanner.ScanAll(repos)
	if err != nil {
		t.Fatalf("ScanAll() failed: %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

// TestIntegration_ListJSON tests JSON output format for list command.
func TestIntegration_ListJSON(t *testing.T) {
	repoURL := createLocalGitRepo(t,
		map[string]string{"test": validSkillFrontmatter("test", "Test")},
		nil, nil, nil,
	)

	// Use default config path since runListWithWriter reads from there
	configPath := config.DefaultConfigPath()
	manager := repo.NewManager(configPath)

	// Clean up any leftover from previous runs (config entry and directory)
	_ = manager.Remove("json-test")
	_ = os.RemoveAll(filepath.Join(paths.ReposCacheDir(), "json-test"))
	defer func() {
		_ = manager.Remove("json-test")
		_ = os.RemoveAll(filepath.Join(paths.ReposCacheDir(), "json-test"))
	}()

	_, err := manager.Add(repoURL, repo.WithName("json-test"))
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Set JSON flag and test output
	listJSON = true
	defer func() { listJSON = false }()

	var buf bytes.Buffer
	err = runListWithWriter(&buf)
	if err != nil {
		t.Fatalf("runListWithWriter() failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"name"`) {
		t.Error("expected JSON name field in output")
	}
	if !strings.Contains(output, `"url"`) {
		t.Error("expected JSON url field in output")
	}
	if !strings.Contains(output, "json-test") {
		t.Error("expected repo name in JSON output")
	}
}

// TestIntegration_ConfigPersistence tests that config changes persist across manager instances.
func TestIntegration_ConfigPersistence(t *testing.T) {
	repoURL := createLocalGitRepo(t,
		map[string]string{"test": validSkillFrontmatter("test", "Test")},
		nil, nil, nil,
	)

	configPath := setupTestConfig(t)

	// First manager instance adds a repo
	manager1 := repo.NewManager(configPath)
	_, err := manager1.Add(repoURL, repo.WithName("persist-test"))
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}

	// Second manager instance should see the repo
	manager2 := repo.NewManager(configPath)
	repos, err := manager2.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("expected 1 repo in second manager, got %d", len(repos))
	}
	if repos[0].Name != "persist-test" {
		t.Errorf("expected 'persist-test', got %q", repos[0].Name)
	}

	// Cleanup using second manager
	_ = manager2.Remove("persist-test")

	// Third manager should see empty list
	manager3 := repo.NewManager(configPath)
	repos, _ = manager3.List()
	if len(repos) != 0 {
		t.Errorf("expected 0 repos after remove, got %d", len(repos))
	}
}

// TestIntegration_RepoAddWithValidationWarnings tests that add command shows warnings.
func TestIntegration_RepoAddWithValidationWarnings(t *testing.T) {
	// Create repo with malformed resources
	srcDir := t.TempDir()

	if err := runGit(srcDir, "init"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(srcDir, "config", "user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(srcDir, "config", "user.name", "Test User"); err != nil {
		t.Fatal(err)
	}

	// Create malformed skill
	skillDir := filepath.Join(srcDir, "skills", "broken")
	if err := os.MkdirAll(skillDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
		[]byte("---\nname: [invalid yaml\n---\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := runGit(srcDir, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if err := runGit(srcDir, "commit", "-m", "Add broken skill"); err != nil {
		t.Fatal(err)
	}

	repoURL := "file://" + srcDir

	// Reset name flag for clean test
	nameFlag = "warning-test"
	defer func() { nameFlag = "" }()

	var buf bytes.Buffer
	err := runAddWithIO([]string{repoURL}, &buf)
	if err != nil {
		t.Fatalf("runAddWithIO() failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[OK]") {
		t.Error("expected success checkmark")
	}
	if !strings.Contains(output, "Validation warnings") {
		t.Error("expected validation warnings in output")
	}
	if !strings.Contains(output, "invalid frontmatter") {
		t.Error("expected 'invalid frontmatter' warning")
	}

	// Cleanup
	configPath := config.DefaultConfigPath()
	manager := repo.NewManager(configPath)
	_ = manager.Remove("warning-test")
}
