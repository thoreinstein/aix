// Package frontmatter provides utilities for parsing and formatting
// YAML frontmatter in markdown files. It wraps github.com/adrg/frontmatter
// and provides additional formatting capabilities.
package frontmatter

import (
	"bytes"
	"io"

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

// Format formats content with YAML frontmatter.
// The matter struct is serialized to YAML and wrapped in "---" delimiters,
// followed by the body content.
func Format(matter any, body string) ([]byte, error) {
	yamlData, err := yaml.Marshal(matter)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlData)
	buf.WriteString("---\n")
	if body != "" {
		buf.WriteString("\n")
		buf.WriteString(body)
		if len(body) > 0 && body[len(body)-1] != '\n' {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes(), nil
}
