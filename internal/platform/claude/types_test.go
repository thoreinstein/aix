package claude

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMCPServer_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		server *MCPServer
	}{
		{
			name: "minimal server",
			server: &MCPServer{
				Name:    "test",
				Command: "test-cmd",
			},
		},
		{
			name: "stdio server with args and env",
			server: &MCPServer{
				Name:    "github",
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-github"},
				Env: map[string]string{
					"GITHUB_TOKEN": "${GITHUB_TOKEN}",
				},
			},
		},
		{
			name: "sse server with headers",
			server: &MCPServer{
				Name:      "remote",
				URL:       "https://api.example.com/mcp",
				Transport: "sse",
				Headers: map[string]string{
					"Authorization": "Bearer ${API_KEY}",
				},
			},
		},
		{
			name: "platform-specific disabled server",
			server: &MCPServer{
				Name:      "linux-only",
				Command:   "linux-tool",
				Platforms: []string{"linux"},
				Disabled:  true,
			},
		},
		{
			name: "full server configuration",
			server: &MCPServer{
				Name:      "full",
				Command:   "full-cmd",
				Args:      []string{"--verbose", "--config", "/etc/full.conf"},
				URL:       "http://localhost:8080",
				Transport: "stdio",
				Env: map[string]string{
					"KEY1": "value1",
					"KEY2": "value2",
				},
				Headers: map[string]string{
					"X-Custom": "header",
				},
				Platforms: []string{"darwin", "linux"},
				Disabled:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.server)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Unmarshal back
			var got MCPServer
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Compare
			if !reflect.DeepEqual(&got, tt.server) {
				t.Errorf("round-trip mismatch:\ngot:  %+v\nwant: %+v", got, tt.server)
			}
		})
	}
}

func TestMCPConfig_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		config *MCPConfig
	}{
		{
			name: "empty config",
			config: &MCPConfig{
				MCPServers: map[string]*MCPServer{},
			},
		},
		{
			name: "single server",
			config: &MCPConfig{
				MCPServers: map[string]*MCPServer{
					"test": {
						Name:    "test",
						Command: "test-cmd",
					},
				},
			},
		},
		{
			name: "multiple servers",
			config: &MCPConfig{
				MCPServers: map[string]*MCPServer{
					"github": {
						Name:    "github",
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-github"},
					},
					"filesystem": {
						Name:    "filesystem",
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Unmarshal back
			var got MCPConfig
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Compare MCPServers
			if !reflect.DeepEqual(got.MCPServers, tt.config.MCPServers) {
				t.Errorf("MCPServers mismatch:\ngot:  %+v\nwant: %+v", got.MCPServers, tt.config.MCPServers)
			}
		})
	}
}

func TestMCPConfig_PreservesUnknownFields(t *testing.T) {
	// JSON with unknown fields at root level
	input := `{
		"mcpServers": {
			"test": {
				"name": "test",
				"command": "test-cmd"
			}
		},
		"futureField": "future value",
		"anotherField": {
			"nested": true
		}
	}`

	// Unmarshal
	var config MCPConfig
	if err := json.Unmarshal([]byte(input), &config); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify known fields parsed correctly
	if len(config.MCPServers) != 1 {
		t.Errorf("MCPServers count = %d, want 1", len(config.MCPServers))
	}
	if config.MCPServers["test"].Command != "test-cmd" {
		t.Errorf("MCPServers[test].Command = %q, want %q", config.MCPServers["test"].Command, "test-cmd")
	}

	// Marshal back
	data, err := json.Marshal(&config)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal to generic map to verify unknown fields preserved
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal to map error = %v", err)
	}

	// Check unknown fields are present
	if _, ok := result["futureField"]; !ok {
		t.Error("futureField not preserved in output")
	}
	if _, ok := result["anotherField"]; !ok {
		t.Error("anotherField not preserved in output")
	}

	// Verify futureField value
	if result["futureField"] != "future value" {
		t.Errorf("futureField = %v, want %q", result["futureField"], "future value")
	}

	// Verify nested unknown field
	nested, ok := result["anotherField"].(map[string]any)
	if !ok {
		t.Fatalf("anotherField is not a map: %T", result["anotherField"])
	}
	if nested["nested"] != true {
		t.Errorf("anotherField.nested = %v, want true", nested["nested"])
	}
}

func TestSkill_YAMLRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		skill *Skill
	}{
		{
			name: "minimal skill",
			skill: &Skill{
				Name:        "test",
				Description: "A test skill",
			},
		},
		{
			name: "skill with version and author",
			skill: &Skill{
				Name:        "code-review",
				Description: "Perform code reviews",
				Version:     "1.0.0",
				Author:      "aix",
			},
		},
		{
			name: "skill with tools and triggers",
			skill: &Skill{
				Name:        "search",
				Description: "Search the codebase",
				Tools:       []string{"Read", "Grep", "Glob"},
				Triggers:    []string{"search", "find", "look for"},
			},
		},
		{
			name: "full skill",
			skill: &Skill{
				Name:        "full-skill",
				Description: "A fully specified skill",
				Version:     "2.1.0",
				Author:      "test-author",
				Tools:       []string{"Bash", "Read", "Write"},
				Triggers:    []string{"trigger1", "trigger2"},
				// Instructions is not serialized
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to YAML
			data, err := yaml.Marshal(tt.skill)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Unmarshal back
			var got Skill
			if err := yaml.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Instructions should remain empty after round-trip (it's marked with "-")
			tt.skill.Instructions = ""

			// Compare
			if !reflect.DeepEqual(&got, tt.skill) {
				t.Errorf("round-trip mismatch:\ngot:  %+v\nwant: %+v", got, tt.skill)
			}
		})
	}
}

func TestSkill_InstructionsNotSerialized(t *testing.T) {
	skill := &Skill{
		Name:         "test",
		Description:  "A test skill",
		Instructions: "These instructions should not appear in YAML or JSON",
	}

	// Test YAML
	yamlData, err := yaml.Marshal(skill)
	if err != nil {
		t.Fatalf("YAML Marshal() error = %v", err)
	}
	yamlStr := string(yamlData)
	if contains(yamlStr, "instructions") || contains(yamlStr, "These instructions") {
		t.Errorf("Instructions field should not be in YAML output:\n%s", yamlStr)
	}

	// Test JSON
	jsonData, err := json.Marshal(skill)
	if err != nil {
		t.Fatalf("JSON Marshal() error = %v", err)
	}
	jsonStr := string(jsonData)
	if contains(jsonStr, "instructions") || contains(jsonStr, "These instructions") {
		t.Errorf("Instructions field should not be in JSON output:\n%s", jsonStr)
	}
}

func TestCommand_Serialization(t *testing.T) {
	tests := []struct {
		name string
		cmd  *Command
	}{
		{
			name: "minimal command",
			cmd: &Command{
				Name: "test",
			},
		},
		{
			name: "command with description",
			cmd: &Command{
				Name:        "build",
				Description: "Build the project",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" YAML", func(t *testing.T) {
			data, err := yaml.Marshal(tt.cmd)
			if err != nil {
				t.Fatalf("YAML Marshal() error = %v", err)
			}

			var got Command
			if err := yaml.Unmarshal(data, &got); err != nil {
				t.Fatalf("YAML Unmarshal() error = %v", err)
			}

			if got.Name != tt.cmd.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.cmd.Name)
			}
			if got.Description != tt.cmd.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.cmd.Description)
			}
		})

		t.Run(tt.name+" JSON", func(t *testing.T) {
			data, err := json.Marshal(tt.cmd)
			if err != nil {
				t.Fatalf("JSON Marshal() error = %v", err)
			}

			var got Command
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("JSON Unmarshal() error = %v", err)
			}

			if got.Name != tt.cmd.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.cmd.Name)
			}
			if got.Description != tt.cmd.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.cmd.Description)
			}
		})
	}
}

func TestCommand_InstructionsNotSerialized(t *testing.T) {
	cmd := &Command{
		Name:         "test",
		Description:  "A test command",
		Instructions: "These instructions should not appear in output",
	}

	// Test YAML
	yamlData, err := yaml.Marshal(cmd)
	if err != nil {
		t.Fatalf("YAML Marshal() error = %v", err)
	}
	if contains(string(yamlData), "instructions") {
		t.Errorf("Instructions field should not be in YAML output:\n%s", yamlData)
	}

	// Test JSON
	jsonData, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("JSON Marshal() error = %v", err)
	}
	if contains(string(jsonData), "instructions") {
		t.Errorf("Instructions field should not be in JSON output:\n%s", jsonData)
	}
}

func TestAgent_Serialization(t *testing.T) {
	tests := []struct {
		name  string
		agent *Agent
	}{
		{
			name: "minimal agent",
			agent: &Agent{
				Name: "test",
			},
		},
		{
			name: "agent with description",
			agent: &Agent{
				Name:        "reviewer",
				Description: "Code review specialist",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" YAML", func(t *testing.T) {
			data, err := yaml.Marshal(tt.agent)
			if err != nil {
				t.Fatalf("YAML Marshal() error = %v", err)
			}

			var got Agent
			if err := yaml.Unmarshal(data, &got); err != nil {
				t.Fatalf("YAML Unmarshal() error = %v", err)
			}

			if got.Name != tt.agent.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.agent.Name)
			}
			if got.Description != tt.agent.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.agent.Description)
			}
		})

		t.Run(tt.name+" JSON", func(t *testing.T) {
			data, err := json.Marshal(tt.agent)
			if err != nil {
				t.Fatalf("JSON Marshal() error = %v", err)
			}

			var got Agent
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("JSON Unmarshal() error = %v", err)
			}

			if got.Name != tt.agent.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.agent.Name)
			}
			if got.Description != tt.agent.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.agent.Description)
			}
		})
	}
}

func TestAgent_InstructionsNotSerialized(t *testing.T) {
	agent := &Agent{
		Name:         "test",
		Description:  "A test agent",
		Instructions: "These instructions should not appear in output",
	}

	// Test YAML
	yamlData, err := yaml.Marshal(agent)
	if err != nil {
		t.Fatalf("YAML Marshal() error = %v", err)
	}
	if contains(string(yamlData), "instructions") {
		t.Errorf("Instructions field should not be in YAML output:\n%s", yamlData)
	}

	// Test JSON
	jsonData, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("JSON Marshal() error = %v", err)
	}
	if contains(string(jsonData), "instructions") {
		t.Errorf("Instructions field should not be in JSON output:\n%s", jsonData)
	}
}

// contains checks if s contains substr (case-sensitive).
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
