// Package fileutil provides file system utilities including atomic write operations.
package fileutil

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/thoreinstein/aix/internal/errors"
)

// AtomicWriteFile writes data to a file atomically using a temp file + rename pattern.
// This ensures interrupted writes leave the original file intact.
//
// The caller is responsible for ensuring the parent directory exists.
// Permissions are applied to the final file via the perm parameter.
func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	// Create temp file in same directory for atomic rename (same filesystem required)
	tmp, err := os.CreateTemp(dir, ".aix-atomic-*.tmp")
	if err != nil {
		return errors.Wrap(err, "creating temp file")
	}

	// Track temp file name for cleanup
	tmpName := tmp.Name()
	defer func() {
		// Only remove if rename failed (file still exists)
		if _, statErr := os.Stat(tmpName); statErr == nil {
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return errors.Wrap(err, "writing temp file")
	}

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return errors.Wrap(err, "setting file permissions")
	}

	if err := tmp.Close(); err != nil {
		return errors.Wrap(err, "closing temp file")
	}

	if err := os.Rename(tmpName, path); err != nil {
		return errors.Wrap(err, "renaming temp file")
	}

	return nil
}

// AtomicWriteJSON writes v as indented JSON to path atomically with specified permissions.
// Uses 2-space indentation and appends a trailing newline for POSIX compliance.
//
// The caller is responsible for ensuring the parent directory exists.
func AtomicWriteJSONWithPerm(path string, v any, perm os.FileMode) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling JSON")
	}

	// Add trailing newline for POSIX compliance
	data = append(data, '\n')

	return AtomicWriteFile(path, data, perm)
}

// AtomicWriteJSON writes v as indented JSON to path atomically.
// Uses 2-space indentation and appends a trailing newline for POSIX compliance.
//
// The caller is responsible for ensuring the parent directory exists.
// The file is created with 0644 permissions.
func AtomicWriteJSON(path string, v any) error {
	return AtomicWriteJSONWithPerm(path, v, 0644)
}

// AtomicWriteYAMLWithPerm writes v as YAML to path atomically with specified permissions.
// Appends a trailing newline for POSIX compliance.
//
// The caller is responsible for ensuring the parent directory exists.
func AtomicWriteYAMLWithPerm(path string, v any, perm os.FileMode) (err error) {
	// yaml.Marshal panics on unmarshalable types; recover and return error
	defer func() {
		if r := recover(); r != nil {
			err = errors.Newf("marshaling YAML: %v", r)
		}
	}()

	data, err := yaml.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "marshaling YAML")
	}

	// yaml.Marshal already includes trailing newline, but ensure it
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	return AtomicWriteFile(path, data, perm)
}

// AtomicWriteYAML writes v as YAML to path atomically.
// Appends a trailing newline for POSIX compliance.
//
// The caller is responsible for ensuring the parent directory exists.
// The file is created with 0644 permissions.
func AtomicWriteYAML(path string, v any) (err error) {
	return AtomicWriteYAMLWithPerm(path, v, 0644)
}
