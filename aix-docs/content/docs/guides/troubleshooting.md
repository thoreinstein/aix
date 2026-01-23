---
title: "Troubleshooting"
description: "Diagnose and fix common aix configuration issues"
summary: "Learn how to use aix doctor, interpret error messages, and recover from configuration issues."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 70
toc: true
seo:
  title: "Troubleshooting Guide - aix"
  description: "Fix common aix errors, permission issues, and configuration syntax problems."
---

If you encounter issues with `aix`, your first step should always be to run the built-in diagnostic tool:

```bash
aix doctor
```

This command checks for permission issues, syntax errors in configuration files, and invalid MCP server setups.

## Common `aix doctor` Issues

### 1. Path Permissions (`path-permissions`)
`aix` requires read/write access to its own configuration directory and those of your AI assistants.

*   **Error: "file is not readable"**
    *   **Cause**: `aix` cannot read a configuration file (e.g., `~/.claude/config.json`).
    *   **Fix**: Run `chmod 644 <path>` or use `aix doctor --fix` to attempt an automatic fix.
*   **Warning: "file is world-writable (security risk)"**
    *   **Cause**: Your configuration files (which may contain API keys) are accessible by other users on the system.
    *   **Fix**: Run `chmod 600 <path>`.

### 2. Configuration Syntax (`config-syntax`)
`aix` validates that your JSON and TOML configuration files are well-formed.

*   **Error: "JSON syntax error at line X"**
    *   **Cause**: A manual edit introduced a trailing comma, missing quote, or unbalanced brace.
    *   **Fix**: Open the file mentioned in the error and correct the syntax. Use a tool like `jq` to verify JSON files.

### 3. MCP Semantics (`config-semantics`)
Even if a file is valid JSON, the *values* inside might be invalid for an MCP server.

*   **Warning: "command not found in PATH"**
    *   **Cause**: An MCP server is configured to run a command (e.g., `npx`, `python`) that isn't installed or isn't in your system's PATH.
    *   **Fix**: Install the missing runtime or use an absolute path to the executable in your `aix mcp add` command.
*   **Error: "server has both command and URL configured"**
    *   **Cause**: Ambiguous transport type.
    *   **Fix**: Use either a command (for local stdio) or a URL (for remote SSE), but not both.

---

## Recovering from Bad Configs

If a configuration change breaks your AI assistant, `aix` provides an automatic safety net.

### Using Backups
Before every write operation, `aix` creates a timestamped backup in `~/.config/aix/backups/`.

1.  **List available backups**:
    ```bash
    aix backup list --platform claude
    ```
2.  **Restore the latest backup**:
    ```bash
    aix backup restore --platform claude
    ```

> **Note**: Restoring a backup will overwrite your current configuration for that platform.

---

## MCP Server Issues

If an MCP server is "Enabled" in `aix` but its tools aren't appearing in your AI assistant:

1.  **Check Logs**: Some platforms log MCP startup errors. Check `~/.claude/logs/` or the equivalent for your tool.
2.  **Test the Command**: Try running the MCP command manually in your terminal to see if it starts without errors.
3.  **Inspect with `mcp show`**:
    ```bash
    aix mcp show <server-name> --show-secrets
    ```
    Verify that the environment variables (like `GITHUB_TOKEN`) are correct.

---

## Still having trouble?

If `aix doctor` passes but you still have issues:
1.  Verify you are on the latest version: `aix version`.
2.  Check the [GitHub Issues](https://github.com/thoreinstein/aix/issues) for similar problems.
3.  Run `aix status --verbose` to see a detailed overview of all active configurations.
