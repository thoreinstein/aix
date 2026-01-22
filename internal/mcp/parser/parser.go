// Package parser provides JSON parsing and writing for canonical MCP configurations.
// It handles loading MCP config files from disk and writing them back with proper
// formatting and atomic file operations.
package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/thoreinstein/aix/internal/mcp"
)

// Sentinel errors for parser operations.
var (
	// ErrInvalidJSON indicates the input is not valid JSON.
	ErrInvalidJSON = errors.New("invalid JSON")

	// ErrInvalidConfig indicates the JSON doesn't represent a valid MCP config.
	ErrInvalidConfig = errors.New("invalid MCP configuration")
)

// ParseError wraps errors that occur during parsing with path context.
type ParseError struct {
	Path string
	Err  error
}

func (e *ParseError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("parsing MCP config %s: %v", e.Path, e.Err)
	}
	return fmt.Sprintf("parsing MCP config: %v", e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// Parse reads a canonical MCP config from JSON bytes.
// Returns an error if the JSON is malformed or doesn't represent a valid config.
func Parse(data []byte) (*mcp.Config, error) {
	if len(data) == 0 {
		return mcp.NewConfig(), nil
	}

	var cfg mcp.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return nil, fmt.Errorf("%w: %v at offset %d", ErrInvalidJSON, err, syntaxErr.Offset)
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	// Initialize Servers map if it was nil in the JSON
	if cfg.Servers == nil {
		cfg.Servers = make(map[string]*mcp.Server)
	}

	return &cfg, nil
}

// ParseFile reads a canonical MCP config from a file path.
// Returns an empty config (not error) if the file doesn't exist, following the
// principle that a missing config file means "no servers configured".
// Returns an error for other file system issues or invalid JSON.
func ParseFile(path string) (*mcp.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Missing file is valid - return empty config
			return mcp.NewConfig(), nil
		}
		return nil, &ParseError{Path: path, Err: err}
	}

	cfg, err := Parse(data)
	if err != nil {
		return nil, &ParseError{Path: path, Err: err}
	}

	return cfg, nil
}

// Write writes a canonical MCP config to JSON bytes with indentation.
// The output is formatted with 2-space indentation for readability.
func Write(cfg *mcp.Config) ([]byte, error) {
	if cfg == nil {
		cfg = mcp.NewConfig()
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling MCP config: %w", err)
	}

	// Append newline for POSIX compliance
	data = append(data, '\n')
	return data, nil
}

// WriteFile writes a canonical MCP config to a file using atomic write.
// It writes to a temporary file first, then renames to the target path.
// This ensures the file is never left in a partial/corrupt state.
// Creates parent directories if they don't exist.
func WriteFile(path string, cfg *mcp.Config) error {
	data, err := Write(cfg)
	if err != nil {
		return &ParseError{Path: path, Err: err}
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return &ParseError{Path: path, Err: fmt.Errorf("creating directory: %w", err)}
	}

	// Write to temp file in same directory for atomic rename
	tmpFile, err := os.CreateTemp(dir, ".mcp-config-*.tmp")
	if err != nil {
		return &ParseError{Path: path, Err: fmt.Errorf("creating temp file: %w", err)}
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on any error
	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return &ParseError{Path: path, Err: fmt.Errorf("writing temp file: %w", err)}
	}

	if err := tmpFile.Close(); err != nil {
		return &ParseError{Path: path, Err: fmt.Errorf("closing temp file: %w", err)}
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return &ParseError{Path: path, Err: fmt.Errorf("renaming temp file: %w", err)}
	}

	// Clear tmpPath so defer doesn't try to remove the renamed file
	tmpPath = ""
	return nil
}
