package gemini

import (
	"github.com/pelletier/go-toml/v2"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/internal/mcp"
)

// MCPTranslator converts between canonical and Gemini CLI MCP formats.
type MCPTranslator struct{}

// NewMCPTranslator creates a new Gemini CLI MCP translator.
func NewMCPTranslator() *MCPTranslator {
	return &MCPTranslator{}
}

// ToCanonical converts Gemini CLI MCP configuration to canonical format.
func (t *MCPTranslator) ToCanonical(platformData []byte) (*mcp.Config, error) {
	// Try to parse as MCPConfig (with mcp key)
	var geminiConfig MCPConfig
	if err := toml.Unmarshal(platformData, &geminiConfig); err != nil {
		// If fails, try parsing as a bare servers map
		var servers map[string]*MCPServer
		if err := toml.Unmarshal(platformData, &servers); err != nil {
			return nil, errors.Wrap(err, "parsing Gemini CLI MCP config")
		}
		geminiConfig.Servers = servers
	}

	if geminiConfig.Servers == nil {
		geminiConfig.Servers = make(map[string]*MCPServer)
	}

	config := mcp.NewConfig()
	for name, geminiServer := range geminiConfig.Servers {
		// Infer transport
		transport := mcp.TransportStdio
		if geminiServer.URL != "" {
			transport = mcp.TransportSSE
		}

		server := &mcp.Server{
			Name:      name,
			Command:   geminiServer.Command,
			Args:      geminiServer.Args,
			URL:       geminiServer.URL,
			Transport: transport,
			Env:       geminiServer.Env,
			Headers:   geminiServer.Headers,
			Disabled:  !geminiServer.Enabled, // Translate Enabled -> Disabled
		}
		config.Servers[name] = server
	}

	return config, nil
}

// FromCanonical converts canonical MCP configuration to Gemini CLI format.
func (t *MCPTranslator) FromCanonical(cfg *mcp.Config) ([]byte, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	geminiConfig := &MCPConfig{
		Servers: make(map[string]*MCPServer, len(cfg.Servers)),
	}

	for name, server := range cfg.Servers {
		geminiServer := &MCPServer{
			Command: server.Command,
			Args:    server.Args,
			URL:     server.URL,
			Env:     server.Env,
			Headers: server.Headers,
			Enabled: !server.Disabled, // Translate Disabled -> Enabled
		}
		geminiConfig.Servers[name] = geminiServer
	}

	data, err := toml.Marshal(geminiConfig)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling Gemini CLI MCP config")
	}

	return data, nil
}

// Platform returns the platform identifier for this translator.
func (t *MCPTranslator) Platform() string {
	return "gemini"
}
