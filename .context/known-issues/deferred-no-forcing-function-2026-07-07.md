---
title: "Known Issue: DEFERRED has no forcing function — blocked items can still be filed ACTIVE"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-07
severity: MEDIUM
value: MEDIUM
themes:
  - GOVERNANCE
related:
  - ../findings/deferred-status-critical-review-2026-07-07.md
  - ../../docs/ADR/adr-060-deferred-lifecycle-status.md
  - ./known-issues-lack-enforcement-2026-07-06.md
---

# Known Issue: DEFERRED has no forcing function

`DEFERRED` (ADR-060) lets authors keep un-actionable work out of the "what's next"
pick, but nothing **detects** that an `ACTIVE` item is actually blocked or gated and
forces it to `DEFERRED`. The benefit depends entirely on author discipline.

**Failure mode:** a plan gated on an external release is filed `ACTIVE` (the habitual
default). It tops the read-protocol pick again, exactly the pollution `DEFERRED`
exists to prevent. Nothing flags it.

This is the same class of gap as
[`known-issues-lack-enforcement-2026-07-06`](known-issues-lack-enforcement-2026-07-06.md):
a convention with no enforcement or forcing function.

**Possible mitigations (not yet chosen):** a lint that flags plan bodies containing
blocked/gated/waiting-on language while `status: ACTIVE`; a review-checklist prompt;
or accepting it as discipline-only and documenting that. Deliberately left open — the
right fix may piggyback on a broader enforcement mechanism for the whole taxonomy.
