package opencode

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thoreinstein/aix/pkg/frontmatter"
)

// Sentinel errors for agent operations.
var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrInvalidAgent       = errors.New("invalid agent: name required")
	ErrAgentDirUnresolved = errors.New("cannot determine agent directory")
)

// AgentManager handles CRUD operations for OpenCode agents.
// Agents are markdown files stored in the agent directory.
type AgentManager struct {
	paths *OpenCodePaths
}

// NewAgentManager creates a new AgentManager with the given path resolver.
func NewAgentManager(paths *OpenCodePaths) *AgentManager {
	return &AgentManager{paths: paths}
}

// List returns all agents in the agent directory.
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
		return nil, fmt.Errorf("reading agent directory: %w", err)
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
			return nil, fmt.Errorf("opening agent file %q: %w", name, err)
		}

		agent := &Agent{Name: name}
		if err := frontmatter.ParseHeader(f, agent); err != nil {
			f.Close()
			return nil, fmt.Errorf("parsing agent header %q: %w", name, err)
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
		return nil, fmt.Errorf("reading agent file: %w", err)
	}

	agent, err := parseAgentContent(data)
	if err != nil {
		return nil, fmt.Errorf("parsing agent %q: %w", name, err)
	}

	agent.Name = name
	return agent, nil
}

// Install writes an agent to the agent directory.
// Creates the agent directory if it doesn't exist.
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
		return fmt.Errorf("creating agent directory: %w", err)
	}

	content, err := formatAgentContent(a)
	if err != nil {
		return fmt.Errorf("formatting agent content: %w", err)
	}
	agentPath := m.paths.AgentPath(a.Name)

	if err := os.WriteFile(agentPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing agent file: %w", err)
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
		return fmt.Errorf("removing agent file: %w", err)
	}

	return nil
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

// Names returns a list of all agent names in the agent directory.
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
		return nil, fmt.Errorf("reading agent directory: %w", err)
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

// AgentDir returns the agent directory path.
// This is a convenience method that delegates to the underlying OpenCodePaths.
func (m *AgentManager) AgentDir() string {
	return m.paths.AgentDir()
}

// AgentPath returns the path to a specific agent file.
// This is a convenience method that delegates to the underlying OpenCodePaths.
func (m *AgentManager) AgentPath(name string) string {
	return m.paths.AgentPath(name)
}

// parseAgentContent parses markdown content with optional YAML frontmatter.
// If frontmatter is present (delimited by ---), it's parsed for metadata.
// The remaining content becomes Instructions.
func parseAgentContent(data []byte) (*Agent, error) {
	agent := &Agent{}

	// Parse with optional frontmatter
	body, err := frontmatter.Parse(bytes.NewReader(data), agent)
	if err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	agent.Instructions = strings.TrimSpace(string(body))
	return agent, nil
}

// agentFrontmatter represents the YAML frontmatter for an OpenCode agent.
// This includes OpenCode-specific fields like Mode and Temperature.
type agentFrontmatter struct {
	Description string  `yaml:"description,omitempty"`
	Mode        string  `yaml:"mode,omitempty"`
	Temperature float64 `yaml:"temperature,omitempty"`
}

// hasFrontmatter returns true if any frontmatter field is set.
func (f *agentFrontmatter) hasFrontmatter() bool {
	return f.Description != "" || f.Mode != "" || f.Temperature != 0
}

// formatAgentContent formats an agent as markdown with optional frontmatter.
// Includes frontmatter if Description, Mode, or Temperature is set.
func formatAgentContent(a *Agent) (string, error) {
	meta := agentFrontmatter{
		Description: a.Description,
		Mode:        a.Mode,
		Temperature: a.Temperature,
	}

	// Only include frontmatter if there's metadata to include
	if !meta.hasFrontmatter() {
		res := a.Instructions
		if !strings.HasSuffix(res, "\n") {
			res += "\n"
		}
		return res, nil
	}

	data, err := frontmatter.Format(meta, a.Instructions)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
