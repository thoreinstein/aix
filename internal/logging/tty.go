package logging

import (
	"io"
	"os"

	"golang.org/x/term"
)

// IsTTY returns true if the given writer is a terminal.
// It supports os.File and any wrapper that provides an Fd() method.
func IsTTY(w io.Writer) bool {
	if f, ok := w.(interface{ Fd() uintptr }); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// SupportsColor returns true if the given writer supports ANSI color codes.
// It returns false if:
//   - The writer is not a TTY
//   - The NO_COLOR environment variable is set
//   - The TERM environment variable is set to "dumb"
func SupportsColor(w io.Writer) bool {
	return supportsColor(w, IsTTY(w))
}

func supportsColor(w io.Writer, isTTY bool) bool {
	// Respect NO_COLOR standard (https://no-color.org)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}

	// Check TERM environment variable
	if term := os.Getenv("TERM"); term == "dumb" {
		return false
	}

	// Must be a TTY
	return isTTY
}
