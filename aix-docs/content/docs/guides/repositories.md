---
title: "Repositories"
description: "Discover and share AI resources using remote git repositories"
summary: "Learn how to use aix repositories to discover, install, and manage skills, commands, and agents from the community."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 60
toc: true
seo:
  title: "Repositories Guide - aix"
  description: "Guide to managing remote Git repositories as sources for AI resources in aix. Learn how to add, list, update, and remove repos."
---

**Repositories** are the backbone of the aix ecosystem. They allow you to discover and share AI resources--such as skills, commands, and agents--via standard Git repositories.

Instead of manually downloading and managing individual files, you can add a repository once, and `aix` will handle the caching, indexing, and installation for you.

## How it Works

`aix` repositories are standard Git repositories that follow a simple directory structure. When you add a repository, `aix` performs a **shallow clone** into a local cache directory (`~/.aix/cache/repos`).

The CLI then scans the repository for resources in the following subdirectories:
*   `skills/`: Reusable agent skills (`SKILL.md`).
*   `commands/`: Custom slash commands.
*   `agents/`: Agent definitions.
*   `mcp/`: Model Context Protocol server configurations.

## Adding a Repository

Use the `aix repo add` command to register a new repository source.

```bash
# Add from GitHub
aix repo add https://github.com/thoreinstein/agents.git

# Add with a custom name
aix repo add https://github.com/user/my-skills.git --name community-skills

# Add a private repository via SSH
aix repo add git@github.com:org/private-resources.git
```

By default, the repository name is derived from the URL (e.g., `agents`). You can override this with the `--name` flag.

## Listing Repositories

To see all registered repositories and where they are cached on your system:

```bash
aix repo list
```

To output the list in JSON format (useful for automation):

```bash
aix repo list --json
```

## Updating Repositories

Since repositories are cached locally, you need to pull the latest changes occasionally to get new resources or updates.

```bash
# Update all repositories
aix repo update

# Update a specific repository
aix repo update agents
```

## Removing a Repository

If you no longer want to use a repository as a source, you can remove it. This also cleans up the locally cached files.

```bash
aix repo remove agents
```

## Discovery (Coming Soon)

Once a repository is added, you can search for resources across all repositories using:

```bash
aix search <query>
```

This makes it easy to find a specific skill or tool without knowing which repository it belongs to.

## Best Practices

1.  **Use Meaningful Names**: When adding multiple repositories, use `--name` to give them clear identifiers like `internal-tools` or `community-core`.
2.  **Organize Your Repos**: If you are creating your own repository, keep resources organized in the standard `skills/`, `commands/`, `agents/`, and `mcp/` folders.
3.  **Keep it Shallow**: `aix` uses `--depth=1` by default to keep the local cache small and fast.
