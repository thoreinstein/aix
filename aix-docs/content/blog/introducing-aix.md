---
title: "Introducing aix: The Unified CLI for AI Coding Assistants"
description: "Manage configurations across Claude Code, OpenCode, and Gemini CLI with one tool."
date: 2026-01-24T10:00:00+00:00
draft: false
weight: 1
categories: ["Announcements"]
tags: ["release", "cli", "ai"]
contributors: []
---

Today we are excited to introduce aix, a unified command-line interface designed to streamline the management of AI coding assistant configurations.

As the ecosystem of AI coding assistants grows, developers often find themselves juggling multiple tools, each with their own configuration formats, file locations, and conventions. Whether you are using Claude Code, OpenCode, or the Gemini CLI, keeping your skills, commands, and MCP servers in sync is a challenge.

aix bridges this gap by providing a single, platform-agnostic control plane.

## Key Features

- **Unified Management**: Manage skills, commands, agents, and MCP servers across multiple platforms using a single set of commands.
- **Write Once, Deploy Everywhere**: Define your skills and commands in a standard format and let aix handle the translation to each platform's native syntax.
- **Automatic Platform Detection**: Run `aix init` and let the tool discover which assistants you have installed.
- **Spec Compliance**: Full support for the Agent Skills and Model Context Protocol (MCP) specifications.

## Supported Platforms

At launch, aix supports:

- **Claude Code**: Full support for skills, commands, and MCP servers.
- **OpenCode**: Support for skills, commands, agents, and MCP servers.
- **Gemini CLI**: Support for skills, commands, and MCP servers.

## Getting Started

Getting started is easy. If you have Homebrew installed, you can install aix with:

```bash
brew install thoreinstein/tap/aix
```

Then, initialize your configuration:

```bash
aix init
```

Check out our [Getting Started Guide](/docs/getting-started/) for more details.

We are just getting started with aix and have many more platforms and features planned. Try it out and let us know what you think!
