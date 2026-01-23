package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/thoreinstein/aix/pkg/fileutil"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Manager handles backup creation, restoration, and management.
type Manager struct {
	rootDir        string
	retentionCount int
}

// Option configures a Manager.
type Option func(*Manager)

// WithBackupDir sets the root backup directory.
func WithBackupDir(dir string) Option {
	return func(m *Manager) {
		m.rootDir = dir
	}
}

// WithRetentionCount sets the number of backups to retain per platform.
func WithRetentionCount(n int) Option {
	return func(m *Manager) {
		if n > 0 {
			m.retentionCount = n
		}
	}
}

// NewManager creates a new backup Manager with the given options.
func NewManager(opts ...Option) *Manager {
	m := &Manager{
		rootDir:        BackupDir(),
		retentionCount: DefaultRetentionCount,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Backup creates a backup of the specified paths for a platform.
// Returns the manifest describing the backup, or an error if the backup fails.
//
// The paths can be files or directories. Directories are backed up recursively.
// Each file is copied with preserved permissions and verified with a SHA256 hash.
func (m *Manager) Backup(platform string, paths []string) (*BackupManifest, error) {
	if platform == "" {
		return nil, errors.New("platform is required")
	}
	if len(paths) == 0 {
		return nil, errors.New("at least one path is required")
	}

	// Generate backup ID from current time
	backupID := time.Now().Format("20060102T150405")
	backupPath := m.backupPath(platform, backupID)

	// Create backup directory
	if err := os.MkdirAll(backupPath, 0o755); err != nil {
		return nil, errors.Wrap(err, "creating backup directory")
	}

	// Track backed up files
	var files []BackupFile

	// Back up each path
	for _, p := range paths {
		// Expand home directory
		expanded := expandHome(p)

		info, err := os.Stat(expanded)
		if err != nil {
			if os.IsNotExist(err) {
				// Skip non-existent paths
				continue
			}
			return nil, errors.Wrapf(err, "stat %s", p)
		}

		if info.IsDir() {
			// Recursively back up directory
			dirFiles, err := m.backupDirectory(expanded, backupPath)
			if err != nil {
				return nil, errors.Wrapf(err, "backing up directory %s", p)
			}
			files = append(files, dirFiles...)
		} else {
			// Back up single file
			bf, err := m.backupFile(expanded, backupPath)
			if err != nil {
				return nil, errors.Wrapf(err, "backing up file %s", p)
			}
			files = append(files, *bf)
		}
	}

	if len(files) == 0 {
		// Clean up empty backup directory
		os.RemoveAll(backupPath)
		return nil, errors.New("no files to back up")
	}

	// Create manifest
	manifest := &BackupManifest{
		Version:    ManifestVersion,
		CreatedAt:  time.Now().UTC(),
		Platform:   platform,
		Files:      files,
		AIXVersion: Version,
		ID:         backupID,
	}

	// Write manifest
	manifestPath := filepath.Join(backupPath, "manifest.json")
	if err := fileutil.AtomicWriteJSON(manifestPath, manifest); err != nil {
		return nil, errors.Wrap(err, "writing manifest")
	}

	return manifest, nil
}

// backupFile copies a single file to the backup directory.
func (m *Manager) backupFile(src, backupPath string) (*BackupFile, error) {
	// Generate relative path for backup storage
	relPath := generateRelPath(src)
	dst := filepath.Join(backupPath, relPath)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return nil, errors.Wrap(err, "creating parent directory")
	}

	// Copy the file and get its hash
	hash, mode, err := copyFile(src, dst)
	if err != nil {
		return nil, err
	}

	return &BackupFile{
		OriginalPath: src,
		RelPath:      relPath,
		SHA256Hash:   hash,
		Mode:         mode,
	}, nil
}

// backupDirectory recursively backs up all files in a directory.
func (m *Manager) backupDirectory(srcDir, backupPath string) ([]BackupFile, error) {
	var files []BackupFile

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories themselves (we only track files)
		if d.IsDir() {
			return nil
		}

		bf, err := m.backupFile(path, backupPath)
		if err != nil {
			return err
		}
		files = append(files, *bf)
		return nil
	})

	return files, err
}

// Restore restores files from a backup to their original locations.
// The backupID should be in timestamp format (e.g., "20260123T100712").
func (m *Manager) Restore(platform, backupID string) error {
	if platform == "" {
		return errors.New("platform is required")
	}
	if backupID == "" {
		return errors.New("backup ID is required")
	}

	// Load the manifest
	manifest, err := m.Get(platform, backupID)
	if err != nil {
		return err
	}

	backupPath := m.backupPath(platform, backupID)

	// Restore each file
	for _, bf := range manifest.Files {
		srcPath := filepath.Join(backupPath, bf.RelPath)

		// Verify integrity before restoring
		hash, err := hashFile(srcPath)
		if err != nil {
			return errors.Wrapf(err, "reading backup file %s", bf.RelPath)
		}
		if hash != bf.SHA256Hash {
			return errors.Wrapf(ErrBackupCorrupted, "file %s hash mismatch", bf.RelPath)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(bf.OriginalPath), 0o755); err != nil {
			return errors.Wrapf(err, "creating directory for %s", bf.OriginalPath)
		}

		// Copy file back to original location
		if _, _, err := copyFile(srcPath, bf.OriginalPath); err != nil {
			return errors.Wrapf(err, "restoring %s", bf.OriginalPath)
		}

		// Restore original permissions
		if err := os.Chmod(bf.OriginalPath, bf.Mode); err != nil {
			return errors.Wrapf(err, "setting permissions for %s", bf.OriginalPath)
		}
	}

	return nil
}

// List returns all available backups for a platform, sorted by date (newest first).
func (m *Manager) List(platform string) ([]BackupManifest, error) {
	if platform == "" {
		return nil, errors.New("platform is required")
	}

	platformDir := m.platformBackupDir(platform)

	entries, err := os.ReadDir(platformDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoBackupsFound
		}
		return nil, errors.Wrap(err, "reading backup directory")
	}

	manifests := make([]BackupManifest, 0, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifest, err := m.Get(platform, entry.Name())
		if err != nil {
			// Skip invalid backup directories
			continue
		}
		manifests = append(manifests, *manifest)
	}

	if len(manifests) == 0 {
		return nil, ErrNoBackupsFound
	}

	// Sort by date, newest first
	slices.SortFunc(manifests, func(a, b BackupManifest) int {
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		}
		if a.CreatedAt.Before(b.CreatedAt) {
			return 1
		}
		return 0
	})

	return manifests, nil
}

// Prune removes old backups beyond the specified retention count.
// Keeps the most recent 'keep' backups for the platform.
func (m *Manager) Prune(platform string, keep int) error {
	if platform == "" {
		return errors.New("platform is required")
	}
	if keep < 0 {
		return errors.New("keep must be non-negative")
	}

	manifests, err := m.List(platform)
	if err != nil {
		if errors.Is(err, ErrNoBackupsFound) {
			return nil // Nothing to prune
		}
		return err
	}

	// Already sorted newest first, delete everything beyond 'keep'
	for i := keep; i < len(manifests); i++ {
		backupPath := m.backupPath(platform, manifests[i].ID)
		if err := os.RemoveAll(backupPath); err != nil {
			return errors.Wrapf(err, "removing backup %s", manifests[i].ID)
		}
	}

	return nil
}

// Get returns the manifest for a specific backup.
func (m *Manager) Get(platform, backupID string) (*BackupManifest, error) {
	if platform == "" {
		return nil, errors.New("platform is required")
	}
	if backupID == "" {
		return nil, errors.New("backup ID is required")
	}

	manifestPath := filepath.Join(m.backupPath(platform, backupID), "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrapf(ErrNoBackupsFound, "backup %s not found", backupID)
		}
		return nil, errors.Wrap(err, "reading manifest")
	}

	var manifest BackupManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, errors.Wrap(err, "parsing manifest")
	}

	manifest.ID = backupID
	return &manifest, nil
}

// backupPath returns the full path to a backup directory.
func (m *Manager) backupPath(platform, backupID string) string {
	return filepath.Join(m.platformBackupDir(platform), backupID)
}

// platformBackupDir returns the backup directory for a platform.
func (m *Manager) platformBackupDir(platform string) string {
	return filepath.Join(m.rootDir, platform)
}

// hashFile computes the SHA256 hash of a file.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "opening file")
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", errors.Wrap(err, "reading file")
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// copyFile copies a file from src to dst, returning the SHA256 hash and mode.
// The destination file is created with 0644 permissions initially,
// then updated to match the source file's permissions.
func copyFile(src, dst string) (hash string, mode fs.FileMode, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return "", 0, errors.Wrap(err, "opening source file")
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return "", 0, errors.Wrap(err, "stat source file")
	}
	mode = srcInfo.Mode()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", 0, errors.Wrap(err, "creating destination file")
	}

	// Compute hash while copying
	h := sha256.New()
	w := io.MultiWriter(dstFile, h)

	if _, err := io.Copy(w, srcFile); err != nil {
		dstFile.Close()
		return "", 0, errors.Wrap(err, "copying file")
	}

	if err := dstFile.Close(); err != nil {
		return "", 0, errors.Wrap(err, "closing destination file")
	}

	// Set permissions to match source
	if err := os.Chmod(dst, mode); err != nil {
		return "", 0, errors.Wrap(err, "setting permissions")
	}

	return hex.EncodeToString(h.Sum(nil)), mode, nil
}

// generateRelPath creates a relative path for storage in the backup directory.
// It converts absolute paths to a consistent format using the base directory name
// and file path.
func generateRelPath(absPath string) string {
	// Use a simple approach: just use the absolute path with slashes replaced
	// This ensures uniqueness and makes it easy to understand the source
	clean := filepath.Clean(absPath)

	// Remove leading slash for Unix or drive letter for Windows
	if filepath.IsAbs(clean) {
		if len(clean) > 0 && clean[0] == filepath.Separator {
			clean = clean[1:]
		}
	}

	return clean
}

// expandHome expands ~ to the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}

	return path
}
