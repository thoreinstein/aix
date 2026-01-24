package backup

import (
	"sync"

	"github.com/thoreinstein/aix/internal/errors"
)

// backupOnce tracks per-platform backup state within a session.
// This prevents redundant backups when multiple operations occur.
var (
	backupOnce  = make(map[string]*sync.Once)
	backupMutex sync.Mutex
)

// EnsureBackedUp ensures a backup exists for the platform before modification.
// Uses sync.Once pattern to ensure only one backup is created per platform per session.
//
// The function is safe for concurrent calls and will only create one backup
// per platform regardless of how many times it's called.
//
// Returns nil if:
//   - A backup was just created successfully
//   - A backup was already created in this session (no-op)
//   - No paths are provided (nothing to back up)
//
// Returns an error if:
//   - The backup creation fails
func EnsureBackedUp(platformName string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	backupMutex.Lock()
	once, exists := backupOnce[platformName]
	if !exists {
		once = &sync.Once{}
		backupOnce[platformName] = once
	}
	backupMutex.Unlock()

	var backupErr error
	once.Do(func() {
		mgr := NewManager()
		_, backupErr = mgr.Backup(platformName, paths)
		if backupErr != nil {
			// If backup fails, we don't want to proceed with modifications
			// Reset the Once so caller can retry
			backupMutex.Lock()
			delete(backupOnce, platformName)
			backupMutex.Unlock()
		}
	})

	if backupErr != nil {
		return errors.Wrapf(backupErr, "creating backup for %s", platformName)
	}

	return nil
}

// ResetBackupState clears the backup state for all platforms.
// This is primarily useful for testing to reset state between tests.
func ResetBackupState() {
	backupMutex.Lock()
	defer backupMutex.Unlock()
	backupOnce = make(map[string]*sync.Once)
}

// ResetPlatformBackupState clears the backup state for a specific platform.
// This allows a new backup to be created for the platform on the next call
// to EnsureBackedUp.
func ResetPlatformBackupState(platformName string) {
	backupMutex.Lock()
	defer backupMutex.Unlock()
	delete(backupOnce, platformName)
}
