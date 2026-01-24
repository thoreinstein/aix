package repo

import (
	"errors"
	"path/filepath"
	"testing"

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

func TestManager_Add_InvalidURL(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	// Try to add with invalid URL (local path)
	_, err := m.Add("/home/user/repo")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}

	// Verify error type
	if !isInvalidURLError(err) {
		t.Errorf("expected ErrInvalidURL, got: %v", err)
	}
}

func TestManager_Add_InvalidName(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	// Try to add with invalid name override
	_, err := m.Add("https://github.com/user/repo.git", WithName("Invalid_Name"))
	if err == nil {
		t.Error("expected error for invalid name, got nil")
	}

	// Verify error type
	if !isInvalidNameError(err) {
		t.Errorf("expected ErrInvalidName, got: %v", err)
	}
}

func TestManager_List_Empty(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	// List should return empty slice when no repos exist
	repos, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if repos == nil {
		t.Error("List() returned nil, want empty slice")
	}

	if len(repos) != 0 {
		t.Errorf("List() returned %d repos, want 0", len(repos))
	}
}

func TestManager_Get_NotFound(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	// Get should return ErrNotFound for non-existent repo
	_, err := m.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent repo, got nil")
	}

	if !isNotFoundError(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestManager_Remove_NotFound(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	// Remove should return ErrNotFound for non-existent repo
	err := m.Remove("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent repo, got nil")
	}

	if !isNotFoundError(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestManager_Update_NotFound(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	// Update with specific name should return ErrNotFound
	err := m.Update("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent repo, got nil")
	}

	if !isNotFoundError(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestManager_Update_EmptyNoError(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	m := NewManager(configPath)

	// Update all with no repos should not error
	err := m.Update("")
	if err != nil {
		t.Errorf("Update(\"\") with no repos should not error, got: %v", err)
	}
}

func TestNewManager(t *testing.T) {
	configPath := "/path/to/config.yaml"
	m := NewManager(configPath)

	if m == nil {
		t.Fatal("NewManager returned nil")
	}

	if m.configPath != configPath {
		t.Errorf("configPath = %q, want %q", m.configPath, configPath)
	}
}

func TestWithName(t *testing.T) {
	var opts addOptions
	WithName("custom-name")(&opts)

	if opts.name != "custom-name" {
		t.Errorf("WithName() set name = %q, want %q", opts.name, "custom-name")
	}
}

// Integration-style test that requires real filesystem but not git
func TestManager_SaveLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "aix", "config.yaml")

	m := NewManager(configPath)

	// Initially, List should return empty
	repos, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("initial List() = %d repos, want 0", len(repos))
	}
}

// Helper functions for error type checking
func isNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func isInvalidURLError(err error) bool {
	return errors.Is(err, ErrInvalidURL)
}

func isInvalidNameError(err error) bool {
	return errors.Is(err, ErrInvalidName)
}
