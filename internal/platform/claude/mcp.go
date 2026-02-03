package claude

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
	paths *ClaudePaths
}

// NewMCPManager creates a new MCPManager instance.
func NewMCPManager(paths *ClaudePaths) *MCPManager {
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

	servers := make([]*MCPServer, 0, len(config.MCPServers))
	for _, server := range config.MCPServers {
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

	server, ok := config.MCPServers[name]
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

	config.MCPServers[server.Name] = server

	return m.saveConfig(config)
}

// Remove removes an MCP server from the configuration by name.
// This operation is idempotent - removing a non-existent server does not error.
func (m *MCPManager) Remove(name string) error {
	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	delete(config.MCPServers, name)

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

	server, ok := config.MCPServers[name]
	if !ok {
		return ErrMCPServerNotFound
	}

	server.Disabled = disabled

	return m.saveConfig(config)
}

// loadConfig reads the MCP configuration from disk.
// Handles nesting for ScopeLocal under the project path key.
func (m *MCPManager) loadConfig() (*MCPConfig, error) {
	configPath := m.paths.MCPConfigPath()
	if configPath == "" {
		return nil, errors.New("MCP config path not configured")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &MCPConfig{
				MCPServers: make(map[string]*MCPServer),
			}, nil
		}
		return nil, errors.Wrap(err, "reading MCP config")
	}

	// For ScopeLocal, we need to extract the section for the current project
	if m.paths.scope == ScopeLocal {
		var fullConfig map[string]json.RawMessage
		if err := json.Unmarshal(data, &fullConfig); err != nil {
			return nil, errors.Wrap(err, "parsing full config")
		}

		projectPath, err := m.projectPath()
		if err != nil {
			return nil, err
		}

		projectData, ok := fullConfig[projectPath]
		if !ok {
			return &MCPConfig{MCPServers: make(map[string]*MCPServer)}, nil
		}

		var config MCPConfig
		if err := json.Unmarshal(projectData, &config); err != nil {
			return nil, errors.Wrap(err, "parsing project config section")
		}

		// Ensure map is initialized and names populated
		if config.MCPServers == nil {
			config.MCPServers = make(map[string]*MCPServer)
		}
		for name, server := range config.MCPServers {
			server.Name = name
		}
		return &config, nil
	}

	// Standard handling for ScopeUser and ScopeProject
	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "parsing MCP config")
	}

	if config.MCPServers == nil {
		config.MCPServers = make(map[string]*MCPServer)
	}

	for name, server := range config.MCPServers {
		server.Name = name
	}

	return &config, nil
}

// saveConfig writes the MCP configuration to disk atomically.
// Handles nesting for ScopeLocal under the project path key.
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

	if m.paths.scope == ScopeLocal {
		// Load the full file first to preserve other sections
		data, err := os.ReadFile(configPath)
		var fullConfig map[string]json.RawMessage
		if err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrap(err, "reading existing config for update")
			}
			fullConfig = make(map[string]json.RawMessage)
		} else {
			if err := json.Unmarshal(data, &fullConfig); err != nil {
				return errors.Wrap(err, "parsing existing config for update")
			}
		}

		projectPath, err := m.projectPath()
		if err != nil {
			return err
		}

		projectData, err := json.Marshal(config)
		if err != nil {
			return errors.Wrap(err, "marshaling project config")
		}

		fullConfig[projectPath] = projectData
		return errors.Wrap(fileutil.AtomicWriteJSON(configPath, fullConfig), "writing nested MCP config")
	}

	return errors.Wrap(fileutil.AtomicWriteJSON(configPath, config), "writing MCP config")
}

// projectPath returns the absolute path of the project for ScopeLocal.
func (m *MCPManager) projectPath() (string, error) {
	if m.paths.projectRoot != "" {
		abs, err := filepath.Abs(m.paths.projectRoot)
		if err != nil {
			return "", errors.Wrapf(err, "getting absolute path for project root %q", m.paths.projectRoot)
		}
		return abs, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "getting current working directory for project path")
	}
	return cwd, nil
}
