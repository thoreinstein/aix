package gemini

import (
	"bytes"
	"io/fs"
	"os"
	"strings"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/pkg/fileutil"
	"github.com/thoreinstein/aix/pkg/frontmatter"
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

	agentCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			agentCount++
		}
	}

	agents := make([]*Agent, 0, agentCount)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		agent, err := m.Get(name)
		if err != nil {
			return nil, errors.Wrapf(err, "loading agent %q", name)
		}

		agents = append(agents, agent)
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

	var meta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}

	body, err := frontmatter.Parse(bytes.NewReader(data), &meta)
	if err != nil {
		return nil, errors.Wrap(err, "parsing agent frontmatter")
	}

	if meta.Name == "" {
		meta.Name = name
	}

	return &Agent{
		Name:         meta.Name,
		Description:  meta.Description,
		Instructions: strings.TrimSpace(string(body)),
	}, nil
}

// Install writes an agent to disk in Markdown format with YAML frontmatter.
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

	// Construct Markdown with YAML frontmatter
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("name: " + a.Name + "\n")
	if a.Description != "" {
		buf.WriteString("description: " + a.Description + "\n")
	}
	buf.WriteString("---\n\n")
	buf.WriteString(a.Instructions)

	agentPath := m.paths.AgentPath(a.Name)
	if err := fileutil.AtomicWriteFile(agentPath, buf.Bytes(), 0o644); err != nil {
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
