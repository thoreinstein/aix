# Agent Init Output Handling Specification

This document specifies the output path resolution, conflict handling, and user feedback for the `aix agent init` command.

## Overview

The `aix agent init` command creates a new agent definition file. Unlike `command init` which creates a directory with `command.md`, agent init creates a single `AGENT.md` file (or uses the provided filename directly if it ends in `.md`).

## Output Path Resolution

### Path Resolution Rules

| Invocation | Output Path | Description |
|------------|-------------|-------------|
| `aix agent init` | `./<name>/AGENT.md` | Interactive; creates directory named after agent |
| `aix agent init <path>` | `<path>/AGENT.md` | Creates AGENT.md in specified directory |
| `aix agent init --name=foo` | `./foo/AGENT.md` | Creates directory from name flag |
| `aix agent init foo.md` | `./foo.md` | Direct file path (ends with `.md`) |
| `aix agent init path/to/agent.md` | `./path/to/agent.md` | Direct file path with directory |

### Path Resolution Algorithm

```go
func resolveAgentOutputPath(args []string, name string) (string, error) {
    var basePath string

    if len(args) > 0 {
        basePath = args[0]
    } else if name != "" {
        basePath = name
    } else {
        // Will be determined interactively
        return "", nil
    }

    // If path ends with .md, use it directly
    if strings.HasSuffix(basePath, ".md") {
        return filepath.Abs(basePath)
    }

    // Otherwise, treat as directory and append AGENT.md
    absPath, err := filepath.Abs(basePath)
    if err != nil {
        return "", fmt.Errorf("resolving path: %w", err)
    }

    return filepath.Join(absPath, "AGENT.md"), nil
}
```

### Examples

```bash
# No arguments - interactive, uses prompted name
$ aix agent init
Agent Name [my-agent]: code-reviewer
# Creates: ./code-reviewer/AGENT.md

# Directory path argument
$ aix agent init my-agents/reviewer
# Creates: ./my-agents/reviewer/AGENT.md

# Direct .md file path
$ aix agent init agents/security-checker.md
# Creates: ./agents/security-checker.md

# Name flag only
$ aix agent init --name=test-helper
# Creates: ./test-helper/AGENT.md

# Name flag with path argument (path takes precedence for location)
$ aix agent init custom-dir --name=my-agent
# Creates: ./custom-dir/AGENT.md (agent name is "my-agent")
```

## Conflict Resolution

### Default Behavior

If the target file already exists, the command fails with an error:

```
Error: file already exists: ./code-reviewer/AGENT.md. Use --force to overwrite.
```

Exit code: 1

### Force Overwrite

The `--force` (or `-f`) flag overwrites existing files without prompting:

```bash
$ aix agent init code-reviewer --force
# Overwrites ./code-reviewer/AGENT.md if it exists
```

### Implementation

```go
if _, err := os.Stat(outputPath); err == nil {
    if !forceFlag {
        fmt.Printf("Error: file already exists: %s. Use --force to overwrite.\n", outputPath)
        return errAgentInitFailed
    }
}
```

## Directory Creation

### Behavior

Parent directories are created automatically as needed, similar to `mkdir -p`:

```bash
$ aix agent init deep/nested/path/agent
# Creates: ./deep/nested/path/agent/AGENT.md
# Also creates: ./deep/, ./deep/nested/, ./deep/nested/path/, ./deep/nested/path/agent/
```

### Permissions

Directories are created with `0755` permissions:

```go
if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
    return fmt.Errorf("creating directory: %w", err)
}
```

### Error Handling

If directory creation fails (e.g., permission denied), the command fails with a descriptive error:

```
Error: creating directory: permission denied
```

## File Permissions

Created agent files use `0644` permissions (owner read/write, group/other read):

```go
if err := os.WriteFile(outputPath, content, 0o644); err != nil {
    return fmt.Errorf("writing agent file: %w", err)
}
```

## Success Message

Upon successful creation, the command outputs:

```
[OK] Agent 'code-reviewer' created at ./code-reviewer/AGENT.md

Next steps:
  1. Edit ./code-reviewer/AGENT.md to customize your agent
  2. Run: aix agent validate ./code-reviewer
  3. Run: aix agent install ./code-reviewer
```

### Message Components

| Component | Description |
|-----------|-------------|
| Checkmark (`[OK]`) | Visual success indicator |
| Agent name | The name from frontmatter (not necessarily the filename) |
| Path | Absolute or relative path to the created file |
| Next steps | Actionable guidance for the user |

### Implementation

```go
fmt.Printf("[OK] Agent '%s' created at %s\n", name, outputPath)
fmt.Println()
fmt.Println("Next steps:")
fmt.Printf("  1. Edit %s to customize your agent\n", outputPath)
fmt.Printf("  2. Run: aix agent validate %s\n", filepath.Dir(outputPath))
fmt.Printf("  3. Run: aix agent install %s\n", filepath.Dir(outputPath))
```

## Name Derivation

The agent name is determined by:

1. **`--name` flag**: If provided, use this value
2. **Interactive prompt**: If not provided, prompt user with a default
3. **Default from path**: Derived from the path argument (sanitized)

### Default Name Derivation

```go
func deriveDefaultAgentName(path string) string {
    base := filepath.Base(path)

    // Strip .md extension if present
    if strings.HasSuffix(base, ".md") {
        base = strings.TrimSuffix(base, ".md")
    }

    // Sanitize to valid agent name
    return sanitizeAgentName(base)
}
```

## Error Cases

| Condition | Error Message | Exit Code |
|-----------|---------------|-----------|
| File exists (no `--force`) | `Error: file already exists: <path>. Use --force to overwrite.` | 1 |
| Invalid agent name | `Error: agent name must be lowercase alphanumeric with hyphens, starting with a letter` | 1 |
| Directory creation fails | `Error: creating directory: <reason>` | 1 |
| File write fails | `Error: writing agent file: <reason>` | 1 |
| Path resolution fails | `Error: resolving path: <reason>` | 1 |

## Comparison with Command Init

| Aspect | `command init` | `agent init` |
|--------|----------------|--------------|
| Output file | `command.md` | `AGENT.md` |
| Directory structure | Always creates directory | Creates directory unless `.md` path given |
| Validation step | `aix command validate` | `aix agent validate` |
| Install step | `aix command install` | `aix agent install` |

## Flag Summary

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--name` | | `string` | `""` | Agent name (prompted if not provided) |
| `--description` | `-d` | `string` | `""` | Short description |
| `--force` | `-f` | `bool` | `false` | Overwrite existing file |
