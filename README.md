# aix

[![CI](https://github.com/thoreinstein/aix/actions/workflows/ci.yml/badge.svg)](https://github.com/thoreinstein/aix/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/thoreinstein/aix/branch/main/graph/badge.svg)](https://codecov.io/gh/thoreinstein/aix)

A unified CLI for managing AI coding assistant configurations across platforms.

## Overview

`aix` provides a single tool to manage skills, MCP servers, and slash commands for multiple AI coding assistants:

- **Claude Code**
- **OpenCode**
- **Codex CLI**
- **Gemini CLI**

Write once, deploy everywhere. Define your configurations in a platform-agnostic format and let `aix` handle the translation to each platform's native format.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install thoreinstein/tap/aix
```

### Standalone Script (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/thoreinstein/aix/main/install.sh | sh
```

### From Source

```bash
go install github.com/thoreinstein/aix/cmd/aix@latest
```

## Getting Started

1.  **Initialize Configuration**

    Bootstrap `aix` by detecting installed AI platforms and creating a configuration file.

    ```bash
    aix init
    ```

    This creates `~/.config/aix/config.yaml` with your detected platforms.

2.  **Verify Configuration**

    ```bash
    aix config list
    ```

## Usage

### Global Options

Target specific platforms using the `--platform` (or `-p`) flag. If omitted, `aix` targets all configured/detected platforms.

```bash
# Target only Claude
aix mcp list --platform claude

# Target OpenCode and Gemini
aix mcp list -p opencode -p gemini
```

### MCP Server Management

Manage Model Context Protocol (MCP) servers.

```bash
# Add a server
aix mcp add github npx -y @modelcontextprotocol/server-github --env GITHUB_TOKEN=ghp_...

# List configured servers
aix mcp list

# Show server details and status
aix mcp show github

# Enable/Disable a server
aix mcp disable github
aix mcp enable github

# Remove a server
aix mcp remove github
```

### Skill Management

Manage reusable skills (prompts/tools) across platforms.

```bash
# Initialize a new skill from a template
aix skill init my-skill

# Install a skill to all platforms
aix skill install ./my-skill.md

# List installed skills
aix skill list

# Show skill details
aix skill show my-skill

# Remove a skill
aix skill remove my-skill
```

### Slash Command Management

Manage custom slash commands.

```bash
# Initialize a new command
aix command init /deploy

# Install a command
aix command install ./deploy.md

# List commands
aix command list

# Remove a command
aix command remove /deploy
```

### Repository Management

Manage remote repositories containing shareable skills, commands, agents, and MCP configurations. See [Repository Documentation](docs/repositories.md) for complete details.

```bash
# Add a community repository
aix repo add https://github.com/example/aix-skills

# Search for resources across all repos
aix search "code review"

# Install a skill from a repo
aix skill install community-repo/code-reviewer
```

Additional repository commands:

```bash
# List configured repositories
aix repo list

# Update repositories to get latest changes
aix repo update

# Remove a repository
aix repo remove agents
```

### Agent Management

Manage AI agent configurations (primarily for Claude Code and OpenCode).

```bash
# List available agents
aix agent list

# Show agent details
aix agent show my-agent
```

### Configuration

Manage `aix`'s own configuration.

```bash
# View all settings
aix config list

# Get a specific value
aix config get default_platforms

# Set a value
aix config set default_platforms claude,opencode,gemini

# Edit config file in your default editor
aix config edit
```

## Architecture

See [docs/adr/001-unified-agent-cli.md](docs/adr/001-unified-agent-cli.md) for the full architecture decision record.

## License

MIT - see [LICENSE](LICENSE) for details.
