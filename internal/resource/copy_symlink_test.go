package resource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyToTemp_SymlinkTraversal(t *testing.T) {
	// Create a temp source directory
	srcDir := t.TempDir()

	// Create a file outside the intended resource directory
	secretFile := filepath.Join(srcDir, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("sensitive data"), 0600); err != nil {
		t.Fatal(err)
	}

	// Create a resource directory
	repoDir := filepath.Join(srcDir, "repo", "skills", "test-skill")
	if err := os.MkdirAll(repoDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create SKILL.md
	if err := os.WriteFile(filepath.Join(repoDir, "SKILL.md"), []byte("..."), 0600); err != nil {
		t.Fatal(err)
	}

	// Create a symlink pointing to the secret file
	if err := os.Symlink(secretFile, filepath.Join(repoDir, "exploit.txt")); err != nil {
		t.Fatal(err)
	}

	// Mock resource
	res := &Resource{
		Name:     "test-skill",
		Type:     TypeSkill,
		RepoName: "repo",
		Path:     "skills/test-skill",
	}

	// Copy to temp - should fail due to symlink
	// We pass srcDir as cacheDir
	tempDir, err := CopyToTempFromCache(res, srcDir)
	if tempDir != "" {
		defer os.RemoveAll(filepath.Dir(tempDir))
	}

	// Expect an error due to symlink rejection
	if err == nil {
		t.Fatal("Expected error due to symlink, but copy succeeded")
	}

	// Verify error message mentions symlink and security
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("Expected error to mention symlink, got: %v", err)
	}
}
