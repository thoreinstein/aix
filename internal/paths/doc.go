// Package paths provides cross-platform path resolution utilities for AI
// coding assistant configuration directories.
//
// This package abstracts the differences between operating systems and AI
// assistant platforms (Claude Code, OpenCode, Codex, Gemini CLI) for
// consistent path resolution across all environments.
//
// # XDG Base Directory Compliance
//
// The package wraps github.com/adrg/xdg for cross-platform XDG Base Directory
// Specification compliance. On Linux and macOS, paths follow XDG conventions
// (~/.config, ~/.local/share, ~/.cache).
//
// # Platform Constants
//
// Use the provided platform constants when calling platform-specific functions:
//
//	paths.GlobalConfigDir(paths.PlatformClaude)   // ~/.claude/
//	paths.GlobalConfigDir(paths.PlatformOpenCode) // ~/.config/opencode/
//
// # Platform Configuration Directories
//
// Each AI assistant platform uses different directory structures:
//
//	| Platform  | Global Config       | Project Config    | Instructions |
//	|-----------|---------------------|-------------------|--------------|
//	| Claude    | ~/.claude/          | .claude/          | CLAUDE.md    |
//	| OpenCode  | ~/.config/opencode/ | (project root)    | AGENTS.md    |
//	| Codex     | ~/.codex/           | .codex/           | AGENTS.md    |
//	| Gemini    | ~/.gemini/          | .gemini/          | GEMINI.md    |
//
// # Standard Directory Helpers
//
// Skills and commands follow consistent patterns relative to the global
// config directory:
//
//	paths.SkillDir(platform)   // <GlobalConfigDir>/skills/
//	paths.CommandDir(platform) // <GlobalConfigDir>/commands/
//
// # Error Handling
//
// Functions that accept a platform parameter return empty strings for
// unknown platforms. Use [ValidPlatform] to check validity before calling:
//
//	if !paths.ValidPlatform(platform) {
//	    return fmt.Errorf("%w: %s", aixerrors.ErrUnknownPlatform, platform)
//	}
package paths
