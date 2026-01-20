package frontmatter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v2"
)

// Sentinel errors for frontmatter parsing.
var (
	// ErrNoFrontmatter indicates the input does not contain frontmatter.
	// This occurs when the content doesn't start with a "---" delimiter.
	ErrNoFrontmatter = errors.New("no frontmatter found")

	// ErrInvalidYAML indicates the frontmatter contains invalid YAML.
	// This occurs when the content between "---" delimiters cannot be
	// parsed as valid YAML or cannot be unmarshaled into the target type.
	ErrInvalidYAML = errors.New("invalid YAML in frontmatter")
)

// Parse reads frontmatter from an io.Reader and unmarshals it into type T.
// It returns the parsed frontmatter, the remaining body content, and any error.
//
// The frontmatter must be delimited by lines containing only "---" at the
// start and end. Both Unix (LF) and Windows (CRLF) line endings are supported.
//
// If no frontmatter is found, returns ErrNoFrontmatter.
// If frontmatter exists but contains invalid YAML, returns ErrInvalidYAML.
func Parse[T any](r io.Reader) (*T, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", fmt.Errorf("reading input: %w", err)
	}

	// Normalize Windows line endings to Unix for consistent parsing
	content := strings.ReplaceAll(string(data), "\r\n", "\n")

	// Check if content starts with frontmatter delimiter
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---") {
		return nil, "", ErrNoFrontmatter
	}

	var result T
	rest, err := frontmatter.Parse(bytes.NewReader([]byte(content)), &result)
	if err != nil {
		// Distinguish between missing frontmatter and invalid YAML
		if strings.Contains(err.Error(), "not found") {
			return nil, "", ErrNoFrontmatter
		}
		// Check if it's a YAML parsing error
		var yamlErr *yaml.TypeError
		if errors.As(err, &yamlErr) {
			return nil, "", fmt.Errorf("%w: %v", ErrInvalidYAML, err)
		}
		return nil, "", fmt.Errorf("%w: %v", ErrInvalidYAML, err)
	}

	// The library returns the entire content as rest when no closing delimiter
	// is found. Detect this by checking if rest equals the original content.
	if string(rest) == content {
		return nil, "", ErrNoFrontmatter
	}

	return &result, string(rest), nil
}

// ParseFile is a convenience wrapper that opens a file and calls Parse.
// It returns the parsed frontmatter, the remaining body content, and any error.
//
// In addition to frontmatter parsing errors, this function may return
// file system errors such as os.ErrNotExist if the file doesn't exist.
func ParseFile[T any](path string) (*T, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("opening file %s: %w", path, err)
	}
	defer f.Close()

	result, body, err := Parse[T](f)
	if err != nil {
		return nil, "", err
	}

	return result, body, nil
}
