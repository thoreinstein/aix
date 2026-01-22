package claude

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/thoreinstein/aix/internal/mcp"
)

// MCPTranslator converts between canonical and Claude Code MCP formats.
//
// Claude Code uses a "mcpServers" key with a structure that closely matches
// the canonical format, making translation mostly 1:1 field mapping.
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
// The translation is nearly 1:1 as Claude Code uses the same field names
// and types as the canonical format.
func (t *MCPTranslator) ToCanonical(platformData []byte) (*mcp.Config, error) {
	// First try to parse as MCPConfig (with mcpServers wrapper)
	var claudeConfig MCPConfig
	if err := json.Unmarshal(platformData, &claudeConfig); err != nil {
		return nil, fmt.Errorf("parsing Claude Code MCP config: %w", err)
	}

	// If mcpServers is nil, try parsing as a bare servers map
	if claudeConfig.MCPServers == nil {
		var servers map[string]*MCPServer
		if err := json.Unmarshal(platformData, &servers); err != nil {
			return nil, fmt.Errorf("parsing Claude Code MCP servers map: %w", err)
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
		server := &mcp.Server{
			Name:      name,
			Command:   claudeServer.Command,
			Args:      claudeServer.Args,
			URL:       claudeServer.URL,
			Transport: claudeServer.Transport,
			Env:       claudeServer.Env,
			Headers:   claudeServer.Headers,
			Platforms: claudeServer.Platforms,
			Disabled:  claudeServer.Disabled,
		}
		config.Servers[name] = server
	}

	return config, nil
}

// FromCanonical converts canonical MCP configuration to Claude Code format.
//
// Output format:
//
//	{"mcpServers": {"name": {...}, ...}}
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
		claudeServer := &MCPServer{
			Name:      name,
			Command:   server.Command,
			Args:      server.Args,
			URL:       server.URL,
			Transport: server.Transport,
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
		return nil, fmt.Errorf("marshaling Claude Code MCP config: %w", err)
	}

	return data, nil
}

// Platform returns the platform identifier for this translator.
func (t *MCPTranslator) Platform() string {
	return "claude"
}
