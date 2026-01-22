// Package claude provides data models for Claude Code configuration entities.
//
// This package defines Go structs that model Claude Code's configuration files,
// including MCP server definitions, skills, commands, and agents. All types
// support JSON and YAML serialization with proper struct tags.
//
// # MCP Configuration
//
// Claude Code uses MCP (Model Context Protocol) servers for extensibility.
// The [MCPConfig] type models the .mcp.json file format:
//
//	config := &claude.MCPConfig{
//	    MCPServers: map[string]*claude.MCPServer{
//	        "github": {
//	            Name:    "github",
//	            Command: "npx",
//	            Args:    []string{"-y", "@modelcontextprotocol/server-github"},
//	            Env: map[string]string{
//	                "GITHUB_TOKEN": "${GITHUB_TOKEN}",
//	            },
//	        },
//	    },
//	}
//
// [MCPConfig] preserves unknown fields during JSON round-trips for forward
// compatibility with future Claude Code versions.
//
// # Skills
//
// Skills are markdown files with YAML frontmatter defining reusable capabilities:
//
//	skill := &claude.Skill{
//	    Name:         "code-review",
//	    Description:  "Perform thorough code reviews",
//	    Version:      "1.0.0",
//	    Tools:        []string{"Read", "Grep", "Glob"},
//	    Triggers:     []string{"review", "code review"},
//	    Instructions: "When reviewing code, check for...",
//	}
//
// The Instructions field contains the markdown body content and is excluded
// from YAML/JSON serialization (marked with "-" tag).
//
// # Commands
//
// Commands define slash commands available in the Claude Code interface:
//
//	cmd := &claude.Command{
//	    Name:         "test",
//	    Description:  "Run test suite",
//	    Instructions: "Execute the test runner...",
//	}
//
// # Agents
//
// Agents define specialized AI assistants with custom instructions:
//
//	agent := &claude.Agent{
//	    Name:         "reviewer",
//	    Description:  "Code review specialist",
//	    Instructions: "You are a code review expert...",
//	}
//
// # Serialization
//
// All types support both JSON and YAML serialization. Use the standard
// library [encoding/json] package or [gopkg.in/yaml.v3] for marshaling:
//
//	// JSON
//	data, err := json.Marshal(config)
//
//	// YAML
//	data, err := yaml.Marshal(skill)
//
// Fields with ",omitempty" tags are omitted when empty to produce
// cleaner output.
package claude
