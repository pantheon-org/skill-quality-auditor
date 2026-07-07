---
title: "Known Issue: ripened DEFERRED items have no automated reactivation sweep"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-07
severity: LOW
value: LOW
themes:
  - GOVERNANCE
related:
  - ../findings/deferred-status-critical-review-2026-07-07.md
  - ../../docs/ADR/adr-060-deferred-lifecycle-status.md
  - ../plans/cross-reference-drift-audit-2026-07-04.md
---

# Known Issue: no automated sweep for ripened DEFERRED items

A date-gated `DEFERRED` item may carry `deferred_until: YYYY-MM-DD`, and the read
protocol surfaces such an item for reactivation once that date has passed. But this
only fires **when an agent happens to scan tier 2** — tier 2 is, by design, read only
after tier 1 is exhausted. There is no scheduled or CI check that proactively surfaces
"this item's gate passed; reactivate it."

**Failure mode:** a `deferred_until` date passes while tier 1 always has work. No agent
scans tier 2, so the ripened item is never reactivated and the gated task silently
never happens — the graveyard risk the field was meant to close, only narrowed.

**Partial mitigation already in place:** the `deferred_until` field plus the read-time
surfacing rule (ADR-060). This issue tracks only the missing *proactive* sweep.

**Possible fix:** a small advisory script that lists `DEFERRED` items with a passed
`deferred_until`, run from the periodic
[`cross-reference-drift-audit`](../plans/cross-reference-drift-audit-2026-07-04.md)
rather than as a blocking gate (a hard gate would fail unrelated commits the day an
item ripens). Low value; defer until an item actually ripens and is missed.
