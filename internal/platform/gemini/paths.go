// Package gemini provides Gemini CLI specific configuration and path handling.
package gemini

import (
	"path/filepath"

	"github.com/thoreinstein/aix/internal/paths"
)

// Scope defines whether paths resolve to user-level or project-level configuration.
type Scope int

const (
	// ScopeUser resolves paths relative to ~/.gemini/
	ScopeUser Scope = iota
	// ScopeProject resolves paths relative to <projectRoot>/.gemini/
	ScopeProject
)

// GeminiPaths provides Gemini-specific path resolution.
// It wraps the generic paths package with Gemini-specific defaults.
type GeminiPaths struct {
	scope       Scope
	projectRoot string
}

// NewGeminiPaths creates a new GeminiPaths instance.
// For ScopeProject, projectRoot must be non-empty.
// For ScopeUser, projectRoot is ignored.
func NewGeminiPaths(scope Scope, projectRoot string) *GeminiPaths {
	return &GeminiPaths{
		scope:       scope,
		projectRoot: projectRoot,
	}
}

// BaseDir returns the base configuration directory.
// For ScopeUser: ~/.gemini/
// For ScopeProject: <projectRoot>/.gemini/
// Returns empty string if projectRoot is empty for ScopeProject.
func (p *GeminiPaths) BaseDir() string {
	switch p.scope {
	case ScopeUser:
		return paths.GlobalConfigDir(paths.PlatformGemini)
	case ScopeProject:
		return paths.ProjectConfigDir(paths.PlatformGemini, p.projectRoot)
	default:
		return ""
	}
}

// SkillDir returns the skills directory.
// Returns <base>/skills/
func (p *GeminiPaths) SkillDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "skills")
}

// CommandDir returns the commands directory.
// Returns <base>/commands/
func (p *GeminiPaths) CommandDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "commands")
}

// AgentDir returns the agents directory.
// Returns <base>/agents/
func (p *GeminiPaths) AgentDir() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "agents")
}

// MCPConfigPath returns the path to the MCP servers configuration file.
// Returns <base>/settings.toml
func (p *GeminiPaths) MCPConfigPath() string {
	base := p.BaseDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "settings.toml")
}

// InstructionsPath returns the path to the GEMINI.md instructions file.
// For ScopeUser: ~/.gemini/GEMINI.md
// For ScopeProject: <projectRoot>/GEMINI.md (note: at project root, not .gemini/)
func (p *GeminiPaths) InstructionsPath() string {
	switch p.scope {
	case ScopeUser:
		base := p.BaseDir()
		if base == "" {
			return ""
		}
		return filepath.Join(base, "GEMINI.md")
	case ScopeProject:
		if p.projectRoot == "" {
			return ""
		}
		return filepath.Join(p.projectRoot, "GEMINI.md")
	default:
		return ""
	}
}

// SkillPath returns the path to a specific skill's SKILL.md file.
// Returns <skills>/<name>/SKILL.md
// Returns empty string if name is empty.
func (p *GeminiPaths) SkillPath(name string) string {
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
// Returns <commands>/<name>.toml
// Returns empty string if name is empty.
func (p *GeminiPaths) CommandPath(name string) string {
	if name == "" {
		return ""
	}
	cmdDir := p.CommandDir()
	if cmdDir == "" {
		return ""
	}
	return filepath.Join(cmdDir, name+".toml")
}

// AgentPath returns the path to a specific agent file.
// Returns <agents>/<name>.md
// Returns empty string if name is empty.
func (p *GeminiPaths) AgentPath(name string) string {
	if name == "" {
		return ""
	}
	agentDir := p.AgentDir()
	if agentDir == "" {
		return ""
	}
	return filepath.Join(agentDir, name+".md")
}
