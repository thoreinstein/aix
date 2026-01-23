---
title: "Platform Configuration"
description: "How aix integrates with Claude, OpenCode, and Gemini"
summary: "Understand where aix stores configurations and how it interacts with your existing AI tools."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 60
toc: true
seo:
  title: "Platform Configuration Guide - aix"
  description: "Learn how aix manages configuration files for Claude Code, OpenCode, and Gemini CLI."
---

`aix` works by managing the **native configuration files** of your installed AI assistants. It does not replace them; rather, it acts as a unified control plane that writes to the locations these tools expect.

This guide explains where `aix` looks for configurations and how it modifies them.

## Claude Code

Anthropic's **Claude Code** stores its configuration in your home directory.

### File Locations

| Scope | Path | Managed By `aix` |
| :--- | :--- | :--- |
| **Global Config** | `~/.claude.json` | ✅ MCP Servers |
| **Instructions** | `~/.claude/CLAUDE.md` | ✅ Skills, Agents (User) |
| **Project Config** | `<project-root>/CLAUDE.md` | ✅ Skills (Project) |
| **History** | `~/.claude/history/` | ❌ (Ignored) |

### Integration Details

*   **MCP Servers**: `aix mcp add` updates the `mcpServers` object in `~/.claude.json`.
*   **Skills & Agents**: `aix` compiles your installed skills into the `CLAUDE.md` system prompt file.
*   **Slash Commands**: Custom commands are generated as executable scripts or instruction blocks that Claude can reference.

---

## OpenCode

**OpenCode** uses the XDG configuration standard on Linux/macOS.

### File Locations

| Scope | Path | Managed By `aix` |
| :--- | :--- | :--- |
| **Global Config** | `~/.config/opencode/opencode.json` | ✅ MCP Servers, Settings |
| **Instructions** | `~/.config/opencode/AGENTS.md` | ✅ Skills, Agents |
| **Agents Dir** | `~/.config/opencode/agent/` | ✅ Agent Definitions |

### Integration Details

*   **MCP Servers**: `aix` manages the `mcpServers` section in `opencode.json`.
*   **Agents**: `aix agent install` writes new agent definitions to the `agent/` subdirectory.
*   **Translation**: `aix` automatically converts generic markdown prompts into OpenCode's specific agent format.

---

## Gemini CLI (Planned)

> **Note:** Gemini CLI support is **planned** and is **not yet available** in the current version of `aix`.

Google's **Gemini CLI** integration in `aix` is planned to be file-based and to use configuration files in `~/.gemini/`.

### Planned File Locations

| Scope | Path | Managed By `aix` (planned) |
| :--- | :--- | :--- |
| **Global Config** | `~/.gemini/config.yaml` | ✅ Settings (planned) |
| **MCP Config** | `~/.gemini/mcp.yaml` | ✅ MCP Servers (planned) |
| **Instructions** | `~/.gemini/GEMINI.md` | ✅ System Prompts (planned) |

### Planned Integration Details

*   **Separate MCP File**: Unlike Claude, Gemini CLI is planned to use a dedicated `mcp.yaml` which `aix` would manage directly.
*   **Extensions**: Future versions of `aix` may install skills as Gemini "Extensions" if supported by the installed Gemini CLI version.

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
