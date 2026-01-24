---
title: "aix Documentation"
description: "Unified CLI for managing AI coding assistant configurations"
lead: "Manage skills, commands, agents, and MCP servers across Claude Code, OpenCode, Codex, and Gemini CLI--all from one tool."
date: 2024-01-01T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
seo:
  title: "aix - Unified CLI for AI Coding Assistants (Claude, Gemini, OpenCode)"
  description: "Manage AI coding assistant configurations across Claude Code, Gemini CLI, and OpenCode. Unified management for MCP servers, skills, and slash commands."
  canonical: ""
  noindex: false
---

## Why aix?

The AI coding assistant landscape is fragmented. Each platform has its own configuration format, file locations, and conventions. aix bridges this gap by providing:

- **Unified Management**: Single interface for all your AI assistants.
- **Write Once, Deploy Everywhere**: Define skills and commands in a platform-agnostic format.
- **Standardization**: Enforce Agent Skills and MCP specifications across tools.

## Core Features

### MCP Servers
Manage Model Context Protocol (MCP) servers to extend your agent's capabilities with external tools and data.
[Learn more about MCP Servers](./docs/guides/mcp/)

### Skills
Create and share reusable agent capabilities (prompts + tools) that work across any supported platform.
[Learn more about Skills](./docs/guides/skills/)

### Slash Commands
Define custom slash commands to automate common workflows and context gathering.
[Learn more about Slash Commands](./docs/guides/commands/)

## Installation

```bash
# Homebrew
brew install thoreinstein/tap/aix

# Go
go install github.com/thoreinstein/aix/cmd/aix@latest
```

## Get Started

Run aix init to detect your installed platforms and create a configuration file.

[Read the full Getting Started Guide](./docs/getting-started/)
