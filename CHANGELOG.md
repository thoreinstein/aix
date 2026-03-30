# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.8.1] - 2026-03-30

- `cb8908f` fix: wire Gemini CLI into mcp add command
- `921bfff` fix: address PR review nits
- `c8759cb` deps: bump github.com/fatih/color from 1.18.0 to 1.19.0
- `7532b0a` ci: bump the actions group with 3 updates
- `7b36f38` chore(deps): bump yaml from 2.8.2 to 2.8.3 in /aix-docs
- `5a44d40` chore(deps): bump picomatch in /aix-docs
- `799c6b7` chore(deps): bump brace-expansion in /aix-docs

## [v0.8.0] - 2026-03-18

- `32f501f` ci: bump the actions group across 1 directory with 5 updates
- `c279119` deps: bump golang.org/x/term from 0.39.0 to 0.41.0
- `c5b21e2` chore(deps-dev): bump rollup from 4.56.0 to 4.59.0 in /aix-docs
- `bc65a4f` chore(deps): bump minimatch in /aix-docs
- `df1ff24` fix: format files
- `59b2c4c` fix: guard CopyDir against same-path and dst-inside-src copies
- `0c6c988` FEAT: copy all supporting files during skill installation
- `4a19e46` ci: bump anchore/sbom-action from 0.22.0 to 0.22.1 in the actions group
- `43dc97d` FEAT: implement Gemini CLI platform adapter updates
- `129bcd1` FEAT: add TOML support for Gemini CLI configuration
- `72a3039` REFACTOR: use shared install logic in resource commands
- `e2b6110` FEAT: extract shared install logic to internal/install
- `67d71a1` FIX security bugs and refactor platform adapters
- `246e8fe` Update docs

## [v0.7.0] - 2026-03-18

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
