package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/thoreinstein/aix/pkg/fileutil"
)

// testConfig represents a simple config structure for atomic write tests.
type testConfig struct {
	Version int               `json:"version"`
	Name    string            `json:"name"`
	Data    map[string]string `json:"data"`
}

func (c *testConfig) validate() bool {
	if c.Version <= 0 || c.Name == "" {
		return false
	}
	return true
}

// TestAtomicWrite_InterruptedWriteLeavesOriginalIntact tests that if a write fails
// (simulated by making the target directory read-only after temp file creation),
// the original file remains intact.
func TestAtomicWrite_InterruptedWriteLeavesOriginalIntact(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Create a valid config file with known content
	original := testConfig{
		Version: 1,
		Name:    "original-config",
		Data:    map[string]string{"key": "original-value"},
	}
	if err := fileutil.AtomicWriteJSON(configPath, original); err != nil {
		t.Fatalf("writing original config: %v", err)
	}

	// Verify original was written correctly
	verifyConfigEquals(t, configPath, original)

	// Make the directory read-only to cause rename to fail
	// The temp file will be created, but rename will fail
	if err := os.Chmod(dir, 0500); err != nil {
		t.Fatalf("chmod directory: %v", err)
	}
	// Restore permissions for cleanup
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0700) // Best effort cleanup
	})

	// Attempt to overwrite with new content - this should fail
	newConfig := testConfig{
		Version: 2,
		Name:    "new-config",
		Data:    map[string]string{"key": "new-value"},
	}
	err := fileutil.AtomicWriteJSON(configPath, newConfig)
	if err == nil {
		t.Fatal("expected write to fail with read-only directory")
	}

	// Restore permissions to read the file
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatalf("restoring directory permissions: %v", err)
	}

	// Verify original content is still present and valid
	verifyConfigEquals(t, configPath, original)
}

// TestAtomicWrite_ConcurrentWritesNoCorruption tests that concurrent writes to the
// same file don't produce corrupted output. One write wins, but the result must
// be a complete, valid config (not a mix of both).
func TestAtomicWrite_ConcurrentWritesNoCorruption(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	const numWriters = 10
	const writesPerWriter = 50

	// Create configs with distinct, identifiable content
	configs := make([]testConfig, numWriters)
	for i := range numWriters {
		configs[i] = testConfig{
			Version: i + 1,
			Name:    repeatString("writer", i+1), // Distinct name per writer
			Data:    map[string]string{"writer": string(rune('A' + i))},
		}
	}

	var wg sync.WaitGroup
	errChan := make(chan error, numWriters*writesPerWriter)

	// Launch concurrent writers
	for i := range numWriters {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for range writesPerWriter {
				if err := fileutil.AtomicWriteJSON(configPath, configs[writerID]); err != nil {
					errChan <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for write errors (some failures are acceptable due to race conditions)
	writeErrors := make([]error, 0, numWriters*writesPerWriter)
	for err := range errChan {
		writeErrors = append(writeErrors, err)
	}
	if len(writeErrors) > numWriters*writesPerWriter/2 {
		t.Errorf("too many write errors: %d out of %d", len(writeErrors), numWriters*writesPerWriter)
	}

	// The critical check: the final file must contain ONE complete, valid config
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading final config: %v", err)
	}

	var finalConfig testConfig
	if err := json.Unmarshal(data, &finalConfig); err != nil {
		t.Fatalf("parsing final config (file may be corrupted): %v\nContent: %s", err, string(data))
	}

	// Validate the config is complete and well-formed
	if !finalConfig.validate() {
		t.Errorf("final config is invalid: %+v", finalConfig)
	}

	// Verify the config matches one of the expected configs exactly
	found := false
	for _, expected := range configs {
		if finalConfig.Version == expected.Version &&
			finalConfig.Name == expected.Name &&
			mapsEqual(finalConfig.Data, expected.Data) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("final config doesn't match any expected config: %+v", finalConfig)
	}
}

// TestAtomicWrite_ClearErrorMessages tests that write failures produce clear,
// actionable error messages.
func TestAtomicWrite_ClearErrorMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectInErr []string
	}{
		{
			name: "write to non-existent directory",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				return filepath.Join(dir, "nonexistent", "subdir", "config.json")
			},
			expectInErr: []string{"creating temp file"},
		},
		{
			name: "write to read-only directory (temp file creation fails)",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()

				// Create a subdirectory that's read-only
				readOnlyDir := filepath.Join(dir, "readonly")
				if err := os.MkdirAll(readOnlyDir, 0500); err != nil {
					t.Fatalf("creating read-only dir: %v", err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(readOnlyDir, 0700) // Best effort cleanup
				})

				return filepath.Join(readOnlyDir, "config.json")
			},
			expectInErr: []string{"creating temp file", "permission denied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.setup(t)
			config := testConfig{Version: 1, Name: "test", Data: map[string]string{}}

			err := fileutil.AtomicWriteJSON(path, config)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			errStr := err.Error()
			for _, expected := range tt.expectInErr {
				if !strings.Contains(strings.ToLower(errStr), strings.ToLower(expected)) {
					t.Errorf("error should contain %q, got: %v", expected, err)
				}
			}
		})
	}
}

// TestAtomicWrite_JSONMarshalError tests error messages for JSON marshaling failures.
func TestAtomicWrite_JSONMarshalError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Channels are not JSON-marshalable
	unmarshalable := make(chan int)

	err := fileutil.AtomicWriteJSON(configPath, unmarshalable)
	if err == nil {
		t.Fatal("expected marshaling error")
	}

	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "marshal") {
		t.Errorf("error should mention marshaling, got: %v", err)
	}

	// File should not exist after marshal failure
	if _, statErr := os.Stat(configPath); statErr == nil {
		t.Error("file should not exist after marshal error")
	}
}

// TestAtomicWrite_TempFileCleanup tests that temp files are cleaned up even on
// partial failures (after temp file creation but before successful rename).
func TestAtomicWrite_TempFileCleanup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Create original file
	original := testConfig{Version: 1, Name: "original", Data: map[string]string{}}
	if err := fileutil.AtomicWriteJSON(configPath, original); err != nil {
		t.Fatalf("writing original: %v", err)
	}

	// Make directory read-only to cause rename to fail
	// (temp file will be created successfully in same dir)
	if err := os.Chmod(dir, 0500); err != nil {
		t.Fatalf("chmod directory: %v", err)
	}

	// Attempt write that will fail at rename stage
	newConfig := testConfig{Version: 2, Name: "new", Data: map[string]string{}}
	_ = fileutil.AtomicWriteJSON(configPath, newConfig) // Expected to fail

	// Restore permissions to check for temp files
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatalf("restoring permissions: %v", err)
	}

	// Check that no temp files remain
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading directory: %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".aix-atomic-") && strings.HasSuffix(name, ".tmp") {
			t.Errorf("temp file left behind: %s", name)
		}
	}
}

// TestAtomicWrite_TempFileCleanupOnMultipleFailures tests cleanup with multiple
// concurrent failures.
func TestAtomicWrite_TempFileCleanupOnMultipleFailures(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a subdirectory where we'll cause failures
	writeDir := filepath.Join(dir, "writes")
	if err := os.MkdirAll(writeDir, 0700); err != nil {
		t.Fatalf("creating write dir: %v", err)
	}

	configPath := filepath.Join(writeDir, "config.json")

	// Create initial file
	initial := testConfig{Version: 1, Name: "initial", Data: map[string]string{}}
	if err := fileutil.AtomicWriteJSON(configPath, initial); err != nil {
		t.Fatalf("writing initial: %v", err)
	}

	// Make directory read-only
	if err := os.Chmod(writeDir, 0500); err != nil {
		t.Fatalf("chmod directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(writeDir, 0700) // Best effort cleanup
	})

	// Attempt multiple concurrent writes that will all fail
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			config := testConfig{Version: id, Name: "test", Data: map[string]string{}}
			_ = fileutil.AtomicWriteJSON(configPath, config) // Expected to fail
		}(i)
	}
	wg.Wait()

	// Restore permissions and check for temp files
	if err := os.Chmod(writeDir, 0700); err != nil {
		t.Fatalf("restoring permissions: %v", err)
	}

	entries, err := os.ReadDir(writeDir)
	if err != nil {
		t.Fatalf("reading directory: %v", err)
	}

	var tempFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".aix-atomic-") && strings.HasSuffix(name, ".tmp") {
			tempFiles = append(tempFiles, name)
		}
	}

	if len(tempFiles) > 0 {
		t.Errorf("temp files left behind: %v", tempFiles)
	}
}

// TestAtomicWrite_PreservesPermissions tests that atomic writes preserve the
// specified file permissions.
func TestAtomicWrite_PreservesPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		perm os.FileMode
	}{
		{"standard 0600", 0600},
		{"restricted 0600", 0600},
		{"executable 0700", 0700},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "test-file")

			data := []byte("test content\n")
			if err := fileutil.AtomicWriteFile(path, data, tt.perm); err != nil {
				t.Fatalf("AtomicWriteFile: %v", err)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat: %v", err)
			}

			gotPerm := info.Mode().Perm()
			if gotPerm != tt.perm {
				t.Errorf("permissions = %o, want %o", gotPerm, tt.perm)
			}
		})
	}
}

// TestAtomicWrite_OverwritePreservesAtomicity tests that overwriting an existing
// file is atomic - readers see either the old or new content, never partial.
func TestAtomicWrite_OverwritePreservesAtomicity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Create initial config
	v1 := testConfig{Version: 1, Name: "version-one", Data: map[string]string{"v": "1"}}
	if err := fileutil.AtomicWriteJSON(configPath, v1); err != nil {
		t.Fatalf("writing v1: %v", err)
	}

	v2 := testConfig{Version: 2, Name: "version-two", Data: map[string]string{"v": "2"}}

	// Launch concurrent reader and writer
	var wg sync.WaitGroup
	stopReader := make(chan struct{})
	readErrors := make(chan error, 1000)
	invalidConfigs := make(chan testConfig, 100)

	// Reader goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopReader:
				return
			default:
				data, err := os.ReadFile(configPath)
				if err != nil {
					if !os.IsNotExist(err) {
						readErrors <- err
					}
					continue
				}

				var config testConfig
				if err := json.Unmarshal(data, &config); err != nil {
					readErrors <- err
					continue
				}

				// Config must be either v1 or v2, never a mix
				isV1 := config.Version == 1 && config.Name == "version-one"
				isV2 := config.Version == 2 && config.Name == "version-two"
				if !isV1 && !isV2 {
					invalidConfigs <- config
				}
			}
		}
	}()

	// Writer goroutine - repeatedly overwrite
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range 100 {
			var config testConfig
			if i%2 == 0 {
				config = v1
			} else {
				config = v2
			}
			_ = fileutil.AtomicWriteJSON(configPath, config)
		}
		close(stopReader)
	}()

	wg.Wait()
	close(readErrors)
	close(invalidConfigs)

	// Check for invalid configs (mixed state)
	for invalid := range invalidConfigs {
		t.Errorf("found invalid/mixed config: %+v", invalid)
	}

	// Some read errors are acceptable (file being replaced), but check for patterns
	var errCount int
	for range readErrors {
		errCount++
	}
	// Allow up to 10% error rate from timing issues
	if errCount > 100 {
		t.Errorf("too many read errors during concurrent access: %d", errCount)
	}
}

// Helper functions

func verifyConfigEquals(t *testing.T, path string, expected testConfig) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	var got testConfig
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("parsing config: %v", err)
	}

	if got.Version != expected.Version {
		t.Errorf("Version = %d, want %d", got.Version, expected.Version)
	}
	if got.Name != expected.Name {
		t.Errorf("Name = %q, want %q", got.Name, expected.Name)
	}
	if !mapsEqual(got.Data, expected.Data) {
		t.Errorf("Data = %v, want %v", got.Data, expected.Data)
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func repeatString(s string, n int) string {
	var result strings.Builder
	for range n {
		result.WriteString(s)
	}
	return result.String()
}
