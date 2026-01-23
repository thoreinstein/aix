---
title: "Slash Commands"
description: "Create custom slash commands with aix"
summary: "Automate common workflows with custom slash commands that work across all your AI assistants."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 40
toc: true
seo:
  title: "Slash Commands Guide - aix"
  description: "Learn how to define and deploy custom slash commands for Claude Code, OpenCode, and Gemini using aix."
---

**Slash Commands** provide a quick, shorthand way to invoke specific actions or context-gathering scripts. Unlike Skills (which are broad capabilities), Commands are typically focused, single-purpose scripts.

## Creating a Command

Initialize a new command with `aix command init`.

```bash
aix command init /deploy
```

This creates a markdown file (e.g., `deploy.md`) with the command definition.

### Command Format

```markdown
---
name: deploy
description: Deploy the application to a specific environment
arguments:
  - name: environment
    description: Target environment (staging | production)
    required: true
---

# Deployment Instruction

Please deploy the current codebase to the **$ARGUMENTS** environment.
Use the `scripts/deploy.sh` script.
```

### Variables
`aix` handles variable interpolation across platforms.
*   `$ARGUMENTS`: The text typed after the command (e.g., `/deploy staging` -> `staging`).
*   `$SELECTION`: The currently selected code in the editor (if supported).

## Installing Commands

### From a Repository
The easiest way to install commands is from a configured [repository](guides/repositories/).

```bash
aix command install deploy
```

### From a Local File
You can also install commands directly from a markdown file.

```bash
aix command install ./deploy.md
```

`aix` will translate this into:
*   **Claude Code**: A custom slash command file.
*   **Gemini CLI**: A command entry in `.gemini/commands/`.
*   **OpenCode**: A command entry in the agent configuration.

## Managing Commands

### Listing
```bash
aix command list
```

### Removing
```bash
aix command remove /deploy
```

## When to use Commands vs. Skills?

| Feature | Commands | Skills |
| :--- | :--- | :--- |
| **Scope** | Single action or prompt | Broad capability / workflow |
| **Complexity** | Low | High |
| **Tools** | Usually relies on existing tools | Can bundle its own tools |
| **Example** | `/fix-lint` | `Git Workflow` |
