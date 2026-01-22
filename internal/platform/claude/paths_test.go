package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudePaths_BaseDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		want        string
	}{
		{
			name:        "user scope returns home claude dir",
			scope:       ScopeUser,
			projectRoot: "",
			want:        filepath.Join(home, ".claude"),
		},
		{
			name:        "user scope ignores project root",
			scope:       ScopeUser,
			projectRoot: "/some/project",
			want:        filepath.Join(home, ".claude"),
		},
		{
			name:        "project scope returns project claude dir",
			scope:       ScopeProject,
			projectRoot: "/some/project",
			want:        filepath.Join("/some/project", ".claude"),
		},
		{
			name:        "project scope with empty root returns empty",
			scope:       ScopeProject,
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.BaseDir()
			if got != tt.want {
				t.Errorf("BaseDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudePaths_SkillDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		want        string
	}{
		{
			name:        "user scope",
			scope:       ScopeUser,
			projectRoot: "",
			want:        filepath.Join(home, ".claude", "skills"),
		},
		{
			name:        "project scope",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", ".claude", "skills"),
		},
		{
			name:        "project scope empty root",
			scope:       ScopeProject,
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.SkillDir()
			if got != tt.want {
				t.Errorf("SkillDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudePaths_CommandDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		want        string
	}{
		{
			name:        "user scope",
			scope:       ScopeUser,
			projectRoot: "",
			want:        filepath.Join(home, ".claude", "commands"),
		},
		{
			name:        "project scope",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", ".claude", "commands"),
		},
		{
			name:        "project scope empty root",
			scope:       ScopeProject,
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.CommandDir()
			if got != tt.want {
				t.Errorf("CommandDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudePaths_AgentDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		want        string
	}{
		{
			name:        "user scope",
			scope:       ScopeUser,
			projectRoot: "",
			want:        filepath.Join(home, ".claude", "agents"),
		},
		{
			name:        "project scope",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", ".claude", "agents"),
		},
		{
			name:        "project scope empty root",
			scope:       ScopeProject,
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.AgentDir()
			if got != tt.want {
				t.Errorf("AgentDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudePaths_MCPConfigPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		want        string
	}{
		{
			name:        "user scope returns ~/.claude.json (not in .claude directory)",
			scope:       ScopeUser,
			projectRoot: "",
			want:        filepath.Join(home, ".claude.json"),
		},
		{
			name:        "project scope returns .mcp.json in .claude directory",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", ".claude", ".mcp.json"),
		},
		{
			name:        "project scope empty root returns empty",
			scope:       ScopeProject,
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.MCPConfigPath()
			if got != tt.want {
				t.Errorf("MCPConfigPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudePaths_InstructionsPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		want        string
	}{
		{
			name:        "user scope returns CLAUDE.md in base dir",
			scope:       ScopeUser,
			projectRoot: "",
			want:        filepath.Join(home, ".claude", "CLAUDE.md"),
		},
		{
			name:        "project scope returns CLAUDE.md at project root",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", "CLAUDE.md"),
		},
		{
			name:        "project scope empty root returns empty",
			scope:       ScopeProject,
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.InstructionsPath()
			if got != tt.want {
				t.Errorf("InstructionsPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClaudePaths_SkillPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		skillName   string
		want        string
	}{
		{
			name:        "user scope with valid name",
			scope:       ScopeUser,
			projectRoot: "",
			skillName:   "debug",
			want:        filepath.Join(home, ".claude", "skills", "debug", "SKILL.md"),
		},
		{
			name:        "project scope with valid name",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			skillName:   "refactor",
			want:        filepath.Join("/my/project", ".claude", "skills", "refactor", "SKILL.md"),
		},
		{
			name:        "empty name returns empty",
			scope:       ScopeUser,
			projectRoot: "",
			skillName:   "",
			want:        "",
		},
		{
			name:        "project scope empty root returns empty",
			scope:       ScopeProject,
			projectRoot: "",
			skillName:   "debug",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.SkillPath(tt.skillName)
			if got != tt.want {
				t.Errorf("SkillPath(%q) = %q, want %q", tt.skillName, got, tt.want)
			}
		})
	}
}

func TestClaudePaths_CommandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		cmdName     string
		want        string
	}{
		{
			name:        "user scope with valid name",
			scope:       ScopeUser,
			projectRoot: "",
			cmdName:     "build",
			want:        filepath.Join(home, ".claude", "commands", "build.md"),
		},
		{
			name:        "project scope with valid name",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			cmdName:     "test",
			want:        filepath.Join("/my/project", ".claude", "commands", "test.md"),
		},
		{
			name:        "empty name returns empty",
			scope:       ScopeUser,
			projectRoot: "",
			cmdName:     "",
			want:        "",
		},
		{
			name:        "project scope empty root returns empty",
			scope:       ScopeProject,
			projectRoot: "",
			cmdName:     "build",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.CommandPath(tt.cmdName)
			if got != tt.want {
				t.Errorf("CommandPath(%q) = %q, want %q", tt.cmdName, got, tt.want)
			}
		})
	}
}

func TestClaudePaths_AgentPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		agentName   string
		want        string
	}{
		{
			name:        "user scope with valid name",
			scope:       ScopeUser,
			projectRoot: "",
			agentName:   "reviewer",
			want:        filepath.Join(home, ".claude", "agents", "reviewer.md"),
		},
		{
			name:        "project scope with valid name",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			agentName:   "planner",
			want:        filepath.Join("/my/project", ".claude", "agents", "planner.md"),
		},
		{
			name:        "empty name returns empty",
			scope:       ScopeUser,
			projectRoot: "",
			agentName:   "",
			want:        "",
		},
		{
			name:        "project scope empty root returns empty",
			scope:       ScopeProject,
			projectRoot: "",
			agentName:   "reviewer",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewClaudePaths(tt.scope, tt.projectRoot)
			got := p.AgentPath(tt.agentName)
			if got != tt.want {
				t.Errorf("AgentPath(%q) = %q, want %q", tt.agentName, got, tt.want)
			}
		})
	}
}
