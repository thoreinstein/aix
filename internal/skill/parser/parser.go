// Package parser provides SKILL.md file parsing functionality.
// It extracts YAML frontmatter and markdown body content from skill files
// according to the Agent Skills Specification.
package parser

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Parser handles SKILL.md file parsing operations.
type Parser struct{}

// New creates a new Parser instance.
func New() *Parser {
	return &Parser{}
}

// ParseFile reads and parses a SKILL.md file from the given path.
// Returns the parsed Skill or an error if parsing fails.
func (p *Parser) ParseFile(path string) (*claude.Skill, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}
	defer f.Close()

	return p.Parse(f, path)
}

// Parse reads and parses a SKILL.md from the given reader.
// The path parameter is used for error context only.
func (p *Parser) Parse(r io.Reader, path string) (*claude.Skill, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}

	return p.ParseBytes(data, path)
}

// ParseBytes parses SKILL.md content from bytes.
// The path parameter is used for error context only.
func (p *Parser) ParseBytes(data []byte, path string) (*claude.Skill, error) {
	var skill claude.Skill
	body, err := frontmatter.MustParse(bytes.NewReader(data), &skill)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}

	skill.Instructions = strings.TrimSpace(string(body))
	return &skill, nil
}

// ParseHeader parses only the frontmatter metadata, stopping at the closing ---.
// This is more efficient for listing skills without reading full content.
func (p *Parser) ParseHeader(path string) (*claude.Skill, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}
	defer f.Close()

	var skill claude.Skill
	if err := frontmatter.ParseHeader(f, &skill); err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}

	return &skill, nil
}
