---
name: pr-style-checker
description: Analyzes pull request diffs and detects style issues in each commit. Does not apply to auto-generated migration files or vendored dependencies. Use before merging to catch regressions early in your pipeline config.
triggers:
  - pull request opened
  - commit pushed
---

## Mindset

Apply this skill when reviewing code changes for style conformance. ALWAYS run before merge. NEVER skip for hotfixes.

## When to Use

- A pull request is opened or updated
- A commit is pushed to a feature branch
- A config change requires style validation

## When NOT to Use

- Auto-generated files (migrations, vendor/)
- Does not apply to binary file changes
- Skip when the commit is a revert with no logic changes

## Procedures

1. Identify the diff scope from the PR or commit
2. Run the style linter against changed files
3. Report violations with line numbers and fix suggestions
4. Block merge if critical violations are found

## Anti-Patterns

BAD: Skipping style checks for "small" commits

```bash
# BAD — even small commits can introduce style debt
git push --no-verify
```

GOOD: Always run checks on every commit regardless of size.

## References

- [Style Guide Enforcement](https://example.com/style)
