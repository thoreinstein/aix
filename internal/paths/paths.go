package paths

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"

	"github.com/thoreinstein/aix/internal/errors"
)

// Platform identifiers for supported AI coding assistants.
const (
	PlatformClaude   = "claude"
	PlatformOpenCode = "opencode"
	PlatformCodex    = "codex"
	PlatformGemini   = "gemini"
)

// platformGlobalConfigs maps platform names to their global config directories.
// Paths are relative to the user's home directory.
var platformGlobalConfigs = map[string]string{
	PlatformClaude:   ".claude",
	PlatformOpenCode: ".config/opencode",
	PlatformCodex:    ".codex",
	PlatformGemini:   ".gemini",
}

// platformProjectConfigs maps platform names to their project config directories.
// Empty string means the project root itself is used.
var platformProjectConfigs = map[string]string{
	PlatformClaude:   ".claude",
	PlatformOpenCode: "", // OpenCode uses project root
	PlatformCodex:    ".codex",
	PlatformGemini:   ".gemini",
}

// platformInstructionFiles maps platform names to their instruction file names.
var platformInstructionFiles = map[string]string{
	PlatformClaude:   "CLAUDE.md",
	PlatformOpenCode: "AGENTS.md",
	PlatformCodex:    "AGENTS.md",
	PlatformGemini:   "GEMINI.md",
}

// platformMCPConfigs maps platform names to their MCP config file paths
// relative to the global config directory.
var platformMCPConfigs = map[string]string{
	PlatformClaude:   ".mcp.json",
	PlatformOpenCode: "opencode.json", // MCP config is in the main config file
	PlatformCodex:    "mcp.json",      // Assumed, may need verification
	PlatformGemini:   "settings.json", // MCP config is in the main settings file
}

// Sentinel errors for path resolution.
var (
	// ErrHomeDirNotFound indicates the user's home directory could not be determined.
	ErrHomeDirNotFound = errors.New("home directory not found")

	// ErrPermissionDenied indicates the operation was rejected due to file system permissions.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrInvalidPath indicates the provided path is malformed or invalid.
	ErrInvalidPath = errors.New("invalid path")
)

// DefaultDirPerm is the default permission for newly created directories (private).
const DefaultDirPerm = 0o700

// EnsureDir creates the directory and any necessary parents with specified permissions.
// If perm is 0, DefaultDirPerm (0700) is used.
// This function is idempotent; it returns nil if the directory already exists.
func EnsureDir(path string, perm os.FileMode) error {
	if perm == 0 {
		perm = DefaultDirPerm
	}
	return errors.Wrapf(os.MkdirAll(path, perm), "creating directory %s", path)
}

// Home returns the user's home directory.
// This is a thin wrapper around os.UserHomeDir for consistency.
// Note: It returns an empty string on error for backward compatibility.
// Use ResolveHome for proper error handling.
func Home() string {
	h, _ := ResolveHome()
	return h
}

// ResolveHome returns the user's home directory.
// Returns ErrHomeDirNotFound if the directory cannot be determined.
func ResolveHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(ErrHomeDirNotFound, err.Error())
	}
	return home, nil
}

// ConfigHome returns the XDG config home directory.
// It respects the AIX_CONFIG_DIR environment variable if set.
// On Linux: ~/.config
// On macOS: ~/.config (overrides Library/Application Support)
// On Windows: %LOCALAPPDATA%
func ConfigHome() string {
	if envDir := os.Getenv("AIX_CONFIG_DIR"); envDir != "" {
		return envDir
	}
	if xdgEnv := os.Getenv("XDG_CONFIG_HOME"); xdgEnv != "" {
		return xdgEnv
	}
	if runtime.GOOS == "darwin" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".config")
		}
	}
	return xdg.ConfigHome
}

// DataHome returns the XDG data home directory.
// On Linux: ~/.local/share
// On macOS: ~/.local/share (overrides Library/Application Support)
// On Windows: %LOCALAPPDATA%
func DataHome() string {
	if xdgEnv := os.Getenv("XDG_DATA_HOME"); xdgEnv != "" {
		return xdgEnv
	}
	if runtime.GOOS == "darwin" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share")
		}
	}
	return xdg.DataHome
}

// CacheHome returns the XDG cache home directory.
// On Linux: ~/.cache
// On macOS: ~/.cache (overrides Library/Caches)
// On Windows: %LOCALAPPDATA%\cache
func CacheHome() string {
	if xdgEnv := os.Getenv("XDG_CACHE_HOME"); xdgEnv != "" {
		return xdgEnv
	}
	if runtime.GOOS == "darwin" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".cache")
		}
	}
	return xdg.CacheHome
}

// ReposCacheDir returns the directory for cached repository clones.
// Returns: <CacheHome>/aix/repos/
func ReposCacheDir() string {
	return filepath.Join(CacheHome(), "aix", "repos")
}

// ValidPlatform returns true if the platform name is recognized.
func ValidPlatform(platform string) bool {
	_, ok := platformGlobalConfigs[platform]
	return ok
}

// Platforms returns a slice of all supported platform identifiers.
func Platforms() []string {
	return []string{
		PlatformClaude,
		PlatformOpenCode,
		PlatformCodex,
		PlatformGemini,
	}
}

// GlobalConfigDir returns the global config directory for a platform.
//
// Platform paths:
//   - claude: ~/.claude/
//   - opencode: ~/.config/opencode/
//   - codex: ~/.codex/
//   - gemini: ~/.gemini/ (or $XDG_CONFIG_HOME/gemini if set)
//
// Returns an empty string for unknown platforms.
func GlobalConfigDir(platform string) string {
	relPath, ok := platformGlobalConfigs[platform]
	if !ok {
		return ""
	}

	// Gemini CLI respects XDG_CONFIG_HOME if set.
	if platform == PlatformGemini {
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return filepath.Join(xdgConfig, "gemini")
		}
	}

	home := Home()
	if home == "" {
		return ""
	}
	return filepath.Join(home, relPath)
}

// ProjectConfigDir returns the project-scoped config directory.
//
// Platform paths:
//   - claude: <projectRoot>/.claude/
//   - opencode: <projectRoot>/ (root of project)
//   - codex: <projectRoot>/.codex/
//   - gemini: <projectRoot>/.gemini/
//
// Returns an empty string for unknown platforms or empty projectRoot.
func ProjectConfigDir(platform, projectRoot string) string {
	if projectRoot == "" {
		return ""
	}
	relPath, ok := platformProjectConfigs[platform]
	if !ok {
		return ""
	}
	// OpenCode uses project root directly
	if relPath == "" {
		return projectRoot
	}
	return filepath.Join(projectRoot, relPath)
}

// InstructionsPath returns the path to the instructions file.
//
// Platform paths:
//   - claude: <projectRoot>/CLAUDE.md
//   - opencode: <projectRoot>/AGENTS.md
//   - codex: <projectRoot>/AGENTS.md
//   - gemini: <projectRoot>/GEMINI.md
//
// Returns an empty string for unknown platforms or empty projectRoot.
func InstructionsPath(platform, projectRoot string) string {
	if projectRoot == "" {
		return ""
	}
	filename, ok := platformInstructionFiles[platform]
	if !ok {
		return ""
	}
	return filepath.Join(projectRoot, filename)
}

// SkillDir returns the skills directory for a platform.
// Always returns: <GlobalConfigDir>/skills/
//
// Returns an empty string for unknown platforms.
func SkillDir(platform string) string {
	globalDir := GlobalConfigDir(platform)
	if globalDir == "" {
		return ""
	}
	return filepath.Join(globalDir, "skills")
}

// CommandDir returns the commands directory for a platform.
// Always returns: <GlobalConfigDir>/commands/
//
// Returns an empty string for unknown platforms.
func CommandDir(platform string) string {
	globalDir := GlobalConfigDir(platform)
	if globalDir == "" {
		return ""
	}
	return filepath.Join(globalDir, "commands")
}

// MCPConfigPath returns the MCP config file path for a platform.
//
// Platform paths:
//   - claude: ~/.claude.json (main user config file, NOT in .claude directory)
//   - opencode: ~/.config/opencode/opencode.json
//   - codex: ~/.codex/mcp.json
//   - gemini: ~/.gemini/settings.toml
//
// Returns an empty string for unknown platforms.
func MCPConfigPath(platform string) string {
	home := Home()
	if home == "" {
		return ""
	}

	// Claude is special: MCP config is in ~/.claude.json (not in .claude directory)
	if platform == PlatformClaude {
		return filepath.Join(home, ".claude.json")
	}

	globalDir := GlobalConfigDir(platform)
	if globalDir == "" {
		return ""
	}
	filename, ok := platformMCPConfigs[platform]
	if !ok {
		return ""
	}
	return filepath.Join(globalDir, filename)
}

// InstructionFilename returns just the instruction filename for a platform,
// without any path components.
//
// Platform filenames:
//   - claude: CLAUDE.md
//   - opencode: AGENTS.md
//   - codex: AGENTS.md
//   - gemini: GEMINI.md
//
// Returns an empty string for unknown platforms.
func InstructionFilename(platform string) string {
	filename, ok := platformInstructionFiles[platform]
	if !ok {
		return ""
	}
	return filename
}
