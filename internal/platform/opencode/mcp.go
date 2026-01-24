package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/pkg/fileutil"
)

// Sentinel errors for MCP operations.
var (
	ErrMCPServerNotFound = errors.New("MCP server not found")
	ErrInvalidMCPServer  = errors.New("invalid MCP server: name required")
)

// MCPManager provides CRUD operations for MCP server configurations.
type MCPManager struct {
	paths *OpenCodePaths
}

// NewMCPManager creates a new MCPManager instance.
func NewMCPManager(paths *OpenCodePaths) *MCPManager {
	return &MCPManager{
		paths: paths,
	}
}

// List returns all MCP servers from the configuration file.
// Returns an empty slice if the config file does not exist.
// The returned servers are sorted by name for deterministic ordering.
func (m *MCPManager) List() ([]*MCPServer, error) {
	config, err := m.loadConfig()
	if err != nil {
		return nil, err
	}

	servers := make([]*MCPServer, 0, len(config.MCP))
	for _, server := range config.MCP {
		servers = append(servers, server)
	}

	// Sort by name for deterministic ordering
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Name < servers[j].Name
	})

	return servers, nil
}

// Get returns a single MCP server by name.
// Returns ErrMCPServerNotFound if the server does not exist.
func (m *MCPManager) Get(name string) (*MCPServer, error) {
	config, err := m.loadConfig()
	if err != nil {
		return nil, err
	}

	server, ok := config.MCP[name]
	if !ok {
		return nil, ErrMCPServerNotFound
	}

	return server, nil
}

// Add adds or updates an MCP server in the configuration.
// Returns ErrInvalidMCPServer if the server name is empty.
func (m *MCPManager) Add(server *MCPServer) error {
	if server == nil || server.Name == "" {
		return ErrInvalidMCPServer
	}

	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	config.MCP[server.Name] = server

	return m.saveConfig(config)
}

// Remove removes an MCP server from the configuration by name.
// This operation is idempotent - removing a non-existent server does not error.
func (m *MCPManager) Remove(name string) error {
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	delete(config.MCP, name)

	return m.saveConfig(config)
}

// Enable sets Enabled=true for the specified server.
// Returns ErrMCPServerNotFound if the server does not exist.
func (m *MCPManager) Enable(name string) error {
	return m.setEnabled(name, true)
}

// Disable sets Enabled=false for the specified server.
// Returns ErrMCPServerNotFound if the server does not exist.
func (m *MCPManager) Disable(name string) error {
	return m.setEnabled(name, false)
}

// setEnabled is a helper to toggle the Enabled field.
func (m *MCPManager) setEnabled(name string, enabled bool) error {
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	server, ok := config.MCP[name]
	if !ok {
		return ErrMCPServerNotFound
	}

	server.Enabled = &enabled

	return m.saveConfig(config)
}

// loadConfig reads the MCP configuration from disk.
// Returns an empty config with initialized MCP map if the file doesn't exist.
func (m *MCPManager) loadConfig() (*MCPConfig, error) {
	configPath := m.paths.MCPConfigPath()
	if configPath == "" {
		return nil, errors.New("MCP config path not configured")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &MCPConfig{
				MCP: make(map[string]*MCPServer),
			}, nil
		}
		return nil, errors.Wrap(err, "reading MCP config")
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "parsing MCP config")
	}

	// Ensure the map is initialized
	if config.MCP == nil {
		config.MCP = make(map[string]*MCPServer)
	}

	// Populate Name field from map keys for consistency
	for name, server := range config.MCP {
		server.Name = name
	}

	return &config, nil
}

// saveConfig writes the MCP configuration to disk atomically.
func (m *MCPManager) saveConfig(config *MCPConfig) error {
	configPath := m.paths.MCPConfigPath()
	if configPath == "" {
		return errors.New("MCP config path not configured")
	}

	// Create parent directory if needed
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrapf(err, "creating directory %s", dir)
	}

	return errors.Wrap(fileutil.AtomicWriteJSON(configPath, config), "writing MCP config")
}
