// Package command provides slash command parsing and validation.
package command

import (
	"path/filepath"
	"strings"
)

// InferName derives a command name from a file path.
// It extracts the filename and strips the .md extension.
//
// Transformation rules:
//   - review.md -> review
//   - my-command.md -> my-command (case preserved)
//   - /path/to/review.md -> review (path stripped)
//   - review -> review (no extension = unchanged)
//   - file.test.md -> file.test (only .md stripped)
func InferName(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".md")
}
