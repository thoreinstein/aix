---
title: "MCP Servers"
description: "Manage Model Context Protocol (MCP) servers with aix"
summary: "Learn how to install, configure, and manage MCP servers to extend your AI assistant's capabilities."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 30
toc: true
seo:
  title: "MCP Servers Guide - aix"
  description: "Complete guide to managing Model Context Protocol servers with aix. Add, remove, and configure MCP tools for Claude, Gemini, and OpenCode."
---

The **Model Context Protocol (MCP)** allows you to connect your AI assistant to external data and tools. `aix` provides a unified interface to manage these servers across all your configured platforms.

## Overview

With `aix`, you can:
- **Add** servers once and deploy them to all supported platforms.
- **Manage** environment variables securely.
- **Enable/Disable** servers on the fly without editing config files manually.

## Adding Servers

To add an MCP server, use the `mcp add` command. You can install servers from NPM, Python packages, or local scripts.

### Syntax

```bash
aix mcp add <name> <command> [args...] [flags]
```

### Examples

**Add the GitHub MCP Server (NPM):**
```bash
aix mcp add github npx -y @modelcontextprotocol/server-github \
  --env GITHUB_TOKEN=ghp_your_token_here
```

**Add a Local Python Server:**
```bash
aix mcp add my-tool python /path/to/server.py
```

**Add with Multiple Environment Variables:**
```bash
aix mcp add postgres npx -y @modelcontextprotocol/server-postgres \
  --env POSTGRES_URL=postgresql://localhost/db \
  --env POSTGRES_USER=admin
```

## Listing Servers

View all configured servers and their status across platforms.

```bash
aix mcp list
```

**Output:**
```text
NAME      COMMAND                     STATUS    PLATFORMS
github    npx @model.../server-github Enabled   [claude, opencode]
postgres  npx @model.../server-postgres Disabled  [claude]
```

## Managing Servers

### Enabling & Disabling
You can temporarily disable a server without removing it.

```bash
# Disable a server
aix mcp disable github

# Enable a server
aix mcp enable github
```

### Removing Servers
To permanently remove a server configuration:

```bash
aix mcp remove github
```

## Configuration Details

`aix` automatically translates your MCP configuration into the native format for each platform:

- **Claude Code**: Updates `~/.claude/config.json` (or project-specific configs).
- **OpenCode**: Updates `mcpServers` in `opencode.json`.
- **Gemini CLI**: Updates `~/.gemini/settings.toml`.

### Environment Variables
Environment variables provided via `--env` are stored securely in the platform's configuration. `aix` ensures that sensitive tokens are passed correctly to the underlying MCP server process.

## Troubleshooting

If a server fails to start, use `mcp show` to inspect its details and logs (if available).

```bash
aix mcp show github
```

```
