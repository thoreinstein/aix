# aix

A unified CLI for managing AI coding assistant configurations across platforms.

## Overview

`aix` provides a single tool to manage skills, MCP servers, and slash commands for multiple AI coding assistants:

- Claude Code
- OpenCode
- Codex CLI
- Gemini CLI

Write once, deploy everywhere. Define your configurations in a platform-agnostic format and let `aix` handle the translation to each platform's native format.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install thoreinstein/tap/aix
```

### From Source

```bash
go install github.com/thoreinstein/aix/cmd/aix@latest
```

## Usage

```bash
# Show version
aix version

# More commands coming soon...
```

## Architecture

See [docs/adr/001-unified-agent-cli.md](docs/adr/001-unified-agent-cli.md) for the full architecture decision record.

## License

MIT - see [LICENSE](LICENSE) for details.
