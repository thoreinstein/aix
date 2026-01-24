# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

As `aix` is currently in pre-1.0 development, security updates are provided for the latest release only. Once we reach 1.0, we will maintain security support for the current major version and the previous major version.

## Reporting a Vulnerability

We take security vulnerabilities seriously. Please report them responsibly.

### Preferred Method: GitHub Security Advisories

1. Go to the [Security Advisories](https://github.com/thoreinstein/aix/security/advisories) page
2. Click "Report a vulnerability"
3. Fill out the form with as much detail as possible

This method allows for private discussion and coordinated disclosure.

### Alternative: Email

If you cannot use GitHub Security Advisories, email your report to:

**thoreinstein8@gmail.com**

Use the subject line: `[SECURITY] aix vulnerability report`

## What to Include in a Report

Please provide as much of the following as possible:

- **Description**: Clear explanation of the vulnerability
- **Impact**: What an attacker could achieve
- **Affected versions**: Which versions are vulnerable
- **Reproduction steps**: Detailed steps to reproduce the issue
- **Proof of concept**: Code or commands that demonstrate the vulnerability
- **Suggested fix**: If you have recommendations for remediation

## Response Timeline

| Stage                          | Timeframe       |
| ------------------------------ | --------------- |
| Acknowledgment                 | Within 48 hours |
| Initial assessment             | Within 7 days   |
| Target fix for critical issues | Within 30 days  |
| Target fix for other issues    | Within 90 days  |

Timelines may vary based on complexity. We will keep you informed of progress.

## Disclosure Policy

We follow coordinated disclosure:

1. **Private report**: You report the vulnerability privately
2. **Acknowledgment**: We confirm receipt and begin investigation
3. **Fix development**: We develop and test a fix
4. **Release**: We release the fix with a security advisory
5. **Public disclosure**: Details are made public after users have had reasonable time to update (typically 30 days post-fix)

We request that you:

- Allow us reasonable time to address the issue before public disclosure
- Avoid exploiting the vulnerability beyond what's necessary for demonstration
- Do not access or modify other users' data

We will:

- Credit you in the security advisory (unless you prefer anonymity)
- Keep you informed throughout the process
- Not take legal action against researchers acting in good faith

## Security Best Practices for Users

### Protecting MCP Configurations

`aix` manages MCP server configurations that may contain sensitive data such as API tokens, authentication headers, and server URLs.

**Recommendations:**

- **Use environment variables** for sensitive values instead of hardcoding tokens in configuration files
- **Restrict file permissions** on configuration files containing credentials (`chmod 600`)
- **Never commit secrets** to version control; use `.gitignore` for local configuration files
- **Rotate credentials** regularly, especially if you suspect exposure
- **Review MCP server configurations** before sharing or publishing them

### Environment Variables

When configuring MCP servers, prefer environment variable references:

```json
{
  "env": {
    "API_TOKEN": "${MY_SERVICE_TOKEN}"
  }
}
```

Rather than hardcoded values:

```json
{
  "env": {
    "API_TOKEN": "sk-abc123..."
  }
}
```

## Security Scanning

This project uses automated security scanning:

- **CodeQL**: Static analysis for security vulnerabilities in Go code, runs on every pull request and push to main
- **Dependabot**: Automated dependency updates to address known vulnerabilities in dependencies

Security scan results are reviewed by maintainers. Critical findings block merges until resolved.

## Questions

For general security questions (not vulnerability reports), open a [GitHub Discussion](https://github.com/thoreinstein/aix/discussions) or issue.
