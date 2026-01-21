// Package opencode provides OpenCode specific configuration and path handling.
package opencode

import (
	"path/filepath"

	"github.com/thoreinstein/aix/internal/paths"
)

// Scope defines whether paths resolve to user-level or project-level configuration.
type Scope int

const (
	// ScopeUser resolves paths relative to ~/.config/opencode/
	ScopeUser Scope = iota
	// ScopeProject resolves paths relative to <projectRoot>/ directly
	ScopeProject
)

// OpenCodePaths provides OpenCode-specific path resolution.
// It wraps the generic paths package with OpenCode-specific defaults.
type OpenCodePaths struct {
	scope       Scope
	projectRoot string
}

// NewOpenCodePaths creates a new OpenCodePaths instance.
// For ScopeProject, projectRoot must be non-empty.
// For ScopeUser, projectRoot is ignored.
func NewOpenCodePaths(scope Scope, projectRoot string) *OpenCodePaths {
	return &OpenCodePaths{
		scope:       scope,
		projectRoot: projectRoot,
	}
}

// BaseDir returns the base configuration directory.
// For ScopeUser: ~/.config/opencode/
// For ScopeProject: <projectRoot>/ (directly, not a subdirectory)
// Returns empty string if projectRoot is empty for ScopeProject.
func (p *OpenCodePaths) BaseDir() string {
	switch p.scope {
	case ScopeUser:
		return paths.GlobalConfigDir(paths.PlatformOpenCode)
	case ScopeProject:
		return paths.ProjectConfigDir(paths.PlatformOpenCode, p.projectRoot)
	default:
		return ""
	}
}

// SkillDir returns the skill directory.
// Returns <base>/skill/ (singular, not "skills")
func (p *OpenCodePaths) SkillDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "skill")
}

// CommandDir returns the commands directory.
// Returns <base>/commands/
func (p *OpenCodePaths) CommandDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "commands")
}

// AgentDir returns the agent directory.
// Returns <base>/agent/ (singular, not "agents")
func (p *OpenCodePaths) AgentDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "agent")
}

// MCPConfigPath returns the path to the MCP servers configuration file.
// Returns <base>/opencode.json (MCP config is embedded in main config file)
func (p *OpenCodePaths) MCPConfigPath() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "opencode.json")
}

// InstructionsPath returns the path to the AGENTS.md instructions file.
// For ScopeUser: ~/.config/opencode/AGENTS.md
// For ScopeProject: <projectRoot>/AGENTS.md
func (p *OpenCodePaths) InstructionsPath() string {
	switch p.scope {
	case ScopeUser:
		base := p.BaseDir()
		if base == "" {
			return ""
		}
		return filepath.Join(base, "AGENTS.md")
	case ScopeProject:
		if p.projectRoot == "" {
			return ""
		}
		return filepath.Join(p.projectRoot, "AGENTS.md")
	default:
		return ""
	}
}

// SkillPath returns the path to a specific skill's SKILL.md file.
// Returns <skill>/<name>/SKILL.md
// Returns empty string if name is empty.
func (p *OpenCodePaths) SkillPath(name string) string {
	if name == "" {
		return ""
	}
	skillDir := p.SkillDir()
	if skillDir == "" {
		return ""
	}
	return filepath.Join(skillDir, name, "SKILL.md")
}

// CommandPath returns the path to a specific command file.
// Returns <commands>/<name>.md
// Returns empty string if name is empty.
func (p *OpenCodePaths) CommandPath(name string) string {
	if name == "" {
		return ""
	}
	cmdDir := p.CommandDir()
	if cmdDir == "" {
		return ""
	}
	return filepath.Join(cmdDir, name+".md")
}

// AgentPath returns the path to a specific agent file.
// Returns <agent>/<name>.md
// Returns empty string if name is empty.
func (p *OpenCodePaths) AgentPath(name string) string {
	if name == "" {
		return ""
	}
	agentDir := p.AgentDir()
	if agentDir == "" {
		return ""
	}
	return filepath.Join(agentDir, name+".md")
}
