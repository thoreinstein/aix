---
title: "Skills"
description: "Create and manage reusable agent skills with aix"
summary: "Learn how to define, package, and deploy portable agent skills across different AI platforms."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 20
toc: true
seo:
  title: "Skills Guide - aix"
  description: "Guide to creating and managing portable Agent Skills with aix. Learn the SKILL.md format and how to deploy skills to Claude, OpenCode, and Gemini."
---

**Skills** are reusable packages of instructions, tools, and prompts that give your AI assistant new capabilities. `aix` allows you to define a skill once and use it everywhere.

## What is a Skill?

A skill is defined by a `SKILL.md` file that follows the [Agent Skills Specification](https://agentskills.io/specification). It typically contains:
*   **Metadata**: Name, version, description.
*   **Tools**: Which MCP tools or shell commands the skill needs.
*   **Instructions**: The prompt that teaches the AI how to use the skill.
*   **Triggers**: Slash commands (e.g., `/git`) that activate the skill.

## Creating a Skill

Use `aix skill init` to scaffold a new skill.

```bash
aix skill init my-feature
```

This creates a directory `my-feature/` with a starter `SKILL.md`:

```markdown
---
name: my-feature
description: Description of what this skill does
version: 1.0.0
tools:
  - Bash
triggers:
  - /feature
---

# My Feature Skill

You are an expert at...
```

## Installing Skills

Once you have a skill defined (or downloaded), install it to your active platforms.

```bash
# Install from a local directory
aix skill install ./my-feature

# Install from a remote URL (planned feature)
aix skill install https://example.com/skills/git-workflow.md
```

### Targeting Platforms
By default, `aix` installs the skill to all configured platforms. You can restrict this with flags:

```bash
# Install only to OpenCode
aix skill install ./my-feature --platform opencode
```

## Managing Skills

### Listing Skills
See what skills are installed and which platforms they are active on.

```bash
aix skill list
```

### Removing Skills
To remove a skill from all platforms:

```bash
aix skill remove my-feature
```

## Best Practices

1.  **Keep it Focused**: A skill should do one thing well (e.g., "Git Workflow" or "Database Migration").
2.  **Define Tools**: Explicitly list required tools in the `tools` frontmatter section so `aix` can validate permissions.
3.  **Use Triggers**: Define clear slash commands (e.g., `/test`, `/deploy`) to make the skill easy to invoke.
