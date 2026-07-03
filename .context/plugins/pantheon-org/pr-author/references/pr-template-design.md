# What Makes a Good PR Template

This document defines the characteristics of an effective PR template, using this repo's `.github/pull_request_template.md` as the reference implementation.

## Structure

A good PR template has these sections:

1. **Summary** — One paragraph explaining *what* the PR changes and *why*. Should be scannable in 30 seconds.
2. **Type of change** — Checklist with conventional-commit prefixes (`feat/`, `fix/`, `refactor/`, `docs/`, `chore/`). Helps reviewers understand scope at a glance.
3. **Checklist** — Grouped by concern, not a flat laundry list:
   - Code quality (lint, tests, coverage)
   - Skill/tile changes (if applicable to the repo)
   - Documentation updates
   - Commit conventions
4. **Merge strategy** — Explicit preference with guidance. Most repos prefer squash for feature branches.
5. **Related issues** — Auto-close syntax (`Closes #NNN`), links to related PRs/ADRs/context files.

## Quality Signals

- **Concrete checklist items:** "`go test ./...` passes" not "tests pass"
- **Grouped concerns:** All skill/tile checks together, all code-quality checks together — agents can skip irrelevant groups
- **Repo-specific conventions:** Includes checks unique to the repo (e.g., `./dist/skill-auditor eval`, `tessl status`, ADR requirements)
- **Character limit awareness:** Notes GitHub's 65536-character body limit and 72-character title truncation in list views
- **HTML comments for guidance:** Invisible instructions to human authors (`<!-- What does this PR change? -->`), stripped by agents when filling

## Anti-Patterns in Template Design

- **Vague checkboxes:** "Code looks good" → no action possible. Use verifiable statements.
- **Missing merge strategy guidance:** Creates ambiguity at merge time. State the preference explicitly.
- **No distinction between code and asset changes:** A PR that only updates `cmd/assets/` does not need `go test ./...` checked.
- **Overly long default sections:** Encourages dumping rather than summarising. The body should be a summary, not a transcript.
- **No linked-issue guidance:** Missing auto-close syntax means issues stay open after merge.

## Reference Implementation

See `.github/pull_request_template.md` in this repo for a complete example.
