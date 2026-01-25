package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/git"
)

func TestDeriveNameFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "HTTPS URL with .git suffix",
			url:  "https://github.com/user/my-repo.git",
			want: "my-repo",
		},
		{
			name: "HTTPS URL without .git suffix",
			url:  "https://github.com/user/my-repo",
			want: "my-repo",
		},
		{
			name: "SSH URL with .git suffix",
			url:  "git@github.com:user/my-repo.git",
			want: "my-repo",
		},
		{
			name: "SSH URL without .git suffix",
			url:  "git@github.com:user/my-repo",
			want: "my-repo",
		},
		{
			name: "URL with uppercase chars",
			url:  "https://github.com/user/MyRepo.git",
			want: "myrepo",
		},
		{
			name: "git protocol URL",
			url:  "git://github.com/user/repo.git",
			want: "repo",
		},
		{
			name: "URL with nested path",
			url:  "https://gitlab.com/group/subgroup/repo.git",
			want: "repo",
		},
		{
			name: "simple .git file",
			url:  "repo.git",
			want: "repo",
		},
		{
			name: "trailing slash stripped by filepath.Base",
			url:  "https://github.com/user/repo/",
			want: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveNameFromURL(tt.url)
			if got != tt.want {
				t.Errorf("deriveNameFromURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsURL_ViaGitPackage(t *testing.T) {
	// These tests verify our expectations of git.IsURL behavior
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "HTTPS URL",
			input: "https://github.com/user/repo.git",
			want:  true,
		},
		{
			name:  "SSH URL",
			input: "git@github.com:user/repo.git",
			want:  true,
		},
		{
			name:  "local path",
			input: "/home/user/repo",
			want:  false,
		},
		{
			name:  "relative path",
			input: "./my-skill",
			want:  false,
		},
		{
			name:  "plain name",
			input: "my-repo",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := git.IsURL(tt.input)
			if got != tt.want {
				t.Errorf("git.IsURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNamePattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "simple lowercase",
			input: "myrepo",
			want:  true,
		},
		{
			name:  "with hyphens",
			input: "my-repo",
			want:  true,
		},
		{
			name:  "with numbers",
			input: "repo123",
			want:  true,
		},
		{
			name:  "with hyphens and numbers",
			input: "my-repo-v2",
			want:  true,
		},
		{
			name:  "single letter",
			input: "a",
			want:  true,
		},
		{
			name:  "uppercase rejected",
			input: "MyRepo",
			want:  false,
		},
		{
			name:  "starts with number rejected",
			input: "123repo",
			want:  false,
		},
		{
			name:  "consecutive hyphens rejected",
			input: "my--repo",
			want:  false,
		},
		{
			name:  "leading hyphen rejected",
			input: "-repo",
			want:  false,
		},
		{
			name:  "trailing hyphen rejected",
			input: "repo-",
			want:  false,
		},
		{
			name:  "underscore rejected",
			input: "my_repo",
			want:  false,
		},
		{
			name:  "dot rejected",
			input: "my.repo",
			want:  false,
		},
		{
			name:  "space rejected",
			input: "my repo",
			want:  false,
		},
		{
			name:  "empty string rejected",
			input: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := namePattern.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("namePattern.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestManager_Add_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cacheDir := filepath.Join(tmpDir, "cache")
	m := NewManager(configPath, WithCacheDir(cacheDir))

	// Create a local source git repo
	repoDir := filepath.Join(tmpDir, "source-repo")
	createLocalGitRepo(t, repoDir)
	repoURL := "file://" + repoDir

	// Test Add
	repo, err := m.Add(repoURL, WithName("test-repo"))
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if repo.Name != "test-repo" {
		t.Errorf("repo name = %q, want %q", repo.Name, "test-repo")
	}

	// Verify it's in List
	repos, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("List() returned %d repos, want 1", len(repos))
	}

	// Test Get
	repo2, err := m.Get("test-repo")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if repo2.URL != repoURL {
		t.Errorf("Get() URL = %q, want %q", repo2.URL, repoURL)
	}

	// Test Remove
	err = m.Remove("test-repo")
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// Verify it's gone
	repos, _ = m.List()
	if len(repos) != 0 {
		t.Errorf("List() returned %d repos after removal, want 0", len(repos))
	}
}

func TestManager_Add_Collision(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cacheDir := filepath.Join(tmpDir, "cache")
	m := NewManager(configPath, WithCacheDir(cacheDir))

	repoDir := filepath.Join(tmpDir, "source-repo")
	createLocalGitRepo(t, repoDir)
	repoURL := "file://" + repoDir

	// Add once
	_, err := m.Add(repoURL, WithName("test-repo"))
	if err != nil {
		t.Fatal(err)
	}

	// Add again with same name
	_, err = m.Add(repoURL, WithName("test-repo"))
	if !isNameCollisionError(err) {
		t.Errorf("expected NameCollision error, got %v", err)
	}
}

func TestManager_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	cacheDir := filepath.Join(tmpDir, "cache")
	m := NewManager(configPath, WithCacheDir(cacheDir))

	repoDir := filepath.Join(tmpDir, "source-repo")
	createLocalGitRepo(t, repoDir)
	repoURL := "file://" + repoDir

	// Add repo
	_, err := m.Add(repoURL, WithName("test-repo"))
	if err != nil {
		t.Fatal(err)
	}

	// Test Update
	err = m.Update("test-repo")
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	// Test Update all
	err = m.Update("")
	if err != nil {
		t.Errorf("Update(\"\") error = %v", err)
	}
}

func createLocalGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, out)
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")

	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# Test Repo"), 0644); err != nil {
		t.Fatal(err)
	}

	runGit("add", "README.md")
	runGit("commit", "-m", "initial commit")
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "repository not found") || errors.Is(err, ErrNotFound)
}

func isInvalidURLError(err error) bool {
	return strings.Contains(err.Error(), "invalid git URL") || errors.Is(err, ErrInvalidURL)
}

func isInvalidNameError(err error) bool {
	return strings.Contains(err.Error(), "invalid repository name") || errors.Is(err, ErrInvalidName)
}

func isNameCollisionError(err error) bool {
	return strings.Contains(err.Error(), "already used by") || errors.Is(err, ErrNameCollision)
}
