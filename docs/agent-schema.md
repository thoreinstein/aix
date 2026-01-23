# Agent Schema Reference

This document describes the canonical agent schema used by `aix` and the platform-specific formats used by Claude Code and OpenCode.

## Overview

An agent file defines a specialized AI assistant with custom instructions and behavior. Agents are markdown files with optional YAML frontmatter that configure the assistant's identity, capabilities, and operational parameters.

### Design Goals

1. **Portable definitions** - Agents should work across supported platforms where possible
2. **Markdown-first** - Instructions are written in markdown for readability
3. **Minimal required fields** - Only `name` and prompt body are strictly required
4. **Platform extensions** - Optional fields enable platform-specific features

## File Format

Agent files use YAML frontmatter followed by a markdown body:

```markdown
---
name: code-reviewer
description: Reviews code for quality, security, and best practices
---

You are a senior code reviewer. When reviewing code:

1. Check for security vulnerabilities
2. Verify error handling is complete
3. Ensure tests cover edge cases
```

The file is stored as `{name}.md` in the platform's agent directory:

- **Claude Code**: `.claude/agents/{name}.md`
- **OpenCode**: `.opencode/agents/{name}.md`

### Frontmatter Rules

- Frontmatter is **optional** - a file with only markdown content is valid
- When present, frontmatter must be delimited by `---` on its own line
- The `name` field in frontmatter is optional; if omitted, the filename (without `.md`) is used

## Required Fields

| Field | Type | Source | Description |
|-------|------|--------|-------------|
| `name` | `string` | Frontmatter or filename | Unique identifier for the agent |
| `description` | `string` | Frontmatter | Brief explanation of the agent's purpose |
| Instructions | `string` | Markdown body | The agent's system prompt and behavioral instructions |

### Name Derivation

The agent name is determined by:

1. **Explicit**: `name` field in YAML frontmatter
2. **Implicit**: Filename without the `.md` extension

When installing via `aix`, the explicit name takes precedence and determines the output filename.

## Optional Fields

| Field | Type | Platform | Default | Description |
|-------|------|----------|---------|-------------|
| `tools` | `[]string` | Both | `nil` | Tool permissions available to the agent |
| `model` | `string` | Both | Platform default | Model to use when invoking this agent |
| `mode` | `string` | OpenCode only | `""` | Operational mode: `primary`, `subagent`, or `all` |
| `temperature` | `float64` | OpenCode only | `0.0` | Response randomness (0.0-1.0) |

### Mode Values (OpenCode)

| Mode | Description |
|------|-------------|
| `primary` | Agent runs as the main assistant |
| `subagent` | Agent runs as a delegated sub-task handler |
| `all` | Agent can run in either mode |

### Temperature Guidelines

| Range | Behavior | Use Case |
|-------|----------|----------|
| 0.0-0.3 | Deterministic, focused | Code generation, factual tasks |
| 0.4-0.6 | Balanced | General assistance |
| 0.7-1.0 | Creative, varied | Brainstorming, creative writing |

## Platform-Specific Mappings

### Field Mapping Table

| Canonical Field | Claude Code | OpenCode | Notes |
|-----------------|-------------|----------|-------|
| `name` | `name` | `name` | Required; derived from filename if not in frontmatter |
| `description` | `description` | `description` | Required in canonical schema |
| `instructions` | Body content | Body content | Markdown after frontmatter |
| `mode` | N/A | `mode` | **LOSSY**: Claude Code does not support mode |
| `temperature` | N/A | `temperature` | **LOSSY**: Claude Code does not support temperature |

### Claude Code Agent Struct

```go
type Agent struct {
    Name         string `yaml:"name" json:"name"`
    Description  string `yaml:"description,omitempty" json:"description,omitempty"`
    Instructions string `yaml:"-" json:"-"` // Markdown body
}
```

### OpenCode Agent Struct

```go
type Agent struct {
    Name         string  `yaml:"name" json:"name"`
    Description  string  `yaml:"description,omitempty" json:"description,omitempty"`
    Mode         string  `yaml:"mode,omitempty" json:"mode,omitempty"`
    Temperature  float64 `yaml:"temperature,omitempty" json:"temperature,omitempty"`
    Instructions string  `yaml:"-" json:"-"` // Markdown body
}
```

## Lossy Conversions

### Mode and Temperature (Claude Code)

The `mode` and `temperature` fields are OpenCode-specific. When translating an OpenCode agent to Claude Code, these fields are lost:

```go
// Original OpenCode agent
agent := &Agent{
    Name:        "creative-writer",
    Description: "Generates creative content",
    Mode:        "subagent",      // ⚠️ Lost in Claude Code
    Temperature: 0.8,             // ⚠️ Lost in Claude Code
}

// After Claude Code round-trip
// agent.Mode is ""
// agent.Temperature is 0.0
```

**Workaround**: If mode or temperature are critical:
- Document the intended settings in the agent's instructions
- Use platform-specific agent files when behavior must differ

## Validation Rules

### Name Validation

Agent names must conform to the following rules:

| Rule | Constraint |
|------|------------|
| Format | Lowercase alphanumeric with hyphens |
| Length | 1-64 characters |
| Pattern | `^[a-z][a-z0-9]*(-[a-z0-9]+)*$` |
| Restrictions | No consecutive hyphens (`--`), no leading/trailing hyphens |

**Valid names:**
- `code-reviewer`
- `test-generator`
- `api-v2-helper`
- `go`

**Invalid names:**
- `Code-Reviewer` (uppercase)
- `-reviewer` (leading hyphen)
- `code--reviewer` (consecutive hyphens)
- `code_reviewer` (underscore not allowed)
- `123-agent` (must start with letter)

### Description Validation

- Required for canonical agents
- Should be a brief, single-line description
- Recommended maximum: 200 characters

### Instructions Validation

- Must not be empty
- Should contain meaningful guidance for the agent
- Markdown formatting is preserved

## Example Agent Files

### Minimal Agent (No Frontmatter)

File: `simple-helper.md`

```markdown
You are a helpful assistant. Answer questions clearly and concisely.
```

The name is derived from the filename: `simple-helper`

### Claude Code Agent

File: `security-reviewer.md`

```markdown
---
description: Reviews code for security vulnerabilities and best practices
---

You are a security-focused code reviewer. For every code review:

## Checklist

1. **Input validation** - Check all external inputs are validated
2. **Authentication** - Verify auth checks are in place
3. **Authorization** - Confirm proper access controls
4. **Secrets** - Ensure no hardcoded credentials
5. **Dependencies** - Flag known vulnerable packages

## Response Format

Provide findings as:
- 🔴 **Critical**: Must fix before merge
- 🟡 **Warning**: Should address soon
- 🟢 **Info**: Suggestions for improvement
```

### OpenCode Agent with Mode and Temperature

File: `brainstorm-agent.md`

```markdown
---
description: Generates creative ideas and explores possibilities
mode: subagent
temperature: 0.9
---

You are a creative brainstorming partner. When asked to brainstorm:

1. Generate at least 5 diverse ideas
2. Include both conventional and unconventional approaches
3. Don't self-censor - wild ideas often spark practical solutions
4. Build on ideas iteratively when asked

Remember: Quantity over quality initially. We'll refine later.
```

### Full-Featured Agent

File: `go-expert.md`

```markdown
---
name: go-expert
description: Expert Go engineer following idiomatic patterns and best practices
mode: primary
temperature: 0.2
---

You are a principal Go engineer embodying the design philosophy of Rob Pike.

## Core Principles

- **Clarity over cleverness** - Simple code is a feature
- **Idiomatic Go** - Standard library first, minimal dependencies
- **Error handling** - Always handle errors explicitly

## Code Style

- Use `gofmt` formatting
- Prefer composition over inheritance
- Keep interfaces small and focused
- Use meaningful variable names

## Quality Checklist

Before declaring code complete:

- [ ] All errors are handled
- [ ] Tests cover critical paths
- [ ] No data races (`go test -race`)
- [ ] Documentation for exported symbols
```

## Implementation Notes

### Frontmatter Parsing

Agent files are parsed using the `pkg/frontmatter` package:

```go
agent := &Agent{}
body, err := frontmatter.Parse(reader, agent)
if err != nil {
    return nil, fmt.Errorf("parsing frontmatter: %w", err)
}
agent.Instructions = strings.TrimSpace(string(body))
```

### Frontmatter Generation

Frontmatter is only written when metadata fields are present:

```go
// Only include frontmatter if there's a description (Claude Code)
if a.Description == "" {
    return a.Instructions + "\n", nil
}

// OpenCode: include if any metadata field is set
if a.Description == "" && a.Mode == "" && a.Temperature == 0 {
    return a.Instructions + "\n", nil
}
```

### File Naming

Agent files always use the `.md` extension:

```go
func (p *Paths) AgentPath(name string) string {
    return filepath.Join(p.AgentDir(), name+".md")
}
```
