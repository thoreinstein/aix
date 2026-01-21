package platform

// Platform defines the contract for platform adapters.
// Each supported AI coding assistant (Claude, OpenCode, Codex, Gemini)
// implements this interface to provide platform-specific functionality.
//
// Implementations must be safe for concurrent use. The methods defined here
// return static configuration data that does not change during the lifetime
// of an adapter instance.
type Platform interface {
	// Name returns the platform identifier (claude, opencode, codex, gemini).
	// The name must match one of the constants in the paths package
	// (e.g., paths.PlatformClaude).
	Name() string

	// GlobalConfigDir returns the path to the global configuration directory.
	// This is where user-level settings, skills, commands, and MCP configs reside.
	//
	// Examples:
	//   - claude: ~/.claude/
	//   - opencode: ~/.config/opencode/
	//   - codex: ~/.codex/
	//   - gemini: ~/.gemini/
	GlobalConfigDir() string

	// MCPConfigPath returns the path to the MCP configuration file.
	// This file configures Model Context Protocol servers for the platform.
	//
	// Examples:
	//   - claude: ~/.claude/mcp_servers.json
	//   - opencode: ~/.config/opencode/opencode.json
	//   - codex: ~/.codex/mcp.json
	//   - gemini: ~/.gemini/settings.toml
	MCPConfigPath() string

	// InstructionFilename returns the platform's instruction file name.
	// This is the filename (without path) used for project-level agent instructions.
	//
	// Examples:
	//   - claude: CLAUDE.md
	//   - opencode: AGENTS.md
	//   - codex: AGENTS.md
	//   - gemini: GEMINI.md
	InstructionFilename() string
}
