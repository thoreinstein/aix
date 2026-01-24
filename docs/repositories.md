# Repository Management Reference

This document describes how `aix` manages remote Git repositories containing shareable resources.

## Overview

Repositories are remote Git repos containing shareable aix resources—skills, commands, agents, and MCP configurations. They enable teams to distribute curated tooling across projects and share best practices through versioned, installable packages.

### Design Goals

1. **Discoverability** - Browse and search resources across multiple repositories
2. **Version control** - Resources are tracked via Git, enabling updates and rollbacks
3. **Isolation** - Each repository is cloned independently, avoiding conflicts
4. **Simplicity** - Standard Git semantics for fetching and updating

## Repository Structure

A valid aix repository contains one or more resource directories:

```
my-aix-repo/
├── skills/           # Skill definitions
│   ├── code-review/
│   │   └── SKILL.md
│   └── security-audit/
│       └── SKILL.md
├── commands/         # Slash command definitions
│   ├── deploy/
│   │   └── command.md
│   └── test-coverage/
│       └── command.md
├── agents/           # Agent definitions
│   ├── go-expert.md
│   └── security-reviewer.md
└── mcp/              # MCP server configurations
    ├── github.json
    └── postgres.json
```

### Directory Purposes

| Directory | Contents | File Format |
|-----------|----------|-------------|
| `skills/` | Skill definitions with prompts and tool configurations | `SKILL.md` in named subdirectory |
| `commands/` | Slash command definitions | `command.md` in named subdirectory |
| `agents/` | Agent definitions with instructions | `{name}.md` files |
| `mcp/` | MCP server configurations | JSON files |

All directories are optional. A repository may contain only skills, only agents, or any combination.

## Configuration

Repositories are tracked in the aix configuration file. Each registered repository has the following fields:

### RepoConfig Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Unique identifier for the repository |
| `url` | `string` | Yes | Git remote URL (HTTPS, SSH, or git protocol) |
| `path` | `string` | Yes | Local filesystem path where the repository is cloned |
| `added_at` | `time.Time` | Yes | Timestamp when the repository was registered |

### Configuration Example

```json
{
  "repositories": [
    {
      "name": "company-tools",
      "url": "https://github.com/acme/aix-resources.git",
      "path": "/Users/dev/.aix/repos/company-tools",
      "added_at": "2024-01-15T10:30:00Z"
    },
    {
      "name": "personal",
      "url": "git@github.com:user/my-aix-tools.git",
      "path": "/Users/dev/.aix/repos/personal",
      "added_at": "2024-02-20T14:45:00Z"
    }
  ]
}
```

## Commands

### Add a Repository

```bash
aix repo add <url>
```

Clones the repository and registers it for use. The repository name is derived from the URL (e.g., `github.com/acme/tools` becomes `tools`).

**Flags:**

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--name` | `-n` | `string` | Override the derived repository name |

**Examples:**

```bash
# Add from HTTPS URL
aix repo add https://github.com/acme/aix-resources.git

# Add from SSH URL with custom name
aix repo add git@github.com:user/tools.git --name personal-tools

# Add from git protocol
aix repo add git://github.com/org/shared-skills.git
```

### List Repositories

```bash
aix repo list
```

Displays all registered repositories with their URLs and local paths.

**Output:**

```
NAME            URL                                         PATH
company-tools   https://github.com/acme/aix-resources.git   ~/.aix/repos/company-tools
personal        git@github.com:user/my-aix-tools.git        ~/.aix/repos/personal
```

### Update Repositories

```bash
aix repo update [name]
```

Pulls the latest changes from the remote. If no name is provided, updates all registered repositories.

**Examples:**

```bash
# Update a specific repository
aix repo update company-tools

# Update all repositories
aix repo update
```

### Remove a Repository

```bash
aix repo remove <name>
```

Unregisters the repository and deletes the local clone.

**Flags:**

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--keep-files` | | `bool` | Unregister without deleting the local clone |

**Examples:**

```bash
# Remove and delete local files
aix repo remove old-tools

# Unregister but keep local clone
aix repo remove archived-tools --keep-files
```

## Examples

### Add a Repository and Install a Skill

```bash
# Register the repository
aix repo add https://github.com/acme/aix-resources.git

# List available skills from all repositories
aix skill list --source=repos

# Install a specific skill
aix skill install company-tools:code-review
```

### Search Across Repositories

```bash
# Find all security-related resources
aix search security --source=repos

# List agents from a specific repository
aix agent list --repo=company-tools
```

### Update and Reinstall

```bash
# Pull latest changes
aix repo update company-tools

# Reinstall to pick up updates
aix skill install company-tools:code-review --force
```

## Validation Rules

### Name Validation

Repository names must conform to the following rules:

| Rule | Constraint |
|------|------------|
| Format | Lowercase alphanumeric with hyphens |
| Length | 1-64 characters |
| Pattern | `^[a-z][a-z0-9]*(-[a-z0-9]+)*$` |
| Restrictions | No consecutive hyphens (`--`), no leading/trailing hyphens |

**Valid names:**
- `company-tools`
- `my-skills`
- `aix-v2-resources`
- `tools`

**Invalid names:**
- `Company-Tools` (uppercase)
- `-tools` (leading hyphen)
- `my--tools` (consecutive hyphens)
- `my_tools` (underscore not allowed)
- `123-tools` (must start with letter)

### URL Validation

Repository URLs must use one of the following protocols:

| Protocol | Example | Notes |
|----------|---------|-------|
| HTTPS | `https://github.com/org/repo.git` | Recommended for public repos |
| SSH | `git@github.com:org/repo.git` | Requires SSH key authentication |
| Git | `git://github.com/org/repo.git` | Read-only, no authentication |

URLs must point to a valid Git repository. The `.git` suffix is optional but recommended for clarity.

## Error Cases

| Error | Constant | Description |
|-------|----------|-------------|
| Repository not found | `ErrNotFound` | The specified repository name is not registered |
| Invalid URL | `ErrInvalidURL` | The URL is malformed or uses an unsupported protocol |
| Name collision | `ErrNameCollision` | A repository with this name is already registered |
| Invalid name | `ErrInvalidName` | The repository name does not match the required pattern |
| Clone failed | `ErrCloneFailed` | Git clone operation failed (network, auth, or permissions) |
| Update failed | `ErrUpdateFailed` | Git pull operation failed (conflicts, network, or permissions) |

### Error Examples

```bash
# Repository not registered
$ aix repo update nonexistent
Error: repository not found: nonexistent

# Name already in use
$ aix repo add https://github.com/other/tools.git --name company-tools
Error: name collision: repository 'company-tools' already exists

# Invalid URL
$ aix repo add not-a-valid-url
Error: invalid URL: must use https://, git@, or git:// protocol

# Invalid name
$ aix repo add https://github.com/org/repo.git --name Invalid_Name
Error: invalid name: must be lowercase alphanumeric with hyphens (e.g., 'my-repo')
```

## Implementation Notes

### Clone Location

Repositories are cloned to `~/.aix/repos/<name>/` by default. This location can be overridden via the `AIX_REPOS_DIR` environment variable.

### Resource Resolution

When installing resources, the syntax `<repo>:<resource>` specifies which repository to use:

```bash
aix skill install company-tools:code-review
#                 └─────┬─────┘ └────┬────┘
#                   repo name    skill name
```

If no repository is specified, aix searches all registered repositories and uses the first match.

### Git Operations

All Git operations use the system's `git` binary. Authentication is handled through standard Git mechanisms (SSH keys, credential helpers, etc.).

## Content Validation

When adding or updating a repository, aix validates the repository content and displays warnings for issues that may cause problems during resource installation.

### Validation Checks

| Directory | Validation | Warning Message |
|-----------|------------|-----------------|
| `skills/` | Each subdirectory must contain `SKILL.md` | "skill directory missing SKILL.md" |
| `skills/` | Valid YAML frontmatter required | "invalid frontmatter: ..." |
| `commands/` | Subdirectories must contain `command.md` | "command directory missing command.md" |
| `commands/` | Valid YAML frontmatter in `.md` files | "invalid frontmatter: ..." |
| `agents/` | Subdirectories must contain `AGENT.md` | "agent directory missing AGENT.md" |
| `agents/` | Valid YAML frontmatter in `.md` files | "invalid frontmatter: ..." |
| `mcp/` | Valid JSON in `.json` files | "invalid JSON: ..." |

### Example Output

```bash
$ aix repo add https://github.com/example/skills.git
✓ Repository 'skills' added from https://github.com/example/skills.git
  Cached at: ~/.aix/repos/skills

⚠ Validation warnings:
  skills/broken/SKILL.md: invalid frontmatter: yaml: line 3: did not find expected '-' indicator
  mcp/server.json: invalid JSON: unexpected end of JSON input
```

### Notes

- Missing optional directories (e.g., `mcp/` when the repo only contains skills) do not generate warnings
- Validation warnings do not block the add or update operation
- Warnings help identify issues before attempting to install resources
