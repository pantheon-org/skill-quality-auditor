---
title: "ADR-049: a `value` frontmatter field makes \"highest value next\" a queryable axis"
status: accepted
date: 2026-07-06
context:
  - path: ".context/plans/context-prioritisation-signal-2026-07-06.md"
  - path: ".context/findings/prioritisation-signal-gap-2026-07-06.md"
  - path: ".context/instructions/value-rubric.md"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

`.context/` entries carried `status`, `date`, `effort` (plans), and `severity`
(known-issues), but no signal for *benefit of action*. So the recurring question
"which item is the highest value to do next?" was re-derived by ad-hoc agent
judgement every time it was asked, with no stored, auditable answer. The finding
`prioritisation-signal-gap-2026-07-06.md` documented the gap; the plan
`context-prioritisation-signal-2026-07-06.md` closes it. The design was settled
through a 3-reviewer plan-review and a guided-interview on four contentious forks.

## Decision

1. **Add a top-level `value` enum (`high`/`medium`/`low`)** to the frontmatter of
   the three action-candidate types — `plan`, `finding`, and `known-issue`. It is
   the benefit-of-action axis. `analysis`, `instruction`, and `audit` are reference
   material and do not carry it.

2. **Keep `effort`, `severity`, and `value` as three distinct axes; do not unify
   them into one number.** They measure different things: `severity` is
   risk-of-inaction, `effort` is cost-of-action, `value` is benefit-of-action.
   Cross-type comparability comes from all three being visible in the index, not
   from collapsing them.

3. **`value` is an authoritative sort key, not an advisory label.** The read
   protocol: filter to `draft`/`active` action-candidates, sort by `value`
   descending, then `effort` ascending where present, and act on the top item
   without re-forming an independent judgement. The index generator exposes the
   fields only; the consumer sorts (no materialised hint or composite rank).

4. **The rubric is canonical and lives at
   `.context/instructions/value-rubric.md`** — grading criteria (leverage,
   consumers unblocked, reversibility), worked examples, and the read protocol.
   `ways-of-working.md` and `AGENTS.md` link to it rather than restating it, so
   there is one source of truth.

5. **`value` is required while `status` is `draft` or `active`; `done` and
   `superseded` are exempt.** The operational sort only ever reads active/draft.
   Historical entries are graded once as a learning corpus, never for the live
   sort. Enforcement lives in `validate-context-frontmatter.sh`.

### Why an ADR now, when `effort` never got one

The earlier `effort` field was added without an ADR. That was an omission, not a
precedent. `value` changes the frontmatter contract for three types, introduces
an authoritative read protocol, and codifies the three-axes model — decisions
that future work must not silently re-litigate. We are correcting practice going
forward: contract-level `.context/` conventions get an ADR.

## Consequences

- **Easier:** "What's next" becomes a deterministic query over `.context/index.yaml`
  instead of a judgement call. The signal is auditable and re-gradeable.
- **Easier:** A future "what's next" skill has a stable field and protocol to
  consume, plus a historical learning corpus (Phase 5) to calibrate against.
- **Harder / ongoing cost:** grades can go stale, so they must be re-graded on
  status transitions and material scope changes (documented in
  `ways-of-working.md`). Grade trust depends on the rubric, the backfill
  calibration pass, and the re-grade discipline — these are load-bearing.
- **Migration:** the field shipped optional-first (schema + validator accept it),
  the active/draft population was backfilled in one serialised pass, and only then
  did the validator flip to required — so the tree stayed green throughout.
- **Accepted redundancy:** a `known-issue` now carries both `severity` and `value`.
  They are distinct axes; the overlap is the cost of one unified sort across types.
