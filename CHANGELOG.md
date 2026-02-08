# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.7.0] - 2026-02-08

- `8bcf3c5` fix: Handle symlinked directories in backup
- `052265e` chore(deps): bump @isaacs/brace-expansion in /aix-docs
- `e1e5154` Feat: Implement local scopes and collision detection for Claude
- `7c4b5b4` Build: Add stretchr/testify dependency
- `bc44d99` Feat: Add IsIgnored git helper
- `1a9b72a` feat(cli): implement scope flag and support
- `b2a9298` feat(core): update Platform interface for configuration scopes
- `0f0f549` ci: bump anchore/sbom-action from 0.22.0 to 0.22.1 in the actions group
- `435b753` FEAT: implement Gemini CLI platform adapter updates
- `258fec5` FEAT: add TOML support for Gemini CLI configuration
- `32b8a0a` REFACTOR: use shared install logic in resource commands
- `eb275a9` FEAT: extract shared install logic to internal/install
- `c697ec6` FIX security bugs and refactor platform adapters
- `0dafda3` Update docs

## [v0.6.1] - 2026-01-28

- `7065195` FIX repo path resolution and agent naming
- `eb8a2c9` FIX config path resolution priority
- `264d2ab` Update blog post
- `ea78760` ci: bump the actions group with 5 updates

## [v0.6.0] - 2026-01-27

- `a05918c` Feat: Add --all-from-repo flag to install commands

## [v0.5.0] - 2026-01-25

- `e0a82ff` feat(search): add interactive fuzzy search

## [v0.4.0] - 2026-01-25

- `ae57518` Update repository documentation with XDG paths
- `ca05bb0` Remove unsupported selection variable from Gemini adapter
- `a31b787` Standardize on XDG paths for macOS
- `c71f641` Fix repository tracking resilience and path handling

## [v0.3.0] - 2026-01-24

- `490c007` refactor(repo): Improve config loading and testability
- `ec06963` docs: Add mandatory session completion workflow
- `adebc9b` feat(config): Harden configuration and repo loading
- `503b8fe` fix(gemini): Correctly serialize multiline TOML strings
- `d991a56` refactor(testing): Use if/else chains in table tests

## [v0.2.0] - 2026-01-24

### Added
- Add Gemini CLI platform support to `command install` and `skill install` commands (7bcd53f)
- Add YAML/TOML bidirectional translation utilities (3134aa9)
- Add integration tests for repo, git, backup, and editor packages (e306b7e)
- Add security review checklist to contribution guide (8659913)

### Fixed
- Fix Gemini argument variable translation (`$ARGUMENTS` → `{{argument}}`) (312494a)

### Changed
- ci: use file-based p12 certificate for release workflow reliability (00910f7)

### Removed
- Remove unused test helper functions from repo manager tests (cd84483)

## [v0.1.0] - 2026-01-24

### Features
- 4152db6 feat(platform): implement Gemini CLI platform adapter

### Fixes
- 53c7467 fix(tests): unset XDG_CONFIG_HOME in path tests to fix CI failure

### Documentation
- 8c27fd0 docs: fix blog post frontmatter
- d04c7ef docs: remove emojis and non-ascii characters for clean release

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
