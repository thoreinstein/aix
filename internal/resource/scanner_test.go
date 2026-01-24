package resource

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/thoreinstein/aix/internal/config"
)

// createTestRepo creates a test repository structure with the given resources.
// Returns the path to the created repository.
func createTestRepo(t *testing.T, skills, commands, agents, mcpServers map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	// Create skills
	for name, content := range skills {
		skillDir := filepath.Join(dir, "skills", name)
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create commands (directory style)
	for name, content := range commands {
		cmdDir := filepath.Join(dir, "commands", name)
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(cmdDir, "command.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create agents (directory style)
	for name, content := range agents {
		agentDir := filepath.Join(dir, "agents", name)
		if err := os.MkdirAll(agentDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create MCP servers
	if len(mcpServers) > 0 {
		mcpDir := filepath.Join(dir, "mcp")
		if err := os.MkdirAll(mcpDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range mcpServers {
			if err := os.WriteFile(filepath.Join(mcpDir, name+".json"), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	return dir
}

// validSkillFrontmatter returns valid SKILL.md content.
func validSkillFrontmatter(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nSkill content here."
}

// validCommandFrontmatter returns valid command.md content.
func validCommandFrontmatter(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nCommand instructions here."
}

// validAgentFrontmatter returns valid AGENT.md content.
func validAgentFrontmatter(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n\nAgent instructions here."
}

// validMCPJSON returns valid MCP server JSON.
func validMCPJSON(name, command string, args []string) string {
	server := map[string]any{
		"name":    name,
		"command": command,
		"args":    args,
	}
	data, _ := json.Marshal(server)
	return string(data)
}

func TestScanner_ScanRepo_HappyPath(t *testing.T) {
	// Create a repo with all resource types
	repoPath := createTestRepo(t,
		map[string]string{
			"code-review":    validSkillFrontmatter("code-review", "Reviews code for quality"),
			"test-generator": validSkillFrontmatter("test-generator", "Generates test cases"),
		},
		map[string]string{
			"deploy":     validCommandFrontmatter("deploy", "Deploy application"),
			"rollback":   validCommandFrontmatter("rollback", "Rollback deployment"),
			"db-migrate": validCommandFrontmatter("db-migrate", "Run database migrations"),
		},
		map[string]string{
			"helper":   validAgentFrontmatter("helper", "A helpful assistant"),
			"reviewer": validAgentFrontmatter("reviewer", "Code review specialist"),
		},
		map[string]string{
			"github":     validMCPJSON("github", "npx", []string{"-y", "@modelcontextprotocol/server-github"}),
			"filesystem": validMCPJSON("filesystem", "npx", []string{"-y", "@modelcontextprotocol/server-filesystem"}),
		},
	)

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(repoPath, "test-repo", "https://github.com/test/repo")

	if err != nil {
		t.Fatalf("ScanRepo() error = %v", err)
	}

	// Count resources by type
	counts := make(map[ResourceType]int)
	for _, r := range resources {
		counts[r.Type]++
	}

	// Verify counts
	if counts[TypeSkill] != 2 {
		t.Errorf("expected 2 skills, got %d", counts[TypeSkill])
	}
	if counts[TypeCommand] != 3 {
		t.Errorf("expected 3 commands, got %d", counts[TypeCommand])
	}
	if counts[TypeAgent] != 2 {
		t.Errorf("expected 2 agents, got %d", counts[TypeAgent])
	}
	if counts[TypeMCP] != 2 {
		t.Errorf("expected 2 MCP servers, got %d", counts[TypeMCP])
	}

	// Verify total count
	expectedTotal := 2 + 3 + 2 + 2
	if len(resources) != expectedTotal {
		t.Errorf("expected %d total resources, got %d", expectedTotal, len(resources))
	}

	// Verify resource fields for a specific resource
	var codeReviewSkill *Resource
	for i := range resources {
		if resources[i].Name == "code-review" && resources[i].Type == TypeSkill {
			codeReviewSkill = &resources[i]
			break
		}
	}
	if codeReviewSkill == nil {
		t.Fatal("expected to find code-review skill")
	}
	if codeReviewSkill.Description != "Reviews code for quality" {
		t.Errorf("unexpected description: %s", codeReviewSkill.Description)
	}
	if codeReviewSkill.RepoName != "test-repo" {
		t.Errorf("unexpected repo name: %s", codeReviewSkill.RepoName)
	}
	if codeReviewSkill.RepoURL != "https://github.com/test/repo" {
		t.Errorf("unexpected repo URL: %s", codeReviewSkill.RepoURL)
	}
	if codeReviewSkill.Path != "skills/code-review" {
		t.Errorf("unexpected path: %s", codeReviewSkill.Path)
	}
}

func TestScanner_ScanRepo_EmptyRepo(t *testing.T) {
	// Create an empty temp directory
	dir := t.TempDir()

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(dir, "empty-repo", "https://github.com/test/empty")

	if err != nil {
		t.Fatalf("ScanRepo() error = %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestScanner_ScanRepo_PartialResources(t *testing.T) {
	tests := []struct {
		name           string
		skills         map[string]string
		commands       map[string]string
		agents         map[string]string
		mcpServers     map[string]string
		expectedCounts map[ResourceType]int
	}{
		{
			name: "skills only",
			skills: map[string]string{
				"debug": validSkillFrontmatter("debug", "Debug helper"),
			},
			expectedCounts: map[ResourceType]int{TypeSkill: 1},
		},
		{
			name: "commands only",
			commands: map[string]string{
				"build": validCommandFrontmatter("build", "Build project"),
			},
			expectedCounts: map[ResourceType]int{TypeCommand: 1},
		},
		{
			name: "agents only",
			agents: map[string]string{
				"qa": validAgentFrontmatter("qa", "QA specialist"),
			},
			expectedCounts: map[ResourceType]int{TypeAgent: 1},
		},
		{
			name: "mcp only",
			mcpServers: map[string]string{
				"memory": validMCPJSON("memory", "npx", []string{"-y", "@modelcontextprotocol/server-memory"}),
			},
			expectedCounts: map[ResourceType]int{TypeMCP: 1},
		},
		{
			name: "skills and commands",
			skills: map[string]string{
				"refactor": validSkillFrontmatter("refactor", "Code refactoring"),
			},
			commands: map[string]string{
				"test": validCommandFrontmatter("test", "Run tests"),
			},
			expectedCounts: map[ResourceType]int{TypeSkill: 1, TypeCommand: 1},
		},
		{
			name: "agents and mcp",
			agents: map[string]string{
				"writer": validAgentFrontmatter("writer", "Documentation writer"),
			},
			mcpServers: map[string]string{
				"brave": validMCPJSON("brave", "npx", []string{"-y", "@modelcontextprotocol/server-brave-search"}),
			},
			expectedCounts: map[ResourceType]int{TypeAgent: 1, TypeMCP: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := createTestRepo(t, tt.skills, tt.commands, tt.agents, tt.mcpServers)

			scanner := NewScanner()
			resources, err := scanner.ScanRepo(repoPath, "partial-repo", "https://github.com/test/partial")

			if err != nil {
				t.Fatalf("ScanRepo() error = %v", err)
			}

			// Count resources by type
			counts := make(map[ResourceType]int)
			for _, r := range resources {
				counts[r.Type]++
			}

			// Verify counts
			for resourceType, expectedCount := range tt.expectedCounts {
				if counts[resourceType] != expectedCount {
					t.Errorf("expected %d %s resources, got %d", expectedCount, resourceType, counts[resourceType])
				}
			}

			// Verify unexpected types have zero count
			for _, resourceType := range []ResourceType{TypeSkill, TypeCommand, TypeAgent, TypeMCP} {
				if _, expected := tt.expectedCounts[resourceType]; !expected && counts[resourceType] != 0 {
					t.Errorf("expected 0 %s resources, got %d", resourceType, counts[resourceType])
				}
			}
		})
	}
}

func TestScanner_ScanRepo_MalformedFrontmatter(t *testing.T) {
	tests := []struct {
		name              string
		skillContent      string
		commandContent    string
		agentContent      string
		mcpContent        string
		expectedValid     int
		expectedMalformed int
	}{
		{
			name:              "malformed skill frontmatter",
			skillContent:      "---\nname: [invalid yaml\n---\n\nContent",
			commandContent:    "",
			expectedValid:     0,
			expectedMalformed: 1,
		},
		{
			name:              "malformed command frontmatter",
			commandContent:    "---\ndescription: [broken\n---\n\nContent",
			expectedValid:     0,
			expectedMalformed: 1,
		},
		{
			name:              "malformed agent frontmatter",
			agentContent:      "---\nname: {unclosed\n---\n\nContent",
			expectedValid:     0,
			expectedMalformed: 1,
		},
		{
			name:              "malformed mcp json",
			mcpContent:        `{"name": "broken", "command": }`,
			expectedValid:     0,
			expectedMalformed: 1,
		},
		{
			name:              "mix valid and malformed skills",
			skillContent:      "---\nname: [invalid\n---",
			expectedValid:     0,
			expectedMalformed: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			// Create malformed skill
			if tt.skillContent != "" {
				skillDir := filepath.Join(dir, "skills", "malformed")
				if err := os.MkdirAll(skillDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(tt.skillContent), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			// Create malformed command
			if tt.commandContent != "" {
				cmdDir := filepath.Join(dir, "commands", "malformed")
				if err := os.MkdirAll(cmdDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(cmdDir, "command.md"), []byte(tt.commandContent), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			// Create malformed agent
			if tt.agentContent != "" {
				agentDir := filepath.Join(dir, "agents", "malformed")
				if err := os.MkdirAll(agentDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(tt.agentContent), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			// Create malformed MCP
			if tt.mcpContent != "" {
				mcpDir := filepath.Join(dir, "mcp")
				if err := os.MkdirAll(mcpDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(mcpDir, "malformed.json"), []byte(tt.mcpContent), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			scanner := NewScanner()
			resources, err := scanner.ScanRepo(dir, "malformed-repo", "https://github.com/test/malformed")

			// Scanner should not return an error - malformed files should be skipped
			if err != nil {
				t.Fatalf("ScanRepo() should not error on malformed files: %v", err)
			}

			// Malformed files should be skipped, so we expect zero valid resources
			if len(resources) != tt.expectedValid {
				t.Errorf("expected %d valid resources, got %d", tt.expectedValid, len(resources))
			}
		})
	}
}

func TestScanner_ScanRepo_MixedValidAndMalformed(t *testing.T) {
	dir := t.TempDir()

	// Create valid skill
	validSkillDir := filepath.Join(dir, "skills", "valid-skill")
	if err := os.MkdirAll(validSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validSkillDir, "SKILL.md"),
		[]byte(validSkillFrontmatter("valid-skill", "A valid skill")), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create malformed skill
	malformedSkillDir := filepath.Join(dir, "skills", "malformed-skill")
	if err := os.MkdirAll(malformedSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(malformedSkillDir, "SKILL.md"),
		[]byte("---\nname: [broken yaml\n---"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create valid command
	validCmdDir := filepath.Join(dir, "commands", "valid-cmd")
	if err := os.MkdirAll(validCmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validCmdDir, "command.md"),
		[]byte(validCommandFrontmatter("valid-cmd", "A valid command")), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create malformed MCP
	mcpDir := filepath.Join(dir, "mcp")
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mcpDir, "malformed.json"),
		[]byte(`{"name": "broken", invalid}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create valid MCP
	if err := os.WriteFile(filepath.Join(mcpDir, "valid.json"),
		[]byte(validMCPJSON("valid-mcp", "npx", []string{"-y", "@mcp/test"})), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(dir, "mixed-repo", "https://github.com/test/mixed")

	if err != nil {
		t.Fatalf("ScanRepo() error = %v", err)
	}

	// Should have 3 valid resources: valid-skill, valid-cmd, valid-mcp
	if len(resources) != 3 {
		t.Errorf("expected 3 valid resources, got %d", len(resources))
	}

	// Verify we have the expected valid resources
	names := make(map[string]bool)
	for _, r := range resources {
		names[r.Name] = true
	}
	if !names["valid-skill"] {
		t.Error("expected valid-skill resource")
	}
	if !names["valid-cmd"] {
		t.Error("expected valid-cmd resource")
	}
	if !names["valid-mcp"] {
		t.Error("expected valid-mcp resource")
	}
}

func TestScanner_ScanRepo_NonExistentRepo(t *testing.T) {
	scanner := NewScanner()
	resources, err := scanner.ScanRepo("/path/that/does/not/exist", "ghost-repo", "")

	// Scanner should not error on non-existent directories - just return empty
	if err != nil {
		t.Fatalf("ScanRepo() unexpected error: %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestScanner_ScanAll(t *testing.T) {
	// Create first repo with skills
	repo1Path := createTestRepo(t,
		map[string]string{
			"skill-a": validSkillFrontmatter("skill-a", "Skill A description"),
			"skill-b": validSkillFrontmatter("skill-b", "Skill B description"),
		},
		nil, nil, nil,
	)

	// Create second repo with commands
	repo2Path := createTestRepo(t,
		nil,
		map[string]string{
			"cmd-x": validCommandFrontmatter("cmd-x", "Command X"),
		},
		nil, nil,
	)

	// Create third repo with agents and MCP
	repo3Path := createTestRepo(t,
		nil, nil,
		map[string]string{
			"agent-1": validAgentFrontmatter("agent-1", "Agent One"),
		},
		map[string]string{
			"mcp-server": validMCPJSON("mcp-server", "npx", []string{"-y", "@mcp/test"}),
		},
	)

	repos := []config.RepoConfig{
		{Path: repo1Path, Name: "repo1", URL: "https://github.com/test/repo1"},
		{Path: repo2Path, Name: "repo2", URL: "https://github.com/test/repo2"},
		{Path: repo3Path, Name: "repo3", URL: "https://github.com/test/repo3"},
	}

	scanner := NewScanner()
	resources, err := scanner.ScanAll(repos)

	if err != nil {
		t.Fatalf("ScanAll() error = %v", err)
	}

	// Total: 2 skills + 1 command + 1 agent + 1 MCP = 5
	if len(resources) != 5 {
		t.Errorf("expected 5 resources, got %d", len(resources))
	}

	// Verify resources are attributed to correct repos
	repoResourceCounts := make(map[string]int)
	for _, r := range resources {
		repoResourceCounts[r.RepoName]++
	}

	if repoResourceCounts["repo1"] != 2 {
		t.Errorf("expected 2 resources from repo1, got %d", repoResourceCounts["repo1"])
	}
	if repoResourceCounts["repo2"] != 1 {
		t.Errorf("expected 1 resource from repo2, got %d", repoResourceCounts["repo2"])
	}
	if repoResourceCounts["repo3"] != 2 {
		t.Errorf("expected 2 resources from repo3, got %d", repoResourceCounts["repo3"])
	}
}

func TestScanner_ScanAll_WithNonExistentRepo(t *testing.T) {
	// Create one valid repo
	validRepoPath := createTestRepo(t,
		map[string]string{
			"test-skill": validSkillFrontmatter("test-skill", "Test skill"),
		},
		nil, nil, nil,
	)

	repos := []config.RepoConfig{
		{Path: validRepoPath, Name: "valid-repo", URL: "https://github.com/test/valid"},
		{Path: "/path/that/does/not/exist", Name: "ghost-repo", URL: ""},
	}

	scanner := NewScanner()
	resources, err := scanner.ScanAll(repos)

	// Should not error
	if err != nil {
		t.Fatalf("ScanAll() error = %v", err)
	}

	// Should still return resources from valid repo
	if len(resources) != 1 {
		t.Errorf("expected 1 resource from valid repo, got %d", len(resources))
	}
	if resources[0].Name != "test-skill" {
		t.Errorf("expected test-skill, got %s", resources[0].Name)
	}
}

func TestScanner_ScanAll_EmptyRepoList(t *testing.T) {
	scanner := NewScanner()
	resources, err := scanner.ScanAll(nil)

	if err != nil {
		t.Fatalf("ScanAll() error = %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestScanner_NameFallback(t *testing.T) {
	t.Run("skill name from directory", func(t *testing.T) {
		dir := t.TempDir()
		skillDir := filepath.Join(dir, "skills", "my-skill-dir")
		if err := os.MkdirAll(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		// Frontmatter without name field
		content := "---\ndescription: A skill without a name field\n---\n\nContent"
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		scanner := NewScanner()
		resources, err := scanner.ScanRepo(dir, "test-repo", "")
		if err != nil {
			t.Fatal(err)
		}

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}
		if resources[0].Name != "my-skill-dir" {
			t.Errorf("expected name 'my-skill-dir', got '%s'", resources[0].Name)
		}
	})

	t.Run("command name from directory", func(t *testing.T) {
		dir := t.TempDir()
		cmdDir := filepath.Join(dir, "commands", "my-cmd-dir")
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatal(err)
		}
		// Frontmatter without name field
		content := "---\ndescription: A command without a name field\n---\n\nContent"
		if err := os.WriteFile(filepath.Join(cmdDir, "command.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		scanner := NewScanner()
		resources, err := scanner.ScanRepo(dir, "test-repo", "")
		if err != nil {
			t.Fatal(err)
		}

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}
		if resources[0].Name != "my-cmd-dir" {
			t.Errorf("expected name 'my-cmd-dir', got '%s'", resources[0].Name)
		}
	})

	t.Run("agent name from directory", func(t *testing.T) {
		dir := t.TempDir()
		agentDir := filepath.Join(dir, "agents", "my-agent-dir")
		if err := os.MkdirAll(agentDir, 0o755); err != nil {
			t.Fatal(err)
		}
		// Frontmatter without name field
		content := "---\ndescription: An agent without a name field\n---\n\nContent"
		if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		scanner := NewScanner()
		resources, err := scanner.ScanRepo(dir, "test-repo", "")
		if err != nil {
			t.Fatal(err)
		}

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}
		if resources[0].Name != "my-agent-dir" {
			t.Errorf("expected name 'my-agent-dir', got '%s'", resources[0].Name)
		}
	})

	t.Run("mcp name from filename", func(t *testing.T) {
		dir := t.TempDir()
		mcpDir := filepath.Join(dir, "mcp")
		if err := os.MkdirAll(mcpDir, 0o755); err != nil {
			t.Fatal(err)
		}
		// JSON without name field
		content := `{"command": "npx", "args": ["-y", "@mcp/test"]}`
		if err := os.WriteFile(filepath.Join(mcpDir, "my-mcp-server.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		scanner := NewScanner()
		resources, err := scanner.ScanRepo(dir, "test-repo", "")
		if err != nil {
			t.Fatal(err)
		}

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}
		if resources[0].Name != "my-mcp-server" {
			t.Errorf("expected name 'my-mcp-server', got '%s'", resources[0].Name)
		}
	})
}

func TestScanner_DirectFileCommands(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create direct .md files in commands/
	if err := os.WriteFile(filepath.Join(cmdDir, "quick.md"),
		[]byte(validCommandFrontmatter("quick", "Quick command")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "another.md"),
		[]byte("---\ndescription: Command without name\n---\n\nInstructions"), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(dir, "test-repo", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	names := make(map[string]bool)
	for _, r := range resources {
		names[r.Name] = true
	}
	if !names["quick"] {
		t.Error("expected 'quick' command")
	}
	if !names["another"] {
		t.Error("expected 'another' command (from filename)")
	}
}

func TestScanner_DirectFileAgents(t *testing.T) {
	dir := t.TempDir()
	agentDir := filepath.Join(dir, "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create direct .md files in agents/
	if err := os.WriteFile(filepath.Join(agentDir, "fast.md"),
		[]byte(validAgentFrontmatter("fast", "Fast agent")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "simple.md"),
		[]byte("---\ndescription: Agent without name\n---\n\nInstructions"), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(dir, "test-repo", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	names := make(map[string]bool)
	for _, r := range resources {
		names[r.Name] = true
	}
	if !names["fast"] {
		t.Error("expected 'fast' agent")
	}
	if !names["simple"] {
		t.Error("expected 'simple' agent (from filename)")
	}
}

func TestScanner_MCPDescriptionGeneration(t *testing.T) {
	tests := []struct {
		name            string
		mcpContent      string
		expectedDescPfx string
	}{
		{
			name:            "local server",
			mcpContent:      `{"name": "local", "command": "/usr/bin/mcp-server"}`,
			expectedDescPfx: "Local MCP server: /usr/bin/mcp-server",
		},
		{
			name:            "remote server",
			mcpContent:      `{"name": "remote", "url": "https://api.example.com/mcp"}`,
			expectedDescPfx: "Remote MCP server: https://api.example.com/mcp",
		},
		{
			name:            "server with neither command nor url",
			mcpContent:      `{"name": "minimal"}`,
			expectedDescPfx: "MCP server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			mcpDir := filepath.Join(dir, "mcp")
			if err := os.MkdirAll(mcpDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(mcpDir, "server.json"), []byte(tt.mcpContent), 0o644); err != nil {
				t.Fatal(err)
			}

			scanner := NewScanner()
			resources, err := scanner.ScanRepo(dir, "test-repo", "")
			if err != nil {
				t.Fatal(err)
			}

			if len(resources) != 1 {
				t.Fatalf("expected 1 resource, got %d", len(resources))
			}
			if resources[0].Description != tt.expectedDescPfx {
				t.Errorf("expected description '%s', got '%s'", tt.expectedDescPfx, resources[0].Description)
			}
		})
	}
}

func TestScanner_IgnoresNonResourceFiles(t *testing.T) {
	dir := t.TempDir()

	// Create skills directory with non-skill files
	skillsDir := filepath.Join(dir, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create a file directly in skills/ (should be ignored)
	if err := os.WriteFile(filepath.Join(skillsDir, "README.md"), []byte("# Skills"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a valid skill
	validSkillDir := filepath.Join(skillsDir, "valid")
	if err := os.MkdirAll(validSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validSkillDir, "SKILL.md"),
		[]byte(validSkillFrontmatter("valid", "Valid skill")), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create skills directory without SKILL.md (should be ignored)
	emptySkillDir := filepath.Join(skillsDir, "empty-skill")
	if err := os.MkdirAll(emptySkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(emptySkillDir, "notes.txt"), []byte("Not a skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create MCP directory with non-JSON files
	mcpDir := filepath.Join(dir, "mcp")
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mcpDir, "README.md"), []byte("# MCP Servers"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mcpDir, "valid.json"),
		[]byte(validMCPJSON("valid-mcp", "npx", []string{"-y", "@mcp/test"})), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create subdirectory in mcp/ (should be ignored)
	if err := os.MkdirAll(filepath.Join(mcpDir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(dir, "test-repo", "")
	if err != nil {
		t.Fatal(err)
	}

	// Should only find valid skill and valid-mcp
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}

	names := make(map[string]bool)
	for _, r := range resources {
		names[r.Name] = true
	}
	if !names["valid"] {
		t.Error("expected 'valid' skill")
	}
	if !names["valid-mcp"] {
		t.Error("expected 'valid-mcp' MCP server")
	}
}

func TestScanner_PathGeneration(t *testing.T) {
	dir := t.TempDir()

	// Create various resources and verify their paths
	skillDir := filepath.Join(dir, "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"),
		[]byte(validSkillFrontmatter("my-skill", "Desc")), 0o644); err != nil {
		t.Fatal(err)
	}

	cmdDir := filepath.Join(dir, "commands", "my-cmd")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "command.md"),
		[]byte(validCommandFrontmatter("my-cmd", "Desc")), 0o644); err != nil {
		t.Fatal(err)
	}

	cmdsDir := filepath.Join(dir, "commands")
	if err := os.WriteFile(filepath.Join(cmdsDir, "direct.md"),
		[]byte(validCommandFrontmatter("direct", "Desc")), 0o644); err != nil {
		t.Fatal(err)
	}

	agentDir := filepath.Join(dir, "agents", "my-agent")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"),
		[]byte(validAgentFrontmatter("my-agent", "Desc")), 0o644); err != nil {
		t.Fatal(err)
	}

	agentsDir := filepath.Join(dir, "agents")
	if err := os.WriteFile(filepath.Join(agentsDir, "quick.md"),
		[]byte(validAgentFrontmatter("quick", "Desc")), 0o644); err != nil {
		t.Fatal(err)
	}

	mcpDir := filepath.Join(dir, "mcp")
	if err := os.MkdirAll(mcpDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mcpDir, "server.json"),
		[]byte(validMCPJSON("server", "npx", []string{})), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(dir, "test-repo", "")
	if err != nil {
		t.Fatal(err)
	}

	// Build path lookup
	paths := make(map[string]string)
	for _, r := range resources {
		paths[r.Name] = r.Path
	}

	// Verify paths
	expectedPaths := map[string]string{
		"my-skill": "skills/my-skill",
		"my-cmd":   "commands/my-cmd",
		"direct":   "commands/direct.md",
		"my-agent": "agents/my-agent",
		"quick":    "agents/quick.md",
		"server":   "mcp/server.json",
	}

	for name, expectedPath := range expectedPaths {
		if paths[name] != expectedPath {
			t.Errorf("resource %s: expected path '%s', got '%s'", name, expectedPath, paths[name])
		}
	}
}

func TestNewScanner(t *testing.T) {
	scanner := NewScanner()
	if scanner == nil {
		t.Fatal("NewScanner() returned nil")
	}
	if scanner.logger == nil {
		t.Error("NewScanner() logger is nil")
	}
}

func TestNewScannerWithLogger(t *testing.T) {
	// Test with nil logger (should still work)
	scanner := NewScannerWithLogger(nil)
	if scanner == nil {
		t.Fatal("NewScannerWithLogger(nil) returned nil")
	}
}

// createBenchmarkRepo creates a repository with the specified number of resources
// for benchmarking purposes.
func createBenchmarkRepo(b *testing.B, numSkills, numCommands, numAgents, numMCP int) string {
	b.Helper()
	dir := b.TempDir()

	// Create skills
	if numSkills > 0 {
		skillsDir := filepath.Join(dir, "skills")
		if err := os.MkdirAll(skillsDir, 0o755); err != nil {
			b.Fatal(err)
		}
		for i := range numSkills {
			skillDir := filepath.Join(skillsDir, fmt.Sprintf("skill-%d", i))
			if err := os.MkdirAll(skillDir, 0o755); err != nil {
				b.Fatal(err)
			}
			content := fmt.Sprintf("---\nname: skill-%d\ndescription: Benchmark skill %d\n---\n\nSkill content here.", i, i)
			if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
				b.Fatal(err)
			}
		}
	}

	// Create commands
	if numCommands > 0 {
		cmdsDir := filepath.Join(dir, "commands")
		if err := os.MkdirAll(cmdsDir, 0o755); err != nil {
			b.Fatal(err)
		}
		for i := range numCommands {
			cmdDir := filepath.Join(cmdsDir, fmt.Sprintf("cmd-%d", i))
			if err := os.MkdirAll(cmdDir, 0o755); err != nil {
				b.Fatal(err)
			}
			content := fmt.Sprintf("---\nname: cmd-%d\ndescription: Benchmark command %d\n---\n\nCommand content.", i, i)
			if err := os.WriteFile(filepath.Join(cmdDir, "command.md"), []byte(content), 0o644); err != nil {
				b.Fatal(err)
			}
		}
	}

	// Create agents
	if numAgents > 0 {
		agentsDir := filepath.Join(dir, "agents")
		if err := os.MkdirAll(agentsDir, 0o755); err != nil {
			b.Fatal(err)
		}
		for i := range numAgents {
			agentDir := filepath.Join(agentsDir, fmt.Sprintf("agent-%d", i))
			if err := os.MkdirAll(agentDir, 0o755); err != nil {
				b.Fatal(err)
			}
			content := fmt.Sprintf("---\nname: agent-%d\ndescription: Benchmark agent %d\n---\n\nAgent instructions.", i, i)
			if err := os.WriteFile(filepath.Join(agentDir, "AGENT.md"), []byte(content), 0o644); err != nil {
				b.Fatal(err)
			}
		}
	}

	// Create MCP servers
	if numMCP > 0 {
		mcpDir := filepath.Join(dir, "mcp")
		if err := os.MkdirAll(mcpDir, 0o755); err != nil {
			b.Fatal(err)
		}
		for i := range numMCP {
			content := fmt.Sprintf(`{"name": "mcp-%d", "command": "npx", "args": ["-y", "@mcp/test-%d"]}`, i, i)
			if err := os.WriteFile(filepath.Join(mcpDir, fmt.Sprintf("mcp-%d.json", i)), []byte(content), 0o644); err != nil {
				b.Fatal(err)
			}
		}
	}

	return dir
}

func BenchmarkScanner_ScanRepo_Small(b *testing.B) {
	// Small repo: 5 of each resource type (20 total)
	repoPath := createBenchmarkRepo(b, 5, 5, 5, 5)
	scanner := NewScanner()

	b.ResetTimer()
	for range b.N {
		resources, err := scanner.ScanRepo(repoPath, "bench-repo", "https://github.com/bench/repo")
		if err != nil {
			b.Fatal(err)
		}
		if len(resources) != 20 {
			b.Fatalf("expected 20 resources, got %d", len(resources))
		}
	}
}

func BenchmarkScanner_ScanRepo_Medium(b *testing.B) {
	// Medium repo: 25 of each resource type (100 total)
	repoPath := createBenchmarkRepo(b, 25, 25, 25, 25)
	scanner := NewScanner()

	b.ResetTimer()
	for range b.N {
		resources, err := scanner.ScanRepo(repoPath, "bench-repo", "https://github.com/bench/repo")
		if err != nil {
			b.Fatal(err)
		}
		if len(resources) != 100 {
			b.Fatalf("expected 100 resources, got %d", len(resources))
		}
	}
}

func BenchmarkScanner_ScanRepo_Large(b *testing.B) {
	// Large repo: 50 of each resource type (200 total)
	repoPath := createBenchmarkRepo(b, 50, 50, 50, 50)
	scanner := NewScanner()

	b.ResetTimer()
	for range b.N {
		resources, err := scanner.ScanRepo(repoPath, "bench-repo", "https://github.com/bench/repo")
		if err != nil {
			b.Fatal(err)
		}
		if len(resources) != 200 {
			b.Fatalf("expected 200 resources, got %d", len(resources))
		}
	}
}

func BenchmarkScanner_ScanAll_Sequential(b *testing.B) {
	// Create 10 repos with 10 resources each (100 total)
	repos := make([]config.RepoConfig, 0, 10)
	for i := range 10 {
		path := createBenchmarkRepo(b, 3, 3, 2, 2)
		repos = append(repos, config.RepoConfig{
			Path: path,
			Name: fmt.Sprintf("repo-%d", i),
			URL:  fmt.Sprintf("https://github.com/bench/repo-%d", i),
		})
	}
	scanner := NewScanner()

	b.ResetTimer()
	for range b.N {
		resources, err := scanner.ScanAll(repos)
		if err != nil {
			b.Fatal(err)
		}
		if len(resources) != 100 {
			b.Fatalf("expected 100 resources, got %d", len(resources))
		}
	}
}

func BenchmarkScanner_ScanAll_ManyRepos(b *testing.B) {
	// Create 50 repos with 4 resources each (200 total)
	// This tests the parallel scanning benefit
	repos := make([]config.RepoConfig, 0, 50)
	for i := range 50 {
		path := createBenchmarkRepo(b, 1, 1, 1, 1)
		repos = append(repos, config.RepoConfig{
			Path: path,
			Name: fmt.Sprintf("repo-%d", i),
			URL:  fmt.Sprintf("https://github.com/bench/repo-%d", i),
		})
	}
	scanner := NewScanner()

	b.ResetTimer()
	for range b.N {
		resources, err := scanner.ScanAll(repos)
		if err != nil {
			b.Fatal(err)
		}
		if len(resources) != 200 {
			b.Fatalf("expected 200 resources, got %d", len(resources))
		}
	}
}

func TestScanner_ScanAll_ConcurrencyRace(t *testing.T) {
	// This test is specifically for detecting race conditions
	// Run with -race flag to verify concurrent access is safe

	// Create multiple repos
	repos := make([]config.RepoConfig, 0, 20)
	for i := range 20 {
		skills := map[string]string{
			fmt.Sprintf("skill-%d", i): validSkillFrontmatter(fmt.Sprintf("skill-%d", i), "Test skill"),
		}
		commands := map[string]string{
			fmt.Sprintf("cmd-%d", i): validCommandFrontmatter(fmt.Sprintf("cmd-%d", i), "Test command"),
		}
		path := createTestRepo(t, skills, commands, nil, nil)
		repos = append(repos, config.RepoConfig{
			Path: path,
			Name: fmt.Sprintf("repo-%d", i),
			URL:  fmt.Sprintf("https://github.com/test/repo-%d", i),
		})
	}

	scanner := NewScanner()

	// Run multiple times to increase chance of detecting races
	for range 5 {
		resources, err := scanner.ScanAll(repos)
		if err != nil {
			t.Fatalf("ScanAll() error = %v", err)
		}

		// Should have 40 resources (20 repos × 2 resources each)
		if len(resources) != 40 {
			t.Errorf("expected 40 resources, got %d", len(resources))
		}
	}
}

func TestScanner_ScanRepo_PermissionDenied(t *testing.T) {
	// Skip on Windows where permission handling is different
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	dir := t.TempDir()

	// Create a skills directory with restricted permissions
	skillsDir := filepath.Join(dir, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a valid skill first
	validSkillDir := filepath.Join(skillsDir, "valid")
	if err := os.MkdirAll(validSkillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validSkillDir, "SKILL.md"),
		[]byte(validSkillFrontmatter("valid", "Valid skill")), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a restricted skill directory
	restrictedSkillDir := filepath.Join(skillsDir, "restricted")
	if err := os.MkdirAll(restrictedSkillDir, 0o000); err != nil {
		t.Fatal(err)
	}

	// Ensure cleanup restores permissions
	t.Cleanup(func() {
		_ = os.Chmod(restrictedSkillDir, 0o755)
	})

	scanner := NewScanner()
	resources, err := scanner.ScanRepo(dir, "perm-repo", "")

	// Should not return an error
	if err != nil {
		t.Fatalf("ScanRepo() should not error on permission denied: %v", err)
	}

	// Should still find the valid skill
	if len(resources) != 1 {
		t.Errorf("expected 1 resource (valid skill), got %d", len(resources))
	}
}
