package gemini

import (
	"io/fs"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/pkg/fileutil"
)

// Sentinel errors for command operations.
var (
	ErrCommandNotFound = errors.New("command not found")
	ErrInvalidCommand  = errors.New("invalid command: name required")
)

// CommandManager provides CRUD operations for Gemini CLI slash commands.
type CommandManager struct {
	paths *GeminiPaths
}

// NewCommandManager creates a new CommandManager with the given paths configuration.
func NewCommandManager(paths *GeminiPaths) *CommandManager {
	return &CommandManager{
		paths: paths,
	}
}

// List returns all commands in the commands directory.
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

	tomlCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".toml") {
			tomlCount++
		}
	}

	commands := make([]*Command, 0, tomlCount)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".toml")
		cmdPath := m.paths.CommandPath(name)

		data, err := os.ReadFile(cmdPath)
		if err != nil {
			return nil, errors.Wrapf(err, "reading command file %q", name)
		}

		var cmd Command
		if err := toml.Unmarshal(data, &cmd); err != nil {
			return nil, errors.Wrapf(err, "unmarshaling command %q", name)
		}

		// Translate instructions back to canonical format
		cmd.Instructions = TranslateToCanonical(cmd.Instructions)

		if cmd.Name == "" {
			cmd.Name = name
		}

		commands = append(commands, &cmd)
	}

	return commands, nil
}

// Get retrieves a command by name.
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

	var cmd Command
	if err := toml.Unmarshal(data, &cmd); err != nil {
		return nil, errors.Wrap(err, "unmarshaling command")
	}

	cmd.Instructions = TranslateToCanonical(cmd.Instructions)

	if cmd.Name == "" {
		cmd.Name = name
	}

	return &cmd, nil
}

// Install writes a command to disk in TOML format.
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

	// Translate variables to Gemini format
	translatedInstructions := TranslateVariables(c.Instructions)

	// Create a copy to avoid mutating the original
	cmdToInstall := *c
	cmdToInstall.Instructions = translatedInstructions

	data, err := toml.Marshal(cmdToInstall)
	if err != nil {
		return errors.Wrap(err, "marshaling command to TOML")
	}

	// HACK: go-toml/v2 v2.2.4 doesn't seem to respect the 'multiline' tag in this context.
	// As a workaround, we marshal the struct and then manually replace the
	// instructions field if it contains newlines.
	if strings.Contains(cmdToInstall.Instructions, "\n") {
		// This is brittle. It assumes `toml.Marshal` produces a specific format.
		// First, create what the marshaler *should* have produced for just the string.
		singleLineInstructions, _ := toml.Marshal(cmdToInstall.Instructions)

		// Construct the field assignment for a single-line string.
		singleLineField := "prompt = " + string(singleLineInstructions)

		// Construct the field assignment for a multi-line string.
		multiLineField := "prompt = \"\"\"\n" + cmdToInstall.Instructions + "\"\"\""

		// Replace the single-line version with the multi-line version.
		data = []byte(strings.Replace(string(data), singleLineField, multiLineField, 1))
	}

	cmdPath := m.paths.CommandPath(c.Name)
	if err := fileutil.AtomicWriteFile(cmdPath, data, 0o644); err != nil {
		return errors.Wrap(err, "writing command file")
	}

	return nil
}

// Uninstall removes a command from disk.
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
			return nil
		}
		return errors.Wrap(err, "removing command file")
	}

	return nil
}
