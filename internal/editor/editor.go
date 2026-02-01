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

	// Split editor command and arguments, respecting quoted strings
	// e.g., `"Visual Studio Code" --wait` -> ["Visual Studio Code", "--wait"]
	parts, err := splitCommand(editorEnv)
	if err != nil {
		return fmt.Errorf("parsing editor command: %w", err)
	}
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
// and available binaries. Fallback chain: $EDITOR -> $VISUAL -> nano -> vi
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

// splitCommand splits a command string into parts, respecting quoted strings.
// Supports both single and double quotes. Backslash escapes the next character
// inside double quotes. Returns an error if quotes are unbalanced.
func splitCommand(cmd string) ([]string, error) {
	var parts []string
	var current strings.Builder
	var inSingleQuote, inDoubleQuote bool

	for i := 0; i < len(cmd); i++ {
		c := cmd[i]

		switch {
		case c == '\\' && inDoubleQuote && i+1 < len(cmd):
			// Backslash escape inside double quotes
			i++
			current.WriteByte(cmd[i])

		case c == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote

		case c == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote

		case (c == ' ' || c == '\t') && !inSingleQuote && !inDoubleQuote:
			// Whitespace outside quotes - end of token
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}

		default:
			current.WriteByte(c)
		}
	}

	if inSingleQuote {
		return nil, errors.New("unbalanced single quotes")
	}
	if inDoubleQuote {
		return nil, errors.New("unbalanced double quotes")
	}

	// Don't forget the last token
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts, nil
}
