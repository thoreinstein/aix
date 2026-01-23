---
title: "Agents"
description: "Manage AI Agent personas and configurations"
summary: "Configure and switch between different agent personas across your AI platforms."
date: 2026-01-23T00:00:00+00:00
lastmod: 2026-01-23T00:00:00+00:00
draft: false
weight: 50
toc: true
seo:
  title: "Agents Guide - aix"
  description: "Learn how to manage AI agent personas and configurations for Claude Code and OpenCode using aix."
---

In `aix`, an **Agent** represents a full configuration set: a system prompt (persona), a set of tools, and specific configuration settings. This is primarily used for platforms like **OpenCode** and **Claude Code** that support switching between different "profiles".

## Listing Agents

To see which agents are currently available on your system:

```bash
aix agent list
```

**Output:**
```text
NAME            PLATFORM    DESCRIPTION
default         opencode    General purpose coding assistant
senior-dev      claude      Strict code reviewer with security focus
python-expert   opencode    Specialized for Python/Django development
```

## Inspecting an Agent

To view the details of a specific agent, including its system prompt and active tools:

```bash
aix agent show senior-dev
```

## Creating Agents

*Support for creating new agents directly via `aix` is currently in preview.*

Currently, `aix` detects agents defined in your platform-specific configuration directories (e.g., `~/.config/opencode/agents/`). Future versions will allow you to define agents in a platform-agnostic format similar to Skills.

## Best Practices

*   **Specialization**: Create separate agents for distinct roles (e.g., "Tech Lead" vs "QA Tester") rather than one giant agent.
*   **Skill Reuse**: Use **Skills** to share capabilities between agents, rather than hardcoding tools into the agent definition.
