// Package parser provides command file parsing functionality.
package parser

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/thoreinstein/aix/internal/command"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Commandable is implemented by command types that can have their Name and Instructions set.
type Commandable interface {
	GetName() string
	SetName(string)
	SetInstructions(string)
}

// Parser handles command file parsing operations.
type Parser[T Commandable] struct{}

// New creates a new Parser instance.
func New[T Commandable]() *Parser[T] {
	return &Parser[T]{}
}

// ParseFile reads and parses a command file from the given path.
// Returns the parsed command or an error if parsing fails.
func (p *Parser[T]) ParseFile(path string) (*T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}
	defer f.Close()

	return p.Parse(f, path)
}

// Parse reads and parses a command from the given reader.
// The path parameter is used for error context and name inference.
func (p *Parser[T]) Parse(r io.Reader, path string) (*T, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}

	return p.ParseBytes(data, path)
}

// ParseBytes parses command content from bytes.
// The path parameter is used for error context and name inference.
func (p *Parser[T]) ParseBytes(data []byte, path string) (*T, error) {
	var cmd T
	body, err := frontmatter.Parse(bytes.NewReader(data), &cmd)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}

	// Set instructions from body
	cmd.SetInstructions(strings.TrimSpace(string(body)))

	// Infer name from path if not set in frontmatter
	if cmd.GetName() == "" && path != "" {
		cmd.SetName(command.InferName(path))
	}

	return &cmd, nil
}

// ParseHeader parses only the frontmatter metadata, stopping at the closing ---.
// This is more efficient for listing commands without reading full content.
func (p *Parser[T]) ParseHeader(path string) (*T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}
	defer f.Close()

	var cmd T
	if err := frontmatter.ParseHeader(f, &cmd); err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}

	// Infer name from path if not set in frontmatter
	if cmd.GetName() == "" && path != "" {
		cmd.SetName(command.InferName(path))
	}

	return &cmd, nil
}
