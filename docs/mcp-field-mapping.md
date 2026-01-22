# MCP Field Mapping Reference

This document describes the field mappings between the canonical MCP configuration format used by `aix` and the platform-specific formats used by Claude Code and OpenCode.

## Overview

The canonical format provides a unified representation of MCP server configurations that can be translated to and from each platform's native format. This enables `aix` to manage MCP servers consistently across different AI coding assistants.

### Design Goals

1. **Lossless round-trip for common fields** - Configurations should survive translation without data loss where possible
2. **Forward compatibility** - Unknown fields are preserved through translation
3. **Clear semantics** - Field names and values have consistent meaning across platforms

## Field Mapping Table

| Canonical Field | Claude Code Field | OpenCode Field | Type | Notes |
|-----------------|-------------------|----------------|------|-------|
| `name` | (map key) | (map key) | `string` | Server identifier; used as the map key in both platforms |
| `command` | `command` | `command[0]` | `string` | Executable path. OpenCode combines with args into single array |
| `args` | `args` | `command[1:]` | `[]string` | Command arguments. OpenCode combines with command |
| `url` | `url` | `url` | `string` | Remote server endpoint for SSE transport |
| `transport` | `transport` | `type` | `string` | See [Transport Mapping](#transport-mapping) below |
| `env` | `env` | `environment` | `map[string]string` | Environment variables for the server process |
| `headers` | `headers` | `headers` | `map[string]string` | HTTP headers for SSE transport |
| `platforms` | `platforms` | N/A | `[]string` | **LOSSY**: OpenCode does not support platform restrictions |
| `disabled` | `disabled` | `disabled` | `bool` | Whether server is temporarily disabled |

### Transport Mapping

Transport types are mapped between different naming conventions:

| Canonical | Claude Code | OpenCode | Description |
|-----------|-------------|----------|-------------|
| `stdio` | `stdio` | `local` | Local process via stdin/stdout |
| `sse` | `sse` | `remote` | Remote server via Server-Sent Events |

## Lossy Conversions

### Platforms Field (OpenCode)

The `platforms` field restricts an MCP server to specific operating systems (e.g., `["darwin"]` for macOS-only servers). **OpenCode does not support this field**, so it is lost during translation.

```go
// Original canonical config
server := &mcp.Server{
    Name:      "macos-tools",
    Command:   "/usr/local/bin/macos-mcp",
    Platforms: []string{"darwin"}, // ⚠️ Lost in OpenCode
}

// After OpenCode round-trip
// server.Platforms is nil
```

**Workaround**: If platform restrictions are critical, consider:
- Using separate configuration files per platform
- Adding platform detection in the MCP server itself
- Documenting platform requirements in server metadata

## Example Configurations

### Local Stdio Server

A typical local MCP server using stdio transport with environment variables.

**Canonical (aix)**:
```json
{
  "servers": {
    "github": {
      "name": "github",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "transport": "stdio",
      "env": {
        "GITHUB_TOKEN": "ghp_xxxxxxxxxxxx"
      }
    }
  }
}
```

**Claude Code**:
```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "transport": "stdio",
      "env": {
        "GITHUB_TOKEN": "ghp_xxxxxxxxxxxx"
      }
    }
  }
}
```

**OpenCode**:
```json
{
  "mcp": {
    "github": {
      "command": ["npx", "-y", "@modelcontextprotocol/server-github"],
      "type": "local",
      "environment": {
        "GITHUB_TOKEN": "ghp_xxxxxxxxxxxx"
      }
    }
  }
}
```

### Remote SSE Server

A remote MCP server using SSE transport with authentication headers.

**Canonical (aix)**:
```json
{
  "servers": {
    "api-gateway": {
      "name": "api-gateway",
      "url": "https://api.example.com/mcp/v1",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer eyJhbGc..."
      }
    }
  }
}
```

**Claude Code**:
```json
{
  "mcpServers": {
    "api-gateway": {
      "url": "https://api.example.com/mcp/v1",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer eyJhbGc..."
      }
    }
  }
}
```

**OpenCode**:
```json
{
  "mcp": {
    "api-gateway": {
      "url": "https://api.example.com/mcp/v1",
      "type": "remote",
      "headers": {
        "Authorization": "Bearer eyJhbGc..."
      }
    }
  }
}
```

### Platform-Restricted Server (Lossy)

A server restricted to macOS. Note that OpenCode loses the platform restriction.

**Canonical (aix)**:
```json
{
  "servers": {
    "macos-tools": {
      "name": "macos-tools",
      "command": "/usr/local/bin/macos-mcp-server",
      "transport": "stdio",
      "platforms": ["darwin"]
    }
  }
}
```

**Claude Code** (preserved):
```json
{
  "mcpServers": {
    "macos-tools": {
      "command": "/usr/local/bin/macos-mcp-server",
      "transport": "stdio",
      "platforms": ["darwin"]
    }
  }
}
```

**OpenCode** (platforms lost):
```json
{
  "mcp": {
    "macos-tools": {
      "command": ["/usr/local/bin/macos-mcp-server"],
      "type": "local"
    }
  }
}
```

## Unknown Field Preservation

Both the canonical format and platform translators preserve unknown fields during round-trips. This ensures forward compatibility when platforms add new configuration options.

```json
// Input with unknown field
{
  "servers": {
    "test": {
      "name": "test",
      "command": "test-cmd",
      "future_feature": "some value"  // Unknown field
    }
  }
}

// Output after round-trip - unknown field preserved
{
  "servers": {
    "test": {
      "name": "test",
      "command": "test-cmd",
      "future_feature": "some value"  // Still present
    }
  }
}
```

## Implementation Notes

### Command/Args Handling

Claude Code and the canonical format separate the executable (`command`) from its arguments (`args`). OpenCode combines them into a single `command` array.

```go
// Canonical → OpenCode
opencode.Command = append([]string{canonical.Command}, canonical.Args...)

// OpenCode → Canonical
canonical.Command = opencode.Command[0]
canonical.Args = opencode.Command[1:]
```

### Transport Type Inference

When `transport` (canonical/Claude) or `type` (OpenCode) is not explicitly set, it can be inferred:

- If `command` is set → `stdio` / `local`
- If `url` is set → `sse` / `remote`

### JSON Output Format

Both translators output formatted JSON with 2-space indentation for readability:

```go
json.MarshalIndent(config, "", "  ")
```
