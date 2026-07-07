---
title: "ADR-058: plan-review gains an execution-location lens, owned by Risk + Strategic"
status: proposed
date: 2026-07-07
context:
  - path: .context/plans/plan-review-execution-location-lens-2026-07-07.md
  - path: .context/known-issues/plan-review-execution-location-blind-spot-2026-07-06.md
  - path: .context/findings/docs-drift-cumulative-mode-ci-gap-2026-07-06.md
---

**Status:** Proposed
**Date:** 2026-07-07

## Context

`plan-review`'s three reviewer lenses (Technical, Strategic, Risk) are thorough on
a plan's internal mechanics but none is prompted to trace a stated
invocation/trigger/execution fact to its external reach — who or what bypasses that
path, and whether the plan's stated goal is actually met given that reach. This gap
let the docs-drift reviewed-baseline plan (ADR-045) ship a mechanism that could
never run in CI: the plan's own text stated it was consumed only by the pre-push
hook, and no lens examined the implication. The gap is recorded in
`plan-review-execution-location-blind-spot-2026-07-06` and the resulting bug in
`docs-drift-cumulative-mode-ci-gap-2026-07-06`.

The binding question the remediation plan had to settle was **which lens owns the
new check**. A three-role plan-review of the remediation plan itself surfaced a real
alternative: the Technical lens ("feasibility — does this work given where it is
invoked") is arguably a natural home, and adding the check there would give
two-angle coverage.

## Decision

Add the execution-location / coverage check to `plan-review` as an extension of
existing reviewer questions, **owned by the Risk lens (primary) with a one-clause
cross-reference on the Strategic lens; the Technical lens is left unchanged.**

- Risk lens: extend the existing BLIND SPOTS question with — *"If the plan states
  where or how a mechanism executes (a hook, a CI job, a trigger, a caller),
  explicitly assess who or what bypasses that path and whether the stated goal is
  actually met given that reach."*
- Strategic lens: append one clause to the COMPLETENESS question referencing
  whether stated execution paths reach far enough to meet the goal.
- The check is folded into existing questions rather than added as a new numbered
  question, to avoid reviewers treating it as optional boilerplate.

The **Technical-lens alternative was considered and rejected** to keep the change
minimal (no third prompt to edit); the Risk + Strategic pairing already covers the
"who bypasses this / is the goal met" question. A future ADR may extend the check to
the Technical lens if evals show the pairing misses feasibility-framed cases.

## Consequences

- **Easier:** future plan reviews are prompted to catch execution-reach / coverage
  gaps of the exact shape that shipped the docs-drift-in-CI bug, reducing the chance
  of a mechanism that "works" but never runs where the goal requires.
- **Easier:** the change is small and reversible — two prompt-string edits in one
  SKILL.md, guarded by a new eval scenario (`scenario-04`) and a regression re-run
  of the existing scenarios.
- **More difficult / risk:** adding text to the Risk BLIND SPOTS question risks
  diluting its existing focus; mitigated by folding into the existing item and by
  the scenario-01..03 regression check.
- **Deferred:** the Technical lens does not gain this check now; if the Risk +
  Strategic pairing proves insufficient, extending Technical is a follow-up decision,
  not a reversal of this one.
