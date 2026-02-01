package gemini

import (
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thoreinstein/aix/internal/errors"
)

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
		for _, part := range strings.Fields(single) {
			*t = append(*t, part)
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
type Skill struct {
	// Name is the skill's unique identifier (required).
	Name string `yaml:"name" json:"name"`

	// Description explains what the skill does (required).
	Description string `yaml:"description" json:"description"`

	// License is the SPDX license identifier (optional).
	License string `yaml:"license,omitempty" json:"license,omitempty"`

	// Compatibility lists compatible AI assistants (optional).
	Compatibility []string `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`

	// Metadata contains optional key-value pairs like author, version, repository.
	Metadata map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// AllowedTools lists the tool permissions required by this skill.
	AllowedTools ToolList `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`

	// Instructions contains the skill's markdown body content.
	// This field is not part of the YAML frontmatter.
	Instructions string `yaml:"-" json:"-"`
}

// Command represents a Gemini CLI slash command definition.
type Command struct {
	// Name is the command's identifier.
	Name string `yaml:"name" json:"name" toml:"-"`

	// Description explains what the command does.
	Description string `yaml:"description,omitempty" json:"description,omitempty" toml:"description,omitempty"`

	// Instructions contains the command's markdown body content.
	Instructions string `yaml:"-" json:"-" toml:"prompt"`
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

// Agent represents a Gemini CLI agent definition.
type Agent struct {
	// Name is the agent's identifier.
	Name string `yaml:"name" json:"name" toml:"name"`

	// Description explains the agent's purpose.
	Description string `yaml:"description,omitempty" json:"description,omitempty" toml:"description,omitempty"`

	// Instructions contains the agent's markdown body content.
	Instructions string `yaml:"-" json:"-" toml:"instructions,multiline"`
}

// MCPServer represents an MCP server configuration for Gemini CLI.
// Gemini uses an "inferred" transport type based on presence of command or url.
type MCPServer struct {
	// Name is the server's identifier, derived from the map key.
	Name string `json:"-" toml:"-"`

	// Command is the executable path for local servers.
	Command string `json:"command,omitempty" toml:"command,omitempty"`

	// Args are command-line arguments for the server process.
	Args []string `json:"args,omitempty" toml:"args,omitempty"`

	// URL is the server endpoint for remote servers.
	URL string `json:"url,omitempty" toml:"url,omitempty"`

	// Env contains environment variables for the server process.
	Env map[string]string `json:"env,omitempty" toml:"env,omitempty"`

	// Headers contains HTTP headers for remote connections.
	Headers map[string]string `json:"headers,omitempty" toml:"headers,omitempty"`

	// Enabled indicates whether the server is active.
	Enabled bool `json:"enabled" toml:"enabled"`
}

// MCPConfig represents the MCP section in Gemini CLI's settings.toml.
type MCPConfig struct {
	// Servers maps server names to their configurations.
	Servers map[string]*MCPServer `json:"servers" toml:"servers"`
}

// ExperimentalConfig represents the experimental section in settings.toml.
type ExperimentalConfig struct {
	// EnableAgents indicates whether custom sub-agents are enabled.
	EnableAgents bool `json:"enableAgents" toml:"enableAgents"`
}

// Settings represents the root structure of Gemini CLI's settings.toml.
type Settings struct {
	// MCP contains the MCP server configurations.
	MCP *MCPConfig `json:"mcp,omitempty" toml:"mcp,omitempty"`

	// Experimental contains experimental feature flags.
	Experimental *ExperimentalConfig `json:"experimental,omitempty" toml:"experimental,omitempty"`

	// Other stores any other fields in settings.toml to preserve them.
	Other map[string]any `json:"-" toml:"-"`
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
