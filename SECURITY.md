# Security Policy

## Reporting a Vulnerability

If you find a security issue in skill-quality-auditor, please report it privately via
[GitHub Security Advisories](https://github.com/pantheon-org/skill-quality-auditor/security/advisories/new)
rather than opening a public issue.

Include as much detail as you can: what you found, how to reproduce it, and what impact you think it has.
You should expect a response within a few days.

## Scope

skill-quality-auditor processes untrusted SKILL.md files and skill packages on the user's machine.
Security-relevant issues include (but are not limited to):

- Path traversal (reading or writing files outside the skill or output directories)
- Command injection through skill content or file paths
- Input that causes the tool to hang, crash, or consume excessive resources
- Denial of service via malformed skill files (e.g. deeply nested YAML, regex backtracking)
- Unintended network requests triggered by skill content
