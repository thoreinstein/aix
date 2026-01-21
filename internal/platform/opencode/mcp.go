package opencode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// Enable sets Disabled=false for the specified server.
// Returns ErrMCPServerNotFound if the server does not exist.
func (m *MCPManager) Enable(name string) error {
	return m.setDisabled(name, false)
}

// Disable sets Disabled=true for the specified server.
// Returns ErrMCPServerNotFound if the server does not exist.
func (m *MCPManager) Disable(name string) error {
	return m.setDisabled(name, true)
}

// setDisabled is a helper to toggle the Disabled field.
func (m *MCPManager) setDisabled(name string, disabled bool) error {
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	server, ok := config.MCP[name]
	if !ok {
		return ErrMCPServerNotFound
	}

	server.Disabled = disabled

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
		return nil, fmt.Errorf("reading MCP config: %w", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing MCP config: %w", err)
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

	return atomicWriteJSON(configPath, config)
}

// atomicWriteJSON writes JSON data to a file atomically using temp file + rename.
func atomicWriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	// Add trailing newline for POSIX compliance
	data = append(data, '\n')

	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Create temp file in same directory for atomic rename
	tmp, err := os.CreateTemp(dir, ".mcp-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	// Clean up temp file on error
	tmpName := tmp.Name()
	defer func() {
		// Only remove if rename failed (file still exists)
		if _, statErr := os.Stat(tmpName); statErr == nil {
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
