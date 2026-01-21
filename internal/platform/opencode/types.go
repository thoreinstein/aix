package opencode

import (
	"encoding/json"
)

// MCPServer represents an MCP (Model Context Protocol) server configuration
// for OpenCode. OpenCode uses a different structure than Claude Code, with
// command as a string array and type-based transport selection.
type MCPServer struct {
	// Name is the server's identifier, typically matching the map key.
	Name string `json:"name"`

	// Command is the executable and its arguments as a single array.
	// Unlike Claude Code which separates command and args, OpenCode combines them.
	Command []string `json:"command,omitempty"`

	// Type specifies the server type: "local" for stdio or "remote" for HTTP/SSE.
	Type string `json:"type,omitempty"`

	// URL is the server endpoint for remote (HTTP/SSE) transport.
	URL string `json:"url,omitempty"`

	// Environment contains environment variables passed to the server process.
	// Note: OpenCode uses "environment" rather than Claude's "env".
	Environment map[string]string `json:"environment,omitempty"`

	// Headers contains HTTP headers for remote transport connections.
	Headers map[string]string `json:"headers,omitempty"`

	// Disabled indicates whether the server is temporarily disabled.
	Disabled bool `json:"disabled,omitempty"`
}

// MCPConfig represents the root structure of OpenCode's MCP configuration.
// In OpenCode, MCP servers are stored under the "mcp" key (not "mcpServers").
// It preserves unknown fields for forward compatibility with future versions.
type MCPConfig struct {
	// MCP maps server names to their configurations.
	MCP map[string]*MCPServer `json:"mcp"`

	// unknownFields stores any JSON fields not explicitly defined in this struct.
	// This ensures forward compatibility when OpenCode adds new top-level fields.
	unknownFields map[string]json.RawMessage
}

// MarshalJSON implements json.Marshaler to include unknown fields in output.
func (c *MCPConfig) MarshalJSON() ([]byte, error) {
	// Build a map with all fields
	result := make(map[string]any)

	// Copy unknown fields first (so known fields take precedence)
	for k, v := range c.unknownFields {
		var val any
		if err := json.Unmarshal(v, &val); err != nil {
			return nil, err
		}
		result[k] = val
	}

	// Add the known field
	result["mcp"] = c.MCP

	return json.Marshal(result)
}

// UnmarshalJSON implements json.Unmarshaler to capture unknown fields.
func (c *MCPConfig) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a generic map to capture all fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract the known field
	if mcpData, ok := raw["mcp"]; ok {
		if err := json.Unmarshal(mcpData, &c.MCP); err != nil {
			return err
		}
		delete(raw, "mcp")
	}

	// Store remaining fields as unknown
	if len(raw) > 0 {
		c.unknownFields = raw
	}

	return nil
}

// Skill represents an OpenCode skill definition.
// Skills are markdown files with YAML frontmatter that define reusable capabilities.
// OpenCode extends the base skill spec with additional fields for tool restrictions
// and platform compatibility.
type Skill struct {
	// Name is the skill's unique identifier.
	Name string `yaml:"name" json:"name"`

	// Description explains what the skill does.
	Description string `yaml:"description" json:"description"`

	// Version is the semantic version of the skill.
	Version string `yaml:"version,omitempty" json:"version,omitempty"`

	// Author is the skill creator's name or identifier.
	Author string `yaml:"author,omitempty" json:"author,omitempty"`

	// Tools lists the tool permissions required by this skill.
	Tools []string `yaml:"tools,omitempty" json:"tools,omitempty"`

	// AllowedTools restricts which tools the skill can use.
	// This is an OpenCode-specific field for fine-grained tool control.
	AllowedTools []string `yaml:"allowed_tools,omitempty" json:"allowedTools,omitempty"`

	// Triggers are phrases that activate this skill.
	Triggers []string `yaml:"triggers,omitempty" json:"triggers,omitempty"`

	// Compatibility maps platform names to compatibility notes or version requirements.
	// This is an OpenCode-specific field for cross-platform skill management.
	Compatibility map[string]string `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`

	// Metadata contains arbitrary key-value pairs for extensibility.
	// This is an OpenCode-specific field for custom skill metadata.
	Metadata map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Instructions contains the skill's markdown body content.
	// This field is not part of the YAML frontmatter.
	Instructions string `yaml:"-" json:"-"`
}

// Command represents an OpenCode slash command definition.
// Commands are markdown files that define custom slash commands.
type Command struct {
	// Name is the command's identifier (used as /name in the interface).
	Name string `yaml:"name" json:"name"`

	// Description explains what the command does.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Instructions contains the command's markdown body content.
	// This field is not part of the YAML frontmatter.
	Instructions string `yaml:"-" json:"-"`
}

// Agent represents an OpenCode agent definition.
// Agents are markdown files that define specialized AI assistants.
// OpenCode extends the base agent spec with mode and temperature controls.
type Agent struct {
	// Name is the agent's identifier.
	Name string `yaml:"name" json:"name"`

	// Description explains the agent's purpose and capabilities.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Mode specifies the agent's operational mode (e.g., "chat", "edit", "review").
	// This is an OpenCode-specific field for agent behavior customization.
	Mode string `yaml:"mode,omitempty" json:"mode,omitempty"`

	// Temperature controls the randomness of the agent's responses.
	// Lower values (0.0-0.3) are more deterministic, higher values (0.7-1.0) are more creative.
	// This is an OpenCode-specific field for fine-tuning agent behavior.
	Temperature float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`

	// Instructions contains the agent's markdown body content.
	// This field is not part of the YAML frontmatter.
	Instructions string `yaml:"-" json:"-"`
}
