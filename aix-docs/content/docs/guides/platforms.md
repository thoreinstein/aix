---
title: "Platform Configuration"
description: "How aix integrates with Claude, OpenCode, Gemini, and Codex"
summary: "Understand where aix stores configurations and how it interacts with your existing AI tools."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 60
toc: true
seo:
  title: "Platform Configuration Guide - aix"
  description: "Learn how aix manages configuration files for Claude Code, OpenCode, Gemini CLI, and Codex."
---

`aix` works by managing the **native configuration files** of your installed AI assistants. It does not replace them; rather, it acts as a unified control plane that writes to the locations these tools expect.

This guide explains where `aix` looks for configurations and how it modifies them.

## Claude Code

Anthropic's Claude Code stores its configuration in your home directory.

### File Locations

| Scope | Path | Managed By aix |
| :--- | :--- | :--- |
| **Global Config** | `~/.claude.json` | Yes (MCP Servers) |
| **Instructions** | `~/.claude/CLAUDE.md` | Yes (Skills, Agents) |
| **Project Config** | `<project-root>/CLAUDE.md` | Yes (Skills) |
| **History** | `~/.claude/history/` | No (Ignored) |

### Integration Details

*   **MCP Servers**: aix mcp add updates the mcpServers object in `~/.claude.json`.
*   **Skills & Agents**: aix compiles your installed skills into the `CLAUDE.md` system prompt file.
*   **Slash Commands**: Custom commands are generated as executable scripts or instruction blocks that Claude can reference.

---

## OpenCode

OpenCode uses the XDG configuration standard on Linux/macOS.

### File Locations

| Scope | Path | Managed By aix |
| :--- | :--- | :--- |
| **Global Config** | `~/.config/opencode/opencode.json` | Yes (MCP Servers, Settings) |
| **Instructions** | `~/.config/opencode/AGENTS.md` | Yes (Skills, Agents) |
| **Agents Dir** | `~/.config/opencode/agent/` | Yes (Agent Definitions) |

### Integration Details

*   **MCP Servers**: aix manages the mcpServers section in `opencode.json`.
*   **Agents**: aix agent install writes new agent definitions to the `agent/` subdirectory.
*   **Translation**: aix automatically converts generic markdown prompts into OpenCode's specific agent format.

---

## Gemini CLI

Google's Gemini CLI stores its configuration in `~/.gemini/` (global) or `.gemini/` (project).

### File Locations

| Scope | Path | Managed By aix |
| :--- | :--- | :--- |
| **Global Config** | `~/.gemini/` | Yes (MCP Servers, Skills, Commands) |
| **Project Config** | `<project-root>/.gemini/` | Yes (Project-scoped resources) |
| **MCP Config** | `~/.gemini/settings.json` | Yes (MCP Servers) |
| **Instructions** | `~/.gemini/GEMINI.md` | Yes (System Prompts) |
| **Skills Dir** | `~/.gemini/skills/` | Yes (Skill Definitions) |
| **Commands Dir** | `~/.gemini/commands/` | Yes (Slash Commands) |

### Integration Details

*   **MCP Servers**: aix manages the mcpServers section in `settings.json`.
*   **Skills**: Installed skills are placed in the `skills/` directory.
*   **Translation**: aix translates skill variables to Gemini's format.

---

## Codex (Planned)

Note: Codex support is planned and is not yet available.

The **Codex** integration is planned to follow the XDG standard, similar to OpenCode but with its own distinct configuration schema.

### Planned File Locations

| Scope | Path | Managed By `aix` (planned) |
| :--- | :--- | :--- |
| Scope | Path | Managed By aix (planned) |
| :--- | :--- | :--- |
| **Global Config** | `~/.config/codex/config.toml` | Yes (MCP Servers) |
| **Instructions** | `~/.config/codex/prompts/` | Yes (System Prompts) |
| **Skills Dir** | `~/.config/codex/skills/` | Yes (Skill Definitions) |

---

## Coexistence & Safety

Can you use `aix` alongside manual editing? **Yes, but with care.**

### How `aix` Updates Files

1.  **Read**: `aix` reads the existing configuration file.
2.  **Patch**: It selectively updates only the sections it manages (e.g., adding a server to `mcpServers`).
3.  **Preserve**: It attempts to preserve comments and unrelated settings, though JSON/YAML serialization limits this.

### Best Practices

*   **Use `aix` for managed sections**: If you used `aix` to add an MCP server, use `aix` to remove it. Do not manually delete it from the JSON file, or `aix` might become out of sync.
*   **Backups**: `aix` automatically creates backups in `~/.config/aix/backups/` before making changes. If something goes wrong, you can restore the previous state:
    ```bash
    aix backup restore
    ```
*   **Manual Edits**: It is safe to manually edit settings that `aix` does not touch (e.g., theme colors, API keys not managed by `aix`).
