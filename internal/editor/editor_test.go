package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDetectEditor_EnvEditor(t *testing.T) {
	t.Setenv("EDITOR", "nvim")
	t.Setenv("VISUAL", "code")

	got := detectEditor()
	if got != "nvim" {
		t.Errorf("detectEditor() = %q, want %q", got, "nvim")
	}
}

func TestDetectEditor_EnvVisual(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "code")

	got := detectEditor()
	if got != "code" {
		t.Errorf("detectEditor() = %q, want %q", got, "code")
	}
}

func TestDetectEditor_FallbackNano(t *testing.T) {
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")

	got := detectEditor()

	// Should be nano if available, otherwise vi
	if _, err := exec.LookPath("nano"); err == nil {
		if got != "nano" {
			t.Errorf("detectEditor() = %q, want %q (nano available)", got, "nano")
		}
	} else {
		if got != "vi" {
			t.Errorf("detectEditor() = %q, want %q (nano not available)", got, "vi")
		}
	}
}

func TestDetectEditor_EmptyEnvTreatedAsUnset(t *testing.T) {
	// Empty string should be treated as unset
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "vscode")

	got := detectEditor()
	if got != "vscode" {
		t.Errorf("detectEditor() = %q, want %q (empty EDITOR should fall through)", got, "vscode")
	}
}

func TestOpen_Integration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping integration test on windows (uses shell script mock)")
	}

	tmpDir := t.TempDir()
	mockEditor := filepath.Join(tmpDir, "mock-editor.sh")
	outputFile := filepath.Join(tmpDir, "output.txt")

	// Create a mock editor that writes its arguments to a file
	script := "#!/bin/sh\necho \"$@\" > " + outputFile + "\n"
	if err := os.WriteFile(mockEditor, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("EDITOR", mockEditor)

	targetFile := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run Open
	if err := Open(targetFile); err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	// Verify mock editor was called with the right argument
	got, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(got), targetFile) {
		t.Errorf("mock editor output = %q, want it to contain %q", string(got), targetFile)
	}
}

func TestOpen_NoEditor(t *testing.T) {
	// This is hard to test because detectEditor always falls back to "vi"
	// but we can try to force a failure by setting EDITOR to a non-existent binary
	t.Setenv("EDITOR", "non-existent-binary-12345")
	t.Setenv("VISUAL", "")

	err := Open("test.txt")
	if err == nil {
		t.Error("expected error for non-existent editor, got nil")
	}
}
