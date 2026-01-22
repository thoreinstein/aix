# Copilot Instructions for aix

## Repository Overview

**aix** is a unified Go CLI for managing AI coding assistant configurations across multiple platforms (Claude Code, OpenCode, Codex CLI, Gemini CLI). It provides a platform-agnostic way to define and deploy skills, MCP servers, and slash commands.

- **Size**: ~1.2MB, 59 Go source files
- **Language**: Go 1.25.5
- **Architecture**: CLI tool using Cobra, adapter pattern for platform-specific implementations
- **Key Dependencies**: spf13/cobra, spf13/viper, adrg/frontmatter, gopkg.in/yaml.v3

## Repository Structure

```
aix/
├── cmd/aix/              # CLI entry point - main.go and commands/
├── internal/             # Private packages (NOT importable externally)
│   ├── config/           # Configuration management
│   ├── errors/           # Error handling utilities
│   ├── logging/          # Logging utilities
│   ├── paths/            # Path resolution for platform configs
│   └── platform/         # Platform adapters (claude/, opencode/)
├── pkg/                  # Public packages (importable)
│   └── frontmatter/      # YAML frontmatter parser
├── docs/adr/             # Architecture Decision Records (see ADR-001)
├── .github/workflows/    # CI/CD pipelines (ci.yml, codeql.yml, release.yml)
├── .golangci.yml         # Linter configuration
├── .pre-commit-config.yaml  # Pre-commit hooks (requires local tools)
└── .beads/               # Issue tracker (bd CLI)
```

## Build, Test, and Lint Commands

### Prerequisites
- Go 1.25.5+ (specified in go.mod)
- Dependencies are automatically downloaded on first build

### Build
```bash
go build ./cmd/aix              # Build CLI binary -> ./aix
go build ./...                  # Build all packages (verification only)
go build -v -o aix cmd/aix/main.go  # Verbose build with custom output name
```

**First build downloads dependencies** (~30 seconds). Subsequent builds are fast (<5 seconds).

### Test
```bash
go test ./...                   # Run all tests (ALWAYS use this)
go test -v -race ./...          # Verbose with race detection (CI requirement)
go test -cover ./...            # With coverage report
go test ./internal/platform/... # Test specific package
```

**Important**:
- Tests pass with exit code 0 and show "ok" for each package
- The "go: no such tool 'covdata'" warning in coverage output is harmless
- Always run `go test ./...` before committing code changes

### Lint and Format
```bash
go fmt ./...                    # Format code (ALWAYS run before commit)
go vet ./...                    # Static analysis
```

**Note**: `golangci-lint` is NOT available in local dev environments. The CI pipeline builds it from source:
```bash
# CI uses golangci-lint v2.7.2 (built from source in CI)
# Local developers: go vet ./... is sufficient for basic checks
# Full linting happens in CI automatically
```

### Running the Binary
```bash
./aix --help                    # Show help
./aix version                   # Show version
./aix init                      # Initialize configuration
```

## CI/CD Pipeline

### Workflows (`.github/workflows/`)

1. **ci.yml** - Runs on push/PR to main
   - **Lint**: Builds golangci-lint v2.7.2 from source, runs with 5m timeout
   - **Test**: `go test -v -race ./...` (requires lint to pass first)
   - **Build**: Builds binary and runs `./aix --help` verification

2. **codeql.yml** - Security scanning (runs on push/PR and weekly)
   - CodeQL analysis for Go code

3. **release.yml** - Release automation (on tag push)
   - Re-runs lint and test checks
   - Uses GoReleaser to build multi-platform binaries
   - **Important**: Tags must be on main branch, not feature branches

### CI Success Criteria
To match CI behavior locally:
```bash
go fmt ./...                    # Format
go vet ./...                    # Static analysis
go test -v -race ./...          # Tests with race detection
go build -v -o aix cmd/aix/main.go  # Build
./aix --help                    # Verify binary works
```

## Code Style and Conventions

### Import Grouping
**ALWAYS** use this order with blank lines between groups:
```go
import (
    // Standard library
    "context"
    "fmt"

    // External dependencies
    "github.com/spf13/cobra"

    // Internal packages
    "github.com/thoreinstein/aix/internal/config"
)
```

### Naming Conventions
- **Packages**: lowercase, single word (skill, platform, mcp)
- **Interfaces**: noun or -er suffix (Platform, Validator, Parser)
- **Constructors**: `New` prefix (NewValidator, NewConfig)
- **Errors**: `Err` prefix (ErrMissingName, ErrNotFound)
- **Test files**: `*_test.go` suffix
- **Test functions**: `Test*` prefix, table-driven pattern preferred

### Error Handling
**ALWAYS** wrap errors with context using `%w`:
```go
if err := doSomething(); err != nil {
    return fmt.Errorf("doing something: %w", err)
}
```

### Testing Pattern
Use table-driven tests (see existing tests for examples):
```go
tests := []struct {
    name    string
    input   string
    want    bool
    wantErr bool
}{
    {name: "valid case", input: "test", want: true, wantErr: false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test implementation
    })
}
```

## Common Issues and Workarounds

### Issue: "go: no such tool 'covdata'" during coverage
**Status**: Harmless warning, tests still pass. Not a failure.

### Issue: golangci-lint not available locally
**Solution**: Use `go vet ./...` for local checks. Full linting runs in CI automatically.

### Issue: pre-commit hooks reference unavailable tools (bd, golangci-lint)
**Solution**: These are optional development tools. CI does not use pre-commit hooks. Manual commands work fine.

### Issue: Building with -buildvcs=false flag
**Context**: Used in .pre-commit-config.yaml to support git worktrees. Standard builds don't need this flag.

## Development Workflow

1. **Make Changes**: Edit Go files in cmd/, internal/, or pkg/
2. **Format**: `go fmt ./...`
3. **Verify**: `go vet ./...`
4. **Test**: `go test ./...`
5. **Build**: `go build ./cmd/aix`
6. **Run**: `./aix --help` to verify binary works

## Key Configuration Files

- **go.mod**: Dependencies and Go version
- **.golangci.yml**: Linter configuration (26 enabled linters, specific exclusions for G102, G115, G306, G402, G404)
- **.gitignore**: Excludes binaries (aix, *.exe), test artifacts (*.out, *.test), build artifacts (dist/), debug files
- **AGENTS.md**: Contains detailed developer instructions (overlaps with this file but more comprehensive)
- **docs/adr/001-unified-agent-cli.md**: Full architecture rationale (34KB, detailed design decisions)

## Additional Resources

- **ADR-001**: See `docs/adr/001-unified-agent-cli.md` for complete architecture and design rationale
- **AGENTS.md**: More detailed development guide with code style examples and issue tracking workflow (bd CLI)
- **README.md**: User-facing documentation for installation and basic usage

## Trust These Instructions

These instructions are validated by running actual commands in the repository. If you encounter issues not documented here, search the codebase or workflows for updates. Otherwise, trust these commands - they work.
