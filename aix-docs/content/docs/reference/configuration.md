---
title: "Configuration File"
description: "Reference for aix configuration file (config.yaml)"
summary: "Detailed reference for config.yaml: settings, overrides, and environment variables."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 100
toc: true
seo:
  title: "Configuration Reference - aix"
  description: "Complete guide to aix config.yaml keys, values, and environment variables."
---

`aix` uses a YAML configuration file to store user preferences and platform overrides. This file is typically located at `~/.config/aix/config.yaml`.

## File Location

`aix` searches for the configuration file in the following order:

1.  **Environment Variable**: The path specified in `AIX_CONFIG_DIR`.
2.  **Current Directory**: `./config.yaml`.
3.  **XDG Config Home**: `~/.config/aix/config.yaml` (Linux/macOS) or `%APPDATA%\aix\config.yaml` (Windows).

You can initialize a new configuration file with defaults by running:

```bash
aix init
```

## Structure

The configuration file is a flat YAML dictionary with the following top-level keys.

```yaml
version: 1
default_platforms:
  - claude
  - opencode
platforms:
  claude:
    config_dir: "/custom/path/to/claude"
```

### `version`
*   **Type**: `integer`
*   **Default**: `1`
*   **Description**: The schema version of the configuration file. Currently, `1` is the only supported version.

### `default_platforms`
*   **Type**: `array` of `string`
*   **Default**: `["claude", "opencode", "codex", "gemini"]`
*   **Description**: A list of platforms to target by default when no `--platform` flag is provided to CLI commands.
*   **Valid Values**: `claude`, `opencode`, `codex`, `gemini`.

### `platforms`
*   **Type**: `dictionary` (Platform Name -> Override Object)
*   **Default**: `{}` (Empty)
*   **Description**: Platform-specific overrides. Useful if you have installed an AI assistant in a non-standard location.

#### `platforms.<name>.config_dir`
*   **Type**: `string`
*   **Description**: The absolute path to the configuration directory for the specified platform.
*   **Example**:
    ```yaml
    platforms:
      claude:
        config_dir: "/Users/me/.config/claude-dev"
    ```

## Environment Variables

You can override configuration values using environment variables prefixed with `AIX_`.

| Variable | Config Key | Description |
| :--- | :--- | :--- |
| `AIX_CONFIG_DIR` | N/A | Directory to search for `config.yaml`. |
| `AIX_VERSION` | `version` | Override schema version. |
| `AIX_DEFAULT_PLATFORMS` | `default_platforms` | Comma-separated list of default platforms. |

**Example:**
```bash
export AIX_DEFAULT_PLATFORMS="opencode"
aix skill list  # Will only list OpenCode skills by default
```
