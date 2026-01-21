package claude

import (
	"encoding/json"
)

// MCPServer represents an MCP (Model Context Protocol) server configuration.
// It supports both stdio-based (Command/Args) and HTTP-based (URL) transports.
type MCPServer struct {
	// Name is the server's identifier, typically matching the map key.
	Name string `json:"name"`

	// Command is the executable to run for stdio transport.
	Command string `json:"command"`

	// Args are command-line arguments passed to the command.
	Args []string `json:"args,omitempty"`

	// URL is the server endpoint for HTTP/SSE transport.
	URL string `json:"url,omitempty"`

	// Transport specifies the protocol: "stdio" (default) or "sse".
	Transport string `json:"transport,omitempty"`

	// Env contains environment variables passed to the server process.
	Env map[string]string `json:"env,omitempty"`

	// Headers contains HTTP headers for SSE transport connections.
	Headers map[string]string `json:"headers,omitempty"`

	// Platforms restricts the server to specific OS platforms (e.g., "darwin", "linux").
	Platforms []string `json:"platforms,omitempty"`

	// Disabled indicates whether the server is temporarily disabled.
	Disabled bool `json:"disabled,omitempty"`
}

// MCPConfig represents the root structure of Claude Code's mcp_servers.json file.
// It preserves unknown fields for forward compatibility with future versions.
type MCPConfig struct {
	// MCPServers maps server names to their configurations.
	MCPServers map[string]*MCPServer `json:"mcpServers"`

	// unknownFields stores any JSON fields not explicitly defined in this struct.
	// This ensures forward compatibility when Claude Code adds new top-level fields.
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
	result["mcpServers"] = c.MCPServers

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
	if serversData, ok := raw["mcpServers"]; ok {
		if err := json.Unmarshal(serversData, &c.MCPServers); err != nil {
			return err
		}
		delete(raw, "mcpServers")
	}

	// Store remaining fields as unknown
	if len(raw) > 0 {
		c.unknownFields = raw
	}

	return nil
}

// Skill represents a skill definition per the Agent Skills Specification.
// Skills are markdown files with YAML frontmatter that define reusable capabilities.
// See: https://agentskills.io/specification
type Skill struct {
	// Name is the skill's unique identifier (required).
	// Must be 1-64 chars, lowercase alphanumeric + hyphens, no --, no start/end -.
	Name string `yaml:"name" json:"name"`

	// Description explains what the skill does (required).
	Description string `yaml:"description" json:"description"`

	// License is the SPDX license identifier (optional).
	License string `yaml:"license,omitempty" json:"license,omitempty"`

	// Compatibility lists compatible AI assistants (optional).
	// E.g., ["claude-code", "opencode", "codex"]
	Compatibility []string `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`

	// Metadata contains optional key-value pairs like author, version, repository.
	Metadata map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// AllowedTools is a space-delimited string of tool permissions (optional).
	// E.g., "Read Write Bash(git:*) Glob"
	AllowedTools string `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`

	// Instructions contains the skill's markdown body content.
	// This field is not part of the YAML frontmatter.
	Instructions string `yaml:"-" json:"-"`
}

// Command represents a Claude Code slash command definition.
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

// Agent represents a Claude Code agent definition.
// Agents are markdown files that define specialized AI assistants.
type Agent struct {
	// Name is the agent's identifier.
	Name string `yaml:"name" json:"name"`

	// Description explains the agent's purpose and capabilities.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Instructions contains the agent's markdown body content.
	// This field is not part of the YAML frontmatter.
	Instructions string `yaml:"-" json:"-"`
}
