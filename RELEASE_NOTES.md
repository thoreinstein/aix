# Release v0.6.0

## Summary

This release introduces bulk installation capabilities for all resource types. Users can now install all agents, commands, skills, or MCP servers from a specific repository in a single operation, significantly streamlining the setup process for new environments.

## New Features

- **Bulk Installation**: Added the `--all-from-repo <repo-name>` flag to `aix agent install`, `aix command install`, `aix skill install`, and `aix mcp install`. This allows users to install every resource of a given type from a configured repository at once (e.g., `aix skill install --all-from-repo official`).

## Operations

- No breaking changes. Existing installation workflows remain unaffected.

---

# Release v0.5.0

## Summary

This release introduces an interactive search experience for discovering resources. Users can now launch an fzf-style fuzzy finder directly from the CLI to browse and filter skills, commands, agents, and MCP servers across all repositories.

## New Features

- **Interactive Fuzzy Search**: Running `aix search` without arguments now launches an interactive, fzf-style fuzzy finder. This interface allows users to filter results in real-time and preview resource details (description, type, repo) before selecting an item. Standard text search (`aix search <query>`) and filters (`--type`, `--repo`) remain fully supported.

## Operations

- No breaking changes. Existing search workflows remain compatible.

---

# Release v0.4.0

## Summary

This release standardizes configuration paths on macOS to follow XDG conventions (`~/.config/aix`, `~/.cache/aix`), aligning the developer experience across Linux and macOS. It also improves the resilience of repository tracking and updates the Gemini CLI adapter to match upstream capabilities.

## New Features

- **Standardized macOS Paths**: `aix` now defaults to XDG-style paths (`~/.config`, `~/.cache`, `~/.local/share`) on macOS, matching Linux behavior. Legacy paths (`~/Library/Application Support`) are still supported as a fallback for existing installations.

## Bug Fixes

- **Repository Tracking**: Fixed an issue where `aix repo add` could fail if the cache directory existed but wasn't tracked, improving robustness against manual filesystem changes.
- **Gemini CLI Adapter**: Removed support for the `{{selection}}` variable in the Gemini adapter as it is not supported by the upstream Gemini CLI tool.

## Operations

- **Configuration**: macOS users setting up `aix` for the first time will now see config files in `~/.config/aix` instead of `~/Library/Application Support/aix`.

---

# Release v0.3.0

## Summary

This release focuses on hardening the configuration loading and repository management logic, making `aix` more robust and predictable. It also includes a key bug fix for Gemini users and several internal refactorings to improve testability and code quality.

## New Features

- **Hardened Configuration Loading**: The tool's configuration handling has been made more secure and stable. `aix` now initializes its configuration earlier in the startup process and no longer searches for `config.yaml` in the current working directory, preventing unexpected behavior during development.
- **Robust Repository Management**: The repository manager is now more resilient to missing configuration files and ensures key data structures are always initialized, preventing potential panics.

## Bug Fixes

- **Gemini Platform**: A bug has been fixed where multiline instructions in skill prompts were incorrectly serialized into single-line strings in the generated TOML command files, breaking their execution.

## Internal Improvements

- **Refactoring**: Several parts of the codebase, including repository command handling and table-driven tests, have been refactored to improve testability, readability, and maintainability.
- **Documentation**: The `AGENTS.md` guide has been updated with a mandatory workflow for completing work sessions to ensure no work is left behind.

---

# v0.2.0

**Released:** 2026-01-24

## Summary

This release adds **Gemini CLI platform support** to aix, completing platform parity for skill and command installation across Claude Code, OpenCode, and Gemini CLI. It also includes a bug fix for Gemini variable translation and improvements to test coverage and documentation.

## New Features

- **Gemini CLI platform support** — The `aix command install` and `aix skill install` commands now support targeting Gemini CLI (`--platform gemini`), enabling configuration management across all three major AI coding assistants.

- **YAML/TOML translation utilities** — Internal translation layer for converting between YAML and TOML configuration formats, supporting platforms with different serialization requirements.

## Bug Fixes

- **Fix Gemini argument variable translation** — Corrected the variable substitution mapping for Gemini CLI. The canonical `$ARGUMENTS` variable now correctly translates to `{{argument}}` instead of the incorrect `{{args}}`.

## Other Changes

- **Tests**: Added comprehensive integration tests for the repo, git, backup, and editor packages; removed unused test helper functions.
- **Docs**: Added security review checklist to the contribution guide with PR self-review guidance.
- **CI**: Improved release workflow reliability by using file-based p12 certificate handling.

## Operations

No breaking changes. No migrations required. Standard upgrade path.

---

## v0.1.0

### Summary

This release adds support for Gemini CLI as a target platform, completing the initial vision of aix as a unified configuration manager across Claude Code, OpenCode, Codex, and now Gemini CLI. The release also includes documentation cleanup and CI improvements for macOS binary signing.

### New Features

- **Gemini CLI Platform Support**: Full platform adapter implementation enabling skill, command, agent, and MCP server installation for Gemini CLI. Includes variable translation between canonical (`$ARGUMENTS`) and Gemini (`{{argument}}`) syntax, MCP configuration translation, and proper config path resolution.

### Bug Fixes

- Fixed CI test failures caused by `XDG_CONFIG_HOME` environment variable interference in path resolution tests.

### Operations

- Release workflow now runs on `macos-latest` with code signing secrets configured for proper macOS binary distribution.
- Documentation cleaned of emojis and non-ASCII characters for consistent, portable output across all terminals.

### CI/CD
- 67b93c3 ci: run release on macos-latest with signing secrets

## [0.1.0] - 2024-01-24

### Added

- Initial release of `aix` - unified CLI for AI coding assistant configurations
- Core skill management (init, validate, install, list)
- Core command management (init, validate, install, list)
- Core agent management (init, validate, install, list)
- MCP server configuration support
- Multi-platform support for Claude Code and OpenCode
- Repository management for shareable resources
- Cross-platform translation layer for configuration formats
- Comprehensive validation for skills, commands, and agents

### Changed

- Nothing yet.

### Deprecated

- Nothing yet.

### Removed

- Nothing yet.

### Fixed

- Nothing yet.

### Security

- Nothing yet.

[Unreleased]: https://github.com/thoreinstein/aix/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/thoreinstein/aix/releases/tag/v0.1.0