// Package parser provides command file parsing functionality.
package parser

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

// Sentinel errors for command parsing.
var (
	ErrEmptyFile            = errors.New("file is empty")
	ErrMalformedFrontmatter = errors.New("malformed frontmatter")
)

// ParseError represents an error that occurred while parsing a command file.
type ParseError struct {
	Path string
	Err  error
}

func (e *ParseError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("parsing command: %v", e.Err)
	}
	return fmt.Sprintf("parsing command %s: %v", e.Path, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
