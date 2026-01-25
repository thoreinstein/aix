package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Valid URLs
		{"https", "https://github.com/user/repo.git", false},
		{"http", "http://github.com/user/repo.git", false},
		{"ssh", "ssh://git@github.com/user/repo.git", false},
		{"git", "git://github.com/user/repo.git", false},
		{"file", "file:///path/to/repo.git", false},
		{"scp-like", "git@github.com:user/repo.git", false},
		{"scp-like subdomain", "git@sub.domain.com:user/repo.git", false},
		{"scp-like user", "user@host.com:path/to/repo.git", false},

		// Invalid URLs
		{"empty", "", true},
		{"argument injection", "-oProxyCommand=touch /tmp/pwned", true},
		{"ext protocol", "ext::sh -c touch% /tmp/pwned", true},
		{"unknown scheme", "ftp://github.com/user/repo.git", true},
		{"missing scheme", "github.com/user/repo.git", true},              // We require scheme or scp-like
		{"scp-like missing git suffix", "git@github.com:user/repo", true}, // Regex requires .git suffix
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestIsURL(t *testing.T) {
	if !IsURL("https://github.com/user/repo.git") {
		t.Error("expected true for valid URL")
	}
	if IsURL("not-a-url") {
		t.Error("expected false for invalid URL")
	}
}

func TestValidateRemote(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Test non-existent path
	err := ValidateRemote(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("expected error for nonexistent path, got nil")
	}

	// 2. Test non-git directory
	err = ValidateRemote(tmpDir)
	if err == nil {
		t.Error("expected error for non-git directory, got nil")
	}

	// 3. Test valid git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	err = ValidateRemote(tmpDir)
	if err != nil {
		t.Errorf("expected nil error for valid git directory, got %v", err)
	}
}

func TestCloneAndPull_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	sourceRepo := filepath.Join(tmpDir, "source")
	destRepo := filepath.Join(tmpDir, "dest")

	// Create source repo
	createLocalGitRepo(t, sourceRepo)

	// Test Clone
	err := Clone("file://"+sourceRepo, destRepo, 1)
	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}

	if err := ValidateRemote(destRepo); err != nil {
		t.Errorf("cloned directory is not a valid git repo: %v", err)
	}

	// Test Pull
	// First, add a commit to source
	if err := os.WriteFile(filepath.Join(sourceRepo, "newfile.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, sourceRepo, "add", "newfile.txt")
	runGit(t, sourceRepo, "commit", "-m", "add newfile")

	// Pull in dest
	err = Pull(destRepo)
	if err != nil {
		t.Errorf("Pull() error = %v", err)
	}

	// Verify pull worked
	if _, err := os.Stat(filepath.Join(destRepo, "newfile.txt")); err != nil {
		t.Errorf("pulled file missing: %v", err)
	}
}

func createLocalGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# Test Repo"), 0644); err != nil {
		t.Fatal(err)
	}

	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial commit")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, out)
	}
}
