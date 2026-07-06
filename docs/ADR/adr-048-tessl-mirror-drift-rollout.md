---
title: "ADR-048: .tessl mirror drift check rollout — advisory-first, gated spike, sibling gating"
status: accepted
date: 2026-07-06
context:
  - path: ".context/plans/tessl-mirror-drift-check-2026-07-06.md"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

`.context/plans/tessl-mirror-drift-check-2026-07-06.md` implements ADR-047's decision
(verify `.tessl/plugins/pantheon-org/**` against `.context/plugins/pantheon-org/**` via
a CI-only diff, not tracked in git). The plan went through a 3-reviewer plan-review
(Technical, Strategic — Claude Sonnet 5; Risk — Claude Haiku 4.5). All three reviewers
independently converged on the same two blocking gaps: the plan asserted hard-fail
enforcement as settled in its Scope section while simultaneously re-opening the same
question in Open Questions, and the assumption that `tessl install` needs no
`TESSL_TOKEN` for `file:`-sourced plugins was inferred from `tessl install --help`
output only, never verified against a real install in a clean CI runner. Three
resolved decisions came out of a follow-up interview and are recorded here rather than
left as implementation detail, matching this repo's convention (ADR-044, ADR-045) of
documenting choices a plan-review surfaces.

## Decision

1. **Advisory rollout, not hard-fail from day one.** The new CI step runs with
   `continue-on-error: true` initially. **Revisit trigger:** after 5 merged PRs that
   exercise the step (or 2 weeks, whichever comes first) with zero false positives, a
   follow-up PR removes `continue-on-error` and the step becomes a hard gate, matching
   ADR-047's original enforcement framing. This is a deliberate, temporary deviation
   from ADR-047's "enforcement, not advisory" framing — justified because the mechanism
   itself (a brand-new CI step, an unverified `tessl install` auth path) is unproven in
   real CI, and zero divergence exists today so there is nothing to protect by rushing
   to hard-fail. Recording the trigger here, not just in the plan, exists specifically
   so "advisory" cannot silently become "permanent" the way
   `.context/known-issues/known-issues-lack-enforcement-2026-07-06.md` already warned
   generic known-issue revisit triggers can.
2. **The step gates on `if: steps.helper_skills.outputs.count != '0'`**, matching the
   existing sibling steps in `quality-gate` (duplication check, batch eval, structural
   eval) exactly, rather than running unconditionally on every job invocation. Accepted
   trade-off: a PR touching only `cmd/assets/**` will not re-verify the mirror on that
   run. Chosen over "always run" because consistency with the job's existing steps
   outweighs closing that narrow gap for a first rollout — it can be revisited once the
   advisory-to-hard-fail flip (Decision 1) happens and the check has a track record.
3. **A gated spike (the plan's Phase 2) runs before any CI-wiring code is written.**
   The spike does nothing but confirm `npm install -g tessl && tessl install` succeeds
   with no `TESSL_TOKEN` in a clean GitHub Actions runner. Chosen over "attempt Phase 3
   directly and handle a failure inline" because the existing `tesslio/setup-tessl`
   action (which does have a working token-based path) is explicitly slated for
   removal — if the tokenless assumption is wrong, discovering that only after the
   CI-wiring PR is already written costs a rewrite, not just a re-run.

## Consequences

- **Easier:** the rollout can absorb an unexpected CI-environment surprise (auth,
  `tessl` version drift, line-ending or symlink edge cases in the diff) without ever
  blocking a real merge during the advisory window.
- **Easier:** the spike isolates the one genuinely unverified assumption (tokenless
  `tessl install` in clean CI) from everything else in the plan, so a wrong assumption
  is caught before, not during, CI-wiring implementation.
- **Harder:** enforcement is not immediate — real drift introduced during the advisory
  window would be visible in CI logs but would not block a merge. Mitigated only by the
  revisit trigger in Decision 1, which is a process commitment, not a technical
  safeguard.
- **Binding for future work:** the follow-up PR that removes `continue-on-error` must
  actually happen per the stated trigger; if this repo's plan- or ADR-review process
  revisits `.context/plans/tessl-mirror-drift-check-2026-07-06.md` and finds that PR
  still outstanding well past the trigger, treat it as overdue technical debt, not a
  quietly-accepted permanent state.
