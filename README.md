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

# Show available commands
aix --help
```

### MCP Server Management

Manage Model Context Protocol (MCP) servers across platforms. MCP servers extend AI coding assistants with additional tools and capabilities.

#### Add an MCP Server

```bash
# Add a local stdio server
aix mcp add github npx -y @modelcontextprotocol/server-github

# Add with environment variables
aix mcp add github npx -y @modelcontextprotocol/server-github \
  --env GITHUB_TOKEN=ghp_xxxx

# Add a remote SSE server
aix mcp add api-gateway --url=https://api.example.com/mcp \
  --headers "Authorization=Bearer token123"

# Add with platform restrictions (for Claude Code; lossy for OpenCode)
aix mcp add macos-tools /usr/local/bin/macos-mcp --platform darwin

# Force overwrite existing server
aix mcp add github npx -y @new-package --force
```

#### List MCP Servers

```bash
# List all configured servers
aix mcp list

# Filter by platform
aix mcp list --platform=opencode

# Output as JSON
aix mcp list --json

# Show secret values (masked by default)
aix mcp list --show-secrets
```

#### Show Server Details

```bash
# Show server configuration across platforms
aix mcp show github

# Output as JSON
aix mcp show github --json

# Reveal masked secrets
aix mcp show github --show-secrets
```

#### Enable/Disable Servers

```bash
# Disable a server without removing it
aix mcp disable github

# Re-enable a disabled server
aix mcp enable github

# Target specific platform
aix mcp disable github --platform=claude
```

#### Remove an MCP Server

```bash
# Remove with confirmation prompt
aix mcp remove github

# Skip confirmation
aix mcp remove github --force

# Remove from specific platform only
aix mcp remove github --platform=opencode
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `aix mcp add` | Add an MCP server configuration |
| `aix mcp list` | List configured MCP servers |
| `aix mcp show` | Show server details across platforms |
| `aix mcp enable` | Enable a disabled server |
| `aix mcp disable` | Disable a server without removing |
| `aix mcp remove` | Remove an MCP server |

## Architecture

See [docs/adr/001-unified-agent-cli.md](docs/adr/001-unified-agent-cli.md) for the full architecture decision record.

## License

MIT - see [LICENSE](LICENSE) for details.
