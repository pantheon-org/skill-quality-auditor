---
title: "ADR-027: Treat .context/audits/ files as first-class context — validated and indexed"
status: accepted
date: 2026-07-03
context:
  - path: .context/plans/audit-frontmatter-validation-remediation-2026-07-03.md
  - path: .context/findings/plan-status-drift-2026-07-01.md
---

**Status:** Accepted
**Date:** 2026-07-03

## Context

Audit files generated under `.context/audits/<skill>/<date>/` by `reporter/analysis.go`
and `reporter/remediation.go` carried canonical frontmatter (`type: audit`, `status: done`)
since commit c1e70ba (PR #83), but two pre-commit hooks (`context-frontmatter`,
`context-index`) explicitly excluded `.context/audits/**`, creating a validation blind spot.

A stale-variable bug in `regenerate-context-index.sh` compounded the problem: audit files
were either silently dropped from the index or miscategorised (one leaked into the `findings`
group with `type: finding`).

## Decision

1. **Remove pre-commit exclusions** — delete `exclude = List(".context/audits/**")` from both
   `context-frontmatter` and `context-index` hooks in `hk.pkl`.
2. **Fix the index script** — remove the stale-variable skip that checked the previous file's
   content; add an explicit `audits` section to `type_group_key`, `type_order`, and `type_label`.
3. **Canonicalise frontmatter** — backfill `title`, `type: audit`, `status: done`, `date` on
   existing audit files that were missing or had incorrect frontmatter (`type: finding`).
4. **Surface in index** — audit files appear under a dedicated `audits:` section in
   `.context/index.yaml`, not in a catch-all `other:` bucket.

## Consequences

- `hk check` now validates audit file frontmatter against the shared schema (`type: audit`
  is in the enum, so validation passes).
- The `adr-undocumented` hook scans `.context/audits/` but audit files do not contain decision
  keywords (`## Decision`, `## Recommendation`), so no false positives arise.
- 8 audit files across 4 skills now appear in the index under `audits:`.
- The stale-variable bug is removed entirely; if audit skipping is ever reintroduced, the
  guard must read the current file's content first.
