package gemini

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/pelletier/go-toml/v2"

	"github.com/thoreinstein/aix/internal/errors"
	"github.com/thoreinstein/aix/pkg/fileutil"
)

// Sentinel errors for MCP operations.
var (
	ErrMCPServerNotFound = errors.New("MCP server not found")
	ErrInvalidMCPServer  = errors.New("invalid MCP server: name required")
)

// MCPManager provides CRUD operations for Gemini CLI MCP server configurations.
type MCPManager struct {
	paths *GeminiPaths
}

// NewMCPManager creates a new MCPManager instance.
func NewMCPManager(paths *GeminiPaths) *MCPManager {
	return &MCPManager{
		paths: paths,
	}
}

// List returns all configured MCP servers.
func (m *MCPManager) List() ([]*MCPServer, error) {
	settings, err := m.loadSettings()
	if err != nil {
		return nil, err
	}

	if settings.MCP == nil || settings.MCP.Servers == nil {
		return []*MCPServer{}, nil
	}

	servers := make([]*MCPServer, 0, len(settings.MCP.Servers))
	for _, server := range settings.MCP.Servers {
		servers = append(servers, server)
	}

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Name < servers[j].Name
	})

	return servers, nil
}

// Get returns a single MCP server by name.
func (m *MCPManager) Get(name string) (*MCPServer, error) {
	settings, err := m.loadSettings()
	if err != nil {
		return nil, err
	}

	if settings.MCP == nil || settings.MCP.Servers == nil {
		return nil, ErrMCPServerNotFound
	}

	server, ok := settings.MCP.Servers[name]
	if !ok {
		return nil, ErrMCPServerNotFound
	}

	return server, nil
}

// Add adds or updates an MCP server configuration.
func (m *MCPManager) Add(server *MCPServer) error {
	if server == nil || server.Name == "" {
		return ErrInvalidMCPServer
	}

	settings, err := m.loadSettings()
	if err != nil {
		return err
	}

	if settings.MCP == nil {
		settings.MCP = &MCPConfig{
			Servers: make(map[string]*MCPServer),
		}
	}
	if settings.MCP.Servers == nil {
		settings.MCP.Servers = make(map[string]*MCPServer)
	}

	settings.MCP.Servers[server.Name] = server

	return m.saveSettings(settings)
}

// Remove removes an MCP server configuration.
func (m *MCPManager) Remove(name string) error {
	settings, err := m.loadSettings()
	if err != nil {
		return err
	}

	if settings.MCP != nil && settings.MCP.Servers != nil {
		delete(settings.MCP.Servers, name)
	}

	return m.saveSettings(settings)
}

// Enable activates an MCP server.
func (m *MCPManager) Enable(name string) error {
	return m.setEnabled(name, true)
}

// Disable deactivates an MCP server.
func (m *MCPManager) Disable(name string) error {
	return m.setEnabled(name, false)
}

func (m *MCPManager) setEnabled(name string, enabled bool) error {
	settings, err := m.loadSettings()
	if err != nil {
		return err
	}

	if settings.MCP == nil || settings.MCP.Servers == nil {
		return ErrMCPServerNotFound
	}

	server, ok := settings.MCP.Servers[name]
	if !ok {
		return ErrMCPServerNotFound
	}

	server.Enabled = enabled
	return m.saveSettings(settings)
}

func (m *MCPManager) loadSettings() (*Settings, error) {
	configPath := m.paths.MCPConfigPath()
	if configPath == "" {
		return nil, errors.New("MCP config path not configured")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Settings{
				Other: make(map[string]any),
			}, nil
		}
		return nil, errors.Wrap(err, "reading settings file")
	}

	// 1. Unmarshal into raw map to preserve everything
	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, errors.Wrap(err, "parsing settings file")
	}

	// 2. Unmarshal into struct for typed access
	var settings Settings
	if err := toml.Unmarshal(data, &settings); err != nil {
		return nil, errors.Wrap(err, "parsing settings file into struct")
	}

	// 3. Store raw map
	settings.Other = raw

	// 4. Set server names from keys (lost during unmarshal because Name has toml:"-")
	if settings.MCP != nil && settings.MCP.Servers != nil {
		for name, server := range settings.MCP.Servers {
			if server != nil {
				server.Name = name
			}
		}
	}

	return &settings, nil
}

func (m *MCPManager) saveSettings(settings *Settings) error {
	configPath := m.paths.MCPConfigPath()
	if configPath == "" {
		return errors.New("MCP config path not configured")
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrapf(err, "creating directory %s", dir)
	}

	// Merge MCP back into Other map
	if settings.Other == nil {
		settings.Other = make(map[string]any)
	}

	// Update the mcp section in the raw map with the typed struct
	if settings.MCP != nil {
		settings.Other["mcp"] = settings.MCP
	} else {
		delete(settings.Other, "mcp")
	}

	// Update experimental section
	if settings.Experimental != nil {
		settings.Other["experimental"] = settings.Experimental
	} else {
		delete(settings.Other, "experimental")
	}

	return errors.Wrap(fileutil.AtomicWriteTOML(configPath, settings.Other), "writing settings file")
}
