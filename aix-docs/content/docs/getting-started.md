---
title: "Getting Started"
description: "Get started with aix in minutes"
lead: "Install aix and configure your first AI coding assistant"
date: 2024-01-01T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 10
toc: true
seo:
  title: "Getting Started with aix"
  description: "Install aix and configure your first AI coding assistant"
  canonical: ""
  noindex: false
---

## What is aix?

`aix` is a unified CLI for managing AI coding assistant configurations across multiple platforms:

- **Claude Code** - Anthropic's Claude-powered coding assistant
- **OpenCode** - Open-source AI coding assistant
- **Codex** - OpenAI's code generation model
- **Gemini CLI** - Google's Gemini-powered CLI

With `aix`, you can manage skills, commands, agents, and MCP servers from a single tool, with automatic translation between platform-specific formats.

## Installation

### Homebrew (macOS/Linux)

Install using the Homebrew tap:

```bash
brew install thoreinstein/tap/aix
```

### Using Go

```bash
go install github.com/thoreinstein/aix/cmd/aix@latest
```

### From Source

```bash
git clone https://github.com/thoreinstein/aix.git
cd aix
go build -o aix ./cmd/aix
```

## Quick Start

### 1. Initialize a skill

Create your first skill definition:

```bash
aix skill init my-skill
```

This creates a skill directory with a `skill.md` file containing the skill definition.

### 2. Install the skill

Install the skill to your preferred platform:

```bash
# Install to Claude Code (default)
aix skill install ./my-skill

# Install to a specific platform
aix skill install ./my-skill --platform opencode
```

### 3. Verify installation

Check that the skill was installed:

```bash
aix skill list
```

## Next Steps

- Learn about [Skills](/docs/guides/skills/) - reusable instruction sets
- Explore [Commands](/docs/guides/commands/) - slash command definitions
- Configure [MCP Servers](/docs/guides/mcp/) - Model Context Protocol servers
- Set up [Agents](/docs/guides/agents/) - specialized AI assistant configurations
