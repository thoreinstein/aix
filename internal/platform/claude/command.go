package claude

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Sentinel errors for command operations.
var (
	// ErrCommandNotFound indicates the requested command does not exist.
	ErrCommandNotFound = errors.New("command not found")

	// ErrInvalidCommand indicates the command is missing required fields.
	ErrInvalidCommand = errors.New("invalid command: name required")
)

// CommandManager provides CRUD operations for Claude Code slash commands.
// Commands are stored as markdown files in the commands directory.
type CommandManager struct {
	paths *ClaudePaths
}

// NewCommandManager creates a new CommandManager with the given paths configuration.
func NewCommandManager(paths *ClaudePaths) *CommandManager {
	return &CommandManager{
		paths: paths,
	}
}

// List returns all commands in the commands directory.
// Returns an empty slice if the directory doesn't exist or contains no .md files.
func (m *CommandManager) List() ([]*Command, error) {
	cmdDir := m.paths.CommandDir()
	if cmdDir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading commands directory: %w", err)
	}

	// Count .md files for pre-allocation
	mdCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			mdCount++
		}
	}

	commands := make([]*Command, 0, mdCount)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		cmd, err := m.Get(name)
		if err != nil {
			return nil, fmt.Errorf("reading command %s: %w", name, err)
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// Get retrieves a command by name.
// Returns ErrCommandNotFound if the command file doesn't exist.
func (m *CommandManager) Get(name string) (*Command, error) {
	if name == "" {
		return nil, ErrInvalidCommand
	}

	cmdPath := m.paths.CommandPath(name)
	if cmdPath == "" {
		return nil, ErrCommandNotFound
	}

	data, err := os.ReadFile(cmdPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrCommandNotFound
		}
		return nil, fmt.Errorf("reading command file: %w", err)
	}

	cmd, err := parseCommandFile(data)
	if err != nil {
		return nil, fmt.Errorf("parsing command file: %w", err)
	}

	// Name is derived from filename, not frontmatter
	cmd.Name = name
	return cmd, nil
}

// Install writes a command to disk.
// Creates the commands directory if it doesn't exist.
// Overwrites any existing command with the same name.
func (m *CommandManager) Install(c *Command) error {
	if c == nil || c.Name == "" {
		return ErrInvalidCommand
	}

	cmdDir := m.paths.CommandDir()
	if cmdDir == "" {
		return errors.New("command directory path is empty")
	}

	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		return fmt.Errorf("creating commands directory: %w", err)
	}

	content := formatCommandFile(c)

	cmdPath := m.paths.CommandPath(c.Name)
	if err := os.WriteFile(cmdPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing command file: %w", err)
	}

	return nil
}

// Uninstall removes a command from disk.
// This operation is idempotent; removing a non-existent command returns nil.
func (m *CommandManager) Uninstall(name string) error {
	if name == "" {
		return ErrInvalidCommand
	}

	cmdPath := m.paths.CommandPath(name)
	if cmdPath == "" {
		return nil
	}

	if err := os.Remove(cmdPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // Idempotent: already gone
		}
		return fmt.Errorf("removing command file: %w", err)
	}

	return nil
}

// parseCommandFile parses a command markdown file.
// Supports optional YAML frontmatter delimited by "---".
// If no frontmatter is present, the entire content is treated as Instructions.
func parseCommandFile(data []byte) (*Command, error) {
	content := string(data)
	cmd := &Command{}

	// Check for frontmatter
	if strings.HasPrefix(content, "---\n") || strings.HasPrefix(content, "---\r\n") {
		frontmatter, body, found := extractFrontmatter(content)
		if found {
			if err := yaml.Unmarshal([]byte(frontmatter), cmd); err != nil {
				return nil, fmt.Errorf("parsing frontmatter: %w", err)
			}
			cmd.Instructions = strings.TrimSpace(body)
			return cmd, nil
		}
	}

	// No frontmatter: entire content is instructions
	cmd.Instructions = strings.TrimSpace(content)
	return cmd, nil
}

// extractFrontmatter extracts YAML frontmatter from markdown content.
// Returns the frontmatter content (without delimiters), the body content, and whether frontmatter was found.
func extractFrontmatter(content string) (frontmatter, body string, found bool) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Must start with ---
	if !scanner.Scan() {
		return "", content, false
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return "", content, false
	}

	// Read until closing ---
	var frontmatterBuf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			// Found closing delimiter
			// Rest is body
			var bodyBuf bytes.Buffer
			for scanner.Scan() {
				bodyBuf.WriteString(scanner.Text())
				bodyBuf.WriteString("\n")
			}
			return frontmatterBuf.String(), bodyBuf.String(), true
		}
		frontmatterBuf.WriteString(line)
		frontmatterBuf.WriteString("\n")
	}

	// No closing delimiter found
	return "", content, false
}

// formatCommandFile formats a Command as a markdown file with optional frontmatter.
// Includes frontmatter only if Description is non-empty.
func formatCommandFile(c *Command) string {
	var buf bytes.Buffer

	// Only include frontmatter if there's metadata to write
	if c.Description != "" {
		buf.WriteString("---\n")
		buf.WriteString("description: ")
		buf.WriteString(c.Description)
		buf.WriteString("\n")
		buf.WriteString("---\n\n")
	}

	buf.WriteString(c.Instructions)

	// Ensure file ends with newline
	if !strings.HasSuffix(c.Instructions, "\n") {
		buf.WriteString("\n")
	}

	return buf.String()
}
