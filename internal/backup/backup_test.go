package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thoreinstein/aix/internal/errors"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	backupDir := t.TempDir()
	m := NewManager(WithBackupDir(backupDir), WithRetentionCount(5))

	if m.rootDir != backupDir {
		t.Errorf("expected rootDir %q, got %q", backupDir, m.rootDir)
	}
	if m.retentionCount != 5 {
		t.Errorf("expected retentionCount 5, got %d", m.retentionCount)
	}
}

func TestManager_BackupAndRestore(t *testing.T) {
	t.Parallel()

	backupDir := t.TempDir()
	srcDir := t.TempDir()
	m := NewManager(WithBackupDir(backupDir))

	// Create test files
	file1 := filepath.Join(srcDir, "file1.txt")
	content1 := []byte("content 1")
	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(srcDir, "subdir", "file2.txt")
	content2 := []byte("content 2")
	if err := os.MkdirAll(filepath.Dir(file2), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, content2, 0600); err != nil {
		t.Fatal(err)
	}

	platform := "test-platform"
	paths := []string{file1, filepath.Dir(file2)}

	// 1. Test Backup
	manifest, err := m.Backup(platform, paths)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	if manifest.Platform != platform {
		t.Errorf("manifest platform = %q, want %q", manifest.Platform, platform)
	}
	if len(manifest.Files) != 2 {
		t.Errorf("expected 2 files in manifest, got %d", len(manifest.Files))
	}

	// Verify backup files exist in backup directory
	backupPath := filepath.Join(backupDir, platform, manifest.ID)
	for _, bf := range manifest.Files {
		p := filepath.Join(backupPath, bf.RelPath)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("backup file missing: %s", p)
		}
	}

	// 2. Modify original files to test restore
	if err := os.WriteFile(file1, []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(file2); err != nil {
		t.Fatal(err)
	}

	// 3. Test Restore
	err = m.Restore(platform, manifest.ID)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify file1 content restored
	restoredContent1, err := os.ReadFile(file1)
	if err != nil {
		t.Fatal(err)
	}
	if string(restoredContent1) != string(content1) {
		t.Errorf("file1 content = %q, want %q", restoredContent1, content1)
	}

	// Verify file2 exists and content restored
	restoredContent2, err := os.ReadFile(file2)
	if err != nil {
		t.Fatal(err)
	}
	if string(restoredContent2) != string(content2) {
		t.Errorf("file2 content = %q, want %q", restoredContent2, content2)
	}

	// Verify permissions restored
	info2, err := os.Stat(file2)
	if err != nil {
		t.Fatal(err)
	}
	if info2.Mode().Perm() != 0600 {
		t.Errorf("file2 perm = %o, want 0600", info2.Mode().Perm())
	}
}

func TestManager_List(t *testing.T) {
	t.Parallel()

	backupDir := t.TempDir()
	srcDir := t.TempDir()
	m := NewManager(WithBackupDir(backupDir))

	file := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(file, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	platform := "list-test"

	// Create 3 backups with small delays to ensure different timestamps
	for range 3 {
		_, err := m.Backup(platform, []string{file})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure unique timestamps if needed, though ID has random suffix
	}

	backups, err := m.List(platform)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("expected 3 backups, got %d", len(backups))
	}

	// Verify sorting (newest first)
	for i := range len(backups) - 1 {
		if backups[i].CreatedAt.Before(backups[i+1].CreatedAt) {
			t.Errorf("backups not sorted newest first at index %d", i)
		}
	}
}

func TestManager_Prune(t *testing.T) {
	t.Parallel()

	backupDir := t.TempDir()
	srcDir := t.TempDir()
	m := NewManager(WithBackupDir(backupDir))

	file := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(file, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	platform := "prune-test"

	// Create 5 backups
	for range 5 {
		_, err := m.Backup(platform, []string{file})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Keep only 2
	err := m.Prune(platform, 2)
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}

	backups, err := m.List(platform)
	if err != nil {
		t.Fatal(err)
	}

	if len(backups) != 2 {
		t.Errorf("expected 2 backups after prune, got %d", len(backups))
	}
}

func TestManager_Get_Errors(t *testing.T) {
	t.Parallel()

	backupDir := t.TempDir()
	m := NewManager(WithBackupDir(backupDir))

	_, err := m.Get("nonexistent", "id")
	if !errors.Is(err, ErrNoBackupsFound) {
		t.Errorf("expected ErrNoBackupsFound, got %v", err)
	}

	err = m.Restore("nonexistent", "id")
	if !errors.Is(err, ErrNoBackupsFound) {
		t.Errorf("expected ErrNoBackupsFound, got %v", err)
	}
}

func TestBackup_Collision(t *testing.T) {
	// Create a temporary directory for backups
	backupDir := t.TempDir()

	// Create a dummy file to backup
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("test content"), 0600); err != nil {
		t.Fatal(err)
	}

	m := NewManager(WithBackupDir(backupDir))

	// Running them sequentially should yield different IDs due to random suffix
	manifest1, err := m.Backup("platform1", []string{srcFile})
	if err != nil {
		t.Fatalf("First backup failed: %v", err)
	}

	manifest2, err := m.Backup("platform1", []string{srcFile})
	if err != nil {
		t.Fatalf("Second backup failed: %v", err)
	}

	if manifest1.ID == manifest2.ID {
		t.Errorf("Backup IDs collided: %s", manifest1.ID)
	}
}

func TestGenerateRelPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/usr/local/bin", "usr/local/bin"},
		{"C:\\Users\\Data", "C\\Users\\Data"}, // Windows style with : removed
		{"file:name", "filename"},             // Arbitrary : removal
	}

	for _, tt := range tests {
		got := generateRelPath(tt.input)

		// The core requirement: NO COLONS.
		for i := range len(got) {
			if got[i] == ':' {
				t.Errorf("generateRelPath(%q) = %q contains colon", tt.input, got)
			}
		}
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("skipping test; user home dir not available")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"/abs/path", "/abs/path"}, // Not expanded if not starting with ~
		{"~", home},                // Just tilde
		{"~/file.txt", filepath.Join(home, "file.txt")},         // Normal expansion
		{"~user/file.txt", "~user/file.txt"},                    // Not supported, returns as is
		{"~/../../../etc/passwd", "~/../../../etc/passwd"},      // Path traversal blocked
		{"~/subdir/../file.txt", "~/subdir/../file.txt"},        // Path traversal blocked
		{"~/..hidden", filepath.Join(home, "..hidden")},         // Filename starting with .. is OK
		{"~/.hidden/file", filepath.Join(home, ".hidden/file")}, // Hidden directories are OK
		{"~/..", "~/.."},                         // ".." as path component is blocked
		{"~/foo/bar/../baz", "~/foo/bar/../baz"}, // ".." in middle is blocked
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandHome(tt.input)
			if got != tt.expected {
				t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
