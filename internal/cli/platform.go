// Package cli provides CLI-specific types and utilities for the aix command.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/internal/platform"
	"github.com/thoreinstein/aix/internal/platform/claude"
	"github.com/thoreinstein/aix/internal/platform/gemini"
	"github.com/thoreinstein/aix/internal/platform/opencode"
)

// Sentinel errors for platform operations.
var (
	// ErrUnknownPlatform is returned when an unknown platform name is provided.
	ErrUnknownPlatform = errors.New("unknown platform")

	// ErrNoPlatformsAvailable is returned when no platforms are detected.
	ErrNoPlatformsAvailable = errors.New("no platforms available")
)

// SkillInfo provides a simplified view of a skill for CLI display.
// This is a platform-agnostic representation used for listing.
type SkillInfo struct {
	// Name is the skill's unique identifier.
	Name string

	// Description explains what the skill does.
	Description string

	// Source indicates where the skill came from: "local" or a git URL.
	Source string
}

// CommandInfo provides a simplified view of a command for CLI display.
// This is a platform-agnostic representation used for listing.
type CommandInfo struct {
	// Name is the command's identifier (used as /name in the interface).
	Name string

	// Description explains what the command does.
	Description string

	// Source indicates where the command came from: file path or "installed".
	Source string
}

// MCPInfo provides platform-agnostic MCP server information for display.
type MCPInfo struct {
	Name      string
	Transport string // "stdio" or "sse"
	Command   string // Executable path (stdio)
	URL       string // Endpoint (sse)
	Disabled  bool
	Env       map[string]string // Environment variables
}

// AgentInfo provides platform-agnostic agent information for display.
type AgentInfo struct {
	Name        string
	Description string
	Source      string // "local" or future: git URL
}

// Scope defines the configuration layer target (User, Project, Local, Managed).
type Scope int

const (
	// ScopeDefault indicates that the platform should use its default behavior
	// (usually merged view for listing, or precedence-based for getting).
	ScopeDefault Scope = iota
	// ScopeUser targets the user's global home directory configuration.
	ScopeUser
	// ScopeProject targets the project/repository configuration (typically committed).
	ScopeProject
	// ScopeLocal targets local overrides within a project (typically gitignored).
	ScopeLocal
	// ScopeManaged targets system-level managed configuration.
	ScopeManaged
)

func (s Scope) String() string {
	switch s {
	case ScopeDefault:
		return "default"
	case ScopeUser:
		return "user"
	case ScopeProject:
		return "project"
	case ScopeLocal:
		return "local"
	case ScopeManaged:
		return "managed"
	default:
		return "default"
	}
}

// ParseScope converts a string to a Scope. Returns ScopeDefault if empty or invalid.
func ParseScope(s string) Scope {
	switch strings.ToLower(s) {
	case "user":
		return ScopeUser
	case "project":
		return ScopeProject
	case "local":
		return ScopeLocal
	case "managed":
		return ScopeManaged
	default:
		return ScopeDefault
	}
}

// Platform defines the interface that platform adapters must implement
// for CLI operations. This is the consumer interface used by CLI commands.
type Platform interface {
	// Name returns the platform identifier (e.g., "claude", "opencode").
	Name() string

	// DisplayName returns a human-readable platform name (e.g., "Claude Code").
	DisplayName() string

	// IsAvailable checks if the platform is installed on this system.
	IsAvailable() bool

	// SkillDir returns the skills directory for the platform.
	SkillDir() string

	// InstallSkill installs a skill to the platform.
	// The skill parameter is platform-specific.
	InstallSkill(skill any, scope Scope) error

	// UninstallSkill removes a skill by name.
	UninstallSkill(name string, scope Scope) error

	// ListSkills returns information about all installed skills.
	ListSkills(scope Scope) ([]SkillInfo, error)

	// GetSkill retrieves a skill by name.
	// Returns the platform-specific skill type.
	GetSkill(name string, scope Scope) (any, error)

	// CommandDir returns the commands directory for the platform.
	CommandDir() string

	// InstallCommand installs a slash command to the platform.
	// The cmd parameter is platform-specific.
	InstallCommand(cmd any, scope Scope) error

	// UninstallCommand removes a command by name.
	UninstallCommand(name string, scope Scope) error

	// ListCommands returns information about all installed commands.
	ListCommands(scope Scope) ([]CommandInfo, error)

	// GetCommand retrieves a command by name.
	// Returns the platform-specific command type.
	GetCommand(name string, scope Scope) (any, error)

	// MCP configuration
	MCPConfigPath() string
	AddMCP(server any, scope Scope) error
	RemoveMCP(name string, scope Scope) error
	ListMCP(scope Scope) ([]MCPInfo, error)
	GetMCP(name string, scope Scope) (any, error)
	EnableMCP(name string) error
	DisableMCP(name string) error

	// Agent configuration
	AgentDir() string
	InstallAgent(agent any, scope Scope) error
	UninstallAgent(name string, scope Scope) error
	ListAgents(scope Scope) ([]AgentInfo, error)
	GetAgent(name string, scope Scope) (any, error)

	// Backup configuration
	// BackupPaths returns all config files/directories that should be backed up.
	// This includes MCP config files and platform-specific directories (skills, commands, agents).
	BackupPaths() []string
}

// claudeAdapter wraps ClaudePlatform to implement the Platform interface.
type claudeAdapter struct {
	p *claude.ClaudePlatform
}

func (a *claudeAdapter) Name() string {
	return a.p.Name()
}

func (a *claudeAdapter) DisplayName() string {
	return a.p.DisplayName()
}

func (a *claudeAdapter) IsAvailable() bool {
	return a.p.IsAvailable()
}

func (a *claudeAdapter) InstallSkill(skill any, scope Scope) error {
	s, ok := skill.(*claude.Skill)
	if !ok {
		return errors.Newf("expected *claude.Skill, got %T", skill)
	}
	return errors.Wrap(a.p.InstallSkill(s), "installing skill to Claude")
}

func (a *claudeAdapter) UninstallSkill(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallSkill(name), "uninstalling skill from Claude")
}

func (a *claudeAdapter) ListSkills(scope Scope) ([]SkillInfo, error) {
	skills, err := a.p.ListSkills()
	if err != nil {
		return nil, errors.Wrap(err, "listing Claude skills")
	}

	infos := make([]SkillInfo, len(skills))
	for i, s := range skills {
		infos[i] = SkillInfo{
			Name:        s.Name,
			Description: s.Description,
			Source:      "local", // TODO: track source in skill metadata
		}
	}
	return infos, nil
}

func (a *claudeAdapter) GetSkill(name string, scope Scope) (any, error) {
	s, err := a.p.GetSkill(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting Claude skill")
	}
	return s, nil
}

func (a *claudeAdapter) SkillDir() string {
	return a.p.SkillDir()
}

func (a *claudeAdapter) CommandDir() string {
	return a.p.CommandDir()
}

func (a *claudeAdapter) InstallCommand(cmd any, scope Scope) error {
	c, ok := cmd.(*claude.Command)
	if !ok {
		return errors.Newf("expected *claude.Command, got %T", cmd)
	}
	return errors.Wrap(a.p.InstallCommand(c), "installing command to Claude")
}

func (a *claudeAdapter) UninstallCommand(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallCommand(name), "uninstalling command from Claude")
}

func (a *claudeAdapter) ListCommands(scope Scope) ([]CommandInfo, error) {
	commands, err := a.p.ListCommands()
	if err != nil {
		return nil, errors.Wrap(err, "listing Claude commands")
	}

	infos := make([]CommandInfo, len(commands))
	for i, c := range commands {
		infos[i] = CommandInfo{
			Name:        c.Name,
			Description: c.Description,
			Source:      "installed",
		}
	}
	return infos, nil
}

func (a *claudeAdapter) GetCommand(name string, scope Scope) (any, error) {
	c, err := a.p.GetCommand(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting Claude command")
	}
	return c, nil
}

func (a *claudeAdapter) MCPConfigPath() string {
	return a.p.MCPConfigPath()
}

func (a *claudeAdapter) AddMCP(server any, scope Scope) error {
	s, ok := server.(*claude.MCPServer)
	if !ok {
		return errors.Newf("expected *claude.MCPServer, got %T", server)
	}
	return errors.Wrap(a.p.AddMCP(s), "adding MCP server to Claude")
}

func (a *claudeAdapter) RemoveMCP(name string, scope Scope) error {
	return errors.Wrap(a.p.RemoveMCP(name), "removing MCP server from Claude")
}

func (a *claudeAdapter) ListMCP(scope Scope) ([]MCPInfo, error) {
	servers, err := a.p.ListMCP()
	if err != nil {
		return nil, errors.Wrap(err, "listing Claude MCP servers")
	}
	infos := make([]MCPInfo, len(servers))
	for i, s := range servers {
		// Map Claude's Type to display transport
		// Claude uses "http" for remote, we display as "sse" for consistency
		var transport string
		switch s.Type {
		case "":
			if s.URL != "" {
				transport = "sse"
			} else {
				transport = "stdio"
			}
		case "http":
			transport = "sse"
		default:
			transport = s.Type
		}
		infos[i] = MCPInfo{
			Name:      s.Name,
			Transport: transport,
			Command:   s.Command,
			URL:       s.URL,
			Disabled:  s.Disabled,
			Env:       s.Env,
		}
	}
	return infos, nil
}

func (a *claudeAdapter) GetMCP(name string, scope Scope) (any, error) {
	s, err := a.p.GetMCP(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting Claude MCP server")
	}
	return s, nil
}

func (a *claudeAdapter) EnableMCP(name string) error {
	return errors.Wrap(a.p.EnableMCP(name), "enabling Claude MCP server")
}

func (a *claudeAdapter) DisableMCP(name string) error {
	return errors.Wrap(a.p.DisableMCP(name), "disabling Claude MCP server")
}

func (a *claudeAdapter) AgentDir() string {
	return a.p.AgentDir()
}

func (a *claudeAdapter) InstallAgent(agent any, scope Scope) error {
	ag, ok := agent.(*claude.Agent)
	if !ok {
		return errors.Newf("expected *claude.Agent, got %T", agent)
	}
	return errors.Wrap(a.p.InstallAgent(ag), "installing agent to Claude")
}

func (a *claudeAdapter) UninstallAgent(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallAgent(name), "uninstalling agent from Claude")
}

func (a *claudeAdapter) ListAgents(scope Scope) ([]AgentInfo, error) {
	agents, err := a.p.ListAgents()
	if err != nil {
		return nil, errors.Wrap(err, "listing Claude agents")
	}

	infos := make([]AgentInfo, len(agents))
	for i, ag := range agents {
		infos[i] = AgentInfo{
			Name:        ag.Name,
			Description: ag.Description,
			Source:      "local", // TODO: track source in agent metadata
		}
	}
	return infos, nil
}

func (a *claudeAdapter) GetAgent(name string, scope Scope) (any, error) {
	ag, err := a.p.GetAgent(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting Claude agent")
	}
	return ag, nil
}

func (a *claudeAdapter) BackupPaths() []string {
	return a.p.BackupPaths()
}

// opencodeAdapter wraps OpenCodePlatform to implement the Platform interface.
type opencodeAdapter struct {
	p *opencode.OpenCodePlatform
}

func (a *opencodeAdapter) Name() string {
	return a.p.Name()
}

func (a *opencodeAdapter) DisplayName() string {
	return a.p.DisplayName()
}

func (a *opencodeAdapter) IsAvailable() bool {
	return a.p.IsAvailable()
}

func (a *opencodeAdapter) InstallSkill(skill any, scope Scope) error {
	s, ok := skill.(*opencode.Skill)
	if !ok {
		return errors.Newf("expected *opencode.Skill, got %T", skill)
	}
	return errors.Wrap(a.p.InstallSkill(s), "installing skill to OpenCode")
}

func (a *opencodeAdapter) UninstallSkill(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallSkill(name), "uninstalling skill from OpenCode")
}

func (a *opencodeAdapter) ListSkills(scope Scope) ([]SkillInfo, error) {
	skills, err := a.p.ListSkills()
	if err != nil {
		return nil, errors.Wrap(err, "listing OpenCode skills")
	}

	infos := make([]SkillInfo, len(skills))
	for i, s := range skills {
		infos[i] = SkillInfo{
			Name:        s.Name,
			Description: s.Description,
			Source:      "local", // TODO: track source in skill metadata
		}
	}
	return infos, nil
}

func (a *opencodeAdapter) GetSkill(name string, scope Scope) (any, error) {
	s, err := a.p.GetSkill(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting OpenCode skill")
	}
	return s, nil
}

func (a *opencodeAdapter) SkillDir() string {
	return a.p.SkillDir()
}

func (a *opencodeAdapter) CommandDir() string {
	return a.p.CommandDir()
}

func (a *opencodeAdapter) InstallCommand(cmd any, scope Scope) error {
	c, ok := cmd.(*opencode.Command)
	if !ok {
		return errors.Newf("expected *opencode.Command, got %T", cmd)
	}
	return errors.Wrap(a.p.InstallCommand(c), "installing command to OpenCode")
}

func (a *opencodeAdapter) UninstallCommand(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallCommand(name), "uninstalling command from OpenCode")
}

func (a *opencodeAdapter) ListCommands(scope Scope) ([]CommandInfo, error) {
	commands, err := a.p.ListCommands()
	if err != nil {
		return nil, errors.Wrap(err, "listing OpenCode commands")
	}

	infos := make([]CommandInfo, len(commands))
	for i, c := range commands {
		infos[i] = CommandInfo{
			Name:        c.Name,
			Description: c.Description,
			Source:      "installed",
		}
	}
	return infos, nil
}

func (a *opencodeAdapter) GetCommand(name string, scope Scope) (any, error) {
	c, err := a.p.GetCommand(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting OpenCode command")
	}
	return c, nil
}

func (a *opencodeAdapter) MCPConfigPath() string {
	return a.p.MCPConfigPath()
}

func (a *opencodeAdapter) AddMCP(server any, scope Scope) error {
	s, ok := server.(*opencode.MCPServer)
	if !ok {
		return errors.Newf("expected *opencode.MCPServer, got %T", server)
	}
	return errors.Wrap(a.p.AddMCP(s), "adding MCP server to OpenCode")
}

func (a *opencodeAdapter) RemoveMCP(name string, scope Scope) error {
	return errors.Wrap(a.p.RemoveMCP(name), "removing MCP server from OpenCode")
}

func (a *opencodeAdapter) ListMCP(scope Scope) ([]MCPInfo, error) {
	servers, err := a.p.ListMCP()
	if err != nil {
		return nil, errors.Wrap(err, "listing OpenCode MCP servers")
	}
	infos := make([]MCPInfo, len(servers))
	for i, s := range servers {
		transport := "stdio"
		if s.Type == "remote" || s.URL != "" {
			transport = "sse"
		}
		cmd := ""
		if len(s.Command) > 0 {
			cmd = s.Command[0]
		}
		// Convert OpenCode's Enabled (positive) to MCPInfo's Disabled (negative)
		disabled := false
		if s.Enabled != nil && !*s.Enabled {
			disabled = true
		}
		infos[i] = MCPInfo{
			Name:      s.Name,
			Transport: transport,
			Command:   cmd,
			URL:       s.URL,
			Disabled:  disabled,
			Env:       s.Environment,
		}
	}
	return infos, nil
}

func (a *opencodeAdapter) GetMCP(name string, scope Scope) (any, error) {
	s, err := a.p.GetMCP(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting OpenCode MCP server")
	}
	return s, nil
}

func (a *opencodeAdapter) EnableMCP(name string) error {
	return errors.Wrap(a.p.EnableMCP(name), "enabling OpenCode MCP server")
}

func (a *opencodeAdapter) DisableMCP(name string) error {
	return errors.Wrap(a.p.DisableMCP(name), "disabling OpenCode MCP server")
}

func (a *opencodeAdapter) AgentDir() string {
	return a.p.AgentDir()
}

func (a *opencodeAdapter) InstallAgent(agent any, scope Scope) error {
	ag, ok := agent.(*opencode.Agent)
	if !ok {
		return errors.Newf("expected *opencode.Agent, got %T", agent)
	}
	return errors.Wrap(a.p.InstallAgent(ag), "installing agent to OpenCode")
}

func (a *opencodeAdapter) UninstallAgent(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallAgent(name), "uninstalling agent from OpenCode")
}

func (a *opencodeAdapter) ListAgents(scope Scope) ([]AgentInfo, error) {
	agents, err := a.p.ListAgents()
	if err != nil {
		return nil, errors.Wrap(err, "listing OpenCode agents")
	}

	infos := make([]AgentInfo, len(agents))
	for i, ag := range agents {
		infos[i] = AgentInfo{
			Name:        ag.Name,
			Description: ag.Description,
			Source:      "local", // TODO: track source in agent metadata
		}
	}
	return infos, nil
}

func (a *opencodeAdapter) GetAgent(name string, scope Scope) (any, error) {
	ag, err := a.p.GetAgent(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting OpenCode agent")
	}
	return ag, nil
}

func (a *opencodeAdapter) BackupPaths() []string {
	return a.p.BackupPaths()
}

// geminiAdapter wraps GeminiPlatform to implement the Platform interface.
type geminiAdapter struct {
	p *gemini.GeminiPlatform
}

func (a *geminiAdapter) Name() string {
	return a.p.Name()
}

func (a *geminiAdapter) DisplayName() string {
	return a.p.DisplayName()
}

func (a *geminiAdapter) IsAvailable() bool {
	return a.p.IsAvailable()
}

func (a *geminiAdapter) InstallSkill(skill any, scope Scope) error {
	s, ok := skill.(*gemini.Skill)
	if !ok {
		return errors.Newf("expected *gemini.Skill, got %T", skill)
	}
	return errors.Wrap(a.p.InstallSkill(s), "installing skill to Gemini")
}

func (a *geminiAdapter) UninstallSkill(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallSkill(name), "uninstalling skill from Gemini")
}

func (a *geminiAdapter) ListSkills(scope Scope) ([]SkillInfo, error) {
	skills, err := a.p.ListSkills()
	if err != nil {
		return nil, errors.Wrap(err, "listing Gemini skills")
	}

	infos := make([]SkillInfo, len(skills))
	for i, s := range skills {
		infos[i] = SkillInfo{
			Name:        s.Name,
			Description: s.Description,
			Source:      "local",
		}
	}
	return infos, nil
}

func (a *geminiAdapter) GetSkill(name string, scope Scope) (any, error) {
	s, err := a.p.GetSkill(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting Gemini skill")
	}
	return s, nil
}

func (a *geminiAdapter) SkillDir() string {
	return a.p.SkillDir()
}

func (a *geminiAdapter) CommandDir() string {
	return a.p.CommandDir()
}

func (a *geminiAdapter) InstallCommand(cmd any, scope Scope) error {
	c, ok := cmd.(*gemini.Command)
	if !ok {
		return errors.Newf("expected *gemini.Command, got %T", cmd)
	}
	return errors.Wrap(a.p.InstallCommand(c), "installing command to Gemini")
}

func (a *geminiAdapter) UninstallCommand(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallCommand(name), "uninstalling command from Gemini")
}

func (a *geminiAdapter) ListCommands(scope Scope) ([]CommandInfo, error) {
	commands, err := a.p.ListCommands()
	if err != nil {
		return nil, errors.Wrap(err, "listing Gemini commands")
	}

	infos := make([]CommandInfo, len(commands))
	for i, c := range commands {
		infos[i] = CommandInfo{
			Name:        c.Name,
			Description: c.Description,
			Source:      "installed",
		}
	}
	return infos, nil
}

func (a *geminiAdapter) GetCommand(name string, scope Scope) (any, error) {
	c, err := a.p.GetCommand(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting Gemini command")
	}
	return c, nil
}

func (a *geminiAdapter) MCPConfigPath() string {
	return a.p.MCPConfigPath()
}

func (a *geminiAdapter) AddMCP(server any, scope Scope) error {
	s, ok := server.(*gemini.MCPServer)
	if !ok {
		return errors.Newf("expected *gemini.MCPServer, got %T", server)
	}
	return errors.Wrap(a.p.AddMCP(s), "adding MCP server to Gemini")
}

func (a *geminiAdapter) RemoveMCP(name string, scope Scope) error {
	return errors.Wrap(a.p.RemoveMCP(name), "removing MCP server from Gemini")
}

func (a *geminiAdapter) ListMCP(scope Scope) ([]MCPInfo, error) {
	servers, err := a.p.ListMCP()
	if err != nil {
		return nil, errors.Wrap(err, "listing Gemini MCP servers")
	}
	infos := make([]MCPInfo, len(servers))
	for i, s := range servers {
		transport := "stdio"
		if s.URL != "" {
			transport = "sse"
		}
		infos[i] = MCPInfo{
			Name:      s.Name,
			Transport: transport,
			Command:   s.Command,
			URL:       s.URL,
			Disabled:  !s.Enabled,
			Env:       s.Env,
		}
	}
	return infos, nil
}

func (a *geminiAdapter) GetMCP(name string, scope Scope) (any, error) {
	s, err := a.p.GetMCP(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting Gemini MCP server")
	}
	return s, nil
}

func (a *geminiAdapter) EnableMCP(name string) error {
	return errors.Wrap(a.p.EnableMCP(name), "enabling Gemini MCP server")
}

func (a *geminiAdapter) DisableMCP(name string) error {
	return errors.Wrap(a.p.DisableMCP(name), "disabling Gemini MCP server")
}

func (a *geminiAdapter) AgentDir() string {
	return a.p.AgentDir()
}

func (a *geminiAdapter) InstallAgent(agent any, scope Scope) error {
	ag, ok := agent.(*gemini.Agent)
	if !ok {
		return errors.Newf("expected *gemini.Agent, got %T", agent)
	}
	return errors.Wrap(a.p.InstallAgent(ag), "installing agent to Gemini")
}

func (a *geminiAdapter) UninstallAgent(name string, scope Scope) error {
	return errors.Wrap(a.p.UninstallAgent(name), "uninstalling agent from Gemini")
}

func (a *geminiAdapter) ListAgents(scope Scope) ([]AgentInfo, error) {
	agents, err := a.p.ListAgents()
	if err != nil {
		return nil, errors.Wrap(err, "listing Gemini agents")
	}

	infos := make([]AgentInfo, len(agents))
	for i, ag := range agents {
		infos[i] = AgentInfo{
			Name:        ag.Name,
			Description: ag.Description,
			Source:      "local",
		}
	}
	return infos, nil
}

func (a *geminiAdapter) GetAgent(name string, scope Scope) (any, error) {
	ag, err := a.p.GetAgent(name)
	if err != nil {
		return nil, errors.Wrap(err, "getting Gemini agent")
	}
	return ag, nil
}

func (a *geminiAdapter) BackupPaths() []string {
	return a.p.BackupPaths()
}

func NewPlatform(name string) (Platform, error) {
	switch name {
	case paths.PlatformClaude:
		return &claudeAdapter{p: claude.NewClaudePlatform()}, nil
	case paths.PlatformOpenCode:
		return &opencodeAdapter{p: opencode.NewOpenCodePlatform()}, nil
	case paths.PlatformGemini:
		return &geminiAdapter{p: gemini.NewGeminiPlatform()}, nil
	default:
		return nil, errors.Wrapf(ErrUnknownPlatform, "platform %q not recognized", name)
	}
}

// ResolvePlatforms returns Platform instances for the given platform names.
// If names is empty, returns all detected/installed platforms.
// Returns an error if any platform name is invalid or if no platforms are available.
func ResolvePlatforms(names []string) ([]Platform, error) {
	// If no names specified, use all detected platforms
	if len(names) == 0 {
		detected := platform.DetectInstalled()
		if len(detected) == 0 {
			return nil, errors.Wrap(ErrNoPlatformsAvailable, "no AI assistants detected on this system")
		}

		platforms := make([]Platform, 0, len(detected))
		for _, d := range detected {
			// Only include platforms we have adapters for
			p, err := NewPlatform(d.Name)
			if err != nil {
				continue // Skip platforms without adapters
			}
			platforms = append(platforms, p)
		}

		if len(platforms) == 0 {
			return nil, errors.Wrap(ErrNoPlatformsAvailable, "no supported AI assistants found")
		}
		return platforms, nil
	}

	// Validate and create platforms for the specified names
	var invalid []string
	platforms := make([]Platform, 0, len(names))

	for _, name := range names {
		if !paths.ValidPlatform(name) {
			invalid = append(invalid, name)
			continue
		}

		p, err := NewPlatform(name)
		if err != nil {
			invalid = append(invalid, name)
			continue
		}
		platforms = append(platforms, p)
	}

	if len(invalid) > 0 {
		return nil, errors.Wrapf(ErrUnknownPlatform, "%s (valid: %s)",
			strings.Join(invalid, ", "),
			strings.Join(paths.Platforms(), ", "))
	}

	return platforms, nil
}

// DetermineScope resolves the configuration scope based on user request,
// environment context (Git), and interactivity.
//
// Precedence:
// 1. Explicit request via flag (--scope)
// 2. Interactive prompt (if in repo and TTY available)
// 3. Project scope default (if in repo but no TTY)
// 4. User scope default (if not in repo)
func DetermineScope(requested string) (Scope, error) {
	if requested != "" {
		return ParseScope(requested), nil
	}

	// Default context detection
	cwd, err := os.Getwd()
	if err != nil {
		return ScopeUser, fmt.Errorf("getting cwd: %w", err) // Fallback to safe default
	}

	inRepo := IsRepo(cwd)

	// If not in a repository, always default to User scope
	if !inRepo {
		return ScopeUser, nil
	}

	// If in a repository, check for interactivity
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// CI/Non-interactive: default to Project scope inside a repo
		return ScopeProject, nil
	}

	// Interactive: Prompt user to select scope
	return promptForScope()
}

func promptForScope() (Scope, error) {
	fmt.Println("\nTarget configuration scope?")
	fmt.Println("  [1] Project (Shared, committed to Git)")
	fmt.Println("  [2] User    (Personal, global)")
	fmt.Println("  [3] Local   (Personal, this project only, gitignored)")
	fmt.Print("Selection [1]: ")

	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return ScopeUser, fmt.Errorf("reading input: %w", err)
	}

	choice = strings.TrimSpace(choice)
	switch choice {
	case "1", "project", "":
		return ScopeProject, nil
	case "2", "user":
		return ScopeUser, nil
	case "3", "local":
		return ScopeLocal, nil
	default:
		fmt.Printf("Invalid selection %q, defaulting to Project scope.\n", choice)
		return ScopeProject, nil
	}
}

// IsRepo returns true if the given path is within a git repository.
func IsRepo(path string) bool {
	// We use git command directly to avoid duplicating logic.
	// This helper is used by DetermineScope.
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}
