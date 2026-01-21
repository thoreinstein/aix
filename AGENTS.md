# Agent Instructions for aix

`aix` is a unified Go CLI for managing AI coding assistant configurations across platforms (Claude Code, OpenCode, Codex, Gemini CLI). See `docs/adr/001-unified-agent-cli.md` for full architecture.

## Development Commands

### Build

```bash
go build ./cmd/aix                      # Build CLI binary
go build ./...                          # Build all packages
```

### Test

```bash
go test ./...                           # All tests
go test ./internal/skill/...            # Single package
go test -run TestSkillParser ./...      # Single test by name
go test -v -run TestSkillParser/valid ./...  # Subtest by name
go test -race ./...                     # With race detection
go test -cover ./...                    # With coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### Lint & Format

```bash
go fmt ./...                            # Format code
goimports -w .                          # Organize imports
go vet ./...                            # Static analysis
golangci-lint run                       # Full linting suite
```

## Code Style

### Imports

Group imports in this order, separated by blank lines:

```go
import (
    // Standard library
    "context"
    "fmt"
    "strings"

    // External dependencies
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"

    // Internal packages
    "github.com/thoreinstein/aix/internal/config"
    "github.com/thoreinstein/aix/internal/platform"
)
```

### Naming Conventions

| Element        | Convention             | Example                                  |
| -------------- | ---------------------- | ---------------------------------------- |
| Packages       | lowercase, single word | `platform`, `skill`, `mcp`               |
| Interfaces     | noun or -er suffix     | `Platform`, `Validator`, `Parser`        |
| Constructors   | `New` prefix           | `NewValidator(strict bool)`              |
| Errors         | `Err` prefix           | `ErrMissingName`, `ErrInvalidToolSyntax` |
| Test files     | `_test.go` suffix      | `validator_test.go`                      |
| Test functions | `Test` prefix          | `TestValidator_Validate`                 |

### Error Handling

Wrap errors with context using `%w`:

```go
// Good - adds context while preserving the error chain
if err := s.loadConfig(); err != nil {
    return fmt.Errorf("loading skill %s: %w", name, err)
}

// Define sentinel errors for expected conditions
var (
    ErrMissingName = errors.New("skill name is required")
    ErrNotFound    = errors.New("resource not found")
)

// Return early on errors
func (p *Parser) Parse(path string) (*Skill, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("reading skill file: %w", err)
    }
    // ... continue processing
}
```

### Types and Structs

```go
// Use struct tags for serialization
type Skill struct {
    Name        string   `yaml:"name" json:"name"`
    Description string   `yaml:"description" json:"description"`
    Tools       []string `yaml:"tools" json:"tools"`
}

// Define interfaces in the package that uses them, not the implementer
// internal/installer/installer.go
type Platform interface {
    InstallSkill(s *skill.Skill) error
    UninstallSkill(name string) error
}
```

### Comments

```go
// Package skill provides Agent Skills Spec compliant skill management.
package skill

// Validator validates skills against the Agent Skills Specification.
// It supports both strict and lenient validation modes.
type Validator struct {
    strict bool
}

// Validate checks a skill for spec compliance.
// Returns a slice of validation errors, or nil if valid.
func (v *Validator) Validate(s *Skill) []error {
    // Implementation
}
```

## Architecture

### Package Layout

```
aix/
├── cmd/aix/              # CLI entry point and Cobra commands
│   ├── main.go
│   └── commands/         # Subcommand implementations
├── internal/             # Private packages
│   ├── config/           # Configuration management
│   ├── platform/         # Platform adapters (claude, opencode, etc.)
│   ├── skill/            # Skill parsing and validation
│   ├── command/          # Slash command management
│   ├── mcp/              # MCP server configuration
│   └── translate/        # Cross-platform translation
└── pkg/                  # Public packages (if any)
    └── frontmatter/      # YAML frontmatter parser
```

### Key Interfaces

```go
// Platform adapter interface - each AI assistant implements this
type Platform interface {
    Name() string
    InstallSkill(s *skill.Skill) error
    ConfigureMCP(server *mcp.Server) error
    TranslateVariables(content string) string
    // ... see ADR for full interface
}
```

### Patterns

- **Adapter pattern**: Each platform (Claude, OpenCode, Codex, Gemini) has its own adapter implementing the `Platform` interface
- **Translation layer**: Handles variable syntax differences (`$ARGUMENTS` vs `{{argument}}`) and config format differences (JSON vs TOML)
- **Validation**: Strong upfront validation prevents misconfiguration

## Testing

### Table-Driven Tests

```go
func TestValidator_Validate(t *testing.T) {
    tests := []struct {
        name    string
        skill   *Skill
        wantErr bool
    }{
        {
            name:    "valid skill",
            skill:   &Skill{Name: "test", Description: "A test skill"},
            wantErr: false,
        },
        {
            name:    "missing name",
            skill:   &Skill{Description: "A test skill"},
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            v := NewValidator(true)
            errs := v.Validate(tt.skill)
            if (len(errs) > 0) != tt.wantErr {
                t.Errorf("Validate() errors = %v, wantErr %v", errs, tt.wantErr)
            }
        })
    }
}
```

### Test Helpers

```go
func setupTestDir(t *testing.T) string {
    t.Helper()
    dir := t.TempDir()
    // ... setup
    return dir
}
```

---

## Issue Tracking

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

---

## Session Completion

**When ending a work session**, complete ALL steps below. Work is NOT complete until `git push` succeeds.

### Mandatory Workflow

1. **File issues** - Create issues for remaining/discovered work
2. **Quality gates** (if code changed):
   ```bash
   go test ./...
   golangci-lint run
   go build ./cmd/aix
   pre-commit run --all-files
   ```
3. **Update issues** - Close finished work, update in-progress items
4. **Push to remote**:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Hand off** - Provide context for next session

### Critical Rules

- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- If push fails, resolve and retry until it succeeds
- NEVER use git commit --no-verify
- NEVER use git commit --amend
