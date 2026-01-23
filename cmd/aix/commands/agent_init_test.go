package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAgentName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple",
			input:   "myagent",
			wantErr: false,
		},
		{
			name:    "valid with hyphen",
			input:   "my-agent",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "agent-v2",
			wantErr: false,
		},
		{
			name:    "valid long segments",
			input:   "my-awesome-agent",
			wantErr: false,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
			errMsg:  "agent name is required",
		},
		{
			name:    "uppercase",
			input:   "MyAgent",
			wantErr: true,
			errMsg:  "lowercase",
		},
		{
			name:    "leading hyphen",
			input:   "-agent",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "trailing hyphen",
			input:   "agent-",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "consecutive hyphens",
			input:   "my--agent",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "underscore",
			input:   "my_agent",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "starts with number",
			input:   "1agent",
			wantErr: true,
			errMsg:  "starting with a letter",
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 65),
			wantErr: true,
			errMsg:  "at most 64 characters",
		},
		{
			name:    "exactly 64 chars is ok",
			input:   strings.Repeat("a", 64),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgentName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateAgentName(%q) expected error, got nil", tt.input)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateAgentName(%q) error = %q, want to contain %q", tt.input, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateAgentName(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

func TestSanitizeDefaultAgentName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "myagent",
			expected: "myagent",
		},
		{
			name:     "uppercase",
			input:    "MyAgent",
			expected: "myagent",
		},
		{
			name:     "spaces",
			input:    "my agent",
			expected: "my-agent",
		},
		{
			name:     "underscores",
			input:    "my_agent",
			expected: "my-agent",
		},
		{
			name:     "special chars",
			input:    "my@agent!",
			expected: "my-agent",
		},
		{
			name:     "leading invalid",
			input:    "-myagent",
			expected: "myagent",
		},
		{
			name:     "trailing invalid",
			input:    "myagent-",
			expected: "myagent",
		},
		{
			name:     "empty after sanitize",
			input:    "---",
			expected: "new-agent",
		},
		{
			name:     "starts with number",
			input:    "123agent",
			expected: "new-agent",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "new-agent",
		},
		{
			name:     "only special chars",
			input:    "!!!",
			expected: "new-agent",
		},
		{
			name:     "mixed case with special",
			input:    "My Cool_Agent!",
			expected: "my-cool-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeDefaultAgentName(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeDefaultAgentName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAgentInitCommand_Metadata(t *testing.T) {
	if agentInitCmd.Use != "init [path]" {
		t.Errorf("Use = %q, want %q", agentInitCmd.Use, "init [path]")
	}

	if agentInitCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// MaximumNArgs(1) allows 0 or 1 args
	if agentInitCmd.Args == nil {
		t.Error("Args validator should be set")
	}
}

func TestAgentNameRegex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "valid lowercase",
			input: "agent",
			want:  true,
		},
		{
			name:  "valid with hyphen",
			input: "my-agent",
			want:  true,
		},
		{
			name:  "valid with numbers",
			input: "v2",
			want:  true,
		},
		{
			name:  "valid number after hyphen",
			input: "agent-v2",
			want:  true,
		},
		{
			name:  "invalid uppercase",
			input: "Agent",
			want:  false,
		},
		{
			name:  "invalid starts with number",
			input: "2agent",
			want:  false,
		},
		{
			name:  "invalid double hyphen",
			input: "my--agent",
			want:  false,
		},
		{
			name:  "invalid trailing hyphen",
			input: "agent-",
			want:  false,
		},
		{
			name:  "invalid leading hyphen",
			input: "-agent",
			want:  false,
		},
		{
			name:  "invalid underscore",
			input: "my_agent",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agentNameRegex.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("agentNameRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAgentInitFlags(t *testing.T) {
	flags := []struct {
		name      string
		shorthand string
	}{
		{"name", ""},
		{"description", "d"},
		{"model", ""},
		{"force", "f"},
	}
	for _, f := range flags {
		flag := agentInitCmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag --%s not found", f.name)
			continue
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag --%s shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}

// resetAgentInitFlags resets all package-level flag variables to their default values.
func resetAgentInitFlags() {
	agentInitName = ""
	agentInitDescription = ""
	agentInitModel = ""
	agentInitForce = false
}

func TestAgentInitCmd_NonInteractive(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "test-agent")

	// Set flags for non-interactive execution
	agentInitName = "test-agent"
	agentInitDescription = "A test agent"
	agentInitModel = ""
	agentInitForce = false
	t.Cleanup(resetAgentInitFlags)

	// Run command with path argument
	err := runAgentInit(agentInitCmd, []string{targetDir})
	if err != nil {
		t.Fatalf("runAgentInit failed: %v", err)
	}

	// Verify file exists
	agentFile := filepath.Join(targetDir, "AGENT.md")
	content, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("failed to read agent file: %v", err)
	}

	// Verify content contains expected values
	if !strings.Contains(string(content), "name: test-agent") {
		t.Error("agent file should contain name")
	}
	if !strings.Contains(string(content), "description: A test agent") {
		t.Error("agent file should contain description")
	}
}

func TestAgentInitCmd_ForceOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "force-test")
	agentFile := filepath.Join(targetDir, "AGENT.md")

	// Create the directory and initial file
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	initialContent := "# Initial content\n"
	if err := os.WriteFile(agentFile, []byte(initialContent), 0o644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	// First, verify that without --force we get an error
	agentInitName = "force-test"
	agentInitDescription = "Force test agent"
	agentInitModel = ""
	agentInitForce = false
	t.Cleanup(resetAgentInitFlags)

	err := runAgentInit(agentInitCmd, []string{targetDir})
	if err == nil {
		t.Error("expected error when file exists and --force is not set")
	}

	// Verify original content is preserved
	content, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("failed to read agent file: %v", err)
	}
	if string(content) != initialContent {
		t.Error("file should not be modified without --force")
	}

	// Now test with --force
	agentInitForce = true
	err = runAgentInit(agentInitCmd, []string{targetDir})
	if err != nil {
		t.Fatalf("runAgentInit with --force failed: %v", err)
	}

	// Verify content was overwritten
	content, err = os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("failed to read agent file after force: %v", err)
	}
	if !strings.Contains(string(content), "name: force-test") {
		t.Error("agent file should contain new name after --force")
	}
	if !strings.Contains(string(content), "description: Force test agent") {
		t.Error("agent file should contain new description after --force")
	}
}
