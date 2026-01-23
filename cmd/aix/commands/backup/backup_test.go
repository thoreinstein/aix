package backup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thoreinstein/aix/cmd/aix/commands/flags"
	"github.com/thoreinstein/aix/internal/backup"
)

func TestBackupList_WithBackups_JSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{"test": true}`), 0o644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	// Create a backup using WithBackupDir
	mgr := backup.NewManager(backup.WithBackupDir(tmpDir))
	_, err := mgr.Backup("claude", []string{testFile})
	if err != nil {
		t.Fatalf("creating backup: %v", err)
	}

	// List backups using the manager directly
	manifests, err := mgr.List("claude")
	if err != nil {
		t.Fatalf("listing backups: %v", err)
	}

	if len(manifests) != 1 {
		t.Errorf("expected 1 backup, got %d", len(manifests))
	}

	// Verify the backup has expected fields
	m := manifests[0]
	if m.ID == "" {
		t.Error("backup ID should not be empty")
	}
	if m.Platform != "claude" {
		t.Errorf("expected platform 'claude', got %q", m.Platform)
	}
	if len(m.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(m.Files))
	}
}

func TestBackupPrune_KeepsCorrectCount(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	// Create multiple backups with different timestamps
	mgr := backup.NewManager(backup.WithBackupDir(tmpDir))
	for i := range 3 {
		_, err := mgr.Backup("claude", []string{testFile})
		if err != nil {
			t.Fatalf("creating backup %d: %v", i, err)
		}
		time.Sleep(time.Second) // Ensure unique timestamps
	}

	// Verify we have 3 backups
	manifests, err := mgr.List("claude")
	if err != nil {
		t.Fatalf("listing backups: %v", err)
	}
	if len(manifests) != 3 {
		t.Fatalf("expected 3 backups, got %d", len(manifests))
	}

	// Prune to keep only 1
	if err := mgr.Prune("claude", 1); err != nil {
		t.Fatalf("pruning: %v", err)
	}

	// Verify only 1 backup remains
	manifests, err = mgr.List("claude")
	if err != nil {
		t.Fatalf("listing backups after prune: %v", err)
	}
	if len(manifests) != 1 {
		t.Errorf("expected 1 backup after prune, got %d", len(manifests))
	}
}

func TestBackupRestore_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file to backup
	testFile := filepath.Join(tmpDir, "config.json")
	originalContent := []byte(`{"original": true}`)
	if err := os.WriteFile(testFile, originalContent, 0o644); err != nil {
		t.Fatalf("creating test file: %v", err)
	}

	// Create a backup
	mgr := backup.NewManager(backup.WithBackupDir(tmpDir))
	manifest, err := mgr.Backup("claude", []string{testFile})
	if err != nil {
		t.Fatalf("creating backup: %v", err)
	}

	// Modify the original file
	if err := os.WriteFile(testFile, []byte(`{"modified": true}`), 0o644); err != nil {
		t.Fatalf("modifying test file: %v", err)
	}

	// Restore the backup
	if err := mgr.Restore("claude", manifest.ID); err != nil {
		t.Fatalf("restoring backup: %v", err)
	}

	// Verify file was restored
	restored, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("reading restored file: %v", err)
	}
	if !bytes.Equal(restored, originalContent) {
		t.Errorf("expected restored content %q, got %q", originalContent, restored)
	}
}

func TestBackupRestore_RequiresPlatform(t *testing.T) {
	origPlatformFlag := flags.GetPlatformFlag()
	defer flags.SetPlatformFlag(origPlatformFlag)
	flags.SetPlatformFlag(nil) // No platform specified

	var buf bytes.Buffer
	err := runRestoreWithWriter(nil, nil, &buf)
	if err == nil {
		t.Error("expected error when no platform specified")
	}
	if err.Error() != "--platform is required for restore" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBackupListOutput_JSON(t *testing.T) {
	// Test JSON output structure
	output := []listOutput{
		{
			Platform: "claude",
			Backups: []infoOutput{
				{
					ID:         "20260123T100712",
					CreatedAt:  time.Now(),
					FileCount:  5,
					AIXVersion: "1.0.0",
				},
			},
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshaling output: %v", err)
	}

	// Verify it's valid JSON
	var parsed []listOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}

	if len(parsed) != 1 || parsed[0].Platform != "claude" {
		t.Errorf("unexpected parsed output: %+v", parsed)
	}
}

func TestBackupPrune_NegativeKeep(t *testing.T) {
	origKeep := pruneKeep
	defer func() { pruneKeep = origKeep }()
	pruneKeep = -1

	var buf bytes.Buffer
	err := runPruneWithWriter(&buf)
	if err == nil {
		t.Error("expected error for negative keep value")
	}
	if err.Error() != "--keep must be non-negative" {
		t.Errorf("unexpected error: %v", err)
	}
}
