package opencode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCodePaths_BaseDir(t *testing.T) {
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
			name:        "user scope returns config opencode dir",
			scope:       ScopeUser,
			projectRoot: "",
			want:        filepath.Join(home, ".config", "opencode"),
		},
		{
			name:        "user scope ignores project root",
			scope:       ScopeUser,
			projectRoot: "/some/project",
			want:        filepath.Join(home, ".config", "opencode"),
		},
		{
			name:        "project scope returns project root directly",
			scope:       ScopeProject,
			projectRoot: "/some/project",
			want:        "/some/project",
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.BaseDir()
			if got != tt.want {
				t.Errorf("BaseDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpenCodePaths_SkillDir(t *testing.T) {
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
			want:        filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			name:        "project scope",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", "skills"),
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.SkillDir()
			if got != tt.want {
				t.Errorf("SkillDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpenCodePaths_CommandDir(t *testing.T) {
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
			want:        filepath.Join(home, ".config", "opencode", "commands"),
		},
		{
			name:        "project scope",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", "commands"),
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.CommandDir()
			if got != tt.want {
				t.Errorf("CommandDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpenCodePaths_AgentDir(t *testing.T) {
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
			want:        filepath.Join(home, ".config", "opencode", "agents"),
		},
		{
			name:        "project scope",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", "agents"),
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.AgentDir()
			if got != tt.want {
				t.Errorf("AgentDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpenCodePaths_MCPConfigPath(t *testing.T) {
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
			want:        filepath.Join(home, ".config", "opencode", "opencode.json"),
		},
		{
			name:        "project scope",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", "opencode.json"),
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.MCPConfigPath()
			if got != tt.want {
				t.Errorf("MCPConfigPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpenCodePaths_InstructionsPath(t *testing.T) {
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
			name:        "user scope returns AGENTS.md in base dir",
			scope:       ScopeUser,
			projectRoot: "",
			want:        filepath.Join(home, ".config", "opencode", "AGENTS.md"),
		},
		{
			name:        "project scope returns AGENTS.md at project root",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			want:        filepath.Join("/my/project", "AGENTS.md"),
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.InstructionsPath()
			if got != tt.want {
				t.Errorf("InstructionsPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOpenCodePaths_SkillPath(t *testing.T) {
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
			want:        filepath.Join(home, ".config", "opencode", "skills", "debug", "SKILL.md"),
		},
		{
			name:        "project scope with valid name",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			skillName:   "refactor",
			want:        filepath.Join("/my/project", "skills", "refactor", "SKILL.md"),
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.SkillPath(tt.skillName)
			if got != tt.want {
				t.Errorf("SkillPath(%q) = %q, want %q", tt.skillName, got, tt.want)
			}
		})
	}
}

func TestOpenCodePaths_CommandPath(t *testing.T) {
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
			want:        filepath.Join(home, ".config", "opencode", "commands", "build.md"),
		},
		{
			name:        "project scope with valid name",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			cmdName:     "test",
			want:        filepath.Join("/my/project", "commands", "test.md"),
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.CommandPath(tt.cmdName)
			if got != tt.want {
				t.Errorf("CommandPath(%q) = %q, want %q", tt.cmdName, got, tt.want)
			}
		})
	}
}

func TestOpenCodePaths_AgentPath(t *testing.T) {
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
			want:        filepath.Join(home, ".config", "opencode", "agents", "reviewer.md"),
		},
		{
			name:        "project scope with valid name",
			scope:       ScopeProject,
			projectRoot: "/my/project",
			agentName:   "planner",
			want:        filepath.Join("/my/project", "agents", "planner.md"),
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
			p := NewOpenCodePaths(tt.scope, tt.projectRoot)
			got := p.AgentPath(tt.agentName)
			if got != tt.want {
				t.Errorf("AgentPath(%q) = %q, want %q", tt.agentName, got, tt.want)
			}
		})
	}
}
