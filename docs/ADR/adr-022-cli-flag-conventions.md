---
title: "ADR-022: Standardize CLI flag naming and output format conventions"
status: accepted
date: 2026-04-29
context:
  - path: .context/plans/standardize-cli-flags-2026-04-29.md
---

**Status:** Accepted
**Date:** 2026-04-29

## Context

Each `skill-auditor` command defined its flags independently, leading to inconsistent naming, overlapping short flags, and unpredictable user experience. The same concept (e.g., JSON output) was a flag on some commands and the default on others. There was no documented convention for flag naming or output format selection.

## Decision

Adopt a standardised flag interface across all commands, implemented in 9 merged PRs:

| Flag | Short | Type | Rule |
| ---- | ----- | ---- | ---- |
| `--json` | `-j` | bool | JSON output (default for most commands) |
| `--markdown` | `-m` | bool | Markdown output instead of JSON |
| `--store` | `-s` | bool | Persist to `.context/` |
| `--repo-root` | `-r` | string | Auto-detected if empty |
| `--dry-run` | `-n` | bool | Preview without side effects |

JSON is the default output format. `--markdown` opts into Markdown. `--json` and `--markdown` are mutually exclusive. A `resolveOutputFormat` helper in `cmd/output.go` centralises the logic.

Short flag conventions: first letter where unambiguous (`-j`, `-m`, `-s`, `-r`, `-d`), Unix dry-run convention (`-n`), uppercase for collisions (`-F` for `--fail`, `-S` for `--strict`).

## Consequences

- Predictable CLI experience across all commands
- 9 PRs coordinated and merged — all commands updated
- JSON default aligns with machine-readability goals
- `resolveOutputFormat` is the single source of truth for format selection
- Short flags are documented and predictable
