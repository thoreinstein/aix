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
