# ADR-001: Unified Agent Configuration CLI (aix)

## Status

**Implemented** | 2026-01-22

## Context

The AI coding assistant landscape has fragmented across multiple platforms, each with distinct configuration formats, locations, and conventions:

| Platform | Config Location | Instructions | Config Format |
|----------|-----------------|--------------|---------------|
| Claude Code | `~/.claude/`, `.claude/` | `CLAUDE.md` | JSON |
| OpenCode | `~/.config/opencode/` | `AGENTS.md` | JSON |
| Codex | `~/.codex/`, `.codex/` | `CODEX.md` / `AGENTS.md` | JSON |
| Gemini CLI | `~/.gemini/`, `.gemini/` | `GEMINI.md` | JSON |

This fragmentation creates friction for developers who:

1. **Use multiple assistants** -- Different projects or organizations mandate different tools
2. **Share configurations** -- Skills and commands must be manually ported between platforms
3. **Maintain consistency** -- MCP servers, tool permissions, and instructions drift across platforms
4. **Follow specifications** -- Agent Skills Spec and MCP Spec compliance varies by platform

Existing tooling is inadequate:

- **skillz (Python)**: Limited feature set, no scoped tool permissions, Python dependency
- **Platform-specific CLIs**: Each requires learning new commands and conventions
- **Manual management**: Error-prone, no validation, tedious cross-platform sync

## Decision

Build a unified Go CLI tool (`aix`) that provides a single interface for managing agent configurations across all major AI coding assistants.

### Name Selection

| Candidate | Pros | Cons |
|-----------|------|------|
| `aix` | Short (3 chars), AI + X (cross-platform), Unix-y feel, memorable | Similar to IBM AIX (ancient, irrelevant) |
| `agentctl` | Follows k8s convention (`kubectl`), clear purpose | Longer to type, verbose |
| `herd` | Memorable, verb-friendly, implies managing multiple things | Too cute, unclear for new users |
| `dot` | Ultra-minimal, references dotfiles | Too generic, conflicts with graphviz |

**Selected: `aix`** -- Short, memorable, captures AI + cross-platform essence

## Decision Drivers

1. **Developer Experience**: One tool, one mental model, all platforms
2. **Spec Compliance**: First-class support for Agent Skills Spec and MCP Spec
3. **Extensibility**: Plugin architecture for new platforms without core changes
4. **Performance**: Fast startup, no runtime dependencies (Go single binary)
5. **Correctness**: Strong validation prevents misconfiguration

## Target Platforms

### Claude Code

```
Config Locations:
  Global:  ~/.claude/
  Project: .claude/

Files:
  Instructions:  CLAUDE.md (project root) or ~/.claude/CLAUDE.md
  Commands:      .claude/commands/<name>.md
  MCP Config:    ~/.claude.json (global) or project specific
  Settings:      ~/.claude/config.json

Variables:
  $ARGUMENTS     - Command arguments
  $SELECTION     - Selected code (IDE context)
```

### OpenCode

```
Config Locations:
  Global:  ~/.config/opencode/
  Project: (project root)

Files:
  Instructions:  AGENTS.md (project root)
  Skills:        ~/.config/opencode/skill/<name>/SKILL.md
  Commands:      ~/.config/opencode/commands/<name>.md
  MCP Config:    opencode.json (mcpServers key)
  Settings:      opencode.json

Variables:
  $ARGUMENTS     - Command arguments
```

### Codex (OpenAI)

```
Config Locations:
  Global:  ~/.codex/
  Project: .codex/

Files:
  Instructions:  CODEX.md or AGENTS.md (project root)
  Commands:      .codex/commands/<name>.md
  MCP Config:    mcp.json (assumed)
```

### Gemini CLI

- **Format:** TOML/JSON
- **Paths:**
  - Global:  `~/.gemini/`
  - Project: `.gemini/`
  - Instructions:  `GEMINI.md` (project root) or `~/.gemini/GEMINI.md`
  - Commands:      `.gemini/commands/<name>.toml`
  - MCP Config:    `~/.gemini/settings.json`
  - Settings:      `~/.gemini/settings.json`
- **Translation:**
  - `{{argument}}` -> `{{argument}}` (Native support)
  - `{{selection}}` -> (Not supported)

## Managed Resources

### 1. Skills

Skills are reusable agent capabilities following the [Agent Skills Specification](https://agentskills.io/specification).

**Structure:**
```
skill-name/
|-- SKILL.md           # Required: Frontmatter + instructions
|-- commands/          # Optional: Skill-specific commands
 |   `--- run.md
`--- templates/         # Optional: Code templates
    `--- component.tsx
```

**SKILL.md Format:**
```markdown
---
name: git-workflow
description: Git workflow automation with PR creation
version: 1.0.0
author: developer@example.com
tools:
  - Bash(git:*)
  - Bash(gh:*)
  - Read
  - Edit
  - Glob
  - Grep
triggers:
  - /git
  - /pr
  - /commit
---

# Git Workflow Skill

You are an expert at Git operations...
```

### 2. Commands

Slash commands provide quick access to common operations.

**Format:**
```markdown
---
name: deploy
description: Deploy to specified environment
arguments:
  - name: environment
    required: true
    description: Target environment (staging|production)
---

Deploy the application to $ARGUMENTS environment.
```

### 3. MCP Servers

Model Context Protocol servers extend agent capabilities.

**Configuration:**
Servers are configured via `aix mcp add` and stored in the platform-native configuration files.

**Supported Transports:**
- `stdio`: Local process execution
- `sse`: Remote Server-Sent Events

### 4. Agents

Full agent personas and configurations, primarily for Claude Code and OpenCode.

## Architecture

### Directory Structure

```
aix/
|-- cmd/
 |   `--- aix/
 |       |-- main.go            # Entry point
 |       `--- commands/          # Cobra command definitions
 |           |-- root.go
 |           |-- skill_*.go     # Skill management
 |           |-- command_*.go   # Command management
 |           |-- mcp_*.go       # MCP management
 |           |-- agent_*.go     # Agent management
 |           |-- config.go      # CLI config
 |           `--- init.go        # Initialization
|-- internal/
 |   |-- cli/
 |    |   |-- platform.go        # Platform interface (Consumer)
 |    |   `--- registry.go        # Platform registry
 |   |-- config/                # Internal configuration
 |   |-- paths/                 # Path resolution
 |   |-- mcp/
 |    |   |-- types.go           # MCP data structures
 |    |   `--- translator.go      # Format translation
 |   |-- skill/                 # Skill logic
 |   |-- command/               # Command logic
 |   `--- platform/
 |       |-- claude/            # Claude Code implementation
 |       |-- opencode/          # OpenCode implementation
 |       `--- ...                # Other platforms
|-- pkg/
 |   `--- frontmatter/           # YAML frontmatter parsing
|-- go.mod
|-- go.sum
`--- README.md
```

### Core Interfaces

The core abstraction is the `Platform` interface defined in `internal/cli/platform.go`. This interface allows the CLI to interact with any underlying AI assistant platform uniformly.

```go
// internal/cli/platform.go

type Platform interface {
    // Identity
    Name() string
    DisplayName() string
    IsAvailable() bool

    // Skills
    SkillDir() string
    InstallSkill(skill any) error
    UninstallSkill(name string) error
    ListSkills() ([]SkillInfo, error)
    GetSkill(name string) (any, error)

    // Commands
    CommandDir() string
    InstallCommand(cmd any) error
    UninstallCommand(name string) error
    ListCommands() ([]CommandInfo, error)
    GetCommand(name string) (any, error)

    // MCP
    MCPConfigPath() string
    AddMCP(server any) error
    RemoveMCP(name string) error
    ListMCP() ([]MCPInfo, error)
    GetMCP(name string) (any, error)
    EnableMCP(name string) error
    DisableMCP(name string) error

    // Agents
    AgentDir() string
    InstallAgent(agent any) error
    UninstallAgent(name string) error
    ListAgents() ([]AgentInfo, error)
    GetAgent(name string) (any, error)
}
```

### Configuration

`aix` maintains its own minimal configuration at `~/.config/aix/config.yaml`.

```yaml
version: 1
default_platforms:
  - claude
  - opencode
  - gemini
```

This configuration controls which platforms are targeted by default when running commands like `aix mcp list`.

## References

- [Agent Skills Specification](https://agentskills.io/specification)
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [Claude Code Documentation](https://docs.anthropic.com/claude-code)
- [OpenCode Repository](https://github.com/opencode-ai/opencode)
- [Gemini CLI Repository](https://github.com/google/gemini-cli)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
