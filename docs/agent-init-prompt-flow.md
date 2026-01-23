# Agent Init Prompt Flow Specification

This document specifies the interactive prompt flow for `aix agent init`, following the established patterns from `command_init.go`.

## Overview

The `aix agent init` command creates a new agent file interactively. It follows the same prompt mechanics as `aix command init`:

- Prompts are skipped when corresponding flags are provided
- Each prompt shows a default value in brackets when available
- Empty input accepts the default value
- Ctrl+C exits cleanly with no output

## Prompt Sequence

### 1. Agent Name

```
Agent Name [<default>]: _
```

| Property | Value |
|----------|-------|
| Label | `Agent Name` |
| Default | Derived from current directory name, sanitized |
| Required | Yes |
| Validation | See [Name Validation](#name-validation) |
| Flag override | `--name` |

**Examples:**
```
Agent Name [my-project]: code-reviewer
Agent Name [my-project]:              # Accepts "my-project"
```

### 2. Description

```
Description [A helpful AI agent]: _
```

| Property | Value |
|----------|-------|
| Label | `Description` |
| Default | `A helpful AI agent` |
| Required | Yes (has default) |
| Validation | Non-empty (default satisfies this) |
| Flag override | `--description` or `-d` |

**Examples:**
```
Description [A helpful AI agent]: Reviews code for security issues
Description [A helpful AI agent]:     # Accepts "A helpful AI agent"
```

### 3. Model (Optional)

```
Model (optional) []: _
```

| Property | Value |
|----------|-------|
| Label | `Model (optional)` |
| Default | Empty string |
| Required | No |
| Validation | None (accepts any string) |
| Flag override | `--model` |

**Examples:**
```
Model (optional) []: claude-3-5-sonnet
Model (optional) []:                   # Accepts empty (platform default)
```

## Input Validation

### Name Validation

Agent names must conform to the pattern defined in `docs/agent-schema.md`:

| Rule | Constraint |
|------|------------|
| Pattern | `^[a-z][a-z0-9]*(-[a-z0-9]+)*$` |
| Length | 1-64 characters |
| Start | Must begin with a lowercase letter |
| Characters | Lowercase letters, digits, and hyphens only |
| Hyphens | No consecutive (`--`), leading, or trailing hyphens |

**Valid names:**
- `code-reviewer`
- `test-generator`
- `api-v2-helper`
- `go`

**Invalid names:**
- `Code-Reviewer` (uppercase)
- `-reviewer` (leading hyphen)
- `code--reviewer` (consecutive hyphens)
- `code_reviewer` (underscore)
- `123-agent` (starts with digit)

### Description Validation

- Required field, but has a default value
- Any non-empty string is valid
- Empty input with no default would trigger an error (not possible with current defaults)

### Model Validation

- Optional field with no validation
- Empty string is valid (uses platform default)
- Any string is accepted

## Error Messages

### Invalid Name Format

When the name fails regex validation:

```
Error: agent name must be lowercase alphanumeric with hyphens (e.g., 'my-agent')
```

### Name Too Long

When the name exceeds 64 characters:

```
Error: agent name must be at most 64 characters (got <N>)
```

### Empty Name

When name is empty (no input and no default):

```
Error: agent name is required
```

### Empty Description (Edge Case)

If description were required without a default:

```
Error: description is required
```

> **Note:** This error is not reachable with current defaults but should be implemented for defensive validation.

## Default Handling

### Name Default

The default name is derived from the current directory name through sanitization:

1. Convert to lowercase
2. Replace invalid characters (non-alphanumeric, non-hyphen) with hyphens
3. Trim leading/trailing hyphens
4. If result is empty or invalid, fallback to `new-agent`

**Sanitization examples:**

| Directory Name | Sanitized Default |
|----------------|-------------------|
| `MyProject` | `myproject` |
| `my_project` | `my-project` |
| `My Cool Agent!` | `my-cool-agent` |
| `123-test` | `new-agent` (invalid start) |
| `---` | `new-agent` (empty after trim) |

### Description Default

Fixed default: `A helpful AI agent`

### Model Default

Empty string (no default). When empty, the platform's default model is used at runtime.

## Ctrl+C Handling

When the user presses Ctrl+C during any prompt:

1. The `bufio.Scanner.Scan()` returns `false`
2. The prompt function returns the default value
3. If there's no meaningful default, initialization may fail gracefully
4. No partial files are written
5. Exit code is 0 (clean exit)

**Implementation pattern** (from `prompt()` function):

```go
if !scanner.Scan() {
    return def
}
```

## Complete Prompt Session Example

### Interactive Session (All Defaults)

```
$ aix agent init
Agent Name [my-project]:
Description [A helpful AI agent]:
Model (optional) []:
Writing agent file...
Agent 'my-project' created at ./my-project.md

  Next steps:
    1. Edit ./my-project.md with your agent's instructions
    2. Run: aix agent install ./my-project.md
```

### Interactive Session (Custom Values)

```
$ aix agent init
Agent Name [my-project]: code-reviewer
Description [A helpful AI agent]: Reviews code for security vulnerabilities
Model (optional) []: claude-3-5-sonnet
Writing agent file...
Agent 'code-reviewer' created at ./code-reviewer.md

  Next steps:
    1. Edit ./code-reviewer.md with your agent's instructions
    2. Run: aix agent install ./code-reviewer.md
```

### Non-Interactive Session (Flags)

```
$ aix agent init --name security-reviewer --description "Security-focused code review" --model claude-3-5-sonnet
Writing agent file...
Agent 'security-reviewer' created at ./security-reviewer.md

  Next steps:
    1. Edit ./security-reviewer.md with your agent's instructions
    2. Run: aix agent install ./security-reviewer.md
```

### Validation Error

```
$ aix agent init
Agent Name [my-project]: Code_Reviewer
Error: agent name must be lowercase alphanumeric with hyphens (e.g., 'my-agent')
$ echo $?
1
```

## Implementation Reference

The prompt implementation follows the pattern in `cmd/aix/commands/skill_init.go`:

```go
func prompt(scanner *bufio.Scanner, label, def string) string {
    fmt.Printf("%s", label)
    if def != "" {
        fmt.Printf(" [%s]", def)
    }
    fmt.Print(": ")

    if !scanner.Scan() {
        return def
    }
    input := strings.TrimSpace(scanner.Text())
    if input == "" {
        return def
    }
    return input
}
```

## Related Documentation

- [Agent Schema Reference](agent-schema.md) - Complete agent file format specification
- `cmd/aix/commands/command_init.go` - Reference implementation for command init
- `cmd/aix/commands/skill_init.go` - Reference implementation for skill init with shared prompt utilities
