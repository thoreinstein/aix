package opencode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewOpenCodePlatform_Defaults(t *testing.T) {
	p := NewOpenCodePlatform()

	if p == nil {
		t.Fatal("NewOpenCodePlatform() returned nil")
	}

	if p.paths == nil {
		t.Error("paths is nil")
	}

	if p.paths.scope != ScopeUser {
		t.Errorf("expected default scope ScopeUser, got %v", p.paths.scope)
	}

	if p.skills == nil {
		t.Error("skills manager is nil")
	}

	if p.commands == nil {
		t.Error("commands manager is nil")
	}

	if p.agents == nil {
		t.Error("agents manager is nil")
	}

	if p.mcp == nil {
		t.Error("mcp manager is nil")
	}
}

func TestNewOpenCodePlatform_WithScope(t *testing.T) {
	tests := []struct {
		name  string
		scope Scope
	}{
		{"user scope", ScopeUser},
		{"project scope", ScopeProject},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewOpenCodePlatform(WithScope(tt.scope))

			if p.paths.scope != tt.scope {
				t.Errorf("expected scope %v, got %v", tt.scope, p.paths.scope)
			}
		})
	}
}

func TestNewOpenCodePlatform_WithProjectRoot(t *testing.T) {
	root := "/test/project"
	p := NewOpenCodePlatform(WithProjectRoot(root))

	if p.paths.projectRoot != root {
		t.Errorf("expected projectRoot %q, got %q", root, p.paths.projectRoot)
	}
}

func TestNewOpenCodePlatform_CombinedOptions(t *testing.T) {
	root := "/test/project"
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(root),
	)

	if p.paths.scope != ScopeProject {
		t.Errorf("expected ScopeProject, got %v", p.paths.scope)
	}
	if p.paths.projectRoot != root {
		t.Errorf("expected projectRoot %q, got %q", root, p.paths.projectRoot)
	}
}

func TestOpenCodePlatform_Identity(t *testing.T) {
	p := NewOpenCodePlatform()

	if got := p.Name(); got != "opencode" {
		t.Errorf("Name() = %q, want %q", got, "opencode")
	}

	if got := p.DisplayName(); got != "OpenCode" {
		t.Errorf("DisplayName() = %q, want %q", got, "OpenCode")
	}
}

func TestOpenCodePlatform_GlobalConfigDir(t *testing.T) {
	p := NewOpenCodePlatform()

	got := p.GlobalConfigDir()
	if got == "" {
		t.Error("GlobalConfigDir() returned empty string")
	}

	// Should contain "opencode" in the path
	if !strings.Contains(got, "opencode") {
		t.Errorf("GlobalConfigDir() = %q, want path containing 'opencode'", got)
	}

	// Should end with opencode (not a subdirectory)
	if filepath.Base(got) != "opencode" {
		t.Errorf("GlobalConfigDir() = %q, want path ending in 'opencode'", got)
	}
}

func TestOpenCodePlatform_ProjectConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform()

	got := p.ProjectConfigDir(tmpDir)

	// OpenCode uses project root directly, not a subdirectory
	if got != tmpDir {
		t.Errorf("ProjectConfigDir(%q) = %q, want %q (project root itself)", tmpDir, got, tmpDir)
	}
}

func TestOpenCodePlatform_PathMethods(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	t.Run("SkillDir", func(t *testing.T) {
		got := p.SkillDir()
		// OpenCode uses plural "skills"
		want := filepath.Join(tmpDir, "skills")
		if got != want {
			t.Errorf("SkillDir() = %q, want %q", got, want)
		}
	})

	t.Run("CommandDir", func(t *testing.T) {
		got := p.CommandDir()
		want := filepath.Join(tmpDir, "commands")
		if got != want {
			t.Errorf("CommandDir() = %q, want %q", got, want)
		}
	})

	t.Run("AgentDir", func(t *testing.T) {
		got := p.AgentDir()
		// OpenCode uses plural "agents"
		want := filepath.Join(tmpDir, "agents")
		if got != want {
			t.Errorf("AgentDir() = %q, want %q", got, want)
		}
	})

	t.Run("MCPConfigPath", func(t *testing.T) {
		got := p.MCPConfigPath()
		// OpenCode uses opencode.json for MCP config
		if !strings.Contains(got, "opencode.json") {
			t.Errorf("MCPConfigPath() = %q, want path containing 'opencode.json'", got)
		}
		want := filepath.Join(tmpDir, "opencode.json")
		if got != want {
			t.Errorf("MCPConfigPath() = %q, want %q", got, want)
		}
	})
}

func TestOpenCodePlatform_InstructionsPath(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("user scope", func(t *testing.T) {
		p := NewOpenCodePlatform(WithScope(ScopeUser))

		// When projectRoot is provided, it should return project-scoped path
		got := p.InstructionsPath(tmpDir)
		want := filepath.Join(tmpDir, "AGENTS.md")
		if got != want {
			t.Errorf("InstructionsPath(%q) = %q, want %q", tmpDir, got, want)
		}
	})

	t.Run("user scope without project root", func(t *testing.T) {
		p := NewOpenCodePlatform(WithScope(ScopeUser))

		got := p.InstructionsPath("")
		// Should return user-scoped path
		if got == "" {
			t.Error("InstructionsPath(\"\") returned empty string for user scope")
		}
		if !strings.Contains(got, "AGENTS.md") {
			t.Errorf("InstructionsPath(\"\") = %q, want path containing 'AGENTS.md'", got)
		}
		if !strings.Contains(got, "opencode") {
			t.Errorf("InstructionsPath(\"\") = %q, want path containing 'opencode'", got)
		}
	})

	t.Run("project scope", func(t *testing.T) {
		p := NewOpenCodePlatform(
			WithScope(ScopeProject),
			WithProjectRoot(tmpDir),
		)

		got := p.InstructionsPath(tmpDir)
		want := filepath.Join(tmpDir, "AGENTS.md")
		if got != want {
			t.Errorf("InstructionsPath(%q) = %q, want %q", tmpDir, got, want)
		}
	})
}

func TestOpenCodePlatform_SkillLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	skill := &Skill{
		Name:         "test-skill",
		Description:  "A test skill",
		Version:      "1.0.0",
		Instructions: "Test instructions",
	}

	// Install
	if err := p.InstallSkill(skill); err != nil {
		t.Fatalf("InstallSkill() error = %v", err)
	}

	// List
	skills, err := p.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("ListSkills() returned %d skills, want 1", len(skills))
	}

	// Get
	got, err := p.GetSkill("test-skill")
	if err != nil {
		t.Fatalf("GetSkill() error = %v", err)
	}
	if got.Name != skill.Name {
		t.Errorf("GetSkill().Name = %q, want %q", got.Name, skill.Name)
	}
	if got.Description != skill.Description {
		t.Errorf("GetSkill().Description = %q, want %q", got.Description, skill.Description)
	}

	// Uninstall
	if err := p.UninstallSkill("test-skill"); err != nil {
		t.Fatalf("UninstallSkill() error = %v", err)
	}

	// Verify uninstalled
	skills, err = p.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() after uninstall error = %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("ListSkills() after uninstall returned %d skills, want 0", len(skills))
	}
}

func TestOpenCodePlatform_CommandLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	cmd := &Command{
		Name:         "test-command",
		Description:  "A test command",
		Instructions: "Test instructions",
	}

	// Install
	if err := p.InstallCommand(cmd); err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}

	// List
	commands, err := p.ListCommands()
	if err != nil {
		t.Fatalf("ListCommands() error = %v", err)
	}
	if len(commands) != 1 {
		t.Errorf("ListCommands() returned %d commands, want 1", len(commands))
	}

	// Get
	got, err := p.GetCommand("test-command")
	if err != nil {
		t.Fatalf("GetCommand() error = %v", err)
	}
	if got.Name != cmd.Name {
		t.Errorf("GetCommand().Name = %q, want %q", got.Name, cmd.Name)
	}
	if got.Description != cmd.Description {
		t.Errorf("GetCommand().Description = %q, want %q", got.Description, cmd.Description)
	}

	// Uninstall
	if err := p.UninstallCommand("test-command"); err != nil {
		t.Fatalf("UninstallCommand() error = %v", err)
	}

	// Verify uninstalled
	commands, err = p.ListCommands()
	if err != nil {
		t.Fatalf("ListCommands() after uninstall error = %v", err)
	}
	if len(commands) != 0 {
		t.Errorf("ListCommands() after uninstall returned %d commands, want 0", len(commands))
	}
}

func TestOpenCodePlatform_AgentLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	agent := &Agent{
		Name:         "test-agent",
		Description:  "A test agent",
		Instructions: "Test instructions",
	}

	// Install
	if err := p.InstallAgent(agent); err != nil {
		t.Fatalf("InstallAgent() error = %v", err)
	}

	// List
	agents, err := p.ListAgents()
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("ListAgents() returned %d agents, want 1", len(agents))
	}

	// Get
	got, err := p.GetAgent("test-agent")
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	if got.Name != agent.Name {
		t.Errorf("GetAgent().Name = %q, want %q", got.Name, agent.Name)
	}
	if got.Description != agent.Description {
		t.Errorf("GetAgent().Description = %q, want %q", got.Description, agent.Description)
	}

	// Uninstall
	if err := p.UninstallAgent("test-agent"); err != nil {
		t.Fatalf("UninstallAgent() error = %v", err)
	}

	// Verify uninstalled
	agents, err = p.ListAgents()
	if err != nil {
		t.Fatalf("ListAgents() after uninstall error = %v", err)
	}
	if len(agents) != 0 {
		t.Errorf("ListAgents() after uninstall returned %d agents, want 0", len(agents))
	}
}

func TestOpenCodePlatform_MCPLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	server := &MCPServer{
		Name:    "test-server",
		Command: []string{"test-cmd", "--arg1", "--arg2"},
		Type:    "local",
	}

	// Add
	if err := p.AddMCP(server); err != nil {
		t.Fatalf("AddMCP() error = %v", err)
	}

	// List
	servers, err := p.ListMCP()
	if err != nil {
		t.Fatalf("ListMCP() error = %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("ListMCP() returned %d servers, want 1", len(servers))
	}

	// Get
	got, err := p.GetMCP("test-server")
	if err != nil {
		t.Fatalf("GetMCP() error = %v", err)
	}
	if got.Name != server.Name {
		t.Errorf("GetMCP().Name = %q, want %q", got.Name, server.Name)
	}
	if len(got.Command) != len(server.Command) {
		t.Errorf("GetMCP().Command length = %d, want %d", len(got.Command), len(server.Command))
	}

	// Disable
	if err := p.DisableMCP("test-server"); err != nil {
		t.Fatalf("DisableMCP() error = %v", err)
	}
	got, _ = p.GetMCP("test-server")
	if !got.Disabled {
		t.Error("DisableMCP() did not set Disabled=true")
	}

	// Enable
	if err := p.EnableMCP("test-server"); err != nil {
		t.Fatalf("EnableMCP() error = %v", err)
	}
	got, _ = p.GetMCP("test-server")
	if got.Disabled {
		t.Error("EnableMCP() did not set Disabled=false")
	}

	// Remove
	if err := p.RemoveMCP("test-server"); err != nil {
		t.Fatalf("RemoveMCP() error = %v", err)
	}

	// Verify removed
	servers, err = p.ListMCP()
	if err != nil {
		t.Fatalf("ListMCP() after remove error = %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("ListMCP() after remove returned %d servers, want 0", len(servers))
	}
}

func TestOpenCodePlatform_Translation(t *testing.T) {
	p := NewOpenCodePlatform()

	t.Run("TranslateVariables passthrough", func(t *testing.T) {
		content := "Use $ARGUMENTS and $SELECTION"
		got := p.TranslateVariables(content)
		if got != content {
			t.Errorf("TranslateVariables() = %q, want %q", got, content)
		}
	})

	t.Run("TranslateToCanonical passthrough", func(t *testing.T) {
		content := "Use $ARGUMENTS and $SELECTION"
		got := p.TranslateToCanonical(content)
		if got != content {
			t.Errorf("TranslateToCanonical() = %q, want %q", got, content)
		}
	})

	t.Run("ValidateVariables valid", func(t *testing.T) {
		content := "Use $ARGUMENTS and $SELECTION"
		if err := p.ValidateVariables(content); err != nil {
			t.Errorf("ValidateVariables() unexpected error = %v", err)
		}
	})

	t.Run("ValidateVariables invalid", func(t *testing.T) {
		content := "Use $UNKNOWN_VAR"
		err := p.ValidateVariables(content)
		if err == nil {
			t.Error("ValidateVariables() expected error for unsupported variable")
		}
	})

	t.Run("ValidateVariables with no variables", func(t *testing.T) {
		content := "Plain text with no variables"
		if err := p.ValidateVariables(content); err != nil {
			t.Errorf("ValidateVariables() unexpected error = %v", err)
		}
	})

	t.Run("ValidateVariables with mixed valid and invalid", func(t *testing.T) {
		content := "Use $ARGUMENTS and $INVALID_VAR"
		err := p.ValidateVariables(content)
		if err == nil {
			t.Error("ValidateVariables() expected error for unsupported variable")
		}
		if !strings.Contains(err.Error(), "$INVALID_VAR") {
			t.Errorf("ValidateVariables() error = %q, want error mentioning '$INVALID_VAR'", err.Error())
		}
	})
}

func TestOpenCodePlatform_IsAvailable(t *testing.T) {
	t.Run("directory exists", func(t *testing.T) {
		// Create a temp dir and set HOME to parent to simulate ~/.config/opencode/
		tmpDir := t.TempDir()
		opencodeDir := filepath.Join(tmpDir, ".config", "opencode")
		if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}

		t.Setenv("HOME", tmpDir)

		p := NewOpenCodePlatform()
		if !p.IsAvailable() {
			t.Error("IsAvailable() = false, want true when ~/.config/opencode/ exists")
		}
	})

	t.Run("directory does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Don't create .config/opencode directory

		t.Setenv("HOME", tmpDir)

		p := NewOpenCodePlatform()
		if p.IsAvailable() {
			t.Error("IsAvailable() = true, want false when ~/.config/opencode/ does not exist")
		}
	})
}

func TestOpenCodePlatform_Version(t *testing.T) {
	p := NewOpenCodePlatform()

	version, err := p.Version()
	if err != nil {
		t.Errorf("Version() unexpected error = %v", err)
	}
	if version != "" {
		t.Errorf("Version() = %q, want empty string (not yet implemented)", version)
	}
}

func TestOpenCodePlatform_SkillOperations_Delegation(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	// Test that operations delegate to the skill manager
	// by verifying they produce actual filesystem effects

	skill := &Skill{
		Name:         "delegation-test",
		Description:  "Test skill",
		Instructions: "Instructions here",
	}

	// InstallSkill should create the skill file
	if err := p.InstallSkill(skill); err != nil {
		t.Fatalf("InstallSkill() error = %v", err)
	}

	// Verify file was created
	skillPath := filepath.Join(tmpDir, "skills", "delegation-test", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Errorf("InstallSkill() did not create skill file at %q", skillPath)
	}

	// GetSkill should read from filesystem
	got, err := p.GetSkill("delegation-test")
	if err != nil {
		t.Fatalf("GetSkill() error = %v", err)
	}
	if got.Name != skill.Name {
		t.Errorf("GetSkill().Name = %q, want %q", got.Name, skill.Name)
	}

	// ListSkills should enumerate from filesystem
	skills, err := p.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
	if len(skills) != 1 {
		t.Errorf("ListSkills() returned %d skills, want 1", len(skills))
	}

	// UninstallSkill should remove from filesystem
	if err := p.UninstallSkill("delegation-test"); err != nil {
		t.Fatalf("UninstallSkill() error = %v", err)
	}
	if _, err := os.Stat(skillPath); !os.IsNotExist(err) {
		t.Errorf("UninstallSkill() did not remove skill file at %q", skillPath)
	}
}

func TestOpenCodePlatform_CommandOperations_Delegation(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	cmd := &Command{
		Name:         "delegation-cmd",
		Description:  "Test command",
		Instructions: "Instructions here",
	}

	// InstallCommand should create the command file
	if err := p.InstallCommand(cmd); err != nil {
		t.Fatalf("InstallCommand() error = %v", err)
	}

	// Verify file was created
	cmdPath := filepath.Join(tmpDir, "commands", "delegation-cmd.md")
	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		t.Errorf("InstallCommand() did not create command file at %q", cmdPath)
	}

	// GetCommand should read from filesystem
	got, err := p.GetCommand("delegation-cmd")
	if err != nil {
		t.Fatalf("GetCommand() error = %v", err)
	}
	if got.Name != cmd.Name {
		t.Errorf("GetCommand().Name = %q, want %q", got.Name, cmd.Name)
	}

	// ListCommands should enumerate from filesystem
	commands, err := p.ListCommands()
	if err != nil {
		t.Fatalf("ListCommands() error = %v", err)
	}
	if len(commands) != 1 {
		t.Errorf("ListCommands() returned %d commands, want 1", len(commands))
	}

	// UninstallCommand should remove from filesystem
	if err := p.UninstallCommand("delegation-cmd"); err != nil {
		t.Fatalf("UninstallCommand() error = %v", err)
	}
	if _, err := os.Stat(cmdPath); !os.IsNotExist(err) {
		t.Errorf("UninstallCommand() did not remove command file at %q", cmdPath)
	}
}

func TestOpenCodePlatform_AgentOperations_Delegation(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	agent := &Agent{
		Name:         "delegation-agent",
		Description:  "Test agent",
		Instructions: "Instructions here",
	}

	// InstallAgent should create the agent file
	if err := p.InstallAgent(agent); err != nil {
		t.Fatalf("InstallAgent() error = %v", err)
	}

	// Verify file was created (OpenCode uses "agents" plural)
	agentPath := filepath.Join(tmpDir, "agents", "delegation-agent.md")
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("InstallAgent() did not create agent file at %q", agentPath)
	}

	// GetAgent should read from filesystem
	got, err := p.GetAgent("delegation-agent")
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	if got.Name != agent.Name {
		t.Errorf("GetAgent().Name = %q, want %q", got.Name, agent.Name)
	}

	// ListAgents should enumerate from filesystem
	agents, err := p.ListAgents()
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("ListAgents() returned %d agents, want 1", len(agents))
	}

	// UninstallAgent should remove from filesystem
	if err := p.UninstallAgent("delegation-agent"); err != nil {
		t.Fatalf("UninstallAgent() error = %v", err)
	}
	if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
		t.Errorf("UninstallAgent() did not remove agent file at %q", agentPath)
	}
}

func TestOpenCodePlatform_MCPOperations_Delegation(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	server := &MCPServer{
		Name:    "delegation-server",
		Command: []string{"test-cmd"},
		Type:    "local",
	}

	// AddMCP should create/update the config file
	if err := p.AddMCP(server); err != nil {
		t.Fatalf("AddMCP() error = %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, "opencode.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("AddMCP() did not create config file at %q", configPath)
	}

	// GetMCP should read from the config file
	got, err := p.GetMCP("delegation-server")
	if err != nil {
		t.Fatalf("GetMCP() error = %v", err)
	}
	if got.Name != server.Name {
		t.Errorf("GetMCP().Name = %q, want %q", got.Name, server.Name)
	}

	// ListMCP should enumerate from config file
	servers, err := p.ListMCP()
	if err != nil {
		t.Fatalf("ListMCP() error = %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("ListMCP() returned %d servers, want 1", len(servers))
	}

	// EnableMCP and DisableMCP should update the config
	if err := p.DisableMCP("delegation-server"); err != nil {
		t.Fatalf("DisableMCP() error = %v", err)
	}
	got, _ = p.GetMCP("delegation-server")
	if !got.Disabled {
		t.Error("DisableMCP() did not set Disabled=true")
	}

	if err := p.EnableMCP("delegation-server"); err != nil {
		t.Fatalf("EnableMCP() error = %v", err)
	}
	got, _ = p.GetMCP("delegation-server")
	if got.Disabled {
		t.Error("EnableMCP() did not set Disabled=false")
	}

	// RemoveMCP should update the config file
	if err := p.RemoveMCP("delegation-server"); err != nil {
		t.Fatalf("RemoveMCP() error = %v", err)
	}

	servers, _ = p.ListMCP()
	if len(servers) != 0 {
		t.Errorf("RemoveMCP() did not remove server, ListMCP() returned %d servers, want 0", len(servers))
	}
}

func TestOpenCodePlatform_MultipleSkills(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	skills := []*Skill{
		{Name: "skill-a", Description: "Skill A", Instructions: "A"},
		{Name: "skill-b", Description: "Skill B", Instructions: "B"},
		{Name: "skill-c", Description: "Skill C", Instructions: "C"},
	}

	// Install all skills
	for _, s := range skills {
		if err := p.InstallSkill(s); err != nil {
			t.Fatalf("InstallSkill(%q) error = %v", s.Name, err)
		}
	}

	// List should return all skills
	listed, err := p.ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
	if len(listed) != len(skills) {
		t.Errorf("ListSkills() returned %d skills, want %d", len(listed), len(skills))
	}

	// Get each skill individually
	for _, s := range skills {
		got, err := p.GetSkill(s.Name)
		if err != nil {
			t.Errorf("GetSkill(%q) error = %v", s.Name, err)
			continue
		}
		if got.Description != s.Description {
			t.Errorf("GetSkill(%q).Description = %q, want %q", s.Name, got.Description, s.Description)
		}
	}
}

func TestOpenCodePlatform_MultipleMCPServers(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewOpenCodePlatform(
		WithScope(ScopeProject),
		WithProjectRoot(tmpDir),
	)

	servers := []*MCPServer{
		{Name: "server-a", Command: []string{"cmd-a"}, Type: "local"},
		{Name: "server-b", Command: []string{"cmd-b"}, Type: "remote", URL: "http://example.com"},
		{Name: "server-c", Command: []string{"cmd-c"}, Type: "local"},
	}

	// Add all servers
	for _, s := range servers {
		if err := p.AddMCP(s); err != nil {
			t.Fatalf("AddMCP(%q) error = %v", s.Name, err)
		}
	}

	// List should return all servers
	listed, err := p.ListMCP()
	if err != nil {
		t.Fatalf("ListMCP() error = %v", err)
	}
	if len(listed) != len(servers) {
		t.Errorf("ListMCP() returned %d servers, want %d", len(listed), len(servers))
	}

	// Get each server individually
	for _, s := range servers {
		got, err := p.GetMCP(s.Name)
		if err != nil {
			t.Errorf("GetMCP(%q) error = %v", s.Name, err)
			continue
		}
		if got.Type != s.Type {
			t.Errorf("GetMCP(%q).Type = %q, want %q", s.Name, got.Type, s.Type)
		}
	}
}
