package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteFile(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		perm    os.FileMode
		wantErr bool
	}{
		{
			name:    "successful write",
			data:    []byte("hello world\n"),
			perm:    0644,
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			perm:    0644,
			wantErr: false,
		},
		{
			name:    "binary data",
			data:    []byte{0x00, 0x01, 0x02, 0xFF},
			perm:    0600,
			wantErr: false,
		},
		{
			name:    "executable permissions",
			data:    []byte("#!/bin/sh\necho hello\n"),
			perm:    0755,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test-file")

			err := AtomicWriteFile(path, tt.data, tt.perm)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AtomicWriteFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			// Verify content
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading file: %v", err)
			}
			if string(got) != string(tt.data) {
				t.Errorf("content = %q, want %q", got, tt.data)
			}

			// Verify permissions
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stating file: %v", err)
			}
			// Mask with 0777 to ignore directory bits
			if gotPerm := info.Mode().Perm(); gotPerm != tt.perm {
				t.Errorf("permissions = %o, want %o", gotPerm, tt.perm)
			}
		})
	}
}

func TestAtomicWriteFile_DirectoryNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "subdir", "file.txt")

	err := AtomicWriteFile(path, []byte("data"), 0600)
	if err == nil {
		t.Error("AtomicWriteFile() expected error for nonexistent directory")
	}
}

func TestAtomicWriteFile_OverwriteExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing-file")

	// Create original file
	original := []byte("original content\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatalf("creating original file: %v", err)
	}

	// Overwrite with new content
	newContent := []byte("new content\n")
	if err := AtomicWriteFile(path, newContent, 0600); err != nil {
		t.Fatalf("AtomicWriteFile() error = %v", err)
	}

	// Verify new content
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(got) != string(newContent) {
		t.Errorf("content = %q, want %q", got, newContent)
	}
}

func TestAtomicWriteFile_NoTempFileLeftOnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent-dir", "file.txt")

	// This should fail because parent directory doesn't exist
	_ = AtomicWriteFile(path, []byte("data"), 0600)

	// Check that no temp files are left in the temp dir
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading directory: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("temp file left behind: %s", entry.Name())
		}
	}
}

func TestAtomicWriteJSON(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		wantJSON string
		wantErr  bool
	}{
		{
			name:     "simple struct",
			value:    struct{ Name string }{Name: "test"},
			wantJSON: "{\n  \"Name\": \"test\"\n}\n",
			wantErr:  false,
		},
		{
			name:     "map",
			value:    map[string]int{"count": 42},
			wantJSON: "{\n  \"count\": 42\n}\n",
			wantErr:  false,
		},
		{
			name:     "slice",
			value:    []string{"a", "b", "c"},
			wantJSON: "[\n  \"a\",\n  \"b\",\n  \"c\"\n]\n",
			wantErr:  false,
		},
		{
			name:     "nested struct",
			value:    struct{ Inner struct{ Value int } }{Inner: struct{ Value int }{Value: 123}},
			wantJSON: "{\n  \"Inner\": {\n    \"Value\": 123\n  }\n}\n",
			wantErr:  false,
		},
		{
			name:    "unmarshalable channel",
			value:   make(chan int),
			wantErr: true,
		},
		{
			name:    "unmarshalable func",
			value:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.json")

			err := AtomicWriteJSON(path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AtomicWriteJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				// Verify file was not created on error
				if _, err := os.Stat(path); err == nil {
					t.Error("file should not exist after marshal error")
				}
				return
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading file: %v", err)
			}
			if string(got) != tt.wantJSON {
				t.Errorf("content = %q, want %q", got, tt.wantJSON)
			}

			// Verify default permissions
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stating file: %v", err)
			}
			if gotPerm := info.Mode().Perm(); gotPerm != 0600 {
				t.Errorf("permissions = %o, want 0600", gotPerm)
			}
		})
	}
}

func TestAtomicWriteYAML(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		wantYAML string
		wantErr  bool
	}{
		{
			name:     "simple struct",
			value:    struct{ Name string }{Name: "test"},
			wantYAML: "name: test\n",
			wantErr:  false,
		},
		{
			name:     "map",
			value:    map[string]int{"count": 42},
			wantYAML: "count: 42\n",
			wantErr:  false,
		},
		{
			name:     "slice",
			value:    []string{"a", "b", "c"},
			wantYAML: "- a\n- b\n- c\n",
			wantErr:  false,
		},
		{
			name: "nested struct",
			value: struct {
				Inner struct {
					Value int
				}
			}{Inner: struct{ Value int }{Value: 123}},
			wantYAML: "inner:\n    value: 123\n",
			wantErr:  false,
		},
		{
			name:    "unmarshalable channel",
			value:   make(chan int),
			wantErr: true,
		},
		{
			name:    "unmarshalable func",
			value:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.yaml")

			err := AtomicWriteYAML(path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AtomicWriteYAML() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				// Verify file was not created on error
				if _, err := os.Stat(path); err == nil {
					t.Error("file should not exist after marshal error")
				}
				return
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading file: %v", err)
			}
			if string(got) != tt.wantYAML {
				t.Errorf("content = %q, want %q", got, tt.wantYAML)
			}

			// Verify default permissions
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stating file: %v", err)
			}
			if gotPerm := info.Mode().Perm(); gotPerm != 0600 {
				t.Errorf("permissions = %o, want 0600", gotPerm)
			}
		})
	}
}

func TestAtomicWriteJSON_DirectoryNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "test.json")

	err := AtomicWriteJSON(path, map[string]string{"key": "value"})
	if err == nil {
		t.Error("AtomicWriteJSON() expected error for nonexistent directory")
	}
}

func TestAtomicWriteYAML_DirectoryNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "test.yaml")

	err := AtomicWriteYAML(path, map[string]string{"key": "value"})
	if err == nil {
		t.Error("AtomicWriteYAML() expected error for nonexistent directory")
	}
}

func TestAtomicWriteJSON_TrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	if err := AtomicWriteJSON(path, "simple"); err != nil {
		t.Fatalf("AtomicWriteJSON() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Error("JSON output should have trailing newline")
	}
}

func TestAtomicWriteYAML_TrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	if err := AtomicWriteYAML(path, "simple"); err != nil {
		t.Fatalf("AtomicWriteYAML() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Error("YAML output should have trailing newline")
	}
}
