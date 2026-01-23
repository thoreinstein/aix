// Package editor provides utilities for launching the user's preferred text editor.
package editor

import (
	"fmt"
	"os"
	"os/exec"
)

// Open launches the user's preferred editor for the given path.
// Uses $EDITOR environment variable, falling back to $VISUAL, then nano, then vi.
func Open(path string) error {
	editorCmd := detectEditor()

	fmt.Printf("Location: %s\n", path)

	cmd := exec.Command(editorCmd, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running editor: %w", err)
	}

	return nil
}

// detectEditor returns the editor command to use based on environment variables
// and available binaries. Fallback chain: $EDITOR → $VISUAL → nano → vi
func detectEditor() string {
	// Check $EDITOR first (most common)
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Then $VISUAL (for full-screen editors)
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	// User-friendly fallback (nano is easier for beginners)
	if _, err := exec.LookPath("nano"); err == nil {
		return "nano"
	}

	// POSIX standard fallback (vi is available on all Unix systems)
	return "vi"
}
