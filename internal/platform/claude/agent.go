package claude

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/thoreinstein/aix/pkg/fileutil"
	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Sentinel errors for agent operations.
var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrInvalidAgent       = errors.New("invalid agent: name required")
	ErrAgentDirUnresolved = errors.New("cannot determine agents directory")
)

// AgentManager handles CRUD operations for Claude Code agents.
// Agents are markdown files stored in the agents directory.
type AgentManager struct {
	paths *ClaudePaths
}

// NewAgentManager creates a new AgentManager with the given path resolver.
func NewAgentManager(paths *ClaudePaths) *AgentManager {
	return &AgentManager{paths: paths}
}

// List returns all agents in the agents directory.
// Returns an empty slice if the directory doesn't exist.
func (m *AgentManager) List() ([]*Agent, error) {
	agentDir := m.paths.AgentDir()
	if agentDir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(agentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "reading agents directory")
	}

	// Count markdown files for pre-allocation
	mdCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			mdCount++
		}
	}

	agents := make([]*Agent, 0, mdCount)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		agentPath := m.paths.AgentPath(name)

		f, err := os.Open(agentPath)
		if err != nil {
			return nil, errors.Wrapf(err, "opening agent file %q", name)
		}

		agent := &Agent{Name: name}
		if err := frontmatter.ParseHeader(f, agent); err != nil {
			f.Close()
			return nil, errors.Wrapf(err, "parsing agent header %q", name)
		}
		f.Close()

		agents = append(agents, agent)
	}

	return agents, nil
}

// Get retrieves an agent by name.
// Returns ErrAgentNotFound if the agent file doesn't exist.
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
		if os.IsNotExist(err) {
			return nil, ErrAgentNotFound
		}
		return nil, errors.Wrap(err, "reading agent file")
	}

	agent, err := parseAgentContent(data)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing agent %q", name)
	}

	agent.Name = name
	return agent, nil
}

// Install writes an agent to the agents directory.
// Creates the agents directory if it doesn't exist.
// Overwrites any existing agent with the same name.
func (m *AgentManager) Install(a *Agent) error {
	if a == nil || a.Name == "" {
		return ErrInvalidAgent
	}

	agentDir := m.paths.AgentDir()
	if agentDir == "" {
		return ErrAgentDirUnresolved
	}

	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		return errors.Wrap(err, "creating agents directory")
	}

	content, err := formatAgentContent(a)
	if err != nil {
		return errors.Wrap(err, "formatting agent content")
	}
	agentPath := m.paths.AgentPath(a.Name)

	if err := fileutil.AtomicWriteFile(agentPath, []byte(content), 0o644); err != nil {
		return errors.Wrap(err, "writing agent file")
	}

	return nil
}

// Uninstall removes an agent file.
// Returns nil if the agent doesn't exist (idempotent).
func (m *AgentManager) Uninstall(name string) error {
	if name == "" {
		return ErrInvalidAgent
	}

	agentPath := m.paths.AgentPath(name)
	if agentPath == "" {
		return nil
	}

	err := os.Remove(agentPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "removing agent file")
	}

	return nil
}

// parseAgentContent parses markdown content with optional YAML frontmatter.
// If frontmatter is present (delimited by ---), it's parsed for metadata.
// The remaining content becomes Instructions.
func parseAgentContent(data []byte) (*Agent, error) {
	agent := &Agent{}

	// Parse with optional frontmatter
	body, err := frontmatter.Parse(bytes.NewReader(data), agent)
	if err != nil {
		return nil, errors.Wrap(err, "parsing frontmatter")
	}

	agent.Instructions = strings.TrimSpace(string(body))
	return agent, nil
}

// formatAgentContent formats an agent as markdown with optional frontmatter.
// Only includes frontmatter if Description is set.
func formatAgentContent(a *Agent) (string, error) {
	// Only include frontmatter if there's a description
	if a.Description == "" {
		res := a.Instructions
		if !strings.HasSuffix(res, "\n") {
			res += "\n"
		}
		return res, nil
	}

	meta := struct {
		Description string `yaml:"description"`
	}{
		Description: a.Description,
	}

	data, err := frontmatter.Format(meta, a.Instructions)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// AgentDir returns the agents directory path.
// This is a convenience method that delegates to the underlying ClaudePaths.
func (m *AgentManager) AgentDir() string {
	return m.paths.AgentDir()
}

// AgentPath returns the path to a specific agent file.
// This is a convenience method that delegates to the underlying ClaudePaths.
func (m *AgentManager) AgentPath(name string) string {
	return m.paths.AgentPath(name)
}

// Exists checks if an agent with the given name exists.
func (m *AgentManager) Exists(name string) bool {
	if name == "" {
		return false
	}

	agentPath := m.paths.AgentPath(name)
	if agentPath == "" {
		return false
	}

	_, err := os.Stat(agentPath)
	return err == nil
}

// Names returns a list of all agent names in the agents directory.
// Returns an empty slice if the directory doesn't exist.
func (m *AgentManager) Names() ([]string, error) {
	agentDir := m.paths.AgentDir()
	if agentDir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(agentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "reading agents directory")
	}

	// Count markdown files for pre-allocation
	mdCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			mdCount++
		}
	}

	names := make([]string, 0, mdCount)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".md" {
			names = append(names, strings.TrimSuffix(name, ".md"))
		}
	}

	return names, nil
}
