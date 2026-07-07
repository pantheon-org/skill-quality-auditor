---
title: "ADR-061: ADR immutability begins at acceptance, not creation"
status: proposed
date: 2026-07-07
context:
  - path: .context/plans/adr-immutability-wording-2026-07-07.md
  - path: .context/findings/adr-immutability-wording-discrepancy-2026-07-07.md
---

**Status:** Proposed
**Date:** 2026-07-07

## Context

The `adr-capture` skill states that an ADR is immutable "once created" — its
banner, a mindset bullet, an anti-pattern, and its eval instructions all say an ADR
must never be edited after creation. Taken literally, that forces an author to
*supersede* an ADR they wrote minutes earlier in the same unmerged branch rather than
simply refining it, spawning a needless superseding ADR for a decision that was never
ratified. This nearly happened during ADR-060 (the DEFERRED lifecycle status), which
was amended in place three times while still `proposed`. See
`.context/findings/adr-immutability-wording-discrepancy-2026-07-07.md`.

## Decision

ADR immutability begins **at acceptance**, not at creation.

- A `proposed`, unmerged ADR is a draft and may be edited freely in place — title,
  body, and context.
- Once an ADR is `accepted`, its title, body, and context are frozen; thereafter only
  `status` and `superseded_by` may change, and replacing the decision is done by
  superseding (a new ADR marking the old one `superseded`).
- The immutability of `accepted` and `superseded` ADRs is unchanged by this decision;
  it only clarifies that the rule does not bind `proposed`/unmerged drafts.

## Consequences

- **Easier:** authors refine a `proposed` ADR in place during the same review cycle,
  without minting a superseding ADR for a decision that was never ratified. Less ADR
  churn; a cleaner history.
- **Harder:** "immutable" is now conditional on status, so the wording must be precise
  everywhere it appears (the `adr-immutability-wording` plan reconciles the skill's
  banner, mindset bullet, anti-pattern, and eval instructions to match).
- **Limitation:** the rule is documentation, not enforcement — nothing prevents an
  agent from editing an `accepted` ADR's body in place. This is accepted; the
  `adr-frontmatter.schema.json` deliberately stays a structural validator rather than
  encoding the behavioural rule.
