---
title: "Value Rubric and Read Protocol"
type: INSTRUCTION
status: ACTIVE
date: 2026-07-06
related:
  - ../plans/context-prioritisation-signal-2026-07-06.md
  - ../findings/prioritisation-signal-gap-2026-07-06.md
---

The `value` frontmatter field records the **benefit-of-action** of a `.context/`
action-candidate entry: how much doing it unblocks or leverages future work. It
is one of three distinct axes, and must not be conflated with the other two:

| Axis | Question it answers | Field |
| ---- | ------------------- | ----- |
| Benefit-of-action | How much good does doing this do? | `value` |
| Cost-of-action | How much work is it? | `effort` (plans only) |
| Risk-of-inaction | How bad is leaving it undone? | `severity` (known-issues only) |

`value` applies to the three action-candidate types: `PLAN`, `FINDING`, and
`KNOWN_ISSUE`. It does not apply to `ANALYSIS`, `INSTRUCTION`, or `AUDIT`, which
are reference material rather than things to do next.

This rubric is the standard that all `value` grades are assigned against. Author
grades before the sort trusts them: `value` is an authoritative sort key (see the
read protocol below), not an advisory label.

## Grading criteria

Grade against three questions, in priority order. When they disagree, leverage
dominates.

1. **Leverage** — Does completing this unblock or simplify other work, or is it a
   leaf that helps only itself? High-leverage items are foundational: other plans
   or skills depend on them, or they retire a class of recurring effort.
2. **Consumers unblocked** — How many future work items, skills, or people can
   proceed once this lands? Count concrete downstream dependents, not hypotheticals.
3. **Reversibility and decay** — Cheap-to-reverse, low-decay work is safer to rate
   up. Work whose benefit evaporates if delayed (a time-boxed fix, a grade that
   goes stale) may warrant a higher grade to capture the closing window.

### `HIGH`

Foundational or broadly-leveraged: several downstream items depend on it, or it
retires a recurring cost, or it closes a gap that keeps re-manifesting. Doing it
changes what else becomes possible.

### `MEDIUM`

Clear standalone benefit with limited leverage: it improves one workflow, closes
one gap, or unblocks one or two consumers, but nothing else is waiting on it.

### `LOW`

Narrow, self-contained, or nice-to-have: benefits a single consumer, is easily
deferred, or is polish rather than capability. Correct to do eventually, not
urgent to do next.

## Worked examples

These grade real `.context/` items against the criteria above.

- **`plans/context-prioritisation-signal-2026-07-06.md` → `HIGH`.** High leverage:
  it produces the `value` signal that a future "what's next" skill and every
  future prioritisation call will consume, and it retires the recurring ad-hoc
  "which is highest value?" judgement. Multiple downstream consumers; foundational.
- **`plans/pr-merge-skill-2026-07-06.md` → `MEDIUM`.** Real standalone benefit
  (closes the `pr-merge-validation-gap` finding and removes a repeated manual
  chore) but nothing else is blocked on it; it unblocks one workflow, not a class
  of future work.
- **`findings/index-yaml-split-review-2026-07-06.md` → `LOW`.** A reviewed-and-
  decided-against investigation. Self-contained, no downstream consumer waiting,
  no recurring cost retired; it documents a closed question rather than enabling
  new work.

## Read protocol (Decision 10)

`value` is an **authoritative sort key**, not an advisory label. To answer "which
item is highest value to do next?", read `.context/index.yaml` and:

1. Filter to `status` in {`DRAFT`, `ACTIVE`} of type `PLAN`, `FINDING`, or
   `KNOWN_ISSUE`. (`DONE`/`SUPERSEDED` grades exist as a learning corpus and never
   enter this sort.)
2. Sort by `value` descending (`HIGH` > `MEDIUM` > `LOW`).
3. Break ties by `effort` ascending (`S` < `M` < `L` < `TBD`) where present.
   Findings and known-issues have no `effort`, so within a bucket they sort by
   `value` alone.
4. Act on the top item **without re-forming an independent judgement**. Relocating
   the judgement to read-time would reopen the gap this field closes.

This protocol only holds if the grades are trustworthy. That is why grading against
this rubric (not ad hoc), the calibration pass on backfill, and re-grading on status
transitions are load-bearing, not optional.
