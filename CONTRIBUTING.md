# Contributing to aix

Thank you for your interest in contributing to `aix`! This document provides guidelines and instructions for contributing to the project.

`aix` is a unified Go CLI for managing AI coding assistant configurations across platforms (Claude Code, OpenCode, Codex, Gemini CLI). Whether you're fixing a bug, adding a feature, or improving documentation, your contributions are welcome.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you agree to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

   ```bash
   git clone git@github.com:YOUR_USERNAME/aix.git
   cd aix
   ```

3. Add the upstream remote:

   ```bash
   git remote add upstream git@github.com:thoreinstein/aix.git
   ```

### Create a Branch

Create a feature branch from `main`:

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

Use descriptive branch names:

- `feature/add-gemini-support`
- `fix/skill-validation-error`
- `docs/improve-mcp-reference`

## Development Environment Setup

### Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.22+ (1.25.x recommended) | Core language |
| golangci-lint | v2.7+ | Linting |
| pre-commit | 3.2+ | Git hooks |
| git | 2.x | Version control |
| GPG | Any | Commit signing |

### Install Dependencies

1. **Install Go** from [go.dev/dl](https://go.dev/dl/) or via your package manager

2. **Install golangci-lint**:

   ```bash
   # macOS
   brew install golangci-lint

   # Linux
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   ```

3. **Install pre-commit**:

   ```bash
   # macOS
   brew install pre-commit

   # pip
   pip install pre-commit
   ```

4. **Set up pre-commit hooks**:

   ```bash
   pre-commit install
   ```

5. **Verify your setup**:

   ```bash
   go version          # Should show 1.22+
   golangci-lint --version
   pre-commit --version
   ```

### GPG Signing Setup

All commits must be GPG signed. Set up signing:

```bash
# Configure Git to sign commits
git config --global commit.gpgsign true
git config --global user.signingkey YOUR_GPG_KEY_ID

# Verify your key is available
gpg --list-secret-keys --keyid-format LONG
```

## Building and Testing

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
```

### Generate Coverage Report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out        # Opens HTML report in browser
```

### Lint

```bash
go fmt ./...                            # Format code
golangci-lint run                       # Full linting suite
golangci-lint run --fix                 # Auto-fix where possible
```

### Pre-commit Checks

Run all pre-commit hooks manually:

```bash
pre-commit run --all-files
```

The pre-commit hooks include:

- `golangci-lint` - Lint Go code
- `golangci-lint fmt` - Format Go code
- `go-mod-tidy` - Clean up go.mod/go.sum
- `go-build` - Verify compilation
- `trailing-whitespace` - Remove trailing whitespace
- `end-of-file-fixer` - Ensure files end with newline
- `check-yaml` - Validate YAML syntax
- `actionlint` - Lint GitHub Actions workflows
- `zizmor` - Security scan GitHub Actions

## Code Style Guidelines

This project follows idiomatic Go conventions. For complete details, see [`AGENTS.md`](AGENTS.md).

### Import Organization

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

| Element | Convention | Example |
|---------|------------|---------|
| Packages | lowercase, single word | `platform`, `skill`, `mcp` |
| Interfaces | noun or -er suffix | `Platform`, `Validator`, `Parser` |
| Constructors | `New` prefix | `NewValidator(strict bool)` |
| Errors | `Err` prefix | `ErrMissingName`, `ErrInvalidToolSyntax` |
| Test files | `_test.go` suffix | `validator_test.go` |
| Test functions | `Test` prefix | `TestValidator_Validate` |

### Error Handling

Use the `internal/errors` package for consistent error handling:

```go
import "github.com/thoreinstein/aix/internal/errors"

// Add context while preserving the error chain
if err := s.loadConfig(); err != nil {
    return errors.Wrapf(err, "loading skill %s", name)
}

// Check for sentinel errors
if errors.Is(err, errors.ErrNotFound) {
    // handle not found case
}
```

### Testing Pattern

Use table-driven tests:

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

### Coverage Requirements

| Metric | Target |
|--------|--------|
| Project coverage | 60% minimum |
| Patch coverage | 70% for new code |

Coverage is enforced via Codecov on pull requests.

## Commit Message Guidelines

### Format

```
<Subject line: capital verb, 50 chars max, no period>

<Body: wrapped at 72 chars, explains WHY not just WHAT>
```

### Subject Line Rules

- Start with a capital verb (Add, Fix, Update, Remove, Refactor)
- Maximum 50 characters
- No period at the end
- Use imperative mood ("Add feature" not "Added feature")

### Body Rules

- Blank line between subject and body
- Wrap at 72 characters
- Explain **why** the change is being made, not just what changed
- Reference issues when applicable

### Examples

Good:

```
Add platform detection for Gemini CLI

Gemini CLI uses a different configuration format than other platforms.
This change adds detection logic to identify Gemini installations and
route configuration appropriately.

Closes #42
```

Bad:

```
added gemini support.
```

### Signing

All commits must be GPG signed. The pre-commit hooks will fail if commits are unsigned.

```bash
# Commits are automatically signed if configured
git commit -m "Add new feature"

# Or explicitly sign
git commit -S -m "Add new feature"
```

## Submitting Changes

### Before You Start

1. **Check existing issues** - Someone may already be working on this
2. **Open an issue first** for significant changes to discuss the approach
3. **Small PRs are better** - Break large changes into smaller, reviewable pieces

### Pull Request Process

1. **Ensure your code passes all checks**:

   ```bash
   go test ./...
   golangci-lint run
   go build ./cmd/aix
   pre-commit run --all-files
   ```

2. **Update documentation** if your change affects user-facing behavior

3. **Push your branch**:

   ```bash
   git push origin feature/your-feature-name
   ```

4. **Open a pull request** against `main` with:
   - Clear title describing the change
   - Description of what and why
   - Link to related issue(s)
   - Screenshots for UI changes (if applicable)

5. **Respond to review feedback** promptly

### Review Expectations

- All PRs require at least one approval
- CI must pass (lint, test, build)
- Coverage must meet thresholds
- Commits should be well-organized

### What to Expect

- Initial response within a few days
- Constructive feedback focused on code quality
- Discussion of alternative approaches when relevant

## Issue Tracking with Beads

This project uses [beads](https://github.com/steveyegge/beads) (`bd`) for issue tracking instead of GitHub Issues.

### Getting Started

```bash
bd onboard                              # First-time setup
```

### Finding Work

```bash
bd ready                                # Show issues ready to work on
bd list --status=open                   # All open issues
bd show <id>                            # Detailed issue view
```

### Working on Issues

```bash
bd update <id> --status in_progress     # Claim work
# ... do the work ...
bd close <id>                           # Mark complete
```

### Creating Issues

```bash
bd create --title="Fix validation bug" --type=bug --priority=2
bd create --title="Add Gemini support" --type=feature --priority=1
```

Priority levels: 0 (critical) to 4 (backlog)

### Syncing

The beads daemon auto-syncs changes. Manual sync if needed:

```bash
bd sync
```

## Release Process

Releases are automated via GitHub Actions and GoReleaser.

### Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes to CLI interface or configuration format
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Creating a Release

Releases are created by maintainers:

1. Tag the release: `git tag v1.2.3`
2. Push the tag: `git push origin v1.2.3`
3. GitHub Actions builds and publishes binaries

### Pre-release Testing

Before tagging a release:

```bash
go test ./...
golangci-lint run
goreleaser release --snapshot --clean   # Test release build
```

## Getting Help

### Documentation

- [`AGENTS.md`](AGENTS.md) - Full development guidelines
- [`docs/`](docs/) - Reference documentation
- [`docs/adr/`](docs/adr/) - Architecture Decision Records

### Communication

- **GitHub Discussions** - Questions and general discussion
- **GitHub Issues** - Bug reports (use beads for tracking work)

### Useful Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [golangci-lint](https://golangci-lint.run/)
- [pre-commit](https://pre-commit.com/)

## Thank You

Your contributions make `aix` better for everyone. We appreciate your time and effort!
