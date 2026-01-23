---
title: "Documentation"
description: "Learn how to use aix to manage AI coding assistant configurations"
summary: "Comprehensive documentation for the aix CLI"
date: 2024-01-01T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 1
toc: true
seo:
  title: "aix Documentation"
  description: "Learn how to use aix to manage AI coding assistant configurations across platforms"
  canonical: ""
  noindex: false
---

**aix** is a unified CLI tool designed to streamline the management of AI coding assistant configurations across multiple platforms. Whether you use **Claude Code**, **OpenCode**, **Codex**, or **Gemini CLI**, `aix` allows you to define your tools, skills, and commands once and deploy them everywhere.

## Why aix?

The AI coding assistant landscape is fragmented. Each platform has its own configuration format, file locations, and conventions. `aix` bridges this gap by providing:

- **Unified Management**: Single interface for all your AI assistants.
- **Write Once, Deploy Everywhere**: Define skills and commands in a platform-agnostic format.
- **Standardization**: Enforce Agent Skills and MCP specifications across tools.

## Getting Started

Ready to unify your AI workflow?

1.  **Install aix**:
    ```bash
    # Homebrew
    brew install thoreinstein/tap/aix

    # Go
    go install github.com/thoreinstein/aix/cmd/aix@latest
    ```

2.  **Initialize**:
    Run `aix init` to detect your installed platforms and create a configuration file.

👉 **[Read the full Getting Started Guide](getting-started/)**

## Core Concepts

Explore the core features of `aix` through our detailed guides:

### 🔌 [MCP Servers](guides/mcp/)
Manage Model Context Protocol (MCP) servers to extend your agent's capabilities with external tools and data.
*   [Add & Configure Servers](guides/mcp/#adding-servers)
*   [Enable/Disable Servers](guides/mcp/#enabling--disabling)

### 🧠 [Skills](guides/skills/)
Create and share reusable agent capabilities (prompts + tools) that work across any supported platform.
*   [Creating Skills](guides/skills/#creating-a-skill)
*   [Installing Skills](guides/skills/#installing-skills)

### ⚡ [Slash Commands](guides/commands/)
Define custom slash commands to automate common workflows and context gathering.
*   [Command Syntax](guides/commands/#syntax)
*   [Distribution](guides/commands/#sharing)

### 📦 [Repositories](guides/repositories/)
Discover and share AI resources using remote git repositories.
*   [Adding Repositories](guides/repositories/#adding-a-repository)
*   [Managing Sources](guides/repositories/#listing-repositories)

## Reference

Need detailed command syntax? Check the **[API Reference](reference/)** for comprehensive documentation on every `aix` command.

*   [`aix mcp`](reference/mcp/)
*   [`aix skill`](reference/skill/)
*   [`aix command`](reference/command/)
*   [`aix repo`](reference/repo/)

### 🛠️ [Troubleshooting](guides/troubleshooting/)
Diagnose issues, fix permissions, and recover from configuration errors using `aix doctor` and `aix backup`.

## Resources

*   [Architecture Decisions](https://github.com/thoreinstein/aix/blob/main/docs/adr/001-unified-agent-cli.md)
*   [Contributing Guidelines](https://github.com/thoreinstein/aix/blob/main/CONTRIBUTING.md)
*   [Privacy Policy](../privacy/)
