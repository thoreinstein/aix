package claude

import (
	"os"
	"path/filepath"
	"testing"
)

// testClaudePaths creates a ClaudePaths for testing with a temporary directory.
type testClaudePaths struct {
	*ClaudePaths
	tmpDir string
}

func newTestClaudePaths(t *testing.T) *testClaudePaths {
	t.Helper()
	tmpDir := t.TempDir()
	return &testClaudePaths{
		ClaudePaths: NewClaudePaths(ScopeProject, tmpDir),
		tmpDir:      tmpDir,
	}
}

func TestAgentManager_List_EmptyDirectory(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agents, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("List() returned %d agents, want 0", len(agents))
	}
}

func TestAgentManager_List_MultipleAgents(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	// Create agents directory
	agentDir := paths.AgentDir()
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Write test agent files
	testAgents := []struct {
		filename string
		content  string
	}{
		{"reviewer.md", "---\ndescription: Code review specialist\n---\n\nReview code carefully.\n"},
		{"planner.md", "Plan and organize tasks.\n"},
		{"debugger.md", "---\ndescription: Debug expert\n---\n\nHelp debug issues.\n"},
	}

	for _, ta := range testAgents {
		path := filepath.Join(agentDir, ta.filename)
		if err := os.WriteFile(path, []byte(ta.content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", ta.filename, err)
		}
	}

	agents, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(agents) != 3 {
		t.Errorf("List() returned %d agents, want 3", len(agents))
	}

	// Verify agent names are set correctly
	names := make(map[string]bool)
	for _, a := range agents {
		names[a.Name] = true
	}

	for _, expected := range []string{"reviewer", "planner", "debugger"} {
		if !names[expected] {
			t.Errorf("List() missing agent %q", expected)
		}
	}
}

func TestAgentManager_List_SkipsNonMarkdownFiles(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agentDir := paths.AgentDir()
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Write various files
	files := map[string]string{
		"valid.md":    "Valid agent content\n",
		"readme.txt":  "Not an agent\n",
		".hidden.md":  "Hidden file\n",
		"config.yaml": "config: value\n",
		"another.md":  "Another valid agent\n",
		"notes.MD":    "Wrong case extension\n", // .MD not .md
	}

	for name, content := range files {
		path := filepath.Join(agentDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	agents, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should only return .md files (case-sensitive)
	if len(agents) != 3 { // valid.md, .hidden.md, another.md
		t.Errorf("List() returned %d agents, want 3", len(agents))
	}
}

func TestAgentManager_List_SkipsDirectories(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agentDir := paths.AgentDir()
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Create a directory that looks like an agent
	subDir := filepath.Join(agentDir, "not-an-agent.md")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create a valid agent
	agentPath := filepath.Join(agentDir, "valid.md")
	if err := os.WriteFile(agentPath, []byte("Valid content\n"), 0o644); err != nil {
		t.Fatalf("failed to write valid.md: %v", err)
	}

	agents, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("List() returned %d agents, want 1", len(agents))
	}
	if len(agents) > 0 && agents[0].Name != "valid" {
		t.Errorf("List() agent name = %q, want %q", agents[0].Name, "valid")
	}
}

func TestAgentManager_Get_ExistingAgent(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agentDir := paths.AgentDir()
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	content := "---\ndescription: A helpful code reviewer\n---\n\nReview all code submissions thoroughly.\n"
	agentPath := filepath.Join(agentDir, "reviewer.md")
	if err := os.WriteFile(agentPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write agent: %v", err)
	}

	agent, err := mgr.Get("reviewer")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if agent.Name != "reviewer" {
		t.Errorf("Get() Name = %q, want %q", agent.Name, "reviewer")
	}
	if agent.Description != "A helpful code reviewer" {
		t.Errorf("Get() Description = %q, want %q", agent.Description, "A helpful code reviewer")
	}
	if agent.Instructions != "Review all code submissions thoroughly." {
		t.Errorf("Get() Instructions = %q, want %q", agent.Instructions, "Review all code submissions thoroughly.")
	}
}

func TestAgentManager_Get_NonExistentAgent(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	_, err := mgr.Get("nonexistent")
	if err != ErrAgentNotFound {
		t.Errorf("Get() error = %v, want ErrAgentNotFound", err)
	}
}

func TestAgentManager_Get_EmptyName(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	_, err := mgr.Get("")
	if err != ErrInvalidAgent {
		t.Errorf("Get() error = %v, want ErrInvalidAgent", err)
	}
}

func TestAgentManager_Install_NewAgent(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agent := &Agent{
		Name:         "planner",
		Description:  "Project planning specialist",
		Instructions: "Help plan and organize development work.",
	}

	if err := mgr.Install(agent); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Verify file was created
	agentPath := paths.AgentPath("planner")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read installed agent: %v", err)
	}

	content := string(data)
	if !contains(content, "description: Project planning specialist") {
		t.Errorf("Install() file missing description")
	}
	if !contains(content, "Help plan and organize development work.") {
		t.Errorf("Install() file missing instructions")
	}
}

func TestAgentManager_Install_OverwriteExisting(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	// Install initial version
	initial := &Agent{
		Name:         "reviewer",
		Description:  "Original description",
		Instructions: "Original instructions.",
	}
	if err := mgr.Install(initial); err != nil {
		t.Fatalf("Install() initial error = %v", err)
	}

	// Install updated version
	updated := &Agent{
		Name:         "reviewer",
		Description:  "Updated description",
		Instructions: "Updated instructions with more detail.",
	}
	if err := mgr.Install(updated); err != nil {
		t.Fatalf("Install() updated error = %v", err)
	}

	// Verify update
	agent, err := mgr.Get("reviewer")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if agent.Description != "Updated description" {
		t.Errorf("Install() overwrite Description = %q, want %q", agent.Description, "Updated description")
	}
	if agent.Instructions != "Updated instructions with more detail." {
		t.Errorf("Install() overwrite Instructions = %q, want %q", agent.Instructions, "Updated instructions with more detail.")
	}
}

func TestAgentManager_Install_CreatesDirectory(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agentDir := paths.AgentDir()
	// Verify directory doesn't exist
	if _, err := os.Stat(agentDir); !os.IsNotExist(err) {
		t.Fatal("agents directory should not exist before install")
	}

	agent := &Agent{
		Name:         "test",
		Instructions: "Test instructions.",
	}
	if err := mgr.Install(agent); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(agentDir); err != nil {
		t.Errorf("Install() did not create agents directory: %v", err)
	}
}

func TestAgentManager_Install_NilAgent(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	err := mgr.Install(nil)
	if err != ErrInvalidAgent {
		t.Errorf("Install(nil) error = %v, want ErrInvalidAgent", err)
	}
}

func TestAgentManager_Install_EmptyName(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agent := &Agent{
		Name:         "",
		Instructions: "Some instructions.",
	}

	err := mgr.Install(agent)
	if err != ErrInvalidAgent {
		t.Errorf("Install() with empty name error = %v, want ErrInvalidAgent", err)
	}
}

func TestAgentManager_Uninstall_ExistingAgent(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	// Install an agent first
	agent := &Agent{
		Name:         "toremove",
		Instructions: "This will be removed.",
	}
	if err := mgr.Install(agent); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Verify it exists
	if !mgr.Exists("toremove") {
		t.Fatal("agent should exist after install")
	}

	// Uninstall
	if err := mgr.Uninstall("toremove"); err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	// Verify it's gone
	if mgr.Exists("toremove") {
		t.Error("Uninstall() did not remove agent")
	}
}

func TestAgentManager_Uninstall_NonExistentAgent(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	// Uninstalling a non-existent agent should succeed (idempotent)
	err := mgr.Uninstall("nonexistent")
	if err != nil {
		t.Errorf("Uninstall() non-existent error = %v, want nil", err)
	}
}

func TestAgentManager_Uninstall_EmptyName(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	err := mgr.Uninstall("")
	if err != ErrInvalidAgent {
		t.Errorf("Uninstall() with empty name error = %v, want ErrInvalidAgent", err)
	}
}

func TestAgentManager_ParseAgentWithFrontmatter(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agentDir := paths.AgentDir()
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	content := `---
description: Multi-line test agent
---

This agent has detailed instructions.

With multiple paragraphs.
`
	agentPath := filepath.Join(agentDir, "multiline.md")
	if err := os.WriteFile(agentPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write agent: %v", err)
	}

	agent, err := mgr.Get("multiline")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if agent.Description != "Multi-line test agent" {
		t.Errorf("Get() Description = %q, want %q", agent.Description, "Multi-line test agent")
	}

	expectedInstructions := "This agent has detailed instructions.\n\nWith multiple paragraphs."
	if agent.Instructions != expectedInstructions {
		t.Errorf("Get() Instructions = %q, want %q", agent.Instructions, expectedInstructions)
	}
}

func TestAgentManager_ParseAgentWithoutFrontmatter(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agentDir := paths.AgentDir()
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	content := `This is a simple agent with no frontmatter.

Just instructions.
`
	agentPath := filepath.Join(agentDir, "simple.md")
	if err := os.WriteFile(agentPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write agent: %v", err)
	}

	agent, err := mgr.Get("simple")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if agent.Description != "" {
		t.Errorf("Get() Description = %q, want empty", agent.Description)
	}

	expectedInstructions := "This is a simple agent with no frontmatter.\n\nJust instructions."
	if agent.Instructions != expectedInstructions {
		t.Errorf("Get() Instructions = %q, want %q", agent.Instructions, expectedInstructions)
	}
}

func TestAgentManager_FormatAgentWithDescription(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agent := &Agent{
		Name:         "test",
		Description:  "Test description",
		Instructions: "Test instructions here.",
	}

	if err := mgr.Install(agent); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(paths.AgentPath("test"))
	if err != nil {
		t.Fatalf("failed to read agent file: %v", err)
	}

	content := string(data)
	if !contains(content, "---\n") {
		t.Error("formatted agent should contain frontmatter delimiters")
	}
	if !contains(content, "description: Test description") {
		t.Error("formatted agent should contain description in frontmatter")
	}
	if !contains(content, "Test instructions here.") {
		t.Error("formatted agent should contain instructions")
	}
}

func TestAgentManager_FormatAgentWithoutDescription(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	agent := &Agent{
		Name:         "minimal",
		Instructions: "Just instructions, no metadata.",
	}

	if err := mgr.Install(agent); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	data, err := os.ReadFile(paths.AgentPath("minimal"))
	if err != nil {
		t.Fatalf("failed to read agent file: %v", err)
	}

	content := string(data)
	// Should not have frontmatter if no description
	if contains(content, "---") {
		t.Error("agent without description should not have frontmatter")
	}
	if !contains(content, "Just instructions, no metadata.") {
		t.Error("agent should contain instructions")
	}
}

func TestAgentManager_Exists(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	// Should not exist initially
	if mgr.Exists("test") {
		t.Error("Exists() should return false for non-existent agent")
	}

	// Install
	agent := &Agent{
		Name:         "test",
		Instructions: "Test.",
	}
	if err := mgr.Install(agent); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Should exist now
	if !mgr.Exists("test") {
		t.Error("Exists() should return true after install")
	}

	// Empty name should return false
	if mgr.Exists("") {
		t.Error("Exists() should return false for empty name")
	}
}

func TestAgentManager_Names(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	// Empty initially
	names, err := mgr.Names()
	if err != nil {
		t.Fatalf("Names() error = %v", err)
	}
	if len(names) != 0 {
		t.Errorf("Names() returned %d names, want 0", len(names))
	}

	// Install some agents
	for _, name := range []string{"alpha", "beta", "gamma"} {
		agent := &Agent{Name: name, Instructions: "Test."}
		if err := mgr.Install(agent); err != nil {
			t.Fatalf("Install(%q) error = %v", name, err)
		}
	}

	names, err = mgr.Names()
	if err != nil {
		t.Fatalf("Names() error = %v", err)
	}

	if len(names) != 3 {
		t.Errorf("Names() returned %d names, want 3", len(names))
	}

	// Verify all names present
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, expected := range []string{"alpha", "beta", "gamma"} {
		if !nameSet[expected] {
			t.Errorf("Names() missing %q", expected)
		}
	}
}

func TestAgentManager_RoundTrip(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	original := &Agent{
		Name:         "roundtrip",
		Description:  "Test round-trip serialization",
		Instructions: "Instructions with special chars: \"quotes\" and 'apostrophes'.\n\nMultiple paragraphs too!",
	}

	if err := mgr.Install(original); err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	retrieved, err := mgr.Get("roundtrip")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Name != original.Name {
		t.Errorf("round-trip Name = %q, want %q", retrieved.Name, original.Name)
	}
	if retrieved.Description != original.Description {
		t.Errorf("round-trip Description = %q, want %q", retrieved.Description, original.Description)
	}
	if retrieved.Instructions != original.Instructions {
		t.Errorf("round-trip Instructions = %q, want %q", retrieved.Instructions, original.Instructions)
	}
}

func TestParseAgentContent_EdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		wantDescription  string
		wantInstructions string
	}{
		{
			name:             "empty content",
			content:          "",
			wantDescription:  "",
			wantInstructions: "",
		},
		{
			name:             "only whitespace",
			content:          "   \n\n  \n",
			wantDescription:  "",
			wantInstructions: "",
		},
		{
			name:             "dashes but not frontmatter",
			content:          "Some content with --- dashes in the middle",
			wantDescription:  "",
			wantInstructions: "Some content with --- dashes in the middle",
		},
		{
			name:             "unclosed frontmatter",
			content:          "---\ndescription: test\nno closing delimiter",
			wantDescription:  "",
			wantInstructions: "---\ndescription: test\nno closing delimiter",
		},
		{
			name:             "empty frontmatter",
			content:          "---\n---\n\nJust instructions",
			wantDescription:  "",
			wantInstructions: "Just instructions",
		},
		{
			name:             "frontmatter only no instructions",
			content:          "---\ndescription: meta only\n---",
			wantDescription:  "meta only",
			wantInstructions: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := parseAgentContent([]byte(tt.content))
			if err != nil {
				t.Fatalf("parseAgentContent() error = %v", err)
			}

			if agent.Description != tt.wantDescription {
				t.Errorf("Description = %q, want %q", agent.Description, tt.wantDescription)
			}
			if agent.Instructions != tt.wantInstructions {
				t.Errorf("Instructions = %q, want %q", agent.Instructions, tt.wantInstructions)
			}
		})
	}
}

func TestAgentManager_AgentDir(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	expected := paths.AgentDir()
	got := mgr.AgentDir()

	if got != expected {
		t.Errorf("AgentDir() = %q, want %q", got, expected)
	}
}

func TestAgentManager_AgentPath(t *testing.T) {
	paths := newTestClaudePaths(t)
	mgr := NewAgentManager(paths.ClaudePaths)

	expected := paths.AgentPath("test")
	got := mgr.AgentPath("test")

	if got != expected {
		t.Errorf("AgentPath() = %q, want %q", got, expected)
	}
}
