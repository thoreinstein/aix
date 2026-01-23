# aix Project Context

## Overview
`aix` is a unified CLI tool for managing AI coding assistant configurations across multiple platforms (Claude Code, OpenCode, Codex, Gemini CLI). It allows developers to define skills, MCP servers, and commands once and deploy them to all supported AI assistants.

## Tech Stack
- **Language:** Go 1.25+
- **CLI Framework:** Cobra + Viper
- **Configuration:** YAML (users), TOML/JSON (platforms)
- **Linter:** golangci-lint
- **Build:** Go standard toolchain + Goreleaser

## Architecture
The project follows a Clean Architecture approach with a strong separation of concerns:

- **`cmd/aix/`**: CLI entry points and command definitions.
- **`internal/cli/`**: Core CLI logic and `Platform` interface definition.
- **`internal/platform/`**: Adapter implementations for each AI assistant (claude, opencode, etc.).
- **`internal/skill/`**: Logic for parsing and validating Agent Skills.
- **`internal/mcp/`**: Logic for managing Model Context Protocol servers.

### Key Concepts
- **Platform Adapter:** Each AI assistant is implemented as a plugin-like adapter satisfying the `Platform` interface.
- **Unified Config:** User configuration lives in `~/.config/aix/config.yaml`.
- **Translation Layer:** The CLI translates abstract definitions (e.g., `$ARGUMENTS`) into platform-specific syntax (e.g., `{{argument}}`).

## Development Workflow

### Build and Run
```bash
# Build the binary
go build ./cmd/aix

# Install locally
go install ./cmd/aix
```

### Testing
Tests are standard Go unit tests, emphasizing table-driven design.
```bash
# Run all tests
go test ./...

# Run with race detection (Recommended)
go test -race ./...

# Run specific package
go test ./internal/skill/...
```

### Linting & Formatting
Strict linting rules are enforced via `.golangci.yml`.
```bash
# Format code
go fmt ./...
goimports -w .

# Run linters
golangci-lint run
```

## Coding Conventions
- **Imports:** Grouped as Standard Library, External (3rd party), and Internal (project).
- **Error Handling:** Always wrap errors with `fmt.Errorf("...: %w", err)` to preserve context and stack traces. Use `errors.Is` for checking.
- **Interfaces:** Define interfaces where they are *used*, not where they are implemented.
- **Testing:**
  - Use `t.Parallel()` for independent tests.
  - Use `t.TempDir()` for file system tests.
  - Use `t.Setenv()` for environment variables.
  - Prefer table-driven tests for logic with multiple edge cases.

## Documentation
- **ADRs:** Architecture Decision Records are in `docs/adr/`. Read `docs/adr/001-unified-agent-cli.md` for deep architectural context.
- **User Docs:** See `README.md` and `docs/` folder.
