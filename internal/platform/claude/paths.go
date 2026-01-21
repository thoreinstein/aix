// Package claude provides Claude Code specific configuration and path handling.
package claude

import (
	"path/filepath"

	"github.com/thoreinstein/aix/internal/paths"
)

// Scope defines whether paths resolve to user-level or project-level configuration.
type Scope int

const (
	// ScopeUser resolves paths relative to ~/.claude/
	ScopeUser Scope = iota
	// ScopeProject resolves paths relative to <projectRoot>/.claude/
	ScopeProject
)

// ClaudePaths provides Claude-specific path resolution.
// It wraps the generic paths package with Claude-specific defaults.
type ClaudePaths struct {
	scope       Scope
	projectRoot string
}

// NewClaudePaths creates a new ClaudePaths instance.
// For ScopeProject, projectRoot must be non-empty.
// For ScopeUser, projectRoot is ignored.
func NewClaudePaths(scope Scope, projectRoot string) *ClaudePaths {
	return &ClaudePaths{
		scope:       scope,
		projectRoot: projectRoot,
	}
}

// BaseDir returns the base configuration directory.
// For ScopeUser: ~/.claude/
// For ScopeProject: <projectRoot>/.claude/
// Returns empty string if projectRoot is empty for ScopeProject.
func (p *ClaudePaths) BaseDir() string {
	switch p.scope {
	case ScopeUser:
		return paths.GlobalConfigDir(paths.PlatformClaude)
	case ScopeProject:
		return paths.ProjectConfigDir(paths.PlatformClaude, p.projectRoot)
	default:
		return ""
	}
}

// SkillDir returns the skills directory.
// Returns <base>/skills/
func (p *ClaudePaths) SkillDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "skills")
}

// CommandDir returns the commands directory.
// Returns <base>/commands/
func (p *ClaudePaths) CommandDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "commands")
}

// AgentDir returns the agents directory.
// Returns <base>/agents/
func (p *ClaudePaths) AgentDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "agents")
}

// MCPConfigPath returns the path to the MCP servers configuration file.
// Returns <base>/mcp_servers.json
func (p *ClaudePaths) MCPConfigPath() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "mcp_servers.json")
}

// InstructionsPath returns the path to the CLAUDE.md instructions file.
// For ScopeUser: ~/.claude/CLAUDE.md
// For ScopeProject: <projectRoot>/CLAUDE.md (note: at project root, not .claude/)
func (p *ClaudePaths) InstructionsPath() string {
	switch p.scope {
	case ScopeUser:
		base := p.BaseDir()
		if base == "" {
			return ""
		}
		return filepath.Join(base, "CLAUDE.md")
	case ScopeProject:
		if p.projectRoot == "" {
			return ""
		}
		return filepath.Join(p.projectRoot, "CLAUDE.md")
	default:
		return ""
	}
}

// SkillPath returns the path to a specific skill's SKILL.md file.
// Returns <skills>/<name>/SKILL.md
// Returns empty string if name is empty.
func (p *ClaudePaths) SkillPath(name string) string {
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
func (p *ClaudePaths) CommandPath(name string) string {
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
// Returns <agents>/<name>.md
// Returns empty string if name is empty.
func (p *ClaudePaths) AgentPath(name string) string {
	if name == "" {
		return ""
	}
	agentDir := p.AgentDir()
	if agentDir == "" {
		return ""
	}
	return filepath.Join(agentDir, name+".md")
}
