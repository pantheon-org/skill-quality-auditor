---
title: "Known Issue: plan-review's reviewer prompts never ask where a mechanism actually executes"
type: KNOWN_ISSUE
status: ACTIVE
date: 2026-07-06
value: MEDIUM
severity: MEDIUM
related:
  - ../findings/docs-drift-cumulative-mode-ci-gap-2026-07-06.md
  - ../plans/docs-drift-reviewed-baseline-2026-07-06.md
  - ../../.context/plugins/pantheon-org/planning/plan-review/SKILL.md
---
# Known Issue: plan-review's reviewer prompts never ask where a mechanism actually executes

> `.context/plans/docs-drift-reviewed-baseline-2026-07-06.md`'s own Goal section stated, in plain text, that cumulative mode was "consumed by `hk.pkl`'s `pre-push` hook, its only caller" — a fact that, if traced to its implication, would have surfaced that the reviewed-baseline mechanism being planned could never run in CI. None of the plan-review's three reviewer prompts (Technical: feasibility/gaps; Strategic: goal alignment/completeness; Risk: blind spots/failure modes) asked "where does this actually execute, and who bypasses it" — so the fact sat in the plan, unexamined, until a direct user question surfaced it after implementation had already shipped (`.context/findings/docs-drift-cumulative-mode-ci-gap-2026-07-06.md`).

## Why this exists

`plan-review`'s three lenses are thorough on a plan's internal mechanics — the same review that missed this execution-location question also caught a genuine rebase-timing bug and a timezone-fragile date-comparison design flaw in the same plan. The gap is narrower than "the review missed things": it's that no lens's prompt explicitly directs a reviewer to trace a stated invocation fact to its coverage/bypass implications. A mechanism-internals review does not automatically ask a mechanism-external-reach question.

## Impact if unfixed

The same class of miss can recur on any future plan-reviewed change: a plan can correctly state where something runs, have that fact go unexamined by all three reviewer lenses, and ship with a coverage gap nobody flagged — exactly what happened here, caught only by luck (a user asking a direct, specific question) rather than by the review process designed to catch this kind of thing.

## Suggested fix (not yet applied — this is the tracked issue, not the fix)

Add an explicit instruction to one or more of `plan-review`'s reviewer prompts (Strategic and/or Risk are the natural fits) directing the reviewer to trace any stated invocation/trigger fact in the plan to its coverage implications — e.g. "if the plan states where or how this mechanism executes, explicitly assess who bypasses that path and whether the stated goal is actually met given that reach." Not applied here because it requires deciding which lens(es) should own this and how to phrase it without diluting the existing lens focus — a small design question, not a large one, but a real decision rather than a one-line fix.
