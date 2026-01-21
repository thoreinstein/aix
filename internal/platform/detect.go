package platform

import (
	"os"

	"github.com/thoreinstein/aix/internal/paths"
)

// InstallStatus indicates the installation state of a platform.
type InstallStatus string

const (
	// StatusInstalled indicates the platform's global config directory exists.
	StatusInstalled InstallStatus = "installed"

	// StatusNotInstalled indicates the platform's global config directory does not exist.
	StatusNotInstalled InstallStatus = "not_installed"

	// StatusPartial indicates a partial installation state.
	// Reserved for future use (e.g., config exists but binary missing).
	StatusPartial InstallStatus = "partial"
)

// DetectionResult contains information about a detected platform.
type DetectionResult struct {
	// Name is the platform identifier (claude, opencode, codex, gemini).
	Name string

	// GlobalConfig is the path to the global configuration directory.
	// This path is always set for valid platforms, even if the directory
	// does not exist.
	GlobalConfig string

	// MCPConfig is the path to the MCP configuration file.
	// This path is always set for valid platforms, even if the file
	// does not exist.
	MCPConfig string

	// Status indicates the installation state of the platform.
	Status InstallStatus
}

// DetectPlatform checks if a specific platform is installed and returns detection info.
// Returns nil if the platform name is invalid.
func DetectPlatform(name string) *DetectionResult {
	if !paths.ValidPlatform(name) {
		return nil
	}

	globalConfig := paths.GlobalConfigDir(name)
	mcpConfig := paths.MCPConfigPath(name)

	status := StatusNotInstalled
	if dirExists(globalConfig) {
		status = StatusInstalled
	}

	return &DetectionResult{
		Name:         name,
		GlobalConfig: globalConfig,
		MCPConfig:    mcpConfig,
		Status:       status,
	}
}

// DetectAll returns detection results for all known platforms.
// Platforms are returned in deterministic order: claude, opencode, codex, gemini.
func DetectAll() []*DetectionResult {
	platforms := paths.Platforms()
	results := make([]*DetectionResult, 0, len(platforms))

	for _, name := range platforms {
		if result := DetectPlatform(name); result != nil {
			results = append(results, result)
		}
	}

	return results
}

// DetectInstalled returns only platforms that are installed (Status == StatusInstalled).
// Platforms are returned in deterministic order.
func DetectInstalled() []*DetectionResult {
	all := DetectAll()
	installed := make([]*DetectionResult, 0, len(all))

	for _, result := range all {
		if result.Status == StatusInstalled {
			installed = append(installed, result)
		}
	}

	return installed
}

// dirExists returns true if the path exists and is a directory.
func dirExists(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
