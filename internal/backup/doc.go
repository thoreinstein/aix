// Package backup provides configuration backup and restore capabilities for aix.
//
// This package implements a backup strategy for AI assistant platform configurations,
// allowing users to safely back up, restore, and manage configuration snapshots
// before making changes.
//
// # Backup Strategy
//
// The backup system uses a directory-based approach where each backup is stored
// in a timestamped directory containing:
//
//   - manifest.json: Metadata about the backup including file hashes for integrity
//   - Copied files: Original configuration files with preserved permissions
//
// Backup locations follow this hierarchy:
//
//	~/.config/aix/backups/
//	└── {platform}/
//	    └── {timestamp}/
//	        ├── manifest.json
//	        └── {copied files...}
//
// # Creating Backups
//
// Use [Manager.Backup] to create a new backup of platform configuration files:
//
//	mgr := backup.NewManager()
//	manifest, err := mgr.Backup("claude", []string{
//	    "~/.claude/config.json",
//	    "~/.claude/agents/",
//	})
//
// The backup captures file contents, permissions, and generates SHA256 checksums
// for integrity verification during restore.
//
// # Restoring Backups
//
// Use [Manager.Restore] to restore a previous backup:
//
//	err := mgr.Restore("claude", "20260123T100712")
//
// Before restoring, the current configuration is backed up automatically to prevent
// data loss. The restore operation verifies file integrity using stored checksums.
//
// # Retention Management
//
// The [Manager.Prune] method removes old backups beyond the configured retention count:
//
//	err := mgr.Prune("claude", 5) // Keep 5 most recent backups
//
// The default retention count is 5 backups per platform.
//
// # Listing Backups
//
// Use [Manager.List] to retrieve available backups sorted by date (newest first):
//
//	manifests, err := mgr.List("claude")
//	for _, m := range manifests {
//	    fmt.Printf("%s: %d files\n", m.CreatedAt.Format(time.RFC3339), len(m.Files))
//	}
//
// # Backup Manifest
//
// Each backup includes a [BackupManifest] containing:
//
//   - Version: Manifest format version for forward compatibility
//   - CreatedAt: Timestamp when the backup was created
//   - Platform: The AI assistant platform (claude, opencode, etc.)
//   - Files: List of backed up files with paths, hashes, and permissions
//   - AIXVersion: Version of aix that created the backup
//
// # Integrity Verification
//
// File integrity is verified using SHA256 checksums stored in the manifest.
// If a backup file's hash doesn't match during restore, [ErrBackupCorrupted]
// is returned.
//
// # Error Handling
//
// The package defines several sentinel errors for specific failure conditions:
//
//   - [ErrNoBackupsFound]: No backups exist for the specified platform
//   - [ErrBackupCorrupted]: Backup file integrity check failed
//   - [ErrRestoreConflict]: Target file has been modified since backup
package backup
