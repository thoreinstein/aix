package opencode

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/thoreinstein/aix/internal/mcp"
)

// Type constants for OpenCode MCP server types.
const (
	// TypeLocal indicates a local process server (maps to canonical "stdio").
	TypeLocal = "local"

	// TypeRemote indicates a remote HTTP/SSE server (maps to canonical "sse").
	TypeRemote = "remote"
)

// MCPTranslator converts between canonical and OpenCode MCP formats.
//
// OpenCode uses different field names and structures:
//   - "mcp" key instead of "mcpServers"
//   - "command" is []string (combined cmd + args) instead of separate fields
//   - "type" ("local"/"remote") instead of "transport" ("stdio"/"sse")
//   - "environment" instead of "env"
//   - "enabled" (positive logic) instead of "disabled" (negative logic)
//   - No "platforms" field (LOSSY: this field is not preserved)
type MCPTranslator struct{}

// NewMCPTranslator creates a new OpenCode MCP translator.
func NewMCPTranslator() *MCPTranslator {
	return &MCPTranslator{}
}

// ToCanonical converts OpenCode MCP configuration to canonical format.
//
// Input format:
//
//	{"mcp": {"name": {...}, ...}}
//
// or just the servers map:
//
//	{"name": {...}, ...}
//
// Field mappings:
//   - Command ([]string) → Command (string) + Args ([]string)
//   - Type "local" → Transport "stdio"
//   - Type "remote" → Transport "sse"
//   - Environment → Env
func (t *MCPTranslator) ToCanonical(platformData []byte) (*mcp.Config, error) {
	// First try to parse as MCPConfig (with mcp wrapper)
	var openConfig MCPConfig
	if err := json.Unmarshal(platformData, &openConfig); err != nil {
		return nil, fmt.Errorf("parsing OpenCode MCP config: %w", err)
	}

	// If mcp is nil, try parsing as a bare servers map
	if openConfig.MCP == nil {
		var servers map[string]*MCPServer
		if err := json.Unmarshal(platformData, &servers); err != nil {
			return nil, fmt.Errorf("parsing OpenCode MCP servers map: %w", err)
		}
		openConfig.MCP = servers
	}

	// Ensure the map is initialized
	if openConfig.MCP == nil {
		openConfig.MCP = make(map[string]*MCPServer)
	}

	// Convert to canonical format
	config := mcp.NewConfig()
	for name, openServer := range openConfig.MCP {
		// Convert Enabled (positive) to Disabled (negative)
		// If Enabled is nil or true, Disabled is false
		// If Enabled is explicitly false, Disabled is true
		disabled := false
		if openServer.Enabled != nil && !*openServer.Enabled {
			disabled = true
		}

		server := &mcp.Server{
			Name:     name,
			URL:      openServer.URL,
			Headers:  openServer.Headers,
			Disabled: disabled,
		}

		// Split Command []string into Command string + Args []string
		if len(openServer.Command) > 0 {
			server.Command = openServer.Command[0]
			if len(openServer.Command) > 1 {
				server.Args = openServer.Command[1:]
			}
		}

		// Map Type to Transport
		switch openServer.Type {
		case TypeLocal:
			server.Transport = mcp.TransportStdio
		case TypeRemote:
			server.Transport = mcp.TransportSSE
		default:
			// Infer transport from context
			if openServer.URL != "" {
				server.Transport = mcp.TransportSSE
			} else if len(openServer.Command) > 0 {
				server.Transport = mcp.TransportStdio
			}
		}

		// Map Environment to Env
		server.Env = openServer.Environment

		config.Servers[name] = server
	}

	return config, nil
}

// FromCanonical converts canonical MCP configuration to OpenCode format.
//
// Output format:
//
//	{"mcp": {"name": {...}, ...}}
//
// Field mappings:
//   - Command (string) + Args ([]string) → Command ([]string)
//   - Transport "stdio" → Type "local"
//   - Transport "sse" → Type "remote"
//   - Env → Environment
//
// NOTE: The Platforms field from canonical format is NOT preserved.
// OpenCode does not support platform restrictions, so this data is lost
// when converting to OpenCode format.
func (t *MCPTranslator) FromCanonical(cfg *mcp.Config) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	// Convert canonical servers to OpenCode format
	openConfig := &MCPConfig{
		MCP: make(map[string]*MCPServer, len(cfg.Servers)),
	}

	for name, server := range cfg.Servers {
		// Convert Disabled (negative) to Enabled (positive)
		// Set Enabled only if explicitly disabled, otherwise omit it
		var enabled *bool
		if server.Disabled {
			f := false
			enabled = &f
		}

		openServer := &MCPServer{
			Name:    name,
			URL:     server.URL,
			Headers: server.Headers,
			Enabled: enabled,
		}

		// Join Command + Args into Command []string
		if server.Command != "" {
			openServer.Command = make([]string, 0, 1+len(server.Args))
			openServer.Command = append(openServer.Command, server.Command)
			openServer.Command = append(openServer.Command, server.Args...)
		}

		// Map Transport to Type
		switch server.Transport {
		case mcp.TransportStdio:
			openServer.Type = TypeLocal
		case mcp.TransportSSE:
			openServer.Type = TypeRemote
		default:
			// Infer type from context
			if server.URL != "" {
				openServer.Type = TypeRemote
			} else if server.Command != "" {
				openServer.Type = TypeLocal
			}
		}

		// Map Env to Environment
		openServer.Environment = server.Env

		// NOTE: server.Platforms is intentionally NOT mapped.
		// OpenCode does not support the platforms field, so it is lost.

		openConfig.MCP[name] = openServer
	}

	// Marshal with indentation
	data, err := json.MarshalIndent(openConfig, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling OpenCode MCP config: %w", err)
	}

	return data, nil
}

// Platform returns the platform identifier for this translator.
func (t *MCPTranslator) Platform() string {
	return "opencode"
}
