---
title: "Privacy Policy"
description: "How aix handles your data"
summary: "aix is a local-first tool that respects your privacy."
date: 2023-09-07T17:19:07+02:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
type: "legal"
---

## Local-First Philosophy

`aix` is designed to be a local-first tool. All configurations, skills, and command definitions are stored on your local machine.

### Data Collection
- **No Tracking**: `aix` does not include any telemetry, tracking, or analytics.
- **Local Storage**: Your configurations are stored in `~/.config/aix/` and the respective configuration directories of your AI assistants (e.g., `~/.claude/`).
- **No Cloud Sync**: `aix` does not upload your configurations to any cloud service unless you explicitly use a git-based command (like `aix skill install <git-url>`), in which case data is handled by your local git client.

### Third-Party Services
When you use `aix` to manage MCP servers or AI assistants, those individual tools and services have their own privacy policies. `aix` only manages the *configuration* that allows those tools to run on your machine.
