package claude

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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
		return nil, fmt.Errorf("reading agents directory: %w", err)
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
		agent, err := m.Get(name)
		if err != nil {
			return nil, fmt.Errorf("reading agent %q: %w", name, err)
		}
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
		return fmt.Errorf("creating agents directory: %w", err)
	}

	content := formatAgentContent(a)
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

// parseAgentContent parses markdown content with optional YAML frontmatter.
// If frontmatter is present (delimited by ---), it's parsed for metadata.
// The remaining content becomes Instructions.
func parseAgentContent(data []byte) (*Agent, error) {
	content := string(data)
	agent := &Agent{}

	// Check for frontmatter delimiter
	if !strings.HasPrefix(content, "---\n") {
		// No frontmatter, entire content is instructions
		agent.Instructions = strings.TrimSpace(content)
		return agent, nil
	}

	// Find the closing delimiter
	rest := content[4:] // Skip opening "---\n"

	// Handle empty frontmatter case (---\n--- immediately)
	if strings.HasPrefix(rest, "---\n") || strings.HasPrefix(rest, "---") && (len(rest) == 3 || rest[3] == '\n') {
		// Empty frontmatter, skip the closing ---
		afterClose := strings.TrimPrefix(rest, "---")
		afterClose = strings.TrimPrefix(afterClose, "\n")
		agent.Instructions = strings.TrimSpace(afterClose)
		return agent, nil
	}

	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		// No closing delimiter, treat as no frontmatter
		agent.Instructions = strings.TrimSpace(content)
		return agent, nil
	}

	// Parse frontmatter
	frontmatter := rest[:endIdx]
	if err := yaml.Unmarshal([]byte(frontmatter), agent); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	// Extract instructions (skip closing --- and newline)
	instructionsStart := endIdx + 4 // len("\n---")
	if instructionsStart < len(rest) {
		instructions := rest[instructionsStart:]
		// Skip leading newline if present
		instructions = strings.TrimPrefix(instructions, "\n")
		agent.Instructions = strings.TrimSpace(instructions)
	}

	return agent, nil
}

// formatAgentContent formats an agent as markdown with optional frontmatter.
// Only includes frontmatter if Description is set.
func formatAgentContent(a *Agent) string {
	var buf bytes.Buffer

	// Only include frontmatter if there's a description
	if a.Description != "" {
		buf.WriteString("---\n")
		buf.WriteString("description: ")
		buf.WriteString(a.Description)
		buf.WriteString("\n---\n\n")
	}

	buf.WriteString(a.Instructions)

	// Ensure trailing newline
	if !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteString("\n")
	}

	return buf.String()
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
		return nil, fmt.Errorf("reading agents directory: %w", err)
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
