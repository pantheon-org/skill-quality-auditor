---
title: "Finding: what the historical value corpus reveals about the rubric and prioritisation patterns"
type: FINDING
status: ACTIVE
date: 2026-07-06
value: MEDIUM
themes:
  - GOVERNANCE
related:
  - ../plans/context-prioritisation-signal-2026-07-06.md
  - ../instructions/value-rubric.md
  - ../../docs/ADR/adr-049-context-value-frontmatter-field.md
---

Phase 5 of the prioritisation-signal plan graded every `done`/`superseded`
`plan`/`finding`/`known-issue` (34 entries) for `value`, using a lesser model
(Haiku) because these grades are calibration and training signal only and never
feed the live sort (Decision 14). This finding captures what the exercise taught
us. It is explicitly an **input to the design of the future "what's next" skill**,
the natural consumer of the `value` signal.

## The two grading passes diverged sharply

| Pass | Grader | high | medium | low |
| ---- | ------ | ---- | ------ | --- |
| Live backfill (active/draft, Phase 3) | Opus, careful + calibration pass | 3 (7%) | 20 (47%) | 20 (47%) |
| Historical corpus (done/superseded, Phase 5) | Haiku, coverage-first | 12 (35%) | 21 (62%) | 1 (3%) |

Same rubric, materially different distributions. The lesser model graded ~5x more
items `high` and almost never used `low`. This is the single most important datum:
**the `low` bucket is fragile.** Without a careful calibration pass, a grader
gravitates to `medium`/`high` and the `low` criteria ("narrow, self-contained,
deferrable, a closed investigation") go under-applied. The sort degrades toward
"everything is high," which is the exact failure the field exists to prevent.

Implication for the "what's next" skill: grade quality is load-bearing for the
live sort (Decision 10), so the live grading path must not use a coverage-first
lesser model, and the calibration pass (Decision 8) is not optional. The corpus
is where coverage-over-precision is acceptable; the live sort is not.

## The rubric has one genuine ambiguity: findings whose action lives in a plan

In the live backfill we graded a finding `low` when its action was captured in a
paired plan — the benefit-of-action sits in the plan, not in re-reading the
diagnosis. The historical grader did not apply this heuristic: it graded findings
like `code-review`, `go-code-review`, and `plan-status-drift` `high` on their
intrinsic merit. Both readings are defensible because **the rubric does not say
which one is correct.** This matters directly for "what's next": if a finding and
its plan both grade `high`, they duplicate in the sort. The rubric needs an
explicit rule — grade a finding net of where its action is captured, or grade it
intrinsically and let the consumer de-duplicate finding/plan pairs.

## Prioritisation patterns worth encoding

- **Foundational/leverage work genuinely got done first.** The high tier is
  dominated by infrastructure: CLI consolidation, index infrastructure, the native
  eval runner, post-merge status sync, flag standardisation. The project's revealed
  preference matches the rubric's leverage-first ranking — a reassuring signal that
  `value` as defined tracks what actually gets prioritised.
- **Campaigns hide in per-item grades.** The nine dimension-improvement plans
  (D1-D8) each grade `medium` in isolation, D9 `high`. But as a batch they were the
  core deliverable of the framework. A per-item sort misses the campaign. The
  "what's next" skill may need to reason about clusters, not just individual rows.
- **Enabling vs. documenting is cleanly separable.** The rubric correctly put
  research/decision artifacts (e.g. `skill-validator-vs-native-eval`) at `low` and
  enabling work at `high`. That distinction is the rubric's strongest axis.

## Hindsight bias was low but not absent

Only one grade carried a hindsight flag (the hk/markdownlint DX migration, graded
`medium` with a note that shipping made it look more important than it was). But
the high skew of the whole historical pass is itself a soft form of hindsight:
completed foundational work reads as obviously-important in retrospect. Treat the
35% `high` rate as inflated by outcome knowledge, not as a target for live grading.

## Recommended inputs to the "what's next" skill

1. Use a careful grading model plus a calibration pass on the live path; reserve
   lesser models for non-authoritative corpora.
2. Resolve the finding-vs-paired-plan ambiguity in the rubric before the skill sorts.
3. Support cluster/campaign reasoning, not just single-row ranking.
4. Expect and correct for a `low`-bucket-avoidance bias in any automated grader.
