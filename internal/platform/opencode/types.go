package opencode

import (
	"encoding/json"
	"strings"

	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
)

// MCPServer represents an MCP (Model Context Protocol) server configuration
// for OpenCode. OpenCode uses a different structure than Claude Code, with
// command as a string array and type-based transport selection.
type MCPServer struct {
	// Name is the server's identifier, derived from the map key.
	// Not serialized to JSON since OpenCode uses the map key as the name.
	Name string `json:"-"`

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

	// Enabled indicates whether the server is active.
	// OpenCode uses positive logic (enabled=true means active).
	// Pointer type to distinguish unset from explicitly false.
	Enabled *bool `json:"enabled,omitempty"`
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

		// Set each server's Name from its map key
		for name, server := range c.MCP {
			if server != nil {
				server.Name = name
			}
		}
	}

	// Store remaining fields as unknown
	if len(raw) > 0 {
		c.unknownFields = raw
	}

	return nil
}

// CompatibilityMap maps platform names to version requirements.
// It supports unmarshaling from both a map (OpenCode format) and a list (Spec format).
type CompatibilityMap map[string]string

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *CompatibilityMap) UnmarshalYAML(value *yaml.Node) error {
	// Try unmarshaling as a map first (standard OpenCode format)
	var m map[string]string
	if err := value.Decode(&m); err == nil {
		*c = m
		return nil
	}

	// Try unmarshaling as a list (Spec format: ["claude >=1.0", "opencode"])
	var l []string
	if err := value.Decode(&l); err == nil {
		if *c == nil {
			*c = make(map[string]string)
		}
		for _, item := range l {
			// Split "platform >=version" or just "platform"
			parts := strings.SplitN(item, " ", 2)
			platform := parts[0]
			version := ""
			if len(parts) > 1 {
				version = strings.TrimSpace(parts[1])
			}
			(*c)[platform] = version
		}
		return nil
	}

	return errors.Newf("compatibility must be a map or list, got %s", value.Tag)
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
	Compatibility CompatibilityMap `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`

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

	// Agent specifies which agent should handle this command.
	// When set, the command will be executed by the named agent instead of the default.
	Agent string `yaml:"agent,omitempty" json:"agent,omitempty"`

	// Model specifies which model should be used for this command.
	// Overrides the default model selection for command execution.
	Model string `yaml:"model,omitempty" json:"model,omitempty"`

	// Subtask indicates whether the command runs as a subtask.
	// When true, the command executes in a separate context from the main conversation.
	Subtask bool `yaml:"subtask,omitempty" json:"subtask,omitempty"`

	// Template defines the output template for the command.
	// Supports variable substitution for structured command output.
	Template string `yaml:"template,omitempty" json:"template,omitempty"`

	// Instructions contains the command's markdown body content.
	// This field is not part of the YAML frontmatter.
	Instructions string `yaml:"-" json:"-"`
}

// GetName returns the command's name.
func (c *Command) GetName() string {
	return c.Name
}

// SetName sets the command's name.
func (c *Command) SetName(name string) {
	c.Name = name
}

// SetInstructions sets the command's instructions.
func (c *Command) SetInstructions(instructions string) {
	c.Instructions = instructions
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

// GetName returns the agent's name.
func (a *Agent) GetName() string {
	return a.Name
}

// GetDescription returns the agent's description.
func (a *Agent) GetDescription() string {
	return a.Description
}

// GetInstructions returns the agent's instructions.
func (a *Agent) GetInstructions() string {
	return a.Instructions
}
