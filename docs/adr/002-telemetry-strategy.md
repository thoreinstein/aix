# ADR-002: Usage Telemetry Strategy

## Status

**Research Complete** | 2026-01-23

## Context

`aix` is a CLI tool that manages AI assistant configurations. Understanding how users interact with the tool could help prioritize features and identify usability issues. However, telemetry in open-source CLI tools is controversial and carries significant trust implications.

This ADR documents the research findings and recommendation for whether to implement telemetry.

## Research Summary

### Backend Options Evaluated

| Option | Type | Cost | Self-Host | CLI-Suited | Complexity |
|--------|------|------|-----------|------------|------------|
| PostHog | Product Analytics | Free tier: 1M events/mo, then usage-based | Yes (AGPL) | Overkill | High |
| Plausible | Web Analytics | From $9/mo | Yes (AGPL) | No | Medium |
| Umami | Web Analytics | Free (self-host only) | Yes | No | Medium |
| Custom (HTTP + SQLite) | DIY | Hosting costs only | Yes | Yes | Medium |
| None | - | $0 | N/A | N/A | None |

#### PostHog

**Summary**: All-in-one product analytics platform with feature flags, session replay, and A/B testing.

**Pros**:
- Comprehensive analytics (events, funnels, cohorts)
- Self-hostable (AGPL license)
- Free tier: 1M events/month, 365-day retention
- Well-documented SDKs

**Cons**:
- Kubernetes/Helm self-hosting deprecated; Docker Compose still works but less supported
- Overkill for simple CLI usage tracking
- Usage-based pricing can surprise at scale
- Complex setup for minimal needs

**Verdict**: Too complex for a CLI tool's modest needs.

#### Plausible

**Summary**: Privacy-focused, lightweight web analytics. GDPR/CCPA compliant by design.

**Pros**:
- Cookie-free, privacy-first
- Simple, clean dashboard
- Self-hostable (AGPL, Community Edition)
- 24K+ GitHub stars, active community

**Cons**:
- Designed for web page views, not CLI events
- No native CLI SDK--would require custom HTTP integration
- Self-hosting requires infrastructure

**Verdict**: Wrong tool for the job (web-focused).

#### Umami

**Summary**: Open-source, privacy-friendly web analytics.

**Pros**:
- Self-hosted only (truly free)
- GDPR compliant, no cookies
- Lightweight (<2KB tracking script)
- PostgreSQL support

**Cons**:
- Web analytics oriented, not event/CLI focused
- Requires self-hosting infrastructure
- MySQL support dropped in v3

**Verdict**: Similar limitations to Plausible--designed for websites.

#### Custom Solution (HTTP + SQLite/PostgreSQL)

**Summary**: Build a minimal telemetry endpoint.

**Pros**:
- Full control over data schema
- Minimal complexity--exactly what's needed
- No vendor dependency
- Can be as privacy-focused as desired

**Cons**:
- Must build and maintain
- Requires hosting infrastructure
- Security responsibility falls on maintainers

**Verdict**: Most appropriate if telemetry is implemented, but requires maintenance commitment.

#### No Telemetry

**Summary**: Rely on alternative feedback mechanisms.

**Pros**:
- Zero trust cost
- No infrastructure to maintain
- No privacy considerations
- Simplest option

**Cons**:
- No quantitative usage data
- Feature prioritization relies on anecdote
- Harder to identify silent failures

**Verdict**: Safest choice for user trust.

### Prior Art: How Other CLIs Handle Telemetry

#### Homebrew (Opt-Out, Controversial)

- **Model**: Opt-out (telemetry ON by default)
- **Disclosure**: Shown at first `brew update`
- **Data Collected**: CPU arch, OS version, formula name, Homebrew version
- **Disable**: `brew analytics off` or `HOMEBREW_NO_ANALYTICS=1`
- **Backend**: InfluxDB (migrated from Google Analytics)
- **Controversy**: 120+ thumbs-up on GitHub issue requesting opt-in instead

#### Go Telemetry Proposal (2023)

- **Proposed Model**: Opt-out with transparent disclosure
- **Community Response**: Strong pushback; revised to opt-in
- **Key Quote**: Users expressed that even minimal, transparent telemetry erodes trust
- **Outcome**: Opt-in model adopted after controversy

#### Rust/Rustup

- **Considered**: Telemetry proposed in 2016
- **Outcome**: Rejected
- **Philosophy**: "No compiler needs to call home"
- **Alternative Considered**: Local metrics with optional manual submission

### Key Insight

The pattern is clear: **opt-out telemetry in developer tools generates backlash**, even when the data collected is minimal and the process is transparent. Homebrew's ongoing controversy and Go's forced revision demonstrate that developer trust is fragile.

---

## Proposed Data Schema (If Implemented)

### Data Collected (Minimal, Anonymous)

| Field | Example | Purpose |
|-------|---------|---------|
| `cli_version` | `1.2.3` | Track adoption of new versions |
| `command` | `skill install` | Understand feature usage |
| `platform_target` | `claude` | Know which platforms are popular |
| `success` | `true` / `false` | Identify failure patterns |
| `os_arch` | `darwin/arm64` | Platform support prioritization |

### Data Explicitly Excluded

| Category | Examples | Rationale |
|----------|----------|-----------|
| **User Identity** | IP address, username, machine ID | Privacy |
| **File Content** | Paths, config values, skill content | Security |
| **MCP Details** | Server names, URLs, API keys | Sensitivity |
| **Timestamps** | Event time, session correlation | Anti-tracking |
| **Location** | Timezone, locale | Privacy |

### Data Retention

- **Maximum**: 90 days raw, then aggregated only
- **Aggregation**: Daily counts by command/platform/os, no individual events
- **Deletion**: On request (if identifiable, which it shouldn't be)

---

## Consent UX Design (If Implemented)

### Principles

1. **Off by default** -- Telemetry is disabled until explicitly enabled
2. **Explicit opt-in** -- No silent consent, no buried settings
3. **Easy verification** -- Users can check status anytime
4. **Instant opt-out** -- Disabling is immediate and permanent
5. **Fail silent** -- Network issues don't affect CLI operation

### Proposed Commands

```bash
# Check current status
aix telemetry status
# Output: "Telemetry is disabled. Run 'aix telemetry enable' to opt in."

# Enable telemetry
aix telemetry enable
# Output: Shows what will be collected, asks for confirmation

# Disable telemetry
aix telemetry disable
# Output: "Telemetry disabled. No data will be collected."
```

### First-Run Experience

**NOT proposed**: A first-run prompt asking about telemetry. This interrupts the user's actual task.

**Proposed**: Silent by default. Users who want to support development can run `aix telemetry enable`. Mentioned in README and `aix --help`.

### Environment Variable Override

```bash
AIX_TELEMETRY=0  # Disable telemetry (overrides config)
AIX_TELEMETRY=1  # Enable telemetry (overrides config)
```

---

## Privacy Statement Drafts

### README Section (Brief)

```markdown
## Privacy

`aix` does not collect any data by default. Optional, anonymous usage telemetry
can be enabled with `aix telemetry enable` to help prioritize development.

When enabled, we collect only: CLI version, command name, target platform,
success/failure, and OS/architecture. We never collect: file paths, config
content, IP addresses, or any identifying information.

Run `aix telemetry status` to check your current setting.
```

### Detailed Privacy Document (docs/privacy.md)

```markdown
# Privacy Policy

## Overview

`aix` respects your privacy. Telemetry is **disabled by default** and only
enabled through explicit user action.

## What We Collect (When Enabled)

When you opt into telemetry via `aix telemetry enable`, we collect:

| Data | Example | Purpose |
|------|---------|---------|
| CLI Version | `1.2.3` | Track version adoption |
| Command | `skill install` | Understand feature usage |
| Platform | `claude` | Prioritize platform support |
| Success | `true`/`false` | Identify problems |
| OS/Architecture | `darwin/arm64` | Platform compatibility |

## What We Never Collect

Regardless of telemetry setting, we **never** collect:

- IP addresses or location data
- File paths or content
- Configuration values
- MCP server names or URLs
- Skill or command content
- User identifiers or machine fingerprints
- Timestamps enabling session correlation

## Data Handling

- **Retention**: Raw events kept 90 days, then aggregated
- **Storage**: [Self-hosted / specific provider]
- **Access**: Only project maintainers
- **Sharing**: Never sold or shared with third parties

## Your Controls

```bash
aix telemetry status   # Check current setting
aix telemetry enable   # Opt in to telemetry
aix telemetry disable  # Opt out (immediate)
```

Environment variable override: `AIX_TELEMETRY=0` disables regardless of config.

## Contact

Privacy questions: [maintainer email or GitHub issues]
```

---

## Trade-off Analysis

### Value of Telemetry

| Benefit | Without Telemetry | With Telemetry |
|---------|-------------------|----------------|
| Feature prioritization | Anecdote, GitHub issues | Quantitative usage data |
| Error detection | User reports only | Failure rate patterns |
| Platform focus | Guesswork | Clear adoption numbers |
| Version adoption | Download counts only | Active usage by version |

### Cost of Telemetry

| Cost | Severity | Mitigation |
|------|----------|------------|
| User trust erosion | High | Strict opt-in, transparency |
| Implementation effort | Medium | Simple custom solution |
| Maintenance burden | Medium | Keep scope minimal |
| Potential controversy | Medium | Follow Go's opt-in lesson |
| Infrastructure cost | Low | Minimal data, cheap hosting |

---

## Recommendation

### Decision: **Conditional Go** (with constraints)

After reviewing the landscape, I recommend proceeding with telemetry **only if** the following constraints are met:

1. **Strict opt-in**: Telemetry is OFF by default, always
2. **No first-run prompt**: Don't interrupt users; mention in docs only
3. **Minimal scope**: Only the 5 fields specified (version, command, platform, success, os/arch)
4. **Self-hosted or privacy-first provider**: No Google Analytics or similar
5. **Open dashboard**: Consider making aggregate stats public (builds trust)
6. **Implementation deferred**: Not a priority for initial release

### Rationale

- The Go telemetry controversy shows that even well-intentioned opt-out telemetry damages trust
- Homebrew's ongoing criticism (7+ years) demonstrates long-term reputation cost
- Opt-in telemetry with strict privacy controls has lower risk but also lower adoption
- For a small CLI tool, GitHub issues and direct user feedback may be sufficient initially

### Alternative: No Telemetry

If the maintenance burden or trust risk seems too high, the recommended alternative is:

1. **GitHub Discussions**: Feature requests and feedback
2. **Version in User-Agent**: When making network requests (e.g., checking for updates), include version in User-Agent string
3. **Periodic surveys**: Announce via README/releases, link to anonymous form
4. **Download analytics**: GitHub releases provide download counts

### Implementation Priority

**Low**. Focus on core CLI functionality first. Telemetry can be added in a later release if quantitative data becomes necessary for decision-making.

---

## Decision

**Recommendation**: Conditional Go -- implement minimal, opt-in telemetry only if constraints above are met. Defer to post-1.0 release.

**Alternative accepted**: "No Telemetry" is also a valid choice given the small scope of this project and the availability of alternative feedback mechanisms.

## Consequences

### If Telemetry Is Implemented

- Must maintain telemetry infrastructure
- Must respond to privacy inquiries
- Must keep data schema minimal and documented
- Gains quantitative insight into usage patterns

### If Telemetry Is Not Implemented

- Zero trust cost
- Zero maintenance burden
- Rely on GitHub issues, discussions, and download stats
- May miss usage patterns that users don't explicitly report
