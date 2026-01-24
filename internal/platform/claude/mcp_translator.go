package claude

import (
	"encoding/json"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/mcp"
)

// Claude Code transport type constants.
const (
	// ClaudeTypeStdio is the Claude Code type for local process servers.
	ClaudeTypeStdio = "stdio"
	// ClaudeTypeHTTP is the Claude Code type for remote HTTP servers.
	// Note: canonical format uses "sse" for this transport type.
	ClaudeTypeHTTP = "http"
)

// MCPTranslator converts between canonical and Claude Code MCP formats.
//
// Claude Code uses a "mcpServers" key with these differences from canonical:
//   - Field "type" instead of "transport"
//   - Value "http" instead of "sse" for remote servers
//   - Name is stored as map key only, not inside server object
type MCPTranslator struct{}

// NewMCPTranslator creates a new Claude Code MCP translator.
func NewMCPTranslator() *MCPTranslator {
	return &MCPTranslator{}
}

// ToCanonical converts Claude Code MCP configuration to canonical format.
//
// Input format:
//
//	{"mcpServers": {"name": {...}, ...}}
//
// or just the servers map:
//
//	{"name": {...}, ...}
//
// Mapping:
//   - Claude "type" → canonical "transport"
//   - Claude "http" → canonical "sse"
//   - Claude "stdio" → canonical "stdio"
func (t *MCPTranslator) ToCanonical(platformData []byte) (*mcp.Config, error) {
	// First try to parse as MCPConfig (with mcpServers wrapper)
	var claudeConfig MCPConfig
	if err := json.Unmarshal(platformData, &claudeConfig); err != nil {
		return nil, errors.Wrap(err, "parsing Claude Code MCP config")
	}

	// If mcpServers is nil, try parsing as a bare servers map
	if claudeConfig.MCPServers == nil {
		var servers map[string]*MCPServer
		if err := json.Unmarshal(platformData, &servers); err != nil {
			return nil, errors.Wrap(err, "parsing Claude Code MCP servers map")
		}
		claudeConfig.MCPServers = servers
	}

	// Ensure the map is initialized
	if claudeConfig.MCPServers == nil {
		claudeConfig.MCPServers = make(map[string]*MCPServer)
	}

	// Convert to canonical format
	config := mcp.NewConfig()
	for name, claudeServer := range claudeConfig.MCPServers {
		// Map Claude Type to canonical Transport
		transport := claudeTypeToCanonicalTransport(claudeServer.Type, claudeServer.URL, claudeServer.Command)

		server := &mcp.Server{
			Name:      name,
			Command:   claudeServer.Command,
			Args:      claudeServer.Args,
			URL:       claudeServer.URL,
			Transport: transport,
			Env:       claudeServer.Env,
			Headers:   claudeServer.Headers,
			Platforms: claudeServer.Platforms,
			Disabled:  claudeServer.Disabled,
		}
		config.Servers[name] = server
	}

	return config, nil
}

// claudeTypeToCanonicalTransport converts Claude Code's "type" to canonical "transport".
// Claude uses "http" for remote servers, canonical uses "sse".
func claudeTypeToCanonicalTransport(claudeType, url, command string) string {
	switch claudeType {
	case ClaudeTypeStdio:
		return mcp.TransportStdio
	case ClaudeTypeHTTP:
		return mcp.TransportSSE
	default:
		// Infer from URL/Command if type not specified
		if url != "" {
			return mcp.TransportSSE
		}
		if command != "" {
			return mcp.TransportStdio
		}
		return ""
	}
}

// FromCanonical converts canonical MCP configuration to Claude Code format.
//
// Output format:
//
//	{"mcpServers": {"name": {...}, ...}}
//
// Mapping:
//   - canonical "transport" → Claude "type"
//   - canonical "sse" → Claude "http"
//   - canonical "stdio" → Claude "stdio"
//
// The output is formatted with 2-space indentation for readability.
func (t *MCPTranslator) FromCanonical(cfg *mcp.Config) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	// Convert canonical servers to Claude format
	claudeConfig := &MCPConfig{
		MCPServers: make(map[string]*MCPServer, len(cfg.Servers)),
	}

	for name, server := range cfg.Servers {
		// Map canonical Transport to Claude Type
		claudeType := canonicalTransportToClaudeType(server.Transport, server.URL, server.Command)

		claudeServer := &MCPServer{
			Name:      name,
			Type:      claudeType,
			Command:   server.Command,
			Args:      server.Args,
			URL:       server.URL,
			Env:       server.Env,
			Headers:   server.Headers,
			Platforms: server.Platforms,
			Disabled:  server.Disabled,
		}
		claudeConfig.MCPServers[name] = claudeServer
	}

	// Marshal with indentation
	data, err := json.MarshalIndent(claudeConfig, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "marshaling Claude Code MCP config")
	}

	return data, nil
}

// canonicalTransportToClaudeType converts canonical "transport" to Claude Code's "type".
// Canonical uses "sse" for remote servers, Claude uses "http".
func canonicalTransportToClaudeType(transport, url, command string) string {
	switch transport {
	case mcp.TransportStdio:
		return ClaudeTypeStdio
	case mcp.TransportSSE:
		return ClaudeTypeHTTP
	default:
		// Infer from URL/Command if transport not specified
		if url != "" {
			return ClaudeTypeHTTP
		}
		if command != "" {
			return ClaudeTypeStdio
		}
		return ""
	}
}

// Platform returns the platform identifier for this translator.
func (t *MCPTranslator) Platform() string {
	return "claude"
}
