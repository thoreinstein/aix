package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPermissionFixer_CanFix(t *testing.T) {
	tests := []struct {
		name   string
		issues []pathIssue
		want   bool
	}{
		{
			name:   "no issues",
			issues: nil,
			want:   false,
		},
		{
			name: "non-fixable issue",
			issues: []pathIssue{
				{
					Path:     "/path/to/file",
					Platform: "test",
					Type:     "file",
					Problem:  "cannot stat file",
					Severity: SeverityError,
					Fixable:  false,
				},
			},
			want: false,
		},
		{
			name: "fixable issue",
			issues: []pathIssue{
				{
					Path:     "/path/to/file",
					Platform: "test",
					Type:     "file",
					Problem:  "world-writable",
					Severity: SeverityWarning,
					Fixable:  true,
					FixHint:  "chmod 600",
				},
			},
			want: true,
		},
		{
			name: "mixed issues",
			issues: []pathIssue{
				{
					Path:     "/path/to/file1",
					Fixable:  false,
					Severity: SeverityError,
				},
				{
					Path:     "/path/to/file2",
					Fixable:  true,
					Severity: SeverityWarning,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &PermissionFixer{}
			f.setIssues(tt.issues)
			if got := f.CanFix(); got != tt.want {
				t.Errorf("CanFix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionFixer_CountFixable(t *testing.T) {
	tests := []struct {
		name   string
		issues []pathIssue
		want   int
	}{
		{
			name:   "no issues",
			issues: nil,
			want:   0,
		},
		{
			name: "no fixable issues",
			issues: []pathIssue{
				{Fixable: false},
				{Fixable: false},
			},
			want: 0,
		},
		{
			name: "all fixable",
			issues: []pathIssue{
				{Fixable: true},
				{Fixable: true},
			},
			want: 2,
		},
		{
			name: "mixed",
			issues: []pathIssue{
				{Fixable: false},
				{Fixable: true},
				{Fixable: false},
				{Fixable: true},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &PermissionFixer{}
			f.setIssues(tt.issues)
			if got := f.CountFixable(); got != tt.want {
				t.Errorf("CountFixable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionFixer_Fix_File(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tempDir := t.TempDir()

	// Create a world-writable file
	testFile := filepath.Join(tempDir, "test.json")
	if err := os.WriteFile(testFile, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(testFile, 0666); err != nil {
		t.Fatal(err)
	}

	// Verify it's world-writable
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0002 == 0 {
		t.Fatal("test file should be world-writable before fix")
	}

	// Set up the fixer
	f := &PermissionFixer{}
	f.setIssues([]pathIssue{
		{
			Path:     testFile,
			Platform: "test",
			Type:     "file",
			Problem:  "file is world-writable",
			Severity: SeverityWarning,
			Fixable:  true,
			FixHint:  "chmod 600 " + testFile,
		},
	})

	// Run the fix
	results := f.Fix()

	// Verify results
	if len(results) != 1 {
		t.Fatalf("Fix() returned %d results, want 1", len(results))
	}

	r := results[0]
	if !r.Fixed {
		t.Errorf("Fix() result.Fixed = false, want true")
	}
	if r.Error != nil {
		t.Errorf("Fix() result.Error = %v, want nil", r.Error)
	}
	if r.Path != testFile {
		t.Errorf("Fix() result.Path = %q, want %q", r.Path, testFile)
	}

	// Verify the file permissions were fixed
	info, err = os.Stat(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("File permissions after fix = %04o, want 0600", info.Mode().Perm())
	}
}

func TestPermissionFixer_Fix_Directory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tempDir := t.TempDir()

	// Create a world-writable directory
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(testDir, 0777); err != nil {
		t.Fatal(err)
	}

	// Verify it's world-writable
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0002 == 0 {
		t.Fatal("test directory should be world-writable before fix")
	}

	// Set up the fixer
	f := &PermissionFixer{}
	f.setIssues([]pathIssue{
		{
			Path:     testDir,
			Platform: "test",
			Type:     "directory",
			Problem:  "directory is world-writable",
			Severity: SeverityWarning,
			Fixable:  true,
			FixHint:  "chmod 700 " + testDir,
		},
	})

	// Run the fix
	results := f.Fix()

	// Verify results
	if len(results) != 1 {
		t.Fatalf("Fix() returned %d results, want 1", len(results))
	}

	r := results[0]
	if !r.Fixed {
		t.Errorf("Fix() result.Fixed = false, want true")
	}
	if r.Error != nil {
		t.Errorf("Fix() result.Error = %v, want nil", r.Error)
	}

	// Verify the directory permissions were fixed
	info, err = os.Stat(testDir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("Directory permissions after fix = %04o, want 0700", info.Mode().Perm())
	}
}

func TestPermissionFixer_Fix_NonExistentFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	// Set up the fixer with a non-existent file
	f := &PermissionFixer{}
	f.setIssues([]pathIssue{
		{
			Path:     "/nonexistent/path/that/does/not/exist.json",
			Platform: "test",
			Type:     "file",
			Problem:  "file is world-writable",
			Severity: SeverityWarning,
			Fixable:  true,
			FixHint:  "chmod 600",
		},
	})

	// Run the fix
	results := f.Fix()

	// Verify results
	if len(results) != 1 {
		t.Fatalf("Fix() returned %d results, want 1", len(results))
	}

	r := results[0]
	if r.Fixed {
		t.Error("Fix() result.Fixed = true, want false for non-existent file")
	}
	if r.Error == nil {
		t.Error("Fix() result.Error = nil, want error for non-existent file")
	}
}

func TestPermissionFixer_Fix_SkipsNonFixable(t *testing.T) {
	f := &PermissionFixer{}
	f.setIssues([]pathIssue{
		{
			Path:     "/path/to/file",
			Platform: "test",
			Type:     "file",
			Problem:  "cannot stat file",
			Severity: SeverityError,
			Fixable:  false, // Not fixable
		},
	})

	// Run the fix
	results := f.Fix()

	// Should return no results since nothing is fixable
	if len(results) != 0 {
		t.Errorf("Fix() returned %d results, want 0 (no fixable issues)", len(results))
	}
}

func TestPermissionFixer_Fix_UnknownType(t *testing.T) {
	// Set up the fixer with an unknown type
	f := &PermissionFixer{}
	f.setIssues([]pathIssue{
		{
			Path:     "/some/path",
			Platform: "test",
			Type:     "unknown",
			Problem:  "some problem",
			Severity: SeverityWarning,
			Fixable:  true,
		},
	})

	// Run the fix
	results := f.Fix()

	// Verify results
	if len(results) != 1 {
		t.Fatalf("Fix() returned %d results, want 1", len(results))
	}

	r := results[0]
	if r.Fixed {
		t.Error("Fix() result.Fixed = true, want false for unknown type")
	}
	if r.Error == nil {
		t.Error("Fix() result.Error = nil, want error for unknown type")
	}
}

func TestPermissionFixer_Fix_MultipleIssues(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tempDir := t.TempDir()

	// Create two problematic files
	file1 := filepath.Join(tempDir, "file1.json")
	file2 := filepath.Join(tempDir, "file2.json")

	for _, f := range []string{file1, file2} {
		if err := os.WriteFile(f, []byte("{}"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(f, 0666); err != nil {
			t.Fatal(err)
		}
	}

	// Set up the fixer
	f := &PermissionFixer{}
	f.setIssues([]pathIssue{
		{
			Path:     file1,
			Platform: "test",
			Type:     "file",
			Problem:  "file is world-writable",
			Severity: SeverityWarning,
			Fixable:  true,
		},
		{
			Path:    "/nonexistent/file",
			Type:    "file",
			Fixable: false, // Not fixable - should be skipped
		},
		{
			Path:     file2,
			Platform: "test",
			Type:     "file",
			Problem:  "file is world-writable",
			Severity: SeverityWarning,
			Fixable:  true,
		},
	})

	// Run the fix
	results := f.Fix()

	// Should have 2 results (only fixable issues)
	if len(results) != 2 {
		t.Fatalf("Fix() returned %d results, want 2", len(results))
	}

	// Both should be fixed
	for i, r := range results {
		if !r.Fixed {
			t.Errorf("Fix() result[%d].Fixed = false, want true", i)
		}
	}

	// Verify permissions
	for _, f := range []string{file1, file2} {
		info, err := os.Stat(f)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("File %s permissions = %04o, want 0600", f, info.Mode().Perm())
		}
	}
}

func TestPathPermissionCheck_Fixer_Integration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	// This test verifies that PathPermissionCheck properly implements
	// the Fixer interface and that the fixer is populated after Run()

	c := NewPathPermissionCheck()

	// Before Run(), CanFix should return false
	if c.CanFix() {
		t.Error("CanFix() before Run() = true, want false")
	}

	// Run the check (on the actual system - may or may not find issues)
	_ = c.Run()

	// After Run(), the fixer should be populated
	// We can't assert specific values without controlling the test environment
	// but we can verify the interface works

	// CanFix() should not panic
	_ = c.CanFix()

	// Fix() should not panic (even if there's nothing to fix)
	_ = c.Fix()
}

func TestFixResult_Fields(t *testing.T) {
	r := FixResult{
		Path:        "/path/to/file",
		Fixed:       true,
		Description: "chmod 0600",
		Error:       nil,
	}

	if r.Path != "/path/to/file" {
		t.Errorf("Path = %q, want %q", r.Path, "/path/to/file")
	}
	if !r.Fixed {
		t.Error("Fixed = false, want true")
	}
	if r.Description != "chmod 0600" {
		t.Errorf("Description = %q, want %q", r.Description, "chmod 0600")
	}
	if r.Error != nil {
		t.Errorf("Error = %v, want nil", r.Error)
	}
}
