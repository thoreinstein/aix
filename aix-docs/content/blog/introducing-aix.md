---
title: "Introducing aix: The Unified CLI for AI Coding Assistants"
description: "Manage configurations across Claude Code, OpenCode, and Gemini CLI with one tool. Now with interactive search and bulk installs."
date: 2026-01-27T10:00:00+00:00
draft: false
weight: 1
categories: ["Announcements"]
tags: ["release", "cli", "ai", "opensource"]
contributors: ["thoreinstein"]
---

Today we are excited to introduce **aix**, a unified command-line interface designed to streamline the management of AI coding assistant configurations.

As the ecosystem of AI coding assistants grows, developers often find themselves juggling multiple tools—**Claude Code**, **OpenCode**, and **Gemini CLI**—each with their own configuration formats, file locations, and conventions. Keeping your custom skills, slash commands, and MCP servers in sync across these environments is a manual, error-prone process.

**aix** bridges this gap by providing a single, platform-agnostic control plane.

## Why aix?

- **Unified Management**: Manage skills, commands, agents, and MCP servers across multiple platforms using a single CLI.
- **Write Once, Deploy Everywhere**: Define your skills and commands in a standard format (Markdown/YAML) and let `aix` handle the translation to each platform's native syntax (TOML, JSON, etc.).
- **Repo-based Distribution**: Share your configurations via Git repositories and install them anywhere.

## What's New in v0.6.0

We've been moving fast. Here are some of the power-user features available today:

### 🔍 Interactive Fuzzy Search
Don't remember the exact name of that skill? Just run `aix search`. We've added an **interactive, fzf-style fuzzy finder** that lets you browse, filter, and preview skills, commands, and agents from all your configured repositories in real-time.

### 📦 Bulk Installation
Setting up a new machine? You can now install entire suites of tools in one go. The new `--all-from-repo` flag lets you pull everything from a trusted source instantly:

```bash
aix skill install --all-from-repo team-shared-skills
```

### 🐧 XDG Compliance
We know you care about dotfile hygiene. `aix` fully respects XDG Base Directory specifications on Linux and macOS, keeping your home directory clean (`~/.config/aix`, `~/.cache/aix`).

## Supported Platforms

`aix` currently provides deep integration with:

- **Claude Code**: Full support for skills, commands, and MCP servers.
- **OpenCode**: Support for skills, commands, agents, and MCP servers.
- **Gemini CLI**: Full support including variable translation (`$ARGUMENTS` → `{{argument}}`).

## Getting Started

If you have Homebrew installed, you can grab the latest version:

```bash
brew install thoreinstein/tap/aix
```

Initialize your environment (we'll auto-detect your installed AI tools):

```bash
aix init
```

Check out our [GitHub repository](https://github.com/thoreinstein/aix) and [Getting Started Guide](/docs/getting-started/).

We are building `aix` to be the standard package manager for the AI coding era. Give it a spin and let us know what you think!
