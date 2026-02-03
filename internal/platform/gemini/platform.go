package gemini

import (
	"os"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/paths"
)

// GeminiPlatform provides the unified platform adapter for Gemini CLI.
type GeminiPlatform struct {
	paths    *GeminiPaths
	skills   *SkillManager
	commands *CommandManager
	agents   *AgentManager
	mcp      *MCPManager
}

// Option configures a GeminiPlatform instance.
type Option func(*GeminiPlatform)

// WithScope sets the scope (user or project) for path resolution.
func WithScope(scope Scope) Option {
	return func(p *GeminiPlatform) {
		p.paths = NewGeminiPaths(scope, p.paths.projectRoot)
	}
}

// WithProjectRoot sets the project root directory.
func WithProjectRoot(root string) Option {
	return func(p *GeminiPlatform) {
		p.paths = NewGeminiPaths(p.paths.scope, root)
	}
}

// NewGeminiPlatform creates a new GeminiPlatform with the given options.
func NewGeminiPlatform(opts ...Option) *GeminiPlatform {
	p := &GeminiPlatform{
		paths: NewGeminiPaths(ScopeUser, ""),
	}

	for _, opt := range opts {
		opt(p)
	}

	p.skills = NewSkillManager(p.paths)
	p.commands = NewCommandManager(p.paths)
	p.agents = NewAgentManager(p.paths)
	p.mcp = NewMCPManager(p.paths)

	return p
}

func (p *GeminiPlatform) Name() string {
	return "gemini"
}

func (p *GeminiPlatform) DisplayName() string {
	return "Gemini CLI"
}

// Path Methods

func (p *GeminiPlatform) GlobalConfigDir() string {
	return paths.GlobalConfigDir(paths.PlatformGemini)
}

func (p *GeminiPlatform) ProjectConfigDir(projectRoot string) string {
	return paths.ProjectConfigDir(paths.PlatformGemini, projectRoot)
}

func (p *GeminiPlatform) SkillDir() string {
	return p.paths.SkillDir()
}

func (p *GeminiPlatform) CommandDir() string {
	return p.paths.CommandDir()
}

func (p *GeminiPlatform) AgentDir() string {
	return p.paths.AgentDir()
}

func (p *GeminiPlatform) MCPConfigPath() string {
	return p.paths.MCPConfigPath()
}

func (p *GeminiPlatform) InstructionsPath(projectRoot string) string {
	if projectRoot != "" {
		projectPaths := NewGeminiPaths(ScopeProject, projectRoot)
		return projectPaths.InstructionsPath()
	}
	return p.paths.InstructionsPath()
}

// Skill Operations

func (p *GeminiPlatform) InstallSkill(s *Skill) error {
	return p.skills.Install(s)
}

func (p *GeminiPlatform) UninstallSkill(name string) error {
	return p.skills.Uninstall(name)
}

func (p *GeminiPlatform) ListSkills() ([]*Skill, error) {
	return p.skills.List()
}

func (p *GeminiPlatform) GetSkill(name string) (*Skill, error) {
	return p.skills.Get(name)
}

// Command Operations

func (p *GeminiPlatform) InstallCommand(c *Command) error {
	return p.commands.Install(c)
}

func (p *GeminiPlatform) UninstallCommand(name string) error {
	return p.commands.Uninstall(name)
}

func (p *GeminiPlatform) ListCommands() ([]*Command, error) {
	return p.commands.List()
}

func (p *GeminiPlatform) GetCommand(name string) (*Command, error) {
	return p.commands.Get(name)
}

// Agent Operations

func (p *GeminiPlatform) InstallAgent(a *Agent) error {
	// Automatically enable agents in settings if not already enabled
	if err := p.EnableAgents(); err != nil {
		return errors.Wrap(err, "enabling agents in settings")
	}
	return p.agents.Install(a)
}

func (p *GeminiPlatform) UninstallAgent(name string) error {
	return p.agents.Uninstall(name)
}

func (p *GeminiPlatform) ListAgents() ([]*Agent, error) {
	return p.agents.List()
}

func (p *GeminiPlatform) GetAgent(name string) (*Agent, error) {
	return p.agents.Get(name)
}

func (p *GeminiPlatform) EnableAgents() error {
	settings, err := p.mcp.loadSettings()
	if err != nil {
		return err
	}

	if settings.Experimental == nil {
		settings.Experimental = &ExperimentalConfig{}
	}

	if settings.Experimental.EnableAgents {
		return nil // Already enabled
	}

	settings.Experimental.EnableAgents = true
	return p.mcp.saveSettings(settings)
}

// MCP Operations

func (p *GeminiPlatform) AddMCP(s *MCPServer) error {
	return p.mcp.Add(s)
}

func (p *GeminiPlatform) RemoveMCP(name string) error {
	return p.mcp.Remove(name)
}

func (p *GeminiPlatform) ListMCP() ([]*MCPServer, error) {
	return p.mcp.List()
}

func (p *GeminiPlatform) GetMCP(name string) (*MCPServer, error) {
	return p.mcp.Get(name)
}

func (p *GeminiPlatform) EnableMCP(name string) error {
	return p.mcp.Enable(name)
}

func (p *GeminiPlatform) DisableMCP(name string) error {
	return p.mcp.Disable(name)
}

// Translation Methods

func (p *GeminiPlatform) TranslateVariables(content string) string {
	return TranslateVariables(content)
}

func (p *GeminiPlatform) TranslateToCanonical(content string) string {
	return TranslateToCanonical(content)
}

func (p *GeminiPlatform) ValidateVariables(content string) error {
	return ValidateVariables(content)
}

// Backup Paths

func (p *GeminiPlatform) BackupPaths() []string {
	return []string{
		p.paths.MCPConfigPath(),
		p.paths.BaseDir(),
	}
}

// IsLocalConfigIgnored checks if the local configuration is ignored by VCS.
// Gemini CLI doesn't currently support local scope, so this always returns true.
func (p *GeminiPlatform) IsLocalConfigIgnored() (bool, error) {
	return true, nil
}

// Status Methods

func (p *GeminiPlatform) IsAvailable() bool {
	globalDir := p.GlobalConfigDir()
	if globalDir == "" {
		return false
	}
	info, err := os.Stat(globalDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (p *GeminiPlatform) Version() (string, error) {
	return "", nil
}
