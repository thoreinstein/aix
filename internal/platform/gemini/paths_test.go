package gemini

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeminiPaths_BaseDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		xdgConfig   string
		want        string
	}{
		{
			name:  "User scope",
			scope: ScopeUser,
			want:  filepath.Join(home, ".gemini"),
		},
		{
			name:      "User scope with XDG_CONFIG_HOME",
			scope:     ScopeUser,
			xdgConfig: "/custom/config",
			want:      filepath.Join("/custom/config", "gemini"),
		},
		{
			name:        "Project scope",
			scope:       ScopeProject,
			projectRoot: "/tmp/project",
			want:        filepath.Join("/tmp/project", ".gemini"),
		},
		{
			name:        "Project scope empty root",
			scope:       ScopeProject,
			projectRoot: "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.xdgConfig != "" {
				t.Setenv("XDG_CONFIG_HOME", tt.xdgConfig)
			} else {
				t.Setenv("XDG_CONFIG_HOME", "")
			}

			p := NewGeminiPaths(tt.scope, tt.projectRoot)
			if got := p.BaseDir(); got != tt.want {
				t.Errorf("BaseDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeminiPaths_SubDirs(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".gemini")
	p := NewGeminiPaths(ScopeUser, "")

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"SkillDir", p.SkillDir(), filepath.Join(base, "skills")},
		{"CommandDir", p.CommandDir(), filepath.Join(base, "commands")},
		{"AgentDir", p.AgentDir(), filepath.Join(base, "agents")},
		{"MCPConfigPath", p.MCPConfigPath(), filepath.Join(base, "settings.toml")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestGeminiPaths_InstructionsPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()

	tests := []struct {
		name        string
		scope       Scope
		projectRoot string
		want        string
	}{
		{
			name:  "User scope",
			scope: ScopeUser,
			want:  filepath.Join(home, ".gemini", "GEMINI.md"),
		},
		{
			name:        "Project scope",
			scope:       ScopeProject,
			projectRoot: "/tmp/project",
			want:        filepath.Join("/tmp/project", "GEMINI.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewGeminiPaths(tt.scope, tt.projectRoot)
			if got := p.InstructionsPath(); got != tt.want {
				t.Errorf("InstructionsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
