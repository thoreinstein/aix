package fileutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thoreinstein/aix/internal/errors"
)

func TestReadFileWithLimit(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{"small file", 100, false},
		{"exact limit", MaxFileSize, false},
		{"too large", MaxFileSize + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tempDir, tt.name)
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}

			// Write dummy data
			if err := f.Truncate(tt.size); err != nil {
				t.Fatal(err)
			}
			f.Close()

			_, err = ReadFileWithLimit(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFileWithLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrFileTooLarge) {
				t.Errorf("expected ErrFileTooLarge, got %v", err)
			}
		})
	}
}
