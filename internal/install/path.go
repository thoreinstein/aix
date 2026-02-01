package install

import (
	"path/filepath"
	"strings"
)

// LooksLikePath returns true if the source appears to be a file path.
func LooksLikePath(source string) bool {
	// Starts with ./ or ../ or /
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") || strings.HasPrefix(source, "/") {
		return true
	}
	// Contains path separator
	if strings.Contains(source, string(filepath.Separator)) {
		return true
	}
	// On Windows, also check for backslash
	if filepath.Separator != '/' && strings.Contains(source, "/") {
		return true
	}
	return false
}

// MightBePath returns true if the input might be a path the user forgot the --file flag for.
// This catches edge cases not handled by LooksLikePath, like Windows-style paths on Unix
// or files with common resource extensions.
func MightBePath(s string, resourceType string) bool {
	// Check for common extensions based on resource type
	lower := strings.ToLower(s)
	switch resourceType {
	case "skill", "agent", "command":
		if strings.HasSuffix(lower, ".md") {
			return true
		}
	case "mcp":
		if strings.HasSuffix(lower, ".json") {
			return true
		}
	}

	// Contains backslash (Windows paths, even on Unix)
	if strings.Contains(s, `\`) {
		return true
	}
	return false
}
