package repo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// createValidatorTestRepo creates a test repository structure for validation tests.
func createValidatorTestRepo(t *testing.T, opts validatorTestOptions) string {
	t.Helper()
	dir := t.TempDir()

	// Create skills
	if opts.createSkillsDir {
		skillsDir := filepath.Join(dir, "skills")
		if err := os.MkdirAll(skillsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range opts.skills {
			skillDir := filepath.Join(skillsDir, name)
			if err := os.MkdirAll(skillDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if content != "" {
				if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	// Create commands
	if opts.createCommandsDir {
		cmdsDir := filepath.Join(dir, "commands")
		if err := os.MkdirAll(cmdsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range opts.commands {
			cmdDir := filepath.Join(cmdsDir, name)
			if err := os.MkdirAll(cmdDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if content != "" {
				if err := os.WriteFile(filepath.Join(cmdDir, "command.md"), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	// Create agents
	if opts.createAgentsDir {
		agentsDir := filepath.Join(dir, "agents")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range opts.agents {
			agentDir := filepath.Join(agentsDir, name)
			if err := os.MkdirAll(agentDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if content != "" {
				if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	// Create MCP servers
	if opts.createMCPDir {
		mcpDir := filepath.Join(dir, "mcp")
		if err := os.MkdirAll(mcpDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range opts.mcpServers {
			if err := os.WriteFile(filepath.Join(mcpDir, name+".json"), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	return dir
}

type validatorTestOptions struct {
	createSkillsDir   bool
	createCommandsDir bool
	createAgentsDir   bool
	createMCPDir      bool
	skills            map[string]string // name -> SKILL.md content (empty string = no SKILL.md)
	commands          map[string]string // name -> command.md content
	agents            map[string]string // name -> AGENT.md content
	mcpServers        map[string]string // name -> JSON content
}

// validSkillContent returns valid SKILL.md content.
func validSkillContent(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nSkill content here."
}

// validCommandContent returns valid command.md content.
func validCommandContent(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nCommand instructions here."
}

// validAgentContent returns valid AGENT.md content.
func validAgentContent(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nAgent instructions here."
}

// validMCPContent returns valid MCP server JSON.
func validMCPContent(name, command string, args []string) string {
	server := map[string]any{
		"name":    name,
		"command": command,
		"args":    args,
	}
	data, _ := json.Marshal(server)
	return string(data)
}

func TestValidateRepoContent_ValidRepo(t *testing.T) {
	repoPath := createValidatorTestRepo(t, validatorTestOptions{
		createSkillsDir:   true,
		createCommandsDir: true,
		createAgentsDir:   true,
		createMCPDir:      true,
		skills: map[string]string{
			"code-review":    validSkillContent("code-review", "Reviews code"),
			"test-generator": validSkillContent("test-generator", "Generates tests"),
		},
		commands: map[string]string{
			"deploy":   validCommandContent("deploy", "Deploy application"),
			"rollback": validCommandContent("rollback", "Rollback deployment"),
		},
		agents: map[string]string{
			"helper":   validAgentContent("helper", "A helpful assistant"),
			"reviewer": validAgentContent("reviewer", "Code review specialist"),
		},
		mcpServers: map[string]string{
			"github":     validMCPContent("github", "npx", []string{"-y", "@mcp/github"}),
			"filesystem": validMCPContent("filesystem", "npx", []string{"-y", "@mcp/filesystem"}),
		},
	})

	warnings := ValidateRepoContent(repoPath)

	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for valid repo, got %d:", len(warnings))
		for _, w := range warnings {
			t.Errorf("  - %s: %s", w.Path, w.Message)
		}
	}
}

func TestValidateRepoContent_MissingDirectories(t *testing.T) {
	tests := []struct {
		name             string
		opts             validatorTestOptions
		expectedWarnings int
		expectedPaths    []string
	}{
		{
			name: "missing all directories",
			opts: validatorTestOptions{
				// No directories created
			},
			expectedWarnings: 4,
			expectedPaths:    []string{"skills", "commands", "agents", "mcp"},
		},
		{
			name: "missing skills directory",
			opts: validatorTestOptions{
				createCommandsDir: true,
				createAgentsDir:   true,
				createMCPDir:      true,
			},
			expectedWarnings: 1,
			expectedPaths:    []string{"skills"},
		},
		{
			name: "missing commands directory",
			opts: validatorTestOptions{
				createSkillsDir: true,
				createAgentsDir: true,
				createMCPDir:    true,
			},
			expectedWarnings: 1,
			expectedPaths:    []string{"commands"},
		},
		{
			name: "missing agents directory",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createMCPDir:      true,
			},
			expectedWarnings: 1,
			expectedPaths:    []string{"agents"},
		},
		{
			name: "missing mcp directory",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createAgentsDir:   true,
			},
			expectedWarnings: 1,
			expectedPaths:    []string{"mcp"},
		},
		{
			name: "missing skills and commands",
			opts: validatorTestOptions{
				createAgentsDir: true,
				createMCPDir:    true,
			},
			expectedWarnings: 2,
			expectedPaths:    []string{"skills", "commands"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := createValidatorTestRepo(t, tt.opts)
			warnings := ValidateRepoContent(repoPath)

			if len(warnings) != tt.expectedWarnings {
				t.Errorf("expected %d warnings, got %d", tt.expectedWarnings, len(warnings))
			}

			// Check that expected paths are present in warnings
			warningPaths := make(map[string]bool)
			for _, w := range warnings {
				warningPaths[w.Path] = true
			}

			for _, expectedPath := range tt.expectedPaths {
				if !warningPaths[expectedPath] {
					t.Errorf("expected warning for path %q", expectedPath)
				}
			}
		})
	}
}

func TestValidateRepoContent_EmptyDirectories(t *testing.T) {
	// Empty directories should not produce warnings (only missing dirs do)
	repoPath := createValidatorTestRepo(t, validatorTestOptions{
		createSkillsDir:   true,
		createCommandsDir: true,
		createAgentsDir:   true,
		createMCPDir:      true,
		// No resources in any directory
	})

	warnings := ValidateRepoContent(repoPath)

	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for empty directories, got %d:", len(warnings))
		for _, w := range warnings {
			t.Errorf("  - %s: %s", w.Path, w.Message)
		}
	}
}

func TestValidateRepoContent_InvalidResources(t *testing.T) {
	tests := []struct {
		name             string
		opts             validatorTestOptions
		expectedWarnings int
		wantPathContains string
		wantMsgContains  string
	}{
		{
			name: "malformed skill frontmatter",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createAgentsDir:   true,
				createMCPDir:      true,
				skills: map[string]string{
					"broken": "---\nname: [invalid yaml\n---\n\nContent",
				},
			},
			expectedWarnings: 1,
			wantPathContains: "skills/broken/SKILL.md",
			wantMsgContains:  "invalid frontmatter",
		},
		{
			name: "malformed command frontmatter",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createAgentsDir:   true,
				createMCPDir:      true,
				commands: map[string]string{
					"broken": "---\nname: {unclosed\n---\n\nContent",
				},
			},
			expectedWarnings: 1,
			wantPathContains: "commands/broken/command.md",
			wantMsgContains:  "invalid frontmatter",
		},
		{
			name: "malformed agent frontmatter",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createAgentsDir:   true,
				createMCPDir:      true,
				agents: map[string]string{
					"broken": "---\nname: [invalid\n---\n\nContent",
				},
			},
			expectedWarnings: 1,
			wantPathContains: "agents/broken/AGENT.md",
			wantMsgContains:  "invalid frontmatter",
		},
		{
			name: "malformed MCP JSON",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createAgentsDir:   true,
				createMCPDir:      true,
				mcpServers: map[string]string{
					"broken": `{"name": "test", "command": }`,
				},
			},
			expectedWarnings: 1,
			wantPathContains: "mcp/broken.json",
			wantMsgContains:  "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := createValidatorTestRepo(t, tt.opts)
			warnings := ValidateRepoContent(repoPath)

			if len(warnings) != tt.expectedWarnings {
				t.Errorf("expected %d warnings, got %d", tt.expectedWarnings, len(warnings))
				for _, w := range warnings {
					t.Logf("  - %s: %s", w.Path, w.Message)
				}
			}

			if len(warnings) > 0 {
				found := false
				for _, w := range warnings {
					if contains(w.Path, tt.wantPathContains) && contains(w.Message, tt.wantMsgContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning containing path %q and message %q", tt.wantPathContains, tt.wantMsgContains)
				}
			}
		})
	}
}

func TestValidateRepoContent_MissingResourceFiles(t *testing.T) {
	tests := []struct {
		name             string
		opts             validatorTestOptions
		expectedWarnings int
		wantMsgContains  string
	}{
		{
			name: "skill directory without SKILL.md",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createAgentsDir:   true,
				createMCPDir:      true,
				skills: map[string]string{
					"empty-skill": "", // Empty string means create dir but no SKILL.md
				},
			},
			expectedWarnings: 1,
			wantMsgContains:  "missing SKILL.md",
		},
		{
			name: "command directory without command.md",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createAgentsDir:   true,
				createMCPDir:      true,
				commands: map[string]string{
					"empty-cmd": "",
				},
			},
			expectedWarnings: 1,
			wantMsgContains:  "missing command.md",
		},
		{
			name: "agent directory without AGENT.md",
			opts: validatorTestOptions{
				createSkillsDir:   true,
				createCommandsDir: true,
				createAgentsDir:   true,
				createMCPDir:      true,
				agents: map[string]string{
					"empty-agent": "",
				},
			},
			expectedWarnings: 1,
			wantMsgContains:  "missing AGENT.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := createValidatorTestRepo(t, tt.opts)
			warnings := ValidateRepoContent(repoPath)

			if len(warnings) != tt.expectedWarnings {
				t.Errorf("expected %d warnings, got %d", tt.expectedWarnings, len(warnings))
				for _, w := range warnings {
					t.Logf("  - %s: %s", w.Path, w.Message)
				}
			}

			if len(warnings) > 0 && tt.wantMsgContains != "" {
				found := false
				for _, w := range warnings {
					if contains(w.Message, tt.wantMsgContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning message containing %q", tt.wantMsgContains)
				}
			}
		})
	}
}

func TestValidateRepoContent_MixedValidAndInvalid(t *testing.T) {
	repoPath := createValidatorTestRepo(t, validatorTestOptions{
		createSkillsDir:   true,
		createCommandsDir: true,
		createAgentsDir:   true,
		createMCPDir:      true,
		skills: map[string]string{
			"valid-skill":  validSkillContent("valid-skill", "A valid skill"),
			"broken-skill": "---\nname: [broken\n---",
		},
		commands: map[string]string{
			"valid-cmd": validCommandContent("valid-cmd", "A valid command"),
		},
		agents: map[string]string{
			"valid-agent": validAgentContent("valid-agent", "A valid agent"),
		},
		mcpServers: map[string]string{
			"valid-mcp":  validMCPContent("valid-mcp", "npx", []string{"-y", "@mcp/test"}),
			"broken-mcp": `{"name": "broken", invalid}`,
		},
	})

	warnings := ValidateRepoContent(repoPath)

	// Should have exactly 2 warnings (broken skill and broken mcp)
	if len(warnings) != 2 {
		t.Errorf("expected 2 warnings, got %d:", len(warnings))
		for _, w := range warnings {
			t.Errorf("  - %s: %s", w.Path, w.Message)
		}
	}

	// Verify both invalid resources are reported
	foundBrokenSkill := false
	foundBrokenMCP := false
	for _, w := range warnings {
		if contains(w.Path, "skills/broken-skill") {
			foundBrokenSkill = true
		}
		if contains(w.Path, "mcp/broken-mcp") {
			foundBrokenMCP = true
		}
	}

	if !foundBrokenSkill {
		t.Error("expected warning for broken-skill")
	}
	if !foundBrokenMCP {
		t.Error("expected warning for broken-mcp")
	}
}

func TestValidateRepoContent_NonExistentRepo(t *testing.T) {
	warnings := ValidateRepoContent("/path/that/does/not/exist")

	// Should have 4 warnings for missing directories
	if len(warnings) != 4 {
		t.Errorf("expected 4 warnings for non-existent repo, got %d", len(warnings))
	}
}

func TestValidationWarning_Fields(t *testing.T) {
	warning := ValidationWarning{
		Path:    "skills/test",
		Message: "test message",
	}

	if warning.Path != "skills/test" {
		t.Errorf("Path = %q, want %q", warning.Path, "skills/test")
	}
	if warning.Message != "test message" {
		t.Errorf("Message = %q, want %q", warning.Message, "test message")
	}
}

func TestValidateRepoContent_DirectFileResources(t *testing.T) {
	// Test validation of direct .md files in commands/ and agents/ directories
	dir := t.TempDir()

	// Create commands directory with direct .md files
	cmdsDir := filepath.Join(dir, "commands")
	if err := os.MkdirAll(cmdsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Valid direct command file
	if err := os.WriteFile(filepath.Join(cmdsDir, "quick.md"),
		[]byte(validCommandContent("quick", "Quick command")), 0o644); err != nil {
		t.Fatal(err)
	}
	// Invalid direct command file
	if err := os.WriteFile(filepath.Join(cmdsDir, "broken.md"),
		[]byte("---\nname: [broken\n---"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create agents directory with direct .md files
	agentsDir := filepath.Join(dir, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Valid direct agent file
	if err := os.WriteFile(filepath.Join(agentsDir, "fast.md"),
		[]byte(validAgentContent("fast", "Fast agent")), 0o644); err != nil {
		t.Fatal(err)
	}
	// Invalid direct agent file
	if err := os.WriteFile(filepath.Join(agentsDir, "invalid.md"),
		[]byte("---\nname: {unclosed\n---"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create empty skills and mcp directories
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "mcp"), 0o755); err != nil {
		t.Fatal(err)
	}

	warnings := ValidateRepoContent(dir)

	// Should have 2 warnings (broken.md and invalid.md)
	if len(warnings) != 2 {
		t.Errorf("expected 2 warnings, got %d:", len(warnings))
		for _, w := range warnings {
			t.Errorf("  - %s: %s", w.Path, w.Message)
		}
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
