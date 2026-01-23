package claude

import (
	"os"

	"github.com/thoreinstein/aix/internal/paths"
)

// ClaudePlatform provides the unified platform adapter for Claude Code.
// It aggregates all Claude-specific managers and provides a consistent
// interface for the aix CLI.
type ClaudePlatform struct {
	paths    *ClaudePaths
	skills   *SkillManager
	commands *CommandManager
	agents   *AgentManager
	mcp      *MCPManager
}

// Option configures a ClaudePlatform instance.
type Option func(*ClaudePlatform)

// WithScope sets the scope (user or project) for path resolution.
func WithScope(scope Scope) Option {
	return func(p *ClaudePlatform) {
		p.paths = NewClaudePaths(scope, p.paths.projectRoot)
	}
}

// WithProjectRoot sets the project root directory for project-scoped paths.
func WithProjectRoot(root string) Option {
	return func(p *ClaudePlatform) {
		p.paths = NewClaudePaths(p.paths.scope, root)
	}
}

// NewClaudePlatform creates a new ClaudePlatform with the given options.
// Default configuration uses ScopeUser with no project root.
func NewClaudePlatform(opts ...Option) *ClaudePlatform {
	// Initialize with defaults
	p := &ClaudePlatform{
		paths: NewClaudePaths(ScopeUser, ""),
	}

	// Apply options
	for _, opt := range opts {
		opt(p)
	}

	// Initialize managers with the configured paths
	p.skills = NewSkillManager(p.paths)
	p.commands = NewCommandManager(p.paths)
	p.agents = NewAgentManager(p.paths)
	p.mcp = NewMCPManager(p.paths)

	return p
}

// Name returns the platform identifier.
func (p *ClaudePlatform) Name() string {
	return "claude"
}

// DisplayName returns a human-readable platform name.
func (p *ClaudePlatform) DisplayName() string {
	return "Claude Code"
}

// --- Path Methods ---

// GlobalConfigDir returns the global configuration directory (~/.claude/).
func (p *ClaudePlatform) GlobalConfigDir() string {
	return paths.GlobalConfigDir(paths.PlatformClaude)
}

// ProjectConfigDir returns the project-scoped configuration directory.
func (p *ClaudePlatform) ProjectConfigDir(projectRoot string) string {
	return paths.ProjectConfigDir(paths.PlatformClaude, projectRoot)
}

// SkillDir returns the skills directory for the current scope.
func (p *ClaudePlatform) SkillDir() string {
	return p.paths.SkillDir()
}

// CommandDir returns the commands directory for the current scope.
func (p *ClaudePlatform) CommandDir() string {
	return p.paths.CommandDir()
}

// AgentDir returns the agents directory for the current scope.
func (p *ClaudePlatform) AgentDir() string {
	return p.paths.AgentDir()
}

// MCPConfigPath returns the path to the MCP servers configuration file.
func (p *ClaudePlatform) MCPConfigPath() string {
	return p.paths.MCPConfigPath()
}

// InstructionsPath returns the path to the instructions file.
// For user scope, this is ~/.claude/CLAUDE.md.
// For project scope, this is <projectRoot>/CLAUDE.md.
func (p *ClaudePlatform) InstructionsPath(projectRoot string) string {
	if projectRoot != "" {
		// Use project-scoped path
		projectPaths := NewClaudePaths(ScopeProject, projectRoot)
		return projectPaths.InstructionsPath()
	}
	return p.paths.InstructionsPath()
}

// --- Skill Operations ---

// InstallSkill installs a skill to the skill directory.
func (p *ClaudePlatform) InstallSkill(s *Skill) error {
	return p.skills.Install(s)
}

// UninstallSkill removes a skill by name.
func (p *ClaudePlatform) UninstallSkill(name string) error {
	return p.skills.Uninstall(name)
}

// ListSkills returns all installed skills.
func (p *ClaudePlatform) ListSkills() ([]*Skill, error) {
	return p.skills.List()
}

// GetSkill retrieves a skill by name.
func (p *ClaudePlatform) GetSkill(name string) (*Skill, error) {
	return p.skills.Get(name)
}

// --- Command Operations ---

// InstallCommand installs a slash command.
func (p *ClaudePlatform) InstallCommand(c *Command) error {
	return p.commands.Install(c)
}

// UninstallCommand removes a command by name.
func (p *ClaudePlatform) UninstallCommand(name string) error {
	return p.commands.Uninstall(name)
}

// ListCommands returns all installed commands.
func (p *ClaudePlatform) ListCommands() ([]*Command, error) {
	return p.commands.List()
}

// GetCommand retrieves a command by name.
func (p *ClaudePlatform) GetCommand(name string) (*Command, error) {
	return p.commands.Get(name)
}

// --- Agent Operations ---

// InstallAgent installs an agent.
func (p *ClaudePlatform) InstallAgent(a *Agent) error {
	return p.agents.Install(a)
}

// UninstallAgent removes an agent by name.
func (p *ClaudePlatform) UninstallAgent(name string) error {
	return p.agents.Uninstall(name)
}

// ListAgents returns all installed agents.
func (p *ClaudePlatform) ListAgents() ([]*Agent, error) {
	return p.agents.List()
}

// GetAgent retrieves an agent by name.
func (p *ClaudePlatform) GetAgent(name string) (*Agent, error) {
	return p.agents.Get(name)
}

// --- MCP Operations ---

// AddMCP adds an MCP server configuration.
func (p *ClaudePlatform) AddMCP(s *MCPServer) error {
	return p.mcp.Add(s)
}

// RemoveMCP removes an MCP server by name.
func (p *ClaudePlatform) RemoveMCP(name string) error {
	return p.mcp.Remove(name)
}

// ListMCP returns all configured MCP servers.
func (p *ClaudePlatform) ListMCP() ([]*MCPServer, error) {
	return p.mcp.List()
}

// GetMCP retrieves an MCP server by name.
func (p *ClaudePlatform) GetMCP(name string) (*MCPServer, error) {
	return p.mcp.Get(name)
}

// EnableMCP enables an MCP server by name.
func (p *ClaudePlatform) EnableMCP(name string) error {
	return p.mcp.Enable(name)
}

// DisableMCP disables an MCP server by name.
func (p *ClaudePlatform) DisableMCP(name string) error {
	return p.mcp.Disable(name)
}

// --- Translation Methods ---

// TranslateVariables converts canonical variable syntax to Claude Code format.
// Since Claude Code uses the canonical format, this is a pass-through.
func (p *ClaudePlatform) TranslateVariables(content string) string {
	return TranslateVariables(content)
}

// TranslateToCanonical converts Claude Code variable syntax to canonical format.
// Since Claude Code uses the canonical format, this is a pass-through.
func (p *ClaudePlatform) TranslateToCanonical(content string) string {
	return TranslateToCanonical(content)
}

// ValidateVariables checks if content contains only supported variables.
func (p *ClaudePlatform) ValidateVariables(content string) error {
	return ValidateVariables(content)
}

// --- Backup Methods ---

// BackupPaths returns all config files/directories that should be backed up.
// For Claude Code, this includes:
//   - ~/.claude.json (MCP config)
//   - ~/.claude/ directory (skills, commands, agents)
func (p *ClaudePlatform) BackupPaths() []string {
	return []string{
		p.paths.MCPConfigPath(),
		p.paths.BaseDir(),
	}
}

// --- Status Methods ---

// IsAvailable checks if Claude Code is available on this system.
// Returns true if the ~/.claude/ directory exists.
func (p *ClaudePlatform) IsAvailable() bool {
	globalDir := paths.GlobalConfigDir(paths.PlatformClaude)
	if globalDir == "" {
		return false
	}
	info, err := os.Stat(globalDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Version returns the Claude Code version.
// Currently returns an empty string as version detection is not yet implemented.
func (p *ClaudePlatform) Version() (string, error) {
	// TODO: Implement version detection by running claude --version or similar
	return "", nil
}
