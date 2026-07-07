---
title: "ADR-060: DEFERRED lifecycle status for parked context items"
status: proposed
date: 2026-07-07
context:
  - path: .context/plans/deferred-status-2026-07-07.md
---

**Status:** Proposed
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

Add `DEFERRED` as a fifth lifecycle status: a real item intentionally parked
(date-gated, externally blocked, or deprioritised), distinct from `ACTIVE`
(pick-up-next) and `DRAFT` (not yet reviewed).

- **Read protocol ranks `DEFERRED` in a strict second tier.** Candidates split into
  tier 1 (`DRAFT`/`ACTIVE`) and tier 2 (`DEFERRED`); tier 1 is always exhausted
  before tier 2. Within each tier the existing sort applies (`value` descending,
  `effort` ascending, `themes[0]`). A `DEFERRED` item never outranks a tier-1 item
  regardless of its `value`. Reactivate to `ACTIVE` when the blocker clears.
- **`value`, `themes`, and `effort` remain required on `DEFERRED`** (same as
  `DRAFT`/`ACTIVE`), enforced by `validate-context-frontmatter.sh`, so a parked item
  re-ranks cleanly on reactivation. Only `DONE`/`SUPERSEDED` are exempt.
- Excluding `DEFERRED` from the pick entirely (listing it separately) was considered
  and rejected: the intent was "lower priority than active, still in the same queue".

## Consequences

- **Easier:** the read protocol stops recommending gated or blocked work; parked
  items stay visible and fully graded, ready to re-rank the moment they reactivate;
  the `ACTIVE` set once again means "work you would genuinely pick up next".
- **Harder:** one more status to reason about, and a re-grade obligation on
  `ACTIVE ↔ DEFERRED` transitions (added to the ways-of-working re-grade rule).
- **Scope:** schema + validator + author/read-protocol docs only. The Go CLI does
  not gate on the status enum, so it is unaffected. The `.tessl` mirror is gitignored
  and regenerated from source in CI.
