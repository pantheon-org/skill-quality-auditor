---
title: "ADR-060: DEFERRED lifecycle status for parked context items"
status: accepted
date: 2026-07-07
context:
  - path: .context/plans/deferred-status-2026-07-07.md
  - path: .context/findings/deferred-status-critical-review-2026-07-07.md
  - path: .context/known-issues/deferred-no-forcing-function-2026-07-07.md
  - path: .context/known-issues/deferred-no-reactivation-sweep-2026-07-07.md
---

**Status:** Accepted
**Date:** 2026-07-07

## Context

The `.context/` frontmatter status enum was `DRAFT | ACTIVE | DONE | SUPERSEDED`.
Real items that are not done but cannot be actioned — date-gated (e.g. the
tessl-eval-decommission Bucket A, gated to ~15-07-2026) or externally blocked
(e.g. the plan-review execution-location lens, blocked on an eval-harness
limitation) — had no home but `ACTIVE`. Filed as `ACTIVE`, they surfaced as top
candidates in the "what's next" read protocol despite being un-actionable, so the
protocol repeatedly recommended work that could not be picked up.

## Decision

Add `DEFERRED` as a fifth lifecycle status: a real item that **cannot be actioned
yet** because it is date-gated or externally blocked, distinct from `ACTIVE`
(pick-up-next) and `DRAFT` (not yet reviewed).

- **Scope is "cannot", not "won't".** `DEFERRED` is reserved for work that is
  genuinely un-actionable right now. Merely low-priority work stays `ACTIVE` with
  `value: LOW` — that axis already sorts it last within tier 1. Overloading
  `DEFERRED` with "deprioritised" was considered and rejected: it would duplicate
  the `value` axis and blur the "can it be done now?" line the status exists to draw.
- **Read protocol ranks `DEFERRED` in a strict second tier.** Candidates split into
  tier 1 (`DRAFT`/`ACTIVE`) and tier 2 (`DEFERRED`); tier 1 is always exhausted
  before tier 2. Within each tier the existing sort applies (`value` descending,
  `effort` ascending, `themes[0]`). A `DEFERRED` item never outranks a tier-1 item
  regardless of its `value`. Reactivate to `ACTIVE` when the blocker clears.
- **Optional `deferred_until` reactivation date.** A date-gated item may carry a
  `deferred_until: YYYY-MM-DD` field (only valid with `status: DEFERRED`, enforced by
  the validator; the context index carries it so the protocol can filter without
  opening files). The read protocol does **not list** an item whose `deferred_until`
  is still in the future — it is hidden from the pick until that date passes, then
  surfaces as reactivation-eligible. The date takes precedence over the
  blocked-but-visible default: an item can be both externally blocked and date-gated,
  and when it is, the date governs visibility. Externally-blocked items with no known
  ripen date omit `deferred_until` and stay visible in tier 2 (below all DRAFT/ACTIVE).
- **`value`, `themes`, and `effort` remain required on `DEFERRED`** (same as
  `DRAFT`/`ACTIVE`), enforced by `validate-context-frontmatter.sh`, so a parked item
  re-ranks cleanly on reactivation. Only `DONE`/`SUPERSEDED` are exempt.
- Excluding `DEFERRED` from the pick entirely (listing it separately) was considered
  and rejected: the intent was "lower priority than active, still in the same queue".

## Consequences

- **Easier:** authors can keep gated or blocked work out of the pick, so the `ACTIVE`
  set once again means "work you would genuinely pick up next"; parked items stay
  fully graded, ready to re-rank the moment they reactivate.
- **Harder:** one more status to reason about, and a re-grade obligation on
  `ACTIVE ↔ DEFERRED` transitions (added to the ways-of-working re-grade rule).
- **Limitation — no forcing function.** Nothing detects that an `ACTIVE` item is
  actually blocked and forces it to `DEFERRED`; the benefit depends on author
  discipline. This shares the gap recorded in
  [`known-issues-lack-enforcement-2026-07-06`](../../.context/known-issues/known-issues-lack-enforcement-2026-07-06.md)
  and is tracked for `DEFERRED` specifically in
  [`deferred-no-forcing-function-2026-07-07`](../../.context/known-issues/deferred-no-forcing-function-2026-07-07.md).
  The `deferred_until` convention likewise has no automated sweep yet
  ([`deferred-no-reactivation-sweep-2026-07-07`](../../.context/known-issues/deferred-no-reactivation-sweep-2026-07-07.md));
  reactivation currently happens at read time when an agent scans tier 2.
- **Adoption:** the change delivers nothing until the actual blocked items are
  migrated. The plan-review execution-location lens (blocked on the eval harness) is
  re-statused to `DEFERRED` in the same changeset; the tessl-eval-decommission
  Bucket A (date-gated to ~15-07-2026) is re-statused with its own branch to avoid a
  cross-branch status conflict.
- **Scope:** schema + validator + author/read-protocol docs only. The Go CLI does
  not gate on the status enum, so it is unaffected. The `.tessl` mirror is gitignored
  and regenerated from source in CI.
