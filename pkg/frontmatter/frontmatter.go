// Package frontmatter provides utilities for parsing and formatting
// YAML frontmatter in markdown files.
package frontmatter

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrMissingFrontmatter is returned by MustParse when no frontmatter is found.
var ErrMissingFrontmatter = errors.New("missing frontmatter")

// Parse extracts YAML frontmatter and body content from a reader.
// If no frontmatter is present, returns empty struct and full content as body.
// This is useful for files where frontmatter is optional (commands, agents).
func Parse[T any](r io.Reader, matter *T) (body []byte, err error) {
	return parse(r, matter, false)
}

// MustParse is like Parse but returns an error if no frontmatter is found.
// This is useful for files where frontmatter is required (skills).
func MustParse[T any](r io.Reader, matter *T) (body []byte, err error) {
	return parse(r, matter, true)
}

func parse[T any](r io.Reader, matter *T, required bool) ([]byte, error) {
	// Read full content
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Check for start delimiter
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		if required {
			return nil, ErrMissingFrontmatter
		}
		return content, nil
	}

	// Find end delimiter
	// We start searching after the first 3 bytes (---)
	// We need to handle \n--- and \r\n---
	startOffset := 3
	if len(content) > 3 && content[3] == '\r' {
		startOffset = 4
	}
	if len(content) > startOffset && content[startOffset] == '\n' {
		startOffset++
	}

	// Search for closing "---" on a new line
	parts := bytes.SplitN(content[startOffset:], []byte("\n---"), 2)
	if len(parts) < 2 {
		// Try CRLF
		parts = bytes.SplitN(content[startOffset:], []byte("\r\n---"), 2)
	}

	if len(parts) < 2 {
		if required {
			return nil, errors.New("missing closing frontmatter delimiter")
		}
		return content, nil
	}

	// Frontmatter is parts[0]
	// Body is parts[1], but we need to trim the newline after --- if present
	fm := parts[0]
	bodyContent := parts[1]

	// Trim leading newline from body (residue from split)
	if len(bodyContent) > 0 {
		if bodyContent[0] == '\r' {
			bodyContent = bodyContent[1:]
		}
		if len(bodyContent) > 0 && bodyContent[0] == '\n' {
			bodyContent = bodyContent[1:]
		}
	}

	if err := yaml.Unmarshal(fm, matter); err != nil {
		return nil, err
	}

	return bodyContent, nil
}

// ParseHeader parses only the frontmatter from the reader.
// It stops reading after the closing delimiter "---".
// The body is not consumed or returned.
// Returns nil if no frontmatter is found (silent success, matter remains empty).
func ParseHeader(r io.Reader, matter any) error {
	scanner := bufio.NewScanner(r)

	// Check first line
	if !scanner.Scan() {
		return scanner.Err()
	}
	line := strings.TrimSpace(scanner.Text())
	if line != "---" {
		// No frontmatter start delimiter
		return nil
	}

	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			// Found closing delimiter
			return yaml.Unmarshal(buf.Bytes(), matter)
		}
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	return scanner.Err()
}

// Format formats content with YAML frontmatter.
// The matter struct is serialized to YAML and wrapped in "---" delimiters,
// followed by the body content.
func Format(matter any, body string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("---\n")

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(matter); err != nil {
		return nil, err
	}

	buf.WriteString("---\n")
	if body != "" {
		buf.WriteString("\n")
		buf.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes(), nil
}
