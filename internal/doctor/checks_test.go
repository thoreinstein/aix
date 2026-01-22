package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPathPermissionCheck_Name(t *testing.T) {
	c := NewPathPermissionCheck()
	if got := c.Name(); got != "path-permissions" {
		t.Errorf("Name() = %q, want %q", got, "path-permissions")
	}
}

func TestPathPermissionCheck_Category(t *testing.T) {
	c := NewPathPermissionCheck()
	if got := c.Category(); got != "filesystem" {
		t.Errorf("Category() = %q, want %q", got, "filesystem")
	}
}

func TestPathPermissionCheck_checkFile(t *testing.T) {
	c := NewPathPermissionCheck()
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		setup        func() string
		wantIssues   int
		wantSeverity Severity
	}{
		{
			name: "readable file",
			setup: func() string {
				path := filepath.Join(tempDir, "readable.json")
				if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantIssues: 0,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent.json")
			},
			wantIssues: 0, // Non-existent files are OK (platform not configured)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			issues := c.checkFile(path, "test-platform")
			if len(issues) != tt.wantIssues {
				t.Errorf("checkFile() returned %d issues, want %d", len(issues), tt.wantIssues)
			}
		})
	}
}

func TestPathPermissionCheck_checkDirectory(t *testing.T) {
	c := NewPathPermissionCheck()
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		setup      func() string
		wantIssues int
	}{
		{
			name: "writable directory",
			setup: func() string {
				dir := filepath.Join(tempDir, "writable")
				if err := os.Mkdir(dir, 0755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantIssues: 0,
		},
		{
			name: "non-existent directory",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			wantIssues: 0, // Non-existent dirs are OK (platform not installed)
		},
		{
			name: "file where directory expected",
			setup: func() string {
				path := filepath.Join(tempDir, "not-a-dir")
				if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			issues := c.checkDirectory(path, "test-platform")
			if len(issues) != tt.wantIssues {
				t.Errorf("checkDirectory() returned %d issues, want %d", len(issues), tt.wantIssues)
			}
		})
	}
}

func TestPathPermissionCheck_checkFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	c := NewPathPermissionCheck()
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		mode       os.FileMode
		filename   string
		wantIssues int
	}{
		{
			name:       "secure permissions 0644",
			mode:       0644,
			filename:   "config.json",
			wantIssues: 0,
		},
		{
			name:       "world-writable file",
			mode:       0666,
			filename:   "world-writable.json",
			wantIssues: 1, // World-writable is a security concern
		},
		{
			name:       "fully permissive",
			mode:       0777,
			filename:   "fully-permissive.json",
			wantIssues: 1, // World-writable + overly permissive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tempDir, tt.filename)
			// Create file with restrictive mode first, then chmod to desired mode
			// This avoids umask interference
			if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, tt.mode); err != nil {
				t.Fatal(err)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}

			issues := c.checkFilePermissions(path, "test-platform", info.Mode())
			if len(issues) < tt.wantIssues {
				t.Errorf("checkFilePermissions() returned %d issues, want at least %d", len(issues), tt.wantIssues)
			}
		})
	}
}

func TestPathPermissionCheck_checkDirectoryPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	c := NewPathPermissionCheck()
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		mode       os.FileMode
		wantIssues int
	}{
		{
			name:       "secure permissions 0755",
			mode:       0755,
			wantIssues: 0,
		},
		{
			name:       "world-writable directory",
			mode:       0777,
			wantIssues: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(tempDir, tt.name)
			// Create directory with restrictive mode first, then chmod to desired mode
			// This avoids umask interference
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(dir, tt.mode); err != nil {
				t.Fatal(err)
			}

			info, err := os.Stat(dir)
			if err != nil {
				t.Fatal(err)
			}

			issues := c.checkDirectoryPermissions(dir, "test-platform", info.Mode())
			if len(issues) != tt.wantIssues {
				t.Errorf("checkDirectoryPermissions() returned %d issues, want %d (mode=%o)", len(issues), tt.wantIssues, info.Mode().Perm())
			}
		})
	}
}

func TestPathPermissionCheck_isDirectoryWritable(t *testing.T) {
	c := NewPathPermissionCheck()
	tempDir := t.TempDir()

	t.Run("writable directory", func(t *testing.T) {
		writable, err := c.isDirectoryWritable(tempDir)
		if err != nil {
			t.Errorf("isDirectoryWritable() error = %v", err)
		}
		if !writable {
			t.Error("isDirectoryWritable() = false, want true")
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := c.isDirectoryWritable("/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Error("isDirectoryWritable() expected error for non-existent directory")
		}
	})
}

func TestPathPermissionCheck_mayContainSecrets(t *testing.T) {
	c := NewPathPermissionCheck()

	tests := []struct {
		path string
		want bool
	}{
		{"/path/to/config.json", true},
		{"/path/to/mcp.json", true},
		{"/path/to/claude.json", true},
		{"/path/to/opencode.json", true},
		{"/path/to/settings.toml", true},
		{"/path/to/readme.md", false},
		{"/path/to/random.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := c.mayContainSecrets(tt.path)
			if got != tt.want {
				t.Errorf("mayContainSecrets(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestPathPermissionCheck_buildResult(t *testing.T) {
	c := NewPathPermissionCheck()

	t.Run("no issues", func(t *testing.T) {
		result := c.buildResult(nil, 5)
		if result.Status != SeverityPass {
			t.Errorf("buildResult() status = %v, want %v", result.Status, SeverityPass)
		}
		if result.Name != "path-permissions" {
			t.Errorf("buildResult() name = %q, want %q", result.Name, "path-permissions")
		}
		if result.Category != "filesystem" {
			t.Errorf("buildResult() category = %q, want %q", result.Category, "filesystem")
		}
	})

	t.Run("with warnings", func(t *testing.T) {
		issues := []pathIssue{
			{
				Path:     "/path/to/file",
				Platform: "test",
				Type:     "file",
				Problem:  "test problem",
				Severity: SeverityWarning,
			},
		}
		result := c.buildResult(issues, 5)
		if result.Status != SeverityWarning {
			t.Errorf("buildResult() status = %v, want %v", result.Status, SeverityWarning)
		}
		if result.Details == nil {
			t.Error("buildResult() details is nil")
		}
	})

	t.Run("with errors", func(t *testing.T) {
		issues := []pathIssue{
			{
				Path:     "/path/to/file",
				Platform: "test",
				Type:     "file",
				Problem:  "warning problem",
				Severity: SeverityWarning,
			},
			{
				Path:     "/path/to/other",
				Platform: "test",
				Type:     "file",
				Problem:  "error problem",
				Severity: SeverityError,
			},
		}
		result := c.buildResult(issues, 5)
		if result.Status != SeverityError {
			t.Errorf("buildResult() status = %v, want %v (error takes precedence)", result.Status, SeverityError)
		}
	})

	t.Run("with fixable issues", func(t *testing.T) {
		issues := []pathIssue{
			{
				Path:     "/path/to/file",
				Platform: "test",
				Type:     "file",
				Problem:  "fixable problem",
				Severity: SeverityWarning,
				Fixable:  true,
				FixHint:  "chmod 644 /path/to/file",
			},
		}
		result := c.buildResult(issues, 5)
		if !result.Fixable {
			t.Error("buildResult() Fixable = false, want true")
		}
		if result.FixHint == "" {
			t.Error("buildResult() FixHint is empty")
		}
	})
}

func TestFormatPermissions(t *testing.T) {
	tests := []struct {
		mode os.FileMode
		want string
	}{
		{0644, "0644"},
		{0755, "0755"},
		{0600, "0600"},
		{0777, "0777"},
		{0000, "0000"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatPermissions(tt.mode)
			if got != tt.want {
				t.Errorf("formatPermissions(%o) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}

func TestFormatOctal(t *testing.T) {
	tests := []struct {
		mode os.FileMode
		want string
	}{
		{0644, "0644"},
		{0755, "0755"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatOctal(tt.mode)
			if got != tt.want {
				t.Errorf("formatOctal(%o) = %q, want %q", tt.mode, got, tt.want)
			}
		})
	}
}
