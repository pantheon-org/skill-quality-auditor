---
title: "Finding: critical review of the DEFERRED lifecycle status (ADR-060)"
type: FINDING
status: DONE
date: 2026-07-07
related:
  - ../../docs/ADR/adr-060-deferred-lifecycle-status.md
  - ../plans/deferred-status-2026-07-07.md
  - ../known-issues/deferred-no-forcing-function-2026-07-07.md
  - ../known-issues/deferred-no-reactivation-sweep-2026-07-07.md
---

# Finding: critical review of the DEFERRED lifecycle status

A critical review of ADR-060 (DEFERRED status) surfaced five points. This finding
records the review and how each point was resolved. It is a completed record; the
two residual gaps are tracked as their own known-issues.

## What holds up

- The **strict tier** (DEFERRED ranked below all DRAFT/ACTIVE, never by value) is
  correct: a blocked high-value item genuinely cannot be done, so it should not
  outrank actionable low-value work. A value-penalty model would let a big blocked
  item float above small actionable work.
- Keeping `value`/`themes`/`effort` required is right for clean reactivation.
- Scope discipline (no Go change, mirror left to CI, immutable ADR + source plan).

## Findings and resolutions

1. **No forcing function (most serious).** Nothing stops a blocked item being filed
   `ACTIVE`, so the pollution the status targets persists whenever an author does not
   reach for `DEFERRED`. Same class as `known-issues-lack-enforcement`. *Resolution:*
   the ADR Consequences were softened to state this honestly (from "stops
   recommending gated work" to "lets authors keep gated work out of the pick"); the
   gap is tracked in
   [`deferred-no-forcing-function-2026-07-07`](../known-issues/deferred-no-forcing-function-2026-07-07.md).
2. **No reactivation trigger / staleness.** Date-gated items had no mechanism to
   return to ACTIVE when ripe, risking a tier-2 graveyard. *Resolution:* added an
   optional `deferred_until` field plus a read-protocol rule that surfaces ripened
   items as reactivation-eligible. Automating a sweep remains a gap, tracked in
   [`deferred-no-reactivation-sweep-2026-07-07`](../known-issues/deferred-no-reactivation-sweep-2026-07-07.md).
3. **Semantic overload — "deprioritised" collided with the `value` axis.** *Resolution:*
   narrowed the definition to "cannot be actioned yet" (date-gated / externally
   blocked); low-priority work stays `ACTIVE` with `value: LOW`. Schema, validator
   docs, and all read-protocol docs updated to match.
4. **Zero adoption shipped.** The motivating items were still `ACTIVE`. *Resolution:*
   the plan-review execution-location lens is migrated to `DEFERRED` in the same
   changeset; tessl-eval-decommission Bucket A is migrated on its own branch to avoid
   a cross-branch status conflict.
5. **Tiering is unenforced convention.** The pick is prose an agent follows against
   `index.yaml`; nothing computes tiers. *Resolution:* accepted as consistent with
   the pre-existing protocol; folded into the enforcement gap (point 1) rather than
   built as new automation.

## Decision on naming

Kept the name `DEFERRED` (it reads as "postponed", which fits date-gated and blocked)
rather than renaming to `BLOCKED`, once the definition was narrowed. The narrowing,
not the name, was the fix for the overload.
