package opencode

import (
	"bytes"
	"io/fs"
	"os"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Sentinel errors for command operations.
var (
	// ErrCommandNotFound indicates the requested command does not exist.
	ErrCommandNotFound = errors.New("command not found")

	// ErrInvalidCommand indicates the command is missing required fields.
	ErrInvalidCommand = errors.New("invalid command: name required")
)

// CommandManager provides CRUD operations for OpenCode slash commands.
// Commands are stored as markdown files in the commands directory.
type CommandManager struct {
	paths *OpenCodePaths
}

// NewCommandManager creates a new CommandManager with the given paths configuration.
func NewCommandManager(paths *OpenCodePaths) *CommandManager {
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
		return nil, errors.Wrap(err, "reading commands directory")
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
		cmdPath := m.paths.CommandPath(name)

		f, err := os.Open(cmdPath)
		if err != nil {
			return nil, errors.Wrapf(err, "opening command file %q", name)
		}

		cmd := &Command{Name: name}
		if err := frontmatter.ParseHeader(f, cmd); err != nil {
			f.Close()
			return nil, errors.Wrapf(err, "parsing command header %q", name)
		}
		f.Close()

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
		return nil, errors.Wrap(err, "reading command file")
	}

	cmd, err := parseCommandFile(data)
	if err != nil {
		return nil, errors.Wrap(err, "parsing command file")
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
		return errors.Wrap(err, "creating commands directory")
	}

	content, err := formatCommandFile(c)
	if err != nil {
		return errors.Wrap(err, "formatting command content")
	}

	cmdPath := m.paths.CommandPath(c.Name)
	if err := os.WriteFile(cmdPath, []byte(content), 0o644); err != nil {
		return errors.Wrap(err, "writing command file")
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
		return errors.Wrap(err, "removing command file")
	}

	return nil
}

// parseCommandFile parses a command markdown file.
// Supports optional YAML frontmatter delimited by "---".
// If no frontmatter is present, the entire content is treated as Instructions.
func parseCommandFile(data []byte) (*Command, error) {
	cmd := &Command{}

	// Parse with optional frontmatter
	body, err := frontmatter.Parse(bytes.NewReader(data), cmd)
	if err != nil {
		return nil, errors.Wrap(err, "parsing frontmatter")
	}

	cmd.Instructions = strings.TrimSpace(string(body))
	return cmd, nil
}

// formatCommandFile formats a Command as a markdown file with optional frontmatter.
// Includes frontmatter only if Description is non-empty.
func formatCommandFile(c *Command) (string, error) {
	// Only include frontmatter if there's metadata to write
	if c.Description == "" {
		res := c.Instructions
		if !strings.HasSuffix(res, "\n") {
			res += "\n"
		}
		return res, nil
	}

	meta := struct {
		Description string `yaml:"description"`
	}{
		Description: c.Description,
	}

	data, err := frontmatter.Format(meta, c.Instructions)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
