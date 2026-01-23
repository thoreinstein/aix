package backup

import (
	"io/fs"
	"time"

	"github.com/cockroachdb/errors"
)

// Manifest format version for forward compatibility.
const ManifestVersion = 1

// Default configuration values.
const (
	// DefaultRetentionCount is the default number of backups to retain per platform.
	DefaultRetentionCount = 5
)

// Sentinel errors for backup operations.
var (
	// ErrNoBackupsFound indicates no backups exist for the specified platform.
	ErrNoBackupsFound = errors.New("no backups found")

	// ErrBackupCorrupted indicates backup file integrity verification failed.
	// This occurs when a file's SHA256 hash doesn't match the manifest.
	ErrBackupCorrupted = errors.New("backup corrupted")

	// ErrRestoreConflict indicates the target file has been modified since the backup.
	// This prevents accidental overwrites of user changes.
	ErrRestoreConflict = errors.New("restore conflict")
)

// BackupManifest contains metadata about a backup.
// It is stored as manifest.json in each backup directory.
type BackupManifest struct {
	// Version is the manifest format version for forward compatibility.
	Version int `json:"version"`

	// CreatedAt is when the backup was created.
	CreatedAt time.Time `json:"created_at"`

	// Platform is the AI assistant platform (claude, opencode, etc.).
	Platform string `json:"platform"`

	// Files contains metadata for each backed up file.
	Files []BackupFile `json:"files"`

	// AIXVersion is the version of aix that created this backup.
	AIXVersion string `json:"aix_version"`

	// ID is the backup identifier (timestamp format: 20260123T100712).
	// This field is populated when loading from disk but not stored in JSON.
	ID string `json:"-"`
}

// BackupFile contains metadata for a single backed up file.
type BackupFile struct {
	// OriginalPath is the absolute path where the file was located.
	OriginalPath string `json:"original_path"`

	// RelPath is the relative path within the backup directory.
	RelPath string `json:"rel_path"`

	// SHA256Hash is the hex-encoded SHA256 hash of the file contents.
	SHA256Hash string `json:"sha256_hash"`

	// Mode is the file's permission bits.
	Mode fs.FileMode `json:"mode"`
}

// BackupConfig holds configuration for the backup manager.
type BackupConfig struct {
	// RetentionCount is the number of backups to retain per platform.
	// When exceeded, older backups are pruned.
	// Defaults to DefaultRetentionCount (5).
	RetentionCount int

	// BackupDir is the root directory for storing backups.
	// Defaults to ~/.config/aix/backups/
	BackupDir string
}
