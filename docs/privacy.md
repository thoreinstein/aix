# Privacy Policy

## Overview

`aix` respects your privacy. Telemetry is **disabled by default** and only enabled through explicit user action.

This document describes what data could be collected if telemetry is implemented and enabled, and provides transparency about our privacy practices.

## Current Status

As of this writing, `aix` does not implement telemetry. This document describes the approved approach should telemetry be added in a future release.

## What We Collect (When Telemetry Is Enabled)

If you opt into telemetry via `aix telemetry enable`, we collect:

| Data | Example | Purpose |
|------|---------|---------|
| CLI Version | `1.2.3` | Track version adoption, identify outdated installs |
| Command | `skill install` | Understand which features are used |
| Platform Target | `claude` | Prioritize platform support and testing |
| Success/Failure | `true`/`false` | Identify reliability problems |
| OS/Architecture | `darwin/arm64` | Guide platform compatibility decisions |

This data is collected anonymously--there is no user identifier linking events together.

## What We Never Collect

Regardless of telemetry setting, we **never** collect:

| Category | Examples | Why We Exclude It |
|----------|----------|-------------------|
| **User Identity** | IP address, username, email, machine ID | Your identity is not our business |
| **File System** | Paths, directories, filenames | Could reveal project structure |
| **Content** | Config values, skill content, command arguments | Could contain sensitive data |
| **MCP Details** | Server names, URLs, API keys, headers | Security-sensitive configuration |
| **Timing** | Precise timestamps, session duration | Enables tracking and correlation |
| **Location** | Timezone, locale, geographic region | Privacy-invasive |
| **System Details** | Hostname, username, process tree | Could fingerprint users |

## Data Handling

### Collection

- Events are sent over HTTPS
- No cookies or local tracking files
- Network failures are silently ignored (telemetry never blocks CLI operation)

### Storage

- Data is stored on infrastructure controlled by project maintainers
- We do not use third-party analytics services that may have their own data practices

### Retention

- Raw events: Maximum 90 days
- After 90 days: Aggregated into daily/weekly counts with no individual event data
- Aggregated data: Retained indefinitely for historical trend analysis

### Access

- Only project maintainers have access to telemetry data
- Data is not sold, shared, or provided to third parties
- Aggregate statistics may be published (e.g., "60% of users target Claude Code")

## Your Controls

### Check Status

```bash
aix telemetry status
```

Shows whether telemetry is enabled or disabled.

### Enable Telemetry

```bash
aix telemetry enable
```

Opts you in to anonymous usage telemetry. You'll see a summary of what will be collected before confirming.

### Disable Telemetry

```bash
aix telemetry disable
```

Immediately stops all telemetry collection. No confirmation required.

### Environment Variable Override

```bash
# Force telemetry off (overrides config file)
export AIX_TELEMETRY=0

# Force telemetry on (overrides config file)
export AIX_TELEMETRY=1
```

The environment variable takes precedence over any configuration file setting.

## GDPR and CCPA Compliance

Because we:
- Collect no personally identifiable information
- Do not track users across sessions
- Do not sell or share data
- Provide easy opt-out

We believe our telemetry approach, if implemented, is compliant with GDPR and CCPA. However, we are not lawyers, and you should consult legal counsel if you have specific compliance requirements.

## Open Source Commitment

`aix` is open source. You can:
- Review exactly what telemetry code does (when implemented)
- Disable telemetry at build time if you compile from source
- Fork the project without telemetry

## Changes to This Policy

If we change our privacy practices:
- This document will be updated
- Changes will be noted in release notes
- Significant changes will be announced in the README

## Contact

For privacy questions or concerns:
- Open a GitHub issue: [repository URL]
- Email: [maintainer email if desired]

## Summary

| Question | Answer |
|----------|--------|
| Is telemetry on by default? | **No** |
| Can I opt out? | **Yes, instantly** |
| Do you collect my IP? | **No** |
| Do you collect file paths? | **No** |
| Do you sell data? | **No** |
| Can I verify the code? | **Yes, it's open source** |
