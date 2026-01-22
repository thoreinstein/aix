// Package cli provides CLI-specific types and utilities for the aix command.
package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/thoreinstein/aix/internal/paths"
	"github.com/thoreinstein/aix/internal/platform"
	"github.com/thoreinstein/aix/internal/platform/claude"
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
	InstallSkill(skill any) error

	// UninstallSkill removes a skill by name.
	UninstallSkill(name string) error

	// ListSkills returns information about all installed skills.
	ListSkills() ([]SkillInfo, error)

	// GetSkill retrieves a skill by name.
	// Returns the platform-specific skill type.
	GetSkill(name string) (any, error)

	// CommandDir returns the commands directory for the platform.
	CommandDir() string

	// InstallCommand installs a slash command to the platform.
	// The cmd parameter is platform-specific.
	InstallCommand(cmd any) error

	// UninstallCommand removes a command by name.
	UninstallCommand(name string) error

	// ListCommands returns information about all installed commands.
	ListCommands() ([]CommandInfo, error)

	// GetCommand retrieves a command by name.
	// Returns the platform-specific command type.
	GetCommand(name string) (any, error)

	// MCP configuration
	MCPConfigPath() string
	AddMCP(server any) error
	RemoveMCP(name string) error
	ListMCP() ([]MCPInfo, error)
	GetMCP(name string) (any, error)
	EnableMCP(name string) error
	DisableMCP(name string) error
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

func (a *claudeAdapter) InstallSkill(skill any) error {
	s, ok := skill.(*claude.Skill)
	if !ok {
		return fmt.Errorf("expected *claude.Skill, got %T", skill)
	}
	return a.p.InstallSkill(s)
}

func (a *claudeAdapter) UninstallSkill(name string) error {
	return a.p.UninstallSkill(name)
}

func (a *claudeAdapter) ListSkills() ([]SkillInfo, error) {
	skills, err := a.p.ListSkills()
	if err != nil {
		return nil, err
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

func (a *claudeAdapter) GetSkill(name string) (any, error) {
	return a.p.GetSkill(name)
}

func (a *claudeAdapter) SkillDir() string {
	return a.p.SkillDir()
}

func (a *claudeAdapter) CommandDir() string {
	return a.p.CommandDir()
}

func (a *claudeAdapter) InstallCommand(cmd any) error {
	c, ok := cmd.(*claude.Command)
	if !ok {
		return fmt.Errorf("expected *claude.Command, got %T", cmd)
	}
	return a.p.InstallCommand(c)
}

func (a *claudeAdapter) UninstallCommand(name string) error {
	return a.p.UninstallCommand(name)
}

func (a *claudeAdapter) ListCommands() ([]CommandInfo, error) {
	commands, err := a.p.ListCommands()
	if err != nil {
		return nil, err
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

func (a *claudeAdapter) GetCommand(name string) (any, error) {
	return a.p.GetCommand(name)
}

func (a *claudeAdapter) MCPConfigPath() string {
	return a.p.MCPConfigPath()
}

func (a *claudeAdapter) AddMCP(server any) error {
	s, ok := server.(*claude.MCPServer)
	if !ok {
		return fmt.Errorf("expected *claude.MCPServer, got %T", server)
	}
	return a.p.AddMCP(s)
}

func (a *claudeAdapter) RemoveMCP(name string) error {
	return a.p.RemoveMCP(name)
}

func (a *claudeAdapter) ListMCP() ([]MCPInfo, error) {
	servers, err := a.p.ListMCP()
	if err != nil {
		return nil, err
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

func (a *claudeAdapter) GetMCP(name string) (any, error) {
	return a.p.GetMCP(name)
}

func (a *claudeAdapter) EnableMCP(name string) error {
	return a.p.EnableMCP(name)
}

func (a *claudeAdapter) DisableMCP(name string) error {
	return a.p.DisableMCP(name)
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

func (a *opencodeAdapter) InstallSkill(skill any) error {
	s, ok := skill.(*opencode.Skill)
	if !ok {
		return fmt.Errorf("expected *opencode.Skill, got %T", skill)
	}
	return a.p.InstallSkill(s)
}

func (a *opencodeAdapter) UninstallSkill(name string) error {
	return a.p.UninstallSkill(name)
}

func (a *opencodeAdapter) ListSkills() ([]SkillInfo, error) {
	skills, err := a.p.ListSkills()
	if err != nil {
		return nil, err
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

func (a *opencodeAdapter) GetSkill(name string) (any, error) {
	return a.p.GetSkill(name)
}

func (a *opencodeAdapter) SkillDir() string {
	return a.p.SkillDir()
}

func (a *opencodeAdapter) CommandDir() string {
	return a.p.CommandDir()
}

func (a *opencodeAdapter) InstallCommand(cmd any) error {
	c, ok := cmd.(*opencode.Command)
	if !ok {
		return fmt.Errorf("expected *opencode.Command, got %T", cmd)
	}
	return a.p.InstallCommand(c)
}

func (a *opencodeAdapter) UninstallCommand(name string) error {
	return a.p.UninstallCommand(name)
}

func (a *opencodeAdapter) ListCommands() ([]CommandInfo, error) {
	commands, err := a.p.ListCommands()
	if err != nil {
		return nil, err
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

func (a *opencodeAdapter) GetCommand(name string) (any, error) {
	return a.p.GetCommand(name)
}

func (a *opencodeAdapter) MCPConfigPath() string {
	return a.p.MCPConfigPath()
}

func (a *opencodeAdapter) AddMCP(server any) error {
	s, ok := server.(*opencode.MCPServer)
	if !ok {
		return fmt.Errorf("expected *opencode.MCPServer, got %T", server)
	}
	return a.p.AddMCP(s)
}

func (a *opencodeAdapter) RemoveMCP(name string) error {
	return a.p.RemoveMCP(name)
}

func (a *opencodeAdapter) ListMCP() ([]MCPInfo, error) {
	servers, err := a.p.ListMCP()
	if err != nil {
		return nil, err
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

func (a *opencodeAdapter) GetMCP(name string) (any, error) {
	return a.p.GetMCP(name)
}

func (a *opencodeAdapter) EnableMCP(name string) error {
	return a.p.EnableMCP(name)
}

func (a *opencodeAdapter) DisableMCP(name string) error {
	return a.p.DisableMCP(name)
}

// NewPlatform creates a Platform adapter for the given platform name.
// Returns ErrUnknownPlatform if the platform name is not recognized.
func NewPlatform(name string) (Platform, error) {
	switch name {
	case paths.PlatformClaude:
		return &claudeAdapter{p: claude.NewClaudePlatform()}, nil
	case paths.PlatformOpenCode:
		return &opencodeAdapter{p: opencode.NewOpenCodePlatform()}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownPlatform, name)
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
			return nil, ErrNoPlatformsAvailable
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
			return nil, ErrNoPlatformsAvailable
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
		return nil, fmt.Errorf("%w: %s (valid: %s)",
			ErrUnknownPlatform,
			strings.Join(invalid, ", "),
			strings.Join(paths.Platforms(), ", "))
	}

	return platforms, nil
}
