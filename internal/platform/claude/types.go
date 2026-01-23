package claude

import (
	"encoding/json"
	"strings"

	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
)

// MCPServer represents an MCP (Model Context Protocol) server configuration.
// It supports both stdio-based (Command/Args) and HTTP-based (URL) transports.
type MCPServer struct {
	// Name is the server's identifier, populated from the map key when loading.
	// Not serialized to JSON as it's the map key itself.
	Name string `json:"-"`

	// Type specifies the server transport type: "stdio" or "http".
	// Note: Claude Code uses "http" for remote servers, while canonical uses "sse".
	Type string `json:"type,omitempty"`

	// Command is the executable to run for stdio transport.
	Command string `json:"command,omitempty"`

	// Args are command-line arguments passed to the command.
	Args []string `json:"args,omitempty"`

	// URL is the server endpoint for HTTP transport.
	URL string `json:"url,omitempty"`

	// Env contains environment variables passed to the server process.
	Env map[string]string `json:"env,omitempty"`

	// Headers contains HTTP headers for HTTP transport connections.
	Headers map[string]string `json:"headers,omitempty"`

	// Platforms restricts the server to specific OS platforms (e.g., "darwin", "linux").
	Platforms []string `json:"platforms,omitempty"`

	// Disabled indicates whether the server is temporarily disabled.
	Disabled bool `json:"disabled,omitempty"`
}

// MCPConfig represents the root structure of Claude Code's .mcp.json file.
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

// ToolList is a list of allowed tools.
// It supports unmarshaling from both a space-delimited string and a list of strings.
type ToolList []string

// UnmarshalYAML implements yaml.Unmarshaler.
func (t *ToolList) UnmarshalYAML(value *yaml.Node) error {
	var multi []string
	if err := value.Decode(&multi); err == nil {
		*t = multi
		return nil
	}

	var single string
	if err := value.Decode(&single); err == nil {
		if single == "" {
			*t = nil
			return nil
		}
		// Split space-delimited string
		for part := range strings.SplitSeq(single, " ") {
			part = strings.TrimSpace(part)
			if part != "" {
				*t = append(*t, part)
			}
		}
		return nil
	}

	return errors.Newf("allowed-tools must be a string or list of strings, got %s", value.Tag)
}

// String returns the space-delimited string representation.
func (t ToolList) String() string {
	return strings.Join(t, " ")
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

	// AllowedTools lists the tool permissions required by this skill.
	// Can be a space-delimited string or a list of strings.
	AllowedTools ToolList `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`

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

	// ArgumentHint provides hint text shown for command arguments in the UI.
	ArgumentHint string `yaml:"argument-hint,omitempty" json:"argumentHint,omitempty"`

	// DisableModelInvocation prevents the model from invoking this command automatically.
	DisableModelInvocation bool `yaml:"disable-model-invocation,omitempty" json:"disableModelInvocation,omitempty"`

	// UserInvocable indicates whether the user can invoke this command directly.
	UserInvocable bool `yaml:"user-invocable,omitempty" json:"userInvocable,omitempty"`

	// AllowedTools lists the tool permissions available to this command.
	// Can be a space-delimited string or a list of strings.
	AllowedTools ToolList `yaml:"allowed-tools,omitempty" json:"allowedTools,omitempty"`

	// Model specifies which model to use when executing this command.
	Model string `yaml:"model,omitempty" json:"model,omitempty"`

	// Context specifies the context mode for this command (e.g., "none", "file", "repo").
	Context string `yaml:"context,omitempty" json:"context,omitempty"`

	// Agent specifies the agent to use when executing this command.
	Agent string `yaml:"agent,omitempty" json:"agent,omitempty"`

	// Hooks lists hooks to run during command execution.
	Hooks []string `yaml:"hooks,omitempty" json:"hooks,omitempty"`

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
