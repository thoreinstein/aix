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
