# Scenario 04: Plan whose mechanism runs only in a narrow execution path

## User Prompt

"Please review the plan at .context/plugins/pantheon-org/planning/plan-review/evals/scenario-04/fixture-plan.md"

## Expected Behavior

1. Read the plan file and compose a self-contained brief.
2. Check model config / ask about models, then spawn the 3 reviewers in parallel.
3. In the consolidated report, at least one reviewer (Risk via BLIND SPOTS, and/or
   Strategic via COMPLETENESS) traces the plan's stated execution path — the link
   checker runs only from the `pre-push` hook, which is its "only caller" — to its
   coverage implication.
4. That reviewer flags that the stated goal ("no broken links ever reach the
   published site") is not actually met: `pre-push` is skippable (`--no-verify`)
   and does not run on merges through CI / the forge, so broken links can still
   ship without ever triggering the checker.

## Success Criteria

- Agent runs the standard plan-review workflow (brief, model step, 3 parallel
  reviewers, consolidated report).
- The report explicitly identifies that the checker's only trigger is the
  `pre-push` hook.
- The report reasons about who or what bypasses that trigger (local `--no-verify`,
  CI / web merges) rather than only critiquing the plan's internal mechanics.
- The report concludes the stated "nothing broken ever ships" goal is not met by a
  pre-push-only mechanism, and recommends a path that also runs where code actually
  ships (e.g. a CI job).

## Failure Conditions

- The review passes the plan without noting the pre-push-only execution path.
- Reviewers critique only internal mechanics (script parsing, fixture location)
  and never question whether the mechanism's reach meets the goal.
- The coverage/bypass finding is absent from the consolidated report.
