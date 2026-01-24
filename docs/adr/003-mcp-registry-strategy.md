# ADR 003: MCP Registry Strategy

## Context
The `aix` CLI needs to search and install MCP servers. The original requirement was to use `mcpservers.org` as the source. However, investigation revealed that `mcpservers.org` (and its likely upstream `punkpeye/awesome-mcp-servers`) is a curated list of links to external repositories/documentation, not a machine-readable registry with standardized installation commands.

MCP servers require specific configuration to run (executable path, arguments, environment variables), which varies by implementation (Node.js vs Python, Docker vs Local). This information is not structured in the "awesome" lists.

## Decision
We will establish a dedicated **aix MCP Registry** as a Git repository.

1.  **Registry Format**: The registry will follow the existing `aix` repository format:
    - Root of the repo contains an `mcp/` directory.
    - Inside `mcp/`, each server is defined in a separate JSON file (e.g., `mcp/github.json`).
    - The JSON schema matches the `mcp.Server` struct in `aix`.

2.  **Community Driven**: We will seed this registry with popular servers found on `mcpservers.org`. We will encourage the community to submit PRs to add more servers.

3.  **Command Updates**:
    - `aix repo add` will support adding this official registry (enabled by default or easily added).
    - `aix mcp search` will query this registry.
    - `aix mcp install` will download the JSON definition from this registry.

4.  **Future Automation**: We may build a scraper/bot that monitors `mcpservers.org` or `awesome-mcp-servers` to identify *new* servers and propose them for inclusion in our registry (as draft PRs requiring human verification of install commands).

## Consequences
- **Pros**:
    - Reliable, verified installation commands.
    - Standardized configuration format.
    - Fast search (local cache of the registry).
- **Cons**:
    - Maintenance burden to keep the registry up-to-date.
    - Disconnect from the "source of truth" (`mcpservers.org`).

## Alternatives Considered
- **Scraping**: Attempting to scrape installation instructions from READMEs. Rejected due to high failure rate and security risk of executing unverified commands.
- **Upstream Push**: Contributing a `registry.json` standard to `mcpservers.org`. This is a valid long-term goal but blocks our immediate progress. We can switch to it if they adopt a standard.
