---
title: "Finding: session-reflection never checks for the agent's own repeated procedures"
type: finding
status: active
date: 2026-07-06
related:
  - ../plugins/pantheon-org/workshop/session-reflection/SKILL.md
  - ../findings/pr-merge-validation-gap-2026-07-06.md
---

# Finding: session-reflection never checks for the agent's own repeated procedures

> `session-reflection`'s two questions target content and shared-understanding risk.
> Neither one orients toward "have I, the agent, manually repeated the same multi-step
> procedure enough times that it should be a skill?" This session ran the same
> validate-and-merge sequence by hand at least eight times without reflection ever
> catching it.

## Summary

Asked directly (by the user, not surfaced by the tool) why the validate-and-merge
pattern documented in `.context/findings/pr-merge-validation-gap-2026-07-06.md` was
never flagged for automation despite running repeatedly. Root cause: `session-reflection`
has no question aimed at the agent's own procedural repetition.

## Detail

`session-reflection`'s two questions, as designed:

1. **Confidence audit** — "What am I least confident about right now?" Oriented at
   under-investigated *content*: assumptions, unverified dependency versions, skipped
   edge cases.
2. **Blind-spot check** — "What's the biggest thing I'm missing about this situation?"
   Oriented at *shared understanding* with the user: unexamined assumptions, unexplored
   alternatives, dropped signals.

Both ran multiple times this session. Neither ever pointed inward at "what have I, the
agent, done the same way more than once by hand." The validate-and-merge sequence
(poll `gh pr checks` until every check is terminal, resolve any generated-file conflict
via the regenerate-don't-hand-merge sequence, then `gh pr merge`) repeated at least
eight times (PRs #187–#199) without producing a confidence-audit item or a blind-spot
item, because every individual run *succeeded* — nothing broke, so nothing looked wrong
enough to interrupt the flow and check "have I typed this exact multi-step sequence
before?"

This is a distinct failure mode from what the skill already catches. A confidence-audit
item is "I did X and I'm not sure it's right." A blind-spot item is "the user might not
know Y." Neither shape captures "I keep doing Z the same manual way and haven't asked
whether Z should be a skill" — that observation isn't about correctness or shared
understanding, it's about the agent's own workflow shape over time, which the skill's
question design doesn't reach.

The other repeated-pattern catches this session (the `.tessl` mirror `.aislop`
exclusion, the jq hard-dependency bug, formalizing `design-debate`) were all triggered
by something breaking or by an explicit user ask — never by `session-reflection` asking
"what have I repeated?"

## Next Steps

Draft a plan to add a procedural-repetition check to `session-reflection` — either a
third question or a fold-in to the existing blind-spot check — that prompts the agent
to scan its own session summary for a manual, multi-step sequence repeated 2+ times and
flag it as a skill-formalization candidate, mirroring the finding-to-plan-to-skill
pattern already used for `design-debate` and (pending) `pr-merge`.
