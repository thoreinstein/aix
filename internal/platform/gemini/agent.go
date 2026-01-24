package gemini

import (
	"io/fs"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/pkg/fileutil"
)

// Sentinel errors for agent operations.
var (
	ErrAgentNotFound = errors.New("agent not found")
	ErrInvalidAgent  = errors.New("invalid agent: name required")
)

// AgentManager provides CRUD operations for Gemini CLI agents.
type AgentManager struct {
	paths *GeminiPaths
}

// NewAgentManager creates a new AgentManager with the given paths configuration.
func NewAgentManager(paths *GeminiPaths) *AgentManager {
	return &AgentManager{
		paths: paths,
	}
}

// List returns all agents in the agents directory.
func (m *AgentManager) List() ([]*Agent, error) {
	agentDir := m.paths.AgentDir()
	if agentDir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(agentDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "reading agents directory")
	}

	tomlCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".toml") {
			tomlCount++
		}
	}

	agents := make([]*Agent, 0, tomlCount)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".toml")
		agentPath := m.paths.AgentPath(name)

		data, err := os.ReadFile(agentPath)
		if err != nil {
			return nil, errors.Wrapf(err, "reading agent file %q", name)
		}

		var agent Agent
		if err := toml.Unmarshal(data, &agent); err != nil {
			return nil, errors.Wrapf(err, "unmarshaling agent %q", name)
		}

		if agent.Name == "" {
			agent.Name = name
		}

		agents = append(agents, &agent)
	}

	return agents, nil
}

// Get retrieves an agent by name.
func (m *AgentManager) Get(name string) (*Agent, error) {
	if name == "" {
		return nil, ErrInvalidAgent
	}

	agentPath := m.paths.AgentPath(name)
	if agentPath == "" {
		return nil, ErrAgentNotFound
	}

	data, err := os.ReadFile(agentPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrAgentNotFound
		}
		return nil, errors.Wrap(err, "reading agent file")
	}

	var agent Agent
	if err := toml.Unmarshal(data, &agent); err != nil {
		return nil, errors.Wrap(err, "unmarshaling agent")
	}

	if agent.Name == "" {
		agent.Name = name
	}

	return &agent, nil
}

// Install writes an agent to disk in TOML format.
func (m *AgentManager) Install(a *Agent) error {
	if a == nil || a.Name == "" {
		return ErrInvalidAgent
	}

	agentDir := m.paths.AgentDir()
	if agentDir == "" {
		return errors.New("agent directory path is empty")
	}

	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		return errors.Wrap(err, "creating agents directory")
	}

	data, err := toml.Marshal(a)
	if err != nil {
		return errors.Wrap(err, "marshaling agent to TOML")
	}

	agentPath := m.paths.AgentPath(a.Name)
	if err := fileutil.AtomicWriteFile(agentPath, data, 0o644); err != nil {
		return errors.Wrap(err, "writing agent file")
	}

	return nil
}

// Uninstall removes an agent from disk.
func (m *AgentManager) Uninstall(name string) error {
	if name == "" {
		return ErrInvalidAgent
	}

	agentPath := m.paths.AgentPath(name)
	if agentPath == "" {
		return nil
	}

	if err := os.Remove(agentPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return errors.Wrap(err, "removing agent file")
	}

	return nil
}
