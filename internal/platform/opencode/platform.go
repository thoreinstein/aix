package opencode

import (
	"os"

	"github.com/thoreinstein/aix/internal/paths"
)

// OpenCodePlatform provides the unified platform adapter for OpenCode.
// It aggregates all OpenCode-specific managers and provides a consistent
// interface for the aix CLI.
type OpenCodePlatform struct {
	paths    *OpenCodePaths
	skills   *SkillManager
	commands *CommandManager
	agents   *AgentManager
	mcp      *MCPManager
}

// Option configures an OpenCodePlatform instance.
type Option func(*OpenCodePlatform)

// WithScope sets the scope (user or project) for path resolution.
func WithScope(scope Scope) Option {
	return func(p *OpenCodePlatform) {
		p.paths = NewOpenCodePaths(scope, p.paths.projectRoot)
	}
}

// WithProjectRoot sets the project root directory for project-scoped paths.
func WithProjectRoot(root string) Option {
	return func(p *OpenCodePlatform) {
		p.paths = NewOpenCodePaths(p.paths.scope, root)
	}
}

// NewOpenCodePlatform creates a new OpenCodePlatform with the given options.
// Default configuration uses ScopeUser with no project root.
func NewOpenCodePlatform(opts ...Option) *OpenCodePlatform {
	// Initialize with defaults
	p := &OpenCodePlatform{
		paths: NewOpenCodePaths(ScopeUser, ""),
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
func (p *OpenCodePlatform) Name() string {
	return "opencode"
}

// DisplayName returns a human-readable platform name.
func (p *OpenCodePlatform) DisplayName() string {
	return "OpenCode"
}

// --- Path Methods ---

// GlobalConfigDir returns the global configuration directory (~/.config/opencode/).
func (p *OpenCodePlatform) GlobalConfigDir() string {
	return paths.GlobalConfigDir(paths.PlatformOpenCode)
}

// ProjectConfigDir returns the project-scoped configuration directory.
func (p *OpenCodePlatform) ProjectConfigDir(projectRoot string) string {
	return paths.ProjectConfigDir(paths.PlatformOpenCode, projectRoot)
}

// SkillDir returns the skills directory for the current scope.
func (p *OpenCodePlatform) SkillDir() string {
	return p.paths.SkillDir()
}

// CommandDir returns the commands directory for the current scope.
func (p *OpenCodePlatform) CommandDir() string {
	return p.paths.CommandDir()
}

// AgentDir returns the agents directory for the current scope.
func (p *OpenCodePlatform) AgentDir() string {
	return p.paths.AgentDir()
}

// MCPConfigPath returns the path to the MCP servers configuration file.
func (p *OpenCodePlatform) MCPConfigPath() string {
	return p.paths.MCPConfigPath()
}

// InstructionsPath returns the path to the instructions file.
// For user scope, this is ~/.config/opencode/AGENTS.md.
// For project scope, this is <projectRoot>/AGENTS.md.
func (p *OpenCodePlatform) InstructionsPath(projectRoot string) string {
	if projectRoot != "" {
		// Use project-scoped path
		projectPaths := NewOpenCodePaths(ScopeProject, projectRoot)
		return projectPaths.InstructionsPath()
	}
	return p.paths.InstructionsPath()
}

// --- Skill Operations ---

// InstallSkill installs a skill to the skill directory.
func (p *OpenCodePlatform) InstallSkill(s *Skill) error {
	return p.skills.Install(s)
}

// UninstallSkill removes a skill by name.
func (p *OpenCodePlatform) UninstallSkill(name string) error {
	return p.skills.Uninstall(name)
}

// ListSkills returns all installed skills.
func (p *OpenCodePlatform) ListSkills() ([]*Skill, error) {
	return p.skills.List()
}

// GetSkill retrieves a skill by name.
func (p *OpenCodePlatform) GetSkill(name string) (*Skill, error) {
	return p.skills.Get(name)
}

// --- Command Operations ---

// InstallCommand installs a slash command.
func (p *OpenCodePlatform) InstallCommand(c *Command) error {
	return p.commands.Install(c)
}

// UninstallCommand removes a command by name.
func (p *OpenCodePlatform) UninstallCommand(name string) error {
	return p.commands.Uninstall(name)
}

// ListCommands returns all installed commands.
func (p *OpenCodePlatform) ListCommands() ([]*Command, error) {
	return p.commands.List()
}

// GetCommand retrieves a command by name.
func (p *OpenCodePlatform) GetCommand(name string) (*Command, error) {
	return p.commands.Get(name)
}

// --- Agent Operations ---

// InstallAgent installs an agent.
func (p *OpenCodePlatform) InstallAgent(a *Agent) error {
	return p.agents.Install(a)
}

// UninstallAgent removes an agent by name.
func (p *OpenCodePlatform) UninstallAgent(name string) error {
	return p.agents.Uninstall(name)
}

// ListAgents returns all installed agents.
func (p *OpenCodePlatform) ListAgents() ([]*Agent, error) {
	return p.agents.List()
}

// GetAgent retrieves an agent by name.
func (p *OpenCodePlatform) GetAgent(name string) (*Agent, error) {
	return p.agents.Get(name)
}

// --- MCP Operations ---

// AddMCP adds an MCP server configuration.
func (p *OpenCodePlatform) AddMCP(s *MCPServer) error {
	return p.mcp.Add(s)
}

// RemoveMCP removes an MCP server by name.
func (p *OpenCodePlatform) RemoveMCP(name string) error {
	return p.mcp.Remove(name)
}

// ListMCP returns all configured MCP servers.
func (p *OpenCodePlatform) ListMCP() ([]*MCPServer, error) {
	return p.mcp.List()
}

// GetMCP retrieves an MCP server by name.
func (p *OpenCodePlatform) GetMCP(name string) (*MCPServer, error) {
	return p.mcp.Get(name)
}

// EnableMCP enables an MCP server by name.
func (p *OpenCodePlatform) EnableMCP(name string) error {
	return p.mcp.Enable(name)
}

// DisableMCP disables an MCP server by name.
func (p *OpenCodePlatform) DisableMCP(name string) error {
	return p.mcp.Disable(name)
}

// --- Translation Methods ---

// TranslateVariables converts canonical variable syntax to OpenCode format.
// Since OpenCode uses the canonical format, this is a pass-through.
func (p *OpenCodePlatform) TranslateVariables(content string) string {
	return TranslateVariables(content)
}

// TranslateToCanonical converts OpenCode variable syntax to canonical format.
// Since OpenCode uses the canonical format, this is a pass-through.
func (p *OpenCodePlatform) TranslateToCanonical(content string) string {
	return TranslateToCanonical(content)
}

// ValidateVariables checks if content contains only supported variables.
func (p *OpenCodePlatform) ValidateVariables(content string) error {
	return ValidateVariables(content)
}

// --- Backup Methods ---

// BackupPaths returns all config files/directories that should be backed up.
// For OpenCode, this includes:
//   - ~/.config/opencode/opencode.json (MCP config)
//   - ~/.config/opencode/ directory (skills, commands, agents)
func (p *OpenCodePlatform) BackupPaths() []string {
	return []string{
		p.paths.MCPConfigPath(),
		p.paths.BaseDir(),
	}
}

// IsLocalConfigIgnored checks if the local configuration is ignored by VCS.
// OpenCode doesn't currently support local scope, so this always returns true.
func (p *OpenCodePlatform) IsLocalConfigIgnored() (bool, error) {
	return true, nil
}

// --- Status Methods ---

// IsAvailable checks if OpenCode is available on this system.
// Returns true if the ~/.config/opencode/ directory exists.
func (p *OpenCodePlatform) IsAvailable() bool {
	globalDir := paths.GlobalConfigDir(paths.PlatformOpenCode)
	if globalDir == "" {
		return false
	}
	info, err := os.Stat(globalDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Version returns the OpenCode version.
// Currently returns an empty string as version detection is not yet implemented.
func (p *OpenCodePlatform) Version() (string, error) {
	// TODO: Implement version detection by running opencode --version or similar
	return "", nil
}
