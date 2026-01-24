package backup

import (
	"os"
	"path/filepath"
	"testing"
)

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

	// We want to simulate two backups in the same second.
	// Since we can't easily force time.Now() to be identical without mocking time,
	// we will run them in a loop until we get a collision or timeout?
	// Actually, running them sequentially is fast enough that they likely happen in the same second.

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
