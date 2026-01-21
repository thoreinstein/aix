// Package frontmatter provides utilities for parsing and formatting
// YAML frontmatter in markdown files. It wraps github.com/adrg/frontmatter
// and provides additional formatting capabilities.
package frontmatter

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

// Parse extracts YAML frontmatter and body content from a reader.
// If no frontmatter is present, returns empty struct and full content as body.
// This is useful for files where frontmatter is optional (commands, agents).
func Parse[T any](r io.Reader, matter *T) (body []byte, err error) {
	return frontmatter.Parse(r, matter)
}

// MustParse is like Parse but returns an error if no frontmatter is found.
// This is useful for files where frontmatter is required (skills).
func MustParse[T any](r io.Reader, matter *T) (body []byte, err error) {
	return frontmatter.MustParse(r, matter)
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
	if strings.TrimSpace(scanner.Text()) != "---" {
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
