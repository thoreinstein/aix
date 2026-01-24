package fileutil

import (
	"io"
	"os"

	"github.com/thoreinstein/aix/internal/errors"
)

// MaxFileSize is the maximum file size we'll read (1MB).
// This prevents memory exhaustion from maliciously large files.
const MaxFileSize = 1024 * 1024 // 1MB

// ErrFileTooLarge indicates that a file exceeded MaxFileSize.
var ErrFileTooLarge = errors.Newf("file exceeds maximum size of %d bytes", MaxFileSize)

// ReadFileWithLimit reads a file up to MaxFileSize.
// It returns an error if the file is larger than the limit.
func ReadFileWithLimit(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "opening file")
	}
	defer f.Close()

	// Get file info to fail fast if size is already too large
	info, err := f.Stat()
	if err == nil {
		if info.Size() > MaxFileSize {
			return nil, ErrFileTooLarge
		}
	}

	// Read with limit
	r := io.LimitReader(f, MaxFileSize+1)
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading file")
	}

	if len(data) > MaxFileSize {
		return nil, ErrFileTooLarge
	}

	return data, nil
}
