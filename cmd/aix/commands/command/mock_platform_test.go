package command

import (
	"github.com/thoreinstein/aix/internal/cli"
	"github.com/thoreinstein/aix/internal/errors"
)

// mockPlatform implements cli.Platform for testing.
type mockPlatform struct {
	name         string
	displayName  string
	skills       map[string]any
	commands     map[string]any
	agents       map[string]any
	agentErr     error
	uninstallErr error
}

func (m *mockPlatform) Name() string        { return m.name }
func (m *mockPlatform) DisplayName() string { return m.displayName }
func (m *mockPlatform) IsAvailable() bool   { return true }
func (m *mockPlatform) SkillDir() string    { return "/mock/skills" }

func (m *mockPlatform) InstallSkill(_ any, _ cli.Scope) error { return nil }

func (m *mockPlatform) UninstallSkill(name string, _ cli.Scope) error {
	if m.uninstallErr != nil {
		return m.uninstallErr
	}
	delete(m.skills, name)
	return nil
}

func (m *mockPlatform) ListSkills() ([]cli.SkillInfo, error) {
	skills := make([]cli.SkillInfo, 0, len(m.skills))
	for name := range m.skills {
		skills = append(skills, cli.SkillInfo{Name: name})
	}
	return skills, nil
}

func (m *mockPlatform) GetSkill(name string) (any, error) {
	skill, ok := m.skills[name]
	if !ok {
		return nil, errors.New("skill not found")
	}
	return skill, nil
}

func (m *mockPlatform) CommandDir() string { return "/mock/commands" }

func (m *mockPlatform) InstallCommand(_ any, _ cli.Scope) error { return nil }

func (m *mockPlatform) UninstallCommand(name string, _ cli.Scope) error {
	if m.uninstallErr != nil {
		return m.uninstallErr
	}
	delete(m.commands, name)
	return nil
}

func (m *mockPlatform) ListCommands() ([]cli.CommandInfo, error) {
	commands := make([]cli.CommandInfo, 0, len(m.commands))
	for name := range m.commands {
		commands = append(commands, cli.CommandInfo{Name: name})
	}
	return commands, nil
}

func (m *mockPlatform) GetCommand(name string) (any, error) {
	cmd, ok := m.commands[name]
	if !ok {
		return nil, errors.New("command not found")
	}
	return cmd, nil
}

// MCP methods for cli.Platform interface
func (m *mockPlatform) MCPConfigPath() string           { return "/mock/mcp.json" }
func (m *mockPlatform) AddMCP(_ any, _ cli.Scope) error              { return nil }
func (m *mockPlatform) RemoveMCP(_ string, _ cli.Scope) error        { return nil }
func (m *mockPlatform) ListMCP() ([]cli.MCPInfo, error) { return nil, nil }
func (m *mockPlatform) GetMCP(_ string) (any, error)    { return nil, errors.New("not found") }
func (m *mockPlatform) EnableMCP(_ string) error        { return nil }
func (m *mockPlatform) DisableMCP(_ string) error       { return nil }

// Agent methods for cli.Platform interface
func (m *mockPlatform) AgentDir() string { return "/mock/agents" }

func (m *mockPlatform) InstallAgent(_ any, _ cli.Scope) error { return nil }

func (m *mockPlatform) UninstallAgent(name string, _ cli.Scope) error {
	if m.uninstallErr != nil {
		return m.uninstallErr
	}
	delete(m.agents, name)
	return nil
}

func (m *mockPlatform) ListAgents() ([]cli.AgentInfo, error) {
	if m.agentErr != nil {
		return nil, m.agentErr
	}
	agents := make([]cli.AgentInfo, 0, len(m.agents))
	for name := range m.agents {
		agents = append(agents, cli.AgentInfo{Name: name})
	}
	return agents, nil
}

func (m *mockPlatform) GetAgent(name string) (any, error) {
	agent, ok := m.agents[name]
	if !ok {
		return nil, errors.New("agent not found")
	}
	return agent, nil
}

// Backup methods for cli.Platform interface
func (m *mockPlatform) BackupPaths() []string { return []string{"/mock/backup"} }
