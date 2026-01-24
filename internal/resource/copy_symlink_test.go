package resource

import (
	"os"
	"path/filepath"
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

	// Copy to temp
	// We pass srcDir as cacheDir
	tempDir, err := CopyToTempFromCache(res, srcDir)
	if err != nil {
		t.Fatalf("CopyToTempFromCache failed: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(tempDir))

	// Verify the symlink was NOT followed or copied
	destExploit := filepath.Join(tempDir, "exploit.txt")

	// If the file exists, check content
	if _, err := os.Stat(destExploit); err == nil {
		content, _ := os.ReadFile(destExploit)
		if string(content) == "sensitive data" {
			t.Error("Symlink traversal succeeded: secret file was copied")
		} else {
			// It might be copied as a symlink or empty file, which is safer but strict rejection is better
			t.Logf("Symlink copied but content differs (or is symlink)")
		}
	} else if !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking destination: %v", err)
	}
	// If it doesn't exist, that's what we want (symlink skipped)
}
