package parser

import "fmt"

// ParseError represents an error that occurred during skill file parsing.
type ParseError struct {
	Path string // Path to the file that failed to parse
	Err  error  // Underlying error
}

func (e *ParseError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("parsing skill: %v", e.Err)
	}
	return fmt.Sprintf("parsing skill %s: %v", e.Path, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
