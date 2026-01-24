// Package editor provides utilities for launching the user's preferred text editor.
package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Open launches the user's preferred editor for the given path.
// Uses $EDITOR environment variable, falling back to $VISUAL, then nano, then vi.
// It safely splits the editor command (e.g. "code -w") and validates the executable.
func Open(path string) error {
	editorEnv := detectEditor()

	// Split editor command and arguments (e.g., "code -w" -> ["code", "-w"])
	// We use a simple fields split which handles most cases.
	// For more complex quoting, shell parsing would be needed but adds complexity/risk.
	parts := strings.Fields(editorEnv)
	if len(parts) == 0 {
		return errors.New("no editor found")
	}

	bin := parts[0]
	args := parts[1:]

	// Validate executable exists
	execPath, err := exec.LookPath(bin)
	if err != nil {
		return fmt.Errorf("editor executable %q not found: %w", bin, err)
	}

	fmt.Printf("Opening %s with %s...\n", path, bin)

	// Append file path to arguments
	args = append(args, path)

	cmd := exec.Command(execPath, args...)
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
