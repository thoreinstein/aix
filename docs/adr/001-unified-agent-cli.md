# ADR-001: Unified Agent Configuration CLI (aix)

## Status

**Proposed** | 2026-01-20

## Context

The AI coding assistant landscape has fragmented across multiple platforms, each with distinct configuration formats, locations, and conventions:

| Platform | Config Location | Instructions | Config Format |
|----------|-----------------|--------------|---------------|
| Claude Code | `~/.claude/`, `.claude/` | `CLAUDE.md` | JSON |
| OpenCode | `~/.config/opencode/` | `AGENTS.md` | JSON |
| Codex | `~/.codex/`, `.codex/` | `CODEX.md` / `AGENTS.md` | JSON |
| Gemini CLI | `~/.gemini/`, `.gemini/` | `GEMINI.md` | TOML |

This fragmentation creates friction for developers who:

1. **Use multiple assistants** - Different projects or organizations mandate different tools
2. **Share configurations** - Skills and commands must be manually ported between platforms
3. **Maintain consistency** - MCP servers, tool permissions, and instructions drift across platforms
4. **Follow specifications** - Agent Skills Spec and MCP Spec compliance varies by platform

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

**Selected: `aix`** - Short, memorable, captures AI + cross-platform essence

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
  MCP Config:    ~/.claude/mcp_servers.json (global)
                 .mcp.json (project)
  Settings:      ~/.claude/settings.json

Variables:
  $ARGUMENTS     - Command arguments
  $SELECTION     - Selected code (IDE context)

Tool Permissions:
  Supports scoped permissions via allowlist in settings.json
  Format: ["Bash(git:*)", "Read", "Write", "Edit"]
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

Tool Permissions:
  Defined in SKILL.md frontmatter
  Format: tools: ["Bash(git:diff)", "Read", "Glob"]
```

### Codex (OpenAI)

```
Config Locations:
  Global:  ~/.codex/
  Project: .codex/

Files:
  Instructions:  CODEX.md or AGENTS.md (project root)
  Commands:      .codex/commands/<name>.md (assumed, verify)
  MCP Config:    TBD - research required

Variables:
  $ARGUMENTS     - Command arguments (assumed compatible)

Notes:
  - Relatively new, spec may evolve
  - Monitor for MCP support announcements
```

### Gemini CLI

```
Config Locations:
  Global:  ~/.gemini/
  Project: .gemini/

Files:
  Instructions:  GEMINI.md (project root)
  Config:        ~/.gemini/settings.toml (TOML format!)
  Commands:      .gemini/commands/<name>.md

Variables:
  {{argument}}   - Different from $ARGUMENTS!
  {{selection}}  - Selected code

Translation Required:
  - $ARGUMENTS → {{argument}}
  - JSON config → TOML config
```

## Managed Resources

### 1. Skills

Skills are reusable agent capabilities following the [Agent Skills Specification](https://agentskills.io/specification).

**Structure:**
```
skill-name/
├── SKILL.md           # Required: Frontmatter + instructions
├── commands/          # Optional: Skill-specific commands
│   └── run.md
└── templates/         # Optional: Code templates
    └── component.tsx
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

## Capabilities

- Create atomic commits with conventional messages
- Generate PR descriptions from commit history
...
```

**Tool Permission Syntax:**
```
Tool Permission     | Meaning
--------------------|------------------------------------------
Read                | Read any file
Read(src/**)        | Read files matching glob
Write               | Write any file
Write(*.md)         | Write only markdown files
Bash                | Execute any bash command
Bash(git:*)         | Execute git commands only
Bash(npm:install)   | Execute only npm install
Edit                | Edit any file
Glob                | Search for files by pattern
Grep                | Search file contents
WebFetch            | Fetch web content
Task                | Spawn sub-agents
```

**Installation Sources:**
- Local path: `aix skill install ./my-skill`
- Git repository: `aix skill install github.com/user/skill-repo`
- Registry (future): `aix skill install @official/git-workflow`

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
  - name: version
    required: false
    description: Version tag (defaults to HEAD)
---

Deploy the application to $ARGUMENTS environment.

1. Run tests first
2. Build the application
3. Deploy using the CI/CD pipeline
4. Verify deployment health
```

**Platform Translation:**
| Source | Claude/OpenCode/Codex | Gemini |
|--------|----------------------|--------|
| `$ARGUMENTS` | `$ARGUMENTS` | `{{argument}}` |
| `$SELECTION` | `$SELECTION` | `{{selection}}` |

### 3. MCP Servers

Model Context Protocol servers extend agent capabilities.

**Configuration:**
```yaml
# ~/.config/aix/mcp.yaml
servers:
  context7:
    command: npx
    args: ["-y", "@upstash/context7-mcp"]
    transport: stdio
    env:
      CONTEXT7_API_KEY: ${CONTEXT7_API_KEY}
    platforms:
      - claude
      - opencode

  filesystem:
    command: /usr/local/bin/mcp-filesystem
    args: ["--root", "/home/user/projects"]
    transport: stdio
    platforms:
      - all

  remote-db:
    url: https://mcp.example.com/db
    transport: sse
    headers:
      Authorization: Bearer ${DB_MCP_TOKEN}
    platforms:
      - claude
```

**Platform-Specific Output:**

Claude (`~/.claude/mcp_servers.json`):
```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"],
      "env": {
        "CONTEXT7_API_KEY": "..."
      }
    }
  }
}
```

OpenCode (`opencode.json`):
```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"],
      "env": {
        "CONTEXT7_API_KEY": "..."
      }
    }
  }
}
```

### 4. Agent Instructions

Platform-specific instruction files that define agent behavior.

**Sync Strategy:**
```
Source of Truth          Platform-Specific Outputs
─────────────────       ─────────────────────────────
AGENTS.md          →    CLAUDE.md (claude)
                   →    AGENTS.md (opencode) [identity]
                   →    CODEX.md (codex)
                   →    GEMINI.md (gemini, translated)
```

**Include Support:**
```markdown
# AGENTS.md

## Project Guidelines
@include(./docs/coding-standards.md)

## Git Rules
@include(./docs/git-conventions.md)

## Platform-Specific
@if(claude)
Use Claude-specific MCP servers for search.
@endif

@if(opencode)
Use OpenCode skills for code review.
@endif
```

## Proposed CLI Interface

### Skill Management

```bash
# Install from various sources
aix skill install ./my-skill                    # Local directory
aix skill install github.com/user/skills/git   # Git repository  
aix skill install @official/code-review        # Registry (future)

# Platform targeting
aix skill install ./skill --platform=claude     # Single platform
aix skill install ./skill --platform=all        # All platforms (default)
aix skill install ./skill --platforms=claude,opencode  # Multiple

# List and inspect
aix skill list                                  # All installed skills
aix skill list --platform=claude               # Platform-specific
aix skill show git-workflow                    # Skill details
aix skill show git-workflow --tools            # Show tool permissions

# Removal
aix skill remove git-workflow                  # Remove from all platforms
aix skill remove git-workflow --platform=claude

# Validation and creation
aix skill validate ./my-skill                  # Validate skill structure
aix skill init my-new-skill                    # Create skill scaffold
```

### Command Management

```bash
# Install commands
aix cmd install ./commands/deploy.md           # Single command
aix cmd install ./commands/                    # Directory of commands
aix cmd install github.com/user/commands       # From git

# List and inspect
aix cmd list
aix cmd show deploy

# Removal
aix cmd remove deploy
aix cmd remove deploy --platform=gemini
```

### MCP Server Management

```bash
# Add servers (stdio transport)
aix mcp add context7 npx -y @upstash/context7-mcp
aix mcp add filesystem /usr/local/bin/mcp-fs --args="--root,/projects"

# Add with environment variables
aix mcp add myserver ./server --env API_KEY=secret --env DEBUG=true

# Add SSE transport
aix mcp add remote-api --url=https://mcp.example.com --transport=sse

# Platform management
aix mcp enable context7 --platform=claude
aix mcp disable context7 --platform=opencode
aix mcp enable context7 --platform=all

# List and status
aix mcp list                                   # All configured servers
aix mcp list --enabled                         # Currently enabled
aix mcp show context7                          # Server details

# Removal
aix mcp remove context7
```

### Instructions Management

```bash
# Sync instructions across platforms
aix instructions sync                          # Sync AGENTS.md to all
aix instructions sync --from=claude --to=opencode
aix instructions sync --dry-run                # Preview changes

# Validation
aix instructions validate                      # Validate all instruction files
aix instructions validate AGENTS.md            # Specific file

# Show differences
aix instructions diff                          # Show platform divergence
```

### General Commands

```bash
# Initialization
aix init                                       # Initialize aix config
aix init --platforms=claude,opencode           # Specify platforms

# Status overview
aix status                                     # Full status report
aix status --json                              # JSON output
aix status --platform=claude                   # Platform-specific

# Configuration
aix config show                                # Show current config
aix config set default_platforms claude,opencode
aix config set registry.primary https://skills.example.com

# Diagnostics
aix doctor                                     # Check installation health
aix version                                    # Version info
```

## Architecture

### Directory Structure

```
aix/
├── cmd/
│   └── aix/
│       ├── main.go
│       └── commands/
│           ├── root.go
│           ├── skill.go
│           ├── cmd.go
│           ├── mcp.go
│           ├── instructions.go
│           └── config.go
├── internal/
│   ├── config/
│   │   ├── config.go          # Configuration management
│   │   └── paths.go           # Platform path resolution
│   ├── platform/
│   │   ├── platform.go        # Platform interface
│   │   ├── claude.go          # Claude Code adapter
│   │   ├── opencode.go        # OpenCode adapter
│   │   ├── codex.go           # Codex adapter
│   │   ├── gemini.go          # Gemini CLI adapter
│   │   └── registry.go        # Platform registry
│   ├── skill/
│   │   ├── skill.go           # Skill data structures
│   │   ├── parser.go          # SKILL.md parsing
│   │   ├── validator.go       # Skill validation
│   │   └── installer.go       # Installation logic
│   ├── command/
│   │   ├── command.go         # Command data structures
│   │   ├── parser.go          # Command parsing
│   │   └── installer.go       # Installation logic
│   ├── mcp/
│   │   ├── server.go          # MCP server config
│   │   ├── manager.go         # Server lifecycle
│   │   └── schema.go          # MCP config schema
│   ├── translate/
│   │   ├── variables.go       # Variable translation
│   │   └── toml.go            # YAML ↔ TOML conversion
│   └── git/
│       └── clone.go           # Git operations for installs
├── pkg/
│   └── frontmatter/
│       └── parser.go          # YAML frontmatter parsing
├── go.mod
├── go.sum
└── README.md
```

### Core Interfaces

```go
// internal/platform/platform.go

package platform

import "github.com/thoreinstein/aix/internal/skill"
import "github.com/thoreinstein/aix/internal/command"
import "github.com/thoreinstein/aix/internal/mcp"

// Platform defines the interface for AI coding assistant platforms
type Platform interface {
    // Identity
    Name() string
    DisplayName() string

    // Path resolution
    GlobalConfigDir() string
    ProjectConfigDir(projectRoot string) string
    SkillDir() string
    CommandDir() string
    MCPConfigPath() string
    InstructionsPath(projectRoot string) string

    // Installation
    InstallSkill(s *skill.Skill) error
    UninstallSkill(name string) error
    ListSkills() ([]*skill.Skill, error)

    InstallCommand(c *command.Command) error
    UninstallCommand(name string) error
    ListCommands() ([]*command.Command, error)

    // MCP management
    ConfigureMCP(server *mcp.Server) error
    RemoveMCP(name string) error
    ListMCP() ([]*mcp.Server, error)
    EnableMCP(name string) error
    DisableMCP(name string) error

    // Translation
    TranslateVariables(content string) string
    TranslateConfig(config map[string]any) ([]byte, error)

    // Validation
    ValidateSkill(s *skill.Skill) error
    ValidateCommand(c *command.Command) error

    // Status
    IsAvailable() bool
    Version() (string, error)
}
```

```go
// internal/skill/skill.go

package skill

import "time"

// Skill represents an Agent Skills Spec compliant skill
type Skill struct {
    // Metadata from frontmatter
    Name        string    `yaml:"name"`
    Description string    `yaml:"description"`
    Version     string    `yaml:"version"`
    Author      string    `yaml:"author"`
    License     string    `yaml:"license,omitempty"`
    Repository  string    `yaml:"repository,omitempty"`

    // Tool permissions
    Tools       []string  `yaml:"tools"`

    // Activation triggers
    Triggers    []string  `yaml:"triggers,omitempty"`

    // Dependencies
    Requires    []string  `yaml:"requires,omitempty"`

    // Content
    Instructions string   `yaml:"-"` // Markdown body

    // Installation metadata
    InstalledAt time.Time `yaml:"-"`
    SourcePath  string    `yaml:"-"`
    SourceType  string    `yaml:"-"` // local, git, registry
}

// ToolPermission represents a parsed tool permission
type ToolPermission struct {
    Tool   string   // Base tool name (Bash, Read, Write, etc.)
    Scopes []string // Optional scopes (git:*, npm:install, etc.)
}

// ParseToolPermissions parses tool permission strings into structured form
func ParseToolPermissions(tools []string) ([]ToolPermission, error) {
    // Implementation
}
```

```go
// internal/mcp/server.go

package mcp

// Transport defines MCP server transport type
type Transport string

const (
    TransportStdio Transport = "stdio"
    TransportSSE   Transport = "sse"
)

// Server represents an MCP server configuration
type Server struct {
    Name      string            `yaml:"name"`
    Command   string            `yaml:"command,omitempty"`   // For stdio
    Args      []string          `yaml:"args,omitempty"`
    URL       string            `yaml:"url,omitempty"`       // For SSE
    Transport Transport         `yaml:"transport"`
    Env       map[string]string `yaml:"env,omitempty"`
    Headers   map[string]string `yaml:"headers,omitempty"`   // For SSE

    // Platform enablement
    Platforms []string          `yaml:"platforms"` // ["all"], ["claude", "opencode"], etc.

    // Status
    Enabled   bool              `yaml:"-"`
}
```

### Configuration

```yaml
# ~/.config/aix/config.yaml

# Version for config migration
version: 1

# Default platforms for operations
default_platforms:
  - claude
  - opencode

# Platform-specific overrides
platforms:
  claude:
    global_config: ~/.claude
    instructions_file: CLAUDE.md
  opencode:
    global_config: ~/.config/opencode
    instructions_file: AGENTS.md
  codex:
    global_config: ~/.codex
    instructions_file: CODEX.md
  gemini:
    global_config: ~/.gemini
    instructions_file: GEMINI.md

# Skill registries (future)
registries:
  - name: official
    url: https://skills.agentskills.io
    priority: 100
  - name: community
    url: https://registry.example.com
    priority: 50

# Default skill settings
skills:
  # Where to install skills by default
  install_scope: global  # global or project

  # Default tool permissions for new skills
  default_tools:
    - Read
    - Glob
    - Grep

# Default MCP servers to enable for new projects
mcp:
  auto_enable:
    - context7
    - filesystem

# Instructions sync settings
instructions:
  source: AGENTS.md
  includes_enabled: true
  conditional_blocks: true
```

### Translation Layer

```go
// internal/translate/variables.go

package translate

import (
    "regexp"
    "strings"
)

// VariableStyle represents platform-specific variable syntax
type VariableStyle int

const (
    StyleDollar   VariableStyle = iota // $VARIABLE
    StyleMustache                       // {{variable}}
)

var (
    dollarPattern   = regexp.MustCompile(`\$([A-Z_]+)`)
    mustachePattern = regexp.MustCompile(`\{\{(\w+)\}\}`)
)

// VariableMap maps canonical names to platform-specific names
var VariableMap = map[string]map[VariableStyle]string{
    "ARGUMENTS": {
        StyleDollar:   "$ARGUMENTS",
        StyleMustache: "{{argument}}",
    },
    "SELECTION": {
        StyleDollar:   "$SELECTION",
        StyleMustache: "{{selection}}",
    },
}

// Translate converts content between variable styles
func Translate(content string, from, to VariableStyle) string {
    if from == to {
        return content
    }

    result := content
    for canonical, styles := range VariableMap {
        fromPattern := styles[from]
        toPattern := styles[to]
        result = strings.ReplaceAll(result, fromPattern, toPattern)
    }
    return result
}

// ToGemini converts $VARIABLE style to {{variable}} style
func ToGemini(content string) string {
    return Translate(content, StyleDollar, StyleMustache)
}

// FromGemini converts {{variable}} style to $VARIABLE style
func FromGemini(content string) string {
    return Translate(content, StyleMustache, StyleDollar)
}
```

```go
// internal/translate/toml.go

package translate

import (
    "github.com/pelletier/go-toml/v2"
    "gopkg.in/yaml.v3"
)

// YAMLToTOML converts YAML bytes to TOML bytes
func YAMLToTOML(yamlData []byte) ([]byte, error) {
    var data map[string]any
    if err := yaml.Unmarshal(yamlData, &data); err != nil {
        return nil, err
    }
    return toml.Marshal(data)
}

// TOMLToYAML converts TOML bytes to YAML bytes
func TOMLToYAML(tomlData []byte) ([]byte, error) {
    var data map[string]any
    if err := toml.Unmarshal(tomlData, &data); err != nil {
        return nil, err
    }
    return yaml.Marshal(data)
}
```

### Validation

```go
// internal/skill/validator.go

package skill

import (
    "errors"
    "fmt"
    "regexp"
)

var (
    ErrMissingName        = errors.New("skill name is required")
    ErrMissingDescription = errors.New("skill description is required")
    ErrInvalidToolSyntax  = errors.New("invalid tool permission syntax")
    ErrUnknownTool        = errors.New("unknown tool")
)

// KnownTools is the list of valid tool names
var KnownTools = map[string]bool{
    "Bash":     true,
    "Read":     true,
    "Write":    true,
    "Edit":     true,
    "Glob":     true,
    "Grep":     true,
    "WebFetch": true,
    "Task":     true,
    "TodoWrite": true,
    // Platform-specific tools
    "mcp":      true,  // MCP server access
}

var toolPattern = regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?$`)

// Validator validates skills against the spec
type Validator struct {
    strict bool
}

// NewValidator creates a new skill validator
func NewValidator(strict bool) *Validator {
    return &Validator{strict: strict}
}

// Validate validates a skill
func (v *Validator) Validate(s *Skill) []error {
    var errs []error

    // Required fields
    if s.Name == "" {
        errs = append(errs, ErrMissingName)
    }
    if s.Description == "" {
        errs = append(errs, ErrMissingDescription)
    }

    // Tool permissions
    for _, tool := range s.Tools {
        if err := v.validateTool(tool); err != nil {
            errs = append(errs, fmt.Errorf("tool %q: %w", tool, err))
        }
    }

    // Version format (semver)
    if s.Version != "" && !isValidSemver(s.Version) {
        errs = append(errs, fmt.Errorf("invalid version format: %s", s.Version))
    }

    return errs
}

func (v *Validator) validateTool(tool string) error {
    matches := toolPattern.FindStringSubmatch(tool)
    if matches == nil {
        return ErrInvalidToolSyntax
    }

    baseTool := matches[1]
    if !KnownTools[baseTool] {
        if v.strict {
            return fmt.Errorf("%w: %s", ErrUnknownTool, baseTool)
        }
    }

    return nil
}
```

## Alternatives Considered

### 1. Extend skillz (Python)

**Approach:** Fork and extend the existing Python-based skillz tool.

**Rejected because:**
- Python runtime dependency creates friction for installation
- Architecture doesn't support scoped tool permissions
- Would require significant rewrite to support multiple platforms
- No type safety for configuration handling

### 2. Platform-specific shell wrappers

**Approach:** Create shell scripts that wrap each platform's native tooling.

**Rejected because:**
- Not portable across operating systems
- Limited validation capabilities
- Complex cross-platform translation logic in shell is error-prone
- Poor developer experience (inconsistent interfaces)

### 3. Node.js CLI with TypeScript

**Approach:** Build the tool in TypeScript for the Node.js ecosystem.

**Rejected because:**
- Node.js runtime dependency
- Slower startup time compared to Go
- Many target users work in non-Node environments
- Go is more common in DevOps/CLI tooling

### 4. Rust CLI

**Approach:** Build in Rust for maximum performance and safety.

**Considered but deferred because:**
- Longer development time
- Smaller contributor pool
- Go is sufficient for this use case
- May revisit for v2 if performance issues arise

### 5. Configuration-only approach (no CLI)

**Approach:** Define a universal config format that platforms read directly.

**Rejected because:**
- Requires buy-in from all platform vendors
- No validation at write-time
- Doesn't solve the translation problem
- Less portable than a CLI tool

## Implementation Plan

### Phase 1: Foundation (Week 1-2)

```
Priority: P0
Goal: Basic CLI structure and config management

Tasks:
- [ ] Initialize Go module with proper structure
- [ ] Implement Cobra CLI scaffolding
- [ ] Create config file management (read/write/migrate)
- [ ] Define Platform interface
- [ ] Implement path resolution utilities
- [ ] Add basic logging and error handling
- [ ] Write unit tests for config management

Deliverable: `aix init`, `aix config`, `aix version`
```

### Phase 2: Claude & OpenCode Adapters (Week 3-4)

```
Priority: P0
Goal: Full support for the two most common platforms

Tasks:
- [ ] Implement ClaudePlatform adapter
- [ ] Implement OpenCodePlatform adapter
- [ ] Add platform detection logic
- [ ] Implement skill directory management
- [ ] Implement command directory management
- [ ] Add MCP config file management
- [ ] Write integration tests

Deliverable: Platform adapters with full read/write support
```

### Phase 3: Skills Management (Week 5-6)

```
Priority: P0
Goal: Complete skill lifecycle management

Tasks:
- [ ] Implement SKILL.md parser with frontmatter
- [ ] Implement tool permission parser and validator
- [ ] Add local path installation
- [ ] Add git repository installation (clone + extract)
- [ ] Implement skill removal
- [ ] Implement skill listing and inspection
- [ ] Add `aix skill init` scaffolding
- [ ] Write comprehensive validation tests

Deliverable: `aix skill install|list|remove|validate|init`
```

### Phase 4: Commands Management (Week 7)

```
Priority: P1
Goal: Command installation and management

Tasks:
- [ ] Implement command parser
- [ ] Add installation logic
- [ ] Add removal logic
- [ ] Implement listing
- [ ] Write tests

Deliverable: `aix cmd install|list|remove`
```

### Phase 5: MCP Server Management (Week 8-9)

```
Priority: P1
Goal: Full MCP server lifecycle management

Tasks:
- [ ] Implement MCP config schema
- [ ] Add server addition with env vars
- [ ] Add SSE transport support
- [ ] Implement enable/disable per platform
- [ ] Add server removal
- [ ] Implement listing with status
- [ ] Write integration tests

Deliverable: `aix mcp add|remove|enable|disable|list`
```

### Phase 6: Translation Layer (Week 10)

```
Priority: P1
Goal: Cross-platform content translation

Tasks:
- [ ] Implement variable translation ($ARGUMENTS ↔ {{argument}})
- [ ] Implement YAML ↔ TOML conversion
- [ ] Add translation to all installation paths
- [ ] Test translation round-trips
- [ ] Document translation behavior

Deliverable: Seamless cross-platform content handling
```

### Phase 7: Instructions Sync (Week 11)

```
Priority: P2
Goal: Synchronize instruction files across platforms

Tasks:
- [ ] Implement instructions parser with includes
- [ ] Add conditional block support
- [ ] Implement sync logic with diff detection
- [ ] Add dry-run mode
- [ ] Implement validation
- [ ] Write tests

Deliverable: `aix instructions sync|validate|diff`
```

### Phase 8: Codex & Gemini Adapters (Week 12-13)

```
Priority: P2
Goal: Extended platform support

Tasks:
- [ ] Research Codex config format and MCP support
- [ ] Implement CodexPlatform adapter
- [ ] Implement GeminiPlatform adapter with TOML
- [ ] Add Gemini variable translation
- [ ] Test cross-platform workflows
- [ ] Update documentation

Deliverable: Full four-platform support
```

### Phase 9: Polish & Release (Week 14)

```
Priority: P1
Goal: Production-ready release

Tasks:
- [ ] Implement `aix doctor` diagnostics
- [ ] Implement `aix status` overview
- [ ] Add shell completions (bash, zsh, fish)
- [ ] Write user documentation
- [ ] Create installation scripts
- [ ] Set up CI/CD for releases
- [ ] Create Homebrew formula
- [ ] Publish v1.0.0

Deliverable: Public v1.0.0 release
```

### Phase 10: Registry Support (Future)

```
Priority: P3
Goal: Skill registry ecosystem

Tasks:
- [ ] Design registry API specification
- [ ] Implement registry client
- [ ] Add search functionality
- [ ] Add version resolution
- [ ] Implement dependency resolution
- [ ] Build reference registry server

Deliverable: `aix skill search|install @scope/name`
```

## Consequences

### Positive

1. **Unified Experience**: Developers learn one tool for all platforms
2. **Reduced Errors**: Strong validation prevents misconfiguration
3. **Portability**: Single Go binary runs everywhere with no dependencies
4. **Consistency**: Same skills/commands work across platforms
5. **Spec Compliance**: First-class Agent Skills Spec and MCP support
6. **Maintainability**: Adding new platforms requires only a new adapter
7. **Automation**: Scriptable for CI/CD and team onboarding

### Negative

1. **Maintenance Burden**: Must track changes across four+ platforms
2. **Translation Complexity**: Gemini's different syntax requires special handling
3. **Platform Lag**: New platform features require aix updates
4. **Initial Investment**: Significant upfront development time
5. **Testing Matrix**: Must test all combinations of platforms × operations

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Platforms diverge significantly | Medium | High | Abstract through adapters, track upstream |
| MCP spec changes incompatibly | Low | Medium | Version MCP support, maintain compatibility |
| Low community adoption | Medium | Medium | Focus on solving real pain points first |
| New dominant platform emerges | Low | Low | Adapter pattern allows quick addition |
| Platform vendor builds competing tool | Medium | High | Differentiate through multi-platform support |

## Success Metrics

1. **Adoption**: 100+ GitHub stars within 6 months
2. **Coverage**: Support for 4 major platforms
3. **Reliability**: <1% error rate on installations
4. **Performance**: <100ms for common operations
5. **Satisfaction**: >4 star average rating in user feedback

## Open Questions

1. **Registry governance**: Who can publish to the official registry?
2. **Skill versioning**: How to handle breaking changes in skills?
3. **Platform authentication**: Do any platforms require auth for config access?
4. **Conflict resolution**: How to handle conflicting skills with same triggers?
5. **Codex MCP**: Does Codex support MCP? If not, when?

## References

- [Agent Skills Specification](https://agentskills.io/specification)
- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [Claude Code Documentation](https://docs.anthropic.com/claude-code)
- [OpenCode Repository](https://github.com/opencode-ai/opencode)
- [Gemini CLI Repository](https://github.com/google/gemini-cli)
- [Codex Documentation](https://platform.openai.com/docs/guides/codex)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Semantic Versioning](https://semver.org/)

---

## Appendix A: Example Workflows

### Installing a Skill Across All Platforms

```bash
# Clone and install a git-based skill
$ aix skill install github.com/example/code-review-skill

Installing skill: code-review
  Source: github.com/example/code-review-skill
  Version: 1.2.0

Validating skill...
  ✓ Required fields present
  ✓ Tool permissions valid: Bash(git:*), Read, Glob, Grep
  ✓ SKILL.md structure valid

Installing to platforms:
  ✓ claude: ~/.config/opencode/skill/code-review/
  ✓ opencode: ~/.config/opencode/skill/code-review/

Skill 'code-review' installed successfully to 2 platforms.
Triggers: /review, /cr
```

### Configuring an MCP Server

```bash
# Add a new MCP server
$ aix mcp add context7 npx -y @upstash/context7-mcp \
    --env CONTEXT7_API_KEY='$CONTEXT7_API_KEY'

Added MCP server: context7
  Command: npx -y @upstash/context7-mcp
  Transport: stdio
  Environment: CONTEXT7_API_KEY (from env)

# Enable for specific platforms
$ aix mcp enable context7 --platform=claude,opencode

Enabled context7 on:
  ✓ claude: Updated ~/.claude/mcp_servers.json
  ✓ opencode: Updated ~/.config/opencode/opencode.json
```

### Syncing Instructions

```bash
# Sync from source of truth to all platforms
$ aix instructions sync --dry-run

Source: AGENTS.md (project root)

Changes to apply:
  claude (CLAUDE.md):
    + 12 lines (includes expanded)
    ~ 3 lines (platform conditionals resolved)

  gemini (GEMINI.md):
    + 12 lines (includes expanded)
    ~ 5 lines (variables translated: $ARGUMENTS → {{argument}})

Run without --dry-run to apply changes.
```

## Appendix B: Configuration File Formats

### Claude MCP Config (`~/.claude/mcp_servers.json`)

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"],
      "env": {
        "CONTEXT7_API_KEY": "key-here"
      }
    },
    "filesystem": {
      "command": "/usr/local/bin/mcp-filesystem",
      "args": ["--root", "/home/user/projects"]
    }
  }
}
```

### OpenCode Config (`~/.config/opencode/opencode.json`)

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"],
      "env": {
        "CONTEXT7_API_KEY": "key-here"
      }
    }
  }
}
```

### Gemini Config (`~/.gemini/settings.toml`)

```toml
[mcp.servers.context7]
command = "npx"
args = ["-y", "@upstash/context7-mcp"]

[mcp.servers.context7.env]
CONTEXT7_API_KEY = "key-here"
```

## Appendix C: Tool Permission Reference

| Permission | Description | Example Scopes |
|------------|-------------|----------------|
| `Read` | Read file contents | `Read(src/**)`, `Read(*.go)` |
| `Write` | Create/overwrite files | `Write(*.md)`, `Write(tests/**)` |
| `Edit` | Modify existing files | `Edit(src/**)` |
| `Bash` | Execute shell commands | `Bash(git:*)`, `Bash(npm:install)`, `Bash(make:*)` |
| `Glob` | Search for files | (no scopes) |
| `Grep` | Search file contents | (no scopes) |
| `WebFetch` | Fetch web content | `WebFetch(docs.*)`, `WebFetch(api.github.com)` |
| `Task` | Spawn sub-agents | (no scopes) |
| `TodoWrite` | Manage task lists | (no scopes) |

### Bash Scope Syntax

```
Bash(command:subcommand)
Bash(command:*)          # All subcommands
Bash(git:status)         # Specific: git status only
Bash(git:diff)           # Specific: git diff only  
Bash(git:*)              # All git commands
Bash(npm:install)        # npm install only
Bash(npm:*)              # All npm commands
```
