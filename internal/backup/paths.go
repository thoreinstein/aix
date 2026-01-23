package backup

import (
	"os"
	"path/filepath"
)

// BackupDir returns the root backup directory for aix.
// Returns ~/.config/aix/backups/
func BackupDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "aix", "backups")
}

// PlatformBackupDir returns the backup directory for a specific platform.
// Returns ~/.config/aix/backups/{platform}/
func PlatformBackupDir(platform string) string {
	root := BackupDir()
	if root == "" {
		return ""
	}
	return filepath.Join(root, platform)
}

// BackupPath returns the full path to a specific backup directory.
// Returns ~/.config/aix/backups/{platform}/{timestamp}/
func BackupPath(platform, timestamp string) string {
	platformDir := PlatformBackupDir(platform)
	if platformDir == "" {
		return ""
	}
	return filepath.Join(platformDir, timestamp)
}
