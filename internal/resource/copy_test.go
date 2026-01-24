package resource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestCacheDir creates a mock repos cache structure for testing.
// Returns the cache directory path that can be passed to CopyToTempFromCache.
func setupTestCacheDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// createSkillInCache creates a skill directory structure in the test cache.
func createSkillInCache(t *testing.T, cacheDir, repoName, skillName string, files map[string]string) {
	t.Helper()
	skillDir := filepath.Join(cacheDir, repoName, "skills", skillName)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		filePath := filepath.Join(skillDir, name)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// createCommandInCache creates a command in the test cache.
// If isDir is true, creates a directory with command.md; otherwise creates a flat .md file.
func createCommandInCache(t *testing.T, cacheDir, repoName, cmdName string, content string, isDir bool) {
	t.Helper()
	var cmdPath string
	if isDir {
		cmdDir := filepath.Join(cacheDir, repoName, "commands", cmdName)
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatal(err)
		}
		cmdPath = filepath.Join(cmdDir, "command.md")
	} else {
		cmdDir := filepath.Join(cacheDir, repoName, "commands")
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatal(err)
		}
		cmdPath = filepath.Join(cmdDir, cmdName+".md")
	}
	if err := os.WriteFile(cmdPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// createAgentInCache creates an agent in the test cache.
// If isDir is true, creates a directory with AGENT.md; otherwise creates a flat .md file.
func createAgentInCache(t *testing.T, cacheDir, repoName, agentName string, content string, isDir bool) {
	t.Helper()
	var agentPath string
	if isDir {
		agentDir := filepath.Join(cacheDir, repoName, "agents", agentName)
		if err := os.MkdirAll(agentDir, 0o755); err != nil {
			t.Fatal(err)
		}
		agentPath = filepath.Join(agentDir, "AGENT.md")
	} else {
		agentDir := filepath.Join(cacheDir, repoName, "agents")
		if err := os.MkdirAll(agentDir, 0o755); err != nil {
			t.Fatal(err)
		}
		agentPath = filepath.Join(agentDir, agentName+".md")
	}
	if err := os.WriteFile(agentPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCopyToTempFromCache_SkillDirectory(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	// Create a skill with multiple files
	skillFiles := map[string]string{
		"SKILL.md":             "---\nname: test-skill\n---\n\nSkill content",
		"helper.md":            "# Helper\n\nHelper content",
		"examples/basic.md":    "Basic example",
		"examples/advanced.md": "Advanced example",
	}
	createSkillInCache(t, cacheDir, "test-repo", "test-skill", skillFiles)

	res := &Resource{
		Name:     "test-skill",
		Type:     TypeSkill,
		RepoName: "test-repo",
		Path:     "skills/test-skill",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err != nil {
		t.Fatalf("CopyToTempFromCache() error = %v", err)
	}
	defer os.RemoveAll(filepath.Dir(tempPath)) // Clean up parent temp dir

	// Verify the temp path was created with expected prefix and preserves resource name
	if !strings.Contains(tempPath, "aix-install-") {
		t.Errorf("expected temp path to contain 'aix-install-', got %s", tempPath)
	}

	// For directory resources, the path should end with the resource name
	if filepath.Base(tempPath) != "test-skill" {
		t.Errorf("expected temp path to end with resource name 'test-skill', got %s", filepath.Base(tempPath))
	}

	// Verify all files were copied
	for relPath, expectedContent := range skillFiles {
		fullPath := filepath.Join(tempPath, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("failed to read copied file %s: %v", relPath, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("file %s content mismatch: got %q, want %q", relPath, string(content), expectedContent)
		}
	}
}

func TestCopyToTempFromCache_CommandDirectory(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	content := "---\nname: deploy\n---\n\nDeploy instructions"
	createCommandInCache(t, cacheDir, "test-repo", "deploy", content, true)

	res := &Resource{
		Name:     "deploy",
		Type:     TypeCommand,
		RepoName: "test-repo",
		Path:     "commands/deploy",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err != nil {
		t.Fatalf("CopyToTempFromCache() error = %v", err)
	}
	defer os.RemoveAll(filepath.Dir(tempPath)) // Clean up parent temp dir

	// For directory resources, the path should end with the resource name
	if filepath.Base(tempPath) != "deploy" {
		t.Errorf("expected temp path to end with resource name 'deploy', got %s", filepath.Base(tempPath))
	}

	// Verify command.md was copied
	cmdContent, err := os.ReadFile(filepath.Join(tempPath, "command.md"))
	if err != nil {
		t.Fatalf("failed to read command.md: %v", err)
	}
	if string(cmdContent) != content {
		t.Errorf("command.md content mismatch: got %q, want %q", string(cmdContent), content)
	}
}

func TestCopyToTempFromCache_CommandFlatFile(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	content := "---\nname: quick-deploy\n---\n\nQuick deploy instructions"
	createCommandInCache(t, cacheDir, "test-repo", "quick-deploy", content, false)

	res := &Resource{
		Name:     "quick-deploy",
		Type:     TypeCommand,
		RepoName: "test-repo",
		Path:     "commands/quick-deploy.md",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err != nil {
		t.Fatalf("CopyToTempFromCache() error = %v", err)
	}
	defer os.RemoveAll(tempPath)

	// Verify the .md file was copied
	cmdContent, err := os.ReadFile(filepath.Join(tempPath, "quick-deploy.md"))
	if err != nil {
		t.Fatalf("failed to read quick-deploy.md: %v", err)
	}
	if string(cmdContent) != content {
		t.Errorf("quick-deploy.md content mismatch: got %q, want %q", string(cmdContent), content)
	}
}

func TestCopyToTempFromCache_AgentDirectory(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	content := "---\nname: reviewer\n---\n\nCode review agent"
	createAgentInCache(t, cacheDir, "test-repo", "reviewer", content, true)

	res := &Resource{
		Name:     "reviewer",
		Type:     TypeAgent,
		RepoName: "test-repo",
		Path:     "agents/reviewer",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err != nil {
		t.Fatalf("CopyToTempFromCache() error = %v", err)
	}
	defer os.RemoveAll(filepath.Dir(tempPath)) // Clean up parent temp dir

	// For directory resources, the path should end with the resource name
	if filepath.Base(tempPath) != "reviewer" {
		t.Errorf("expected temp path to end with resource name 'reviewer', got %s", filepath.Base(tempPath))
	}

	// Verify AGENT.md was copied
	agentContent, err := os.ReadFile(filepath.Join(tempPath, "AGENT.md"))
	if err != nil {
		t.Fatalf("failed to read AGENT.md: %v", err)
	}
	if string(agentContent) != content {
		t.Errorf("AGENT.md content mismatch: got %q, want %q", string(agentContent), content)
	}
}

func TestCopyToTempFromCache_AgentFlatFile(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	content := "---\nname: helper\n---\n\nHelper agent"
	createAgentInCache(t, cacheDir, "test-repo", "helper", content, false)

	res := &Resource{
		Name:     "helper",
		Type:     TypeAgent,
		RepoName: "test-repo",
		Path:     "agents/helper.md",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err != nil {
		t.Fatalf("CopyToTempFromCache() error = %v", err)
	}
	defer os.RemoveAll(tempPath)

	// Verify the .md file was copied
	agentContent, err := os.ReadFile(filepath.Join(tempPath, "helper.md"))
	if err != nil {
		t.Fatalf("failed to read helper.md: %v", err)
	}
	if string(agentContent) != content {
		t.Errorf("helper.md content mismatch: got %q, want %q", string(agentContent), content)
	}
}

func TestCopyToTempFromCache_NonExistentResource(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	res := &Resource{
		Name:     "ghost-skill",
		Type:     TypeSkill,
		RepoName: "nonexistent-repo",
		Path:     "skills/ghost-skill",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err == nil {
		// Clean up if somehow it succeeded
		os.RemoveAll(tempPath)
		t.Fatal("CopyToTempFromCache() expected error for non-existent resource")
	}

	// Verify error wraps ErrResourceNotFound
	if !strings.Contains(err.Error(), "resource not found") {
		t.Errorf("expected error to contain 'resource not found', got: %v", err)
	}

	// Verify no temp directory was left behind
	if tempPath != "" {
		if _, err := os.Stat(tempPath); err == nil {
			t.Errorf("expected no temp directory to be left behind, but found: %s", tempPath)
		}
	}
}

func TestCopyToTempFromCache_PreservesFilePermissions(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	// Create a skill with a specific file permission
	skillDir := filepath.Join(cacheDir, "test-repo", "skills", "perm-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a file with executable permissions
	scriptPath := filepath.Join(skillDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho hello"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Create SKILL.md
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: perm-skill\n---\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := &Resource{
		Name:     "perm-skill",
		Type:     TypeSkill,
		RepoName: "test-repo",
		Path:     "skills/perm-skill",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err != nil {
		t.Fatalf("CopyToTempFromCache() error = %v", err)
	}
	defer os.RemoveAll(filepath.Dir(tempPath)) // Clean up parent temp dir

	// Verify script permissions were preserved
	copiedScript := filepath.Join(tempPath, "script.sh")
	info, err := os.Stat(copiedScript)
	if err != nil {
		t.Fatalf("failed to stat copied script: %v", err)
	}

	// Check executable bit (at least one execute bit should be set)
	if info.Mode()&0o111 == 0 {
		t.Errorf("expected executable permissions to be preserved, got %o", info.Mode())
	}
}

func TestCopyToTempFromCache_TempDirectoryCreation(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	createSkillInCache(t, cacheDir, "test-repo", "temp-test", map[string]string{
		"SKILL.md": "---\nname: temp-test\n---\n\n",
	})

	res := &Resource{
		Name:     "temp-test",
		Type:     TypeSkill,
		RepoName: "test-repo",
		Path:     "skills/temp-test",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err != nil {
		t.Fatalf("CopyToTempFromCache() error = %v", err)
	}
	defer os.RemoveAll(filepath.Dir(tempPath)) // Clean up parent temp dir

	// Verify temp path exists and is a directory
	info, err := os.Stat(tempPath)
	if err != nil {
		t.Fatalf("temp path does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("temp path is not a directory")
	}

	// For directory resources, the returned path ends with the resource name
	if filepath.Base(tempPath) != "temp-test" {
		t.Errorf("expected temp path to end with resource name 'temp-test', got: %s", filepath.Base(tempPath))
	}

	// The parent directory should have the expected prefix
	parentDir := filepath.Base(filepath.Dir(tempPath))
	if !strings.HasPrefix(parentDir, "aix-install-") {
		t.Errorf("expected parent temp dir to have 'aix-install-' prefix, got: %s", parentDir)
	}
}

func TestIsDirectoryResource(t *testing.T) {
	tests := []struct {
		name     string
		resource *Resource
		want     bool
	}{
		{
			name: "skill is always directory",
			resource: &Resource{
				Type: TypeSkill,
				Path: "skills/my-skill",
			},
			want: true,
		},
		{
			name: "command directory",
			resource: &Resource{
				Type: TypeCommand,
				Path: "commands/my-cmd",
			},
			want: true,
		},
		{
			name: "command flat file",
			resource: &Resource{
				Type: TypeCommand,
				Path: "commands/my-cmd.md",
			},
			want: false,
		},
		{
			name: "agent directory",
			resource: &Resource{
				Type: TypeAgent,
				Path: "agents/my-agent",
			},
			want: true,
		},
		{
			name: "agent flat file",
			resource: &Resource{
				Type: TypeAgent,
				Path: "agents/my-agent.md",
			},
			want: false,
		},
		{
			name: "MCP is always flat file",
			resource: &Resource{
				Type: TypeMCP,
				Path: "mcp/server.json",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDirectoryResource(tt.resource)
			if got != tt.want {
				t.Errorf("IsDirectoryResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyDir_NestedDirectories(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create nested structure
	nestedDir := filepath.Join(srcDir, "level1", "level2", "level3")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create files at various levels
	files := map[string]string{
		"root.txt":                    "root content",
		"level1/l1.txt":               "level 1 content",
		"level1/level2/l2.txt":        "level 2 content",
		"level1/level2/level3/l3.txt": "level 3 content",
	}
	for relPath, content := range files {
		fullPath := filepath.Join(srcDir, relPath)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Copy
	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir() error = %v", err)
	}

	// Verify all files were copied
	for relPath, expectedContent := range files {
		fullPath := filepath.Join(dstDir, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("failed to read copied file %s: %v", relPath, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("file %s content mismatch: got %q, want %q", relPath, string(content), expectedContent)
		}
	}
}

func TestCopyFile_Basic(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "test.txt")
	dstFile := filepath.Join(dstDir, "test.txt")

	content := "Hello, World!"
	if err := os.WriteFile(srcFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(srcFile, dstFile); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	// Verify content
	copiedContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(copiedContent) != content {
		t.Errorf("content mismatch: got %q, want %q", string(copiedContent), content)
	}
}

func TestCopyFile_NonExistentSource(t *testing.T) {
	dstDir := t.TempDir()
	dstFile := filepath.Join(dstDir, "test.txt")

	err := copyFile("/nonexistent/path/file.txt", dstFile)
	if err == nil {
		t.Error("copyFile() expected error for non-existent source")
	}
}

// TestCopyToTempFromCache_CleansUpOnError verifies that the temp directory is removed
// if the copy operation fails.
func TestCopyToTempFromCache_CleansUpOnError(t *testing.T) {
	cacheDir := setupTestCacheDir(t)

	// Create a skill directory but make it unreadable
	skillDir := filepath.Join(cacheDir, "test-repo", "skills", "broken-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create the SKILL.md file
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory and make it unreadable to force a copy error
	unreadableDir := filepath.Join(skillDir, "unreadable")
	if err := os.MkdirAll(unreadableDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(unreadableDir, 0o755)
	})

	res := &Resource{
		Name:     "broken-skill",
		Type:     TypeSkill,
		RepoName: "test-repo",
		Path:     "skills/broken-skill",
	}

	tempPath, err := CopyToTempFromCache(res, cacheDir)
	if err == nil {
		// If it somehow succeeded (e.g., running as root), clean up and skip
		os.RemoveAll(tempPath)
		t.Skip("copy succeeded unexpectedly, possibly running as root")
	}

	// The temp path should be empty when an error occurs
	if tempPath != "" {
		if _, statErr := os.Stat(tempPath); statErr == nil {
			os.RemoveAll(tempPath) // Clean up for test hygiene
			t.Error("temp directory was not cleaned up after copy error")
		}
	}
}
