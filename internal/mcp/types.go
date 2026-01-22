package mcp

import (
	"encoding/json"
)

// Transport type constants for MCP server communication.
const (
	// TransportStdio indicates local process communication via stdin/stdout.
	// This is the default transport when a Command is specified.
	TransportStdio = "stdio"

	// TransportSSE indicates remote server communication via Server-Sent Events.
	// This transport is used when connecting to remote MCP servers via HTTP.
	TransportSSE = "sse"
)

// Server represents a canonical MCP server configuration that can be
// translated to and from platform-specific formats.
type Server struct {
	// Name is the server's unique identifier.
	// This is typically used as the map key in configuration files.
	Name string `json:"name"`

	// Command is the executable path for local (stdio) servers.
	// Required for local servers, empty for remote servers.
	Command string `json:"command,omitempty"`

	// Args are command-line arguments passed to the Command executable.
	// Only applicable for local servers.
	Args []string `json:"args,omitempty"`

	// URL is the server endpoint for remote (SSE) servers.
	// Required for remote servers, empty for local servers.
	URL string `json:"url,omitempty"`

	// Transport specifies the communication protocol: "stdio" or "sse".
	// Defaults to "stdio" if Command is set, "sse" if URL is set.
	Transport string `json:"transport,omitempty"`

	// Env contains environment variables passed to the server process.
	// Only applicable for local servers.
	Env map[string]string `json:"env,omitempty"`

	// Headers contains HTTP headers for SSE transport connections.
	// Only applicable for remote servers.
	Headers map[string]string `json:"headers,omitempty"`

	// Platforms restricts the server to specific OS platforms.
	// Valid values: "darwin", "linux", "windows".
	// Empty means all platforms.
	Platforms []string `json:"platforms,omitempty"`

	// Disabled indicates whether the server is temporarily disabled.
	Disabled bool `json:"disabled,omitempty"`

	// unknownFields stores JSON fields not explicitly defined in this struct.
	// This ensures forward compatibility when MCP adds new server fields.
	unknownFields map[string]json.RawMessage
}

// IsLocal returns true if this server uses local (stdio) transport.
// A server is considered local if it has a Command or explicit stdio transport.
func (s *Server) IsLocal() bool {
	if s.Transport == TransportStdio {
		return true
	}
	if s.Transport == "" && s.Command != "" {
		return true
	}
	return false
}

// IsRemote returns true if this server uses remote (SSE) transport.
// A server is considered remote if it has a URL or explicit SSE transport.
func (s *Server) IsRemote() bool {
	if s.Transport == TransportSSE {
		return true
	}
	if s.Transport == "" && s.URL != "" && s.Command == "" {
		return true
	}
	return false
}

// MarshalJSON implements json.Marshaler to include unknown fields in output.
func (s *Server) MarshalJSON() ([]byte, error) {
	// Build a map with all fields
	result := make(map[string]any)

	// Copy unknown fields first (so known fields take precedence)
	for k, v := range s.unknownFields {
		var val any
		if err := json.Unmarshal(v, &val); err != nil {
			return nil, err
		}
		result[k] = val
	}

	// Add known fields (only if non-zero to match omitempty behavior)
	result["name"] = s.Name
	if s.Command != "" {
		result["command"] = s.Command
	}
	if len(s.Args) > 0 {
		result["args"] = s.Args
	}
	if s.URL != "" {
		result["url"] = s.URL
	}
	if s.Transport != "" {
		result["transport"] = s.Transport
	}
	if len(s.Env) > 0 {
		result["env"] = s.Env
	}
	if len(s.Headers) > 0 {
		result["headers"] = s.Headers
	}
	if len(s.Platforms) > 0 {
		result["platforms"] = s.Platforms
	}
	if s.Disabled {
		result["disabled"] = s.Disabled
	}

	return json.Marshal(result)
}

// UnmarshalJSON implements json.Unmarshaler to capture unknown fields.
func (s *Server) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a generic map to capture all fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract known fields
	if v, ok := raw["name"]; ok {
		if err := json.Unmarshal(v, &s.Name); err != nil {
			return err
		}
		delete(raw, "name")
	}
	if v, ok := raw["command"]; ok {
		if err := json.Unmarshal(v, &s.Command); err != nil {
			return err
		}
		delete(raw, "command")
	}
	if v, ok := raw["args"]; ok {
		if err := json.Unmarshal(v, &s.Args); err != nil {
			return err
		}
		delete(raw, "args")
	}
	if v, ok := raw["url"]; ok {
		if err := json.Unmarshal(v, &s.URL); err != nil {
			return err
		}
		delete(raw, "url")
	}
	if v, ok := raw["transport"]; ok {
		if err := json.Unmarshal(v, &s.Transport); err != nil {
			return err
		}
		delete(raw, "transport")
	}
	if v, ok := raw["env"]; ok {
		if err := json.Unmarshal(v, &s.Env); err != nil {
			return err
		}
		delete(raw, "env")
	}
	if v, ok := raw["headers"]; ok {
		if err := json.Unmarshal(v, &s.Headers); err != nil {
			return err
		}
		delete(raw, "headers")
	}
	if v, ok := raw["platforms"]; ok {
		if err := json.Unmarshal(v, &s.Platforms); err != nil {
			return err
		}
		delete(raw, "platforms")
	}
	if v, ok := raw["disabled"]; ok {
		if err := json.Unmarshal(v, &s.Disabled); err != nil {
			return err
		}
		delete(raw, "disabled")
	}

	// Store remaining fields as unknown
	if len(raw) > 0 {
		s.unknownFields = raw
	}

	return nil
}

// Config represents a canonical MCP configuration containing server definitions.
// It can be translated to and from platform-specific configuration formats.
type Config struct {
	// Servers maps server names to their configurations.
	Servers map[string]*Server `json:"servers"`

	// unknownFields stores JSON fields not explicitly defined in this struct.
	// This ensures forward compatibility when MCP adds new top-level fields.
	unknownFields map[string]json.RawMessage
}

// NewConfig creates a new Config with initialized maps.
func NewConfig() *Config {
	return &Config{
		Servers: make(map[string]*Server),
	}
}

// MarshalJSON implements json.Marshaler to include unknown fields in output.
func (c *Config) MarshalJSON() ([]byte, error) {
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
	result["servers"] = c.Servers

	return json.Marshal(result)
}

// UnmarshalJSON implements json.Unmarshaler to capture unknown fields.
func (c *Config) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a generic map to capture all fields
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract the known field
	if serversData, ok := raw["servers"]; ok {
		if err := json.Unmarshal(serversData, &c.Servers); err != nil {
			return err
		}
		delete(raw, "servers")
	}

	// Store remaining fields as unknown
	if len(raw) > 0 {
		c.unknownFields = raw
	}

	return nil
}
