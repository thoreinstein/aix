# Testing Guide

## Overview

aix uses Go's standard testing package with table-driven tests. All tests are unit tests located alongside their source files.

## Coverage Requirements

| Metric | Target | Notes |
|--------|--------|-------|
| Project coverage | 60% | Overall codebase minimum |
| Patch coverage | 70% | New code must meet this threshold |
| Threshold | 2% | Variance allowed for minor fluctuations |

Coverage is enforced via Codecov on every pull request. Targets will be increased as coverage improves.

## Test Organization

| Directory | Type | Description |
|-----------|------|-------------|
| `cmd/aix/commands/*_test.go` | Unit | CLI command tests |
| `internal/*_test.go` | Unit | Core logic tests |
| `pkg/*_test.go` | Unit | Public package tests |

All tests are unit tests. There are no separate integration or end-to-end test directories.

## Running Tests

### Quick Test Run

```bash
go test ./...
```

### With Race Detection

```bash
go test -race ./...
```

### With Coverage

```bash
go test -cover ./...                    # Summary per package
go test -coverprofile=c.out ./...       # Generate profile
go tool cover -html=c.out               # View HTML report
```

### Single Package

```bash
go test ./internal/skill/...
```

### Single Test

```bash
go test -run TestSkillParser ./...
```

### Single Subtest

```bash
go test -v -run TestSkillParser/valid ./...
```

## Test Patterns

### Table-Driven Tests

The standard pattern for tests in this project:

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

Use `t.Helper()` for reusable test setup functions:

```go
func setupTestDir(t *testing.T) string {
    t.Helper()
    dir := t.TempDir()
    // ... setup
    return dir
}
```

### Mocking

Mock implementations are in `mock_*_test.go` files alongside the tests that use them. Example: `mock_platform_test.go` provides a mock `Platform` interface for testing command handlers.

## CI Integration

Tests run automatically on:
- Every push to `main`
- Every pull request targeting `main`

The CI pipeline runs in this order:
1. **Lint** - golangci-lint with 26+ enabled linters
2. **Test** - `go test -v -race` with coverage collection
3. **Build** - Binary compilation and smoke test

Coverage is uploaded to Codecov after tests pass. PR comments show coverage diff.

## Best Practices

1. **Use `t.TempDir()`** for filesystem tests - automatically cleaned up
2. **Use `t.Setenv()`** for environment variable tests - automatically restored
3. **Use `t.Parallel()`** where safe - speeds up test execution
4. **Name tests descriptively** - `Test<Type>_<Method>` or `Test<Function>_<Scenario>`
5. **Keep tests focused** - One assertion per test case when possible
