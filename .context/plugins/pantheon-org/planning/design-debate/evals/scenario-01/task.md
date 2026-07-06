# Scenario 01: Basic Two-Sided Debate

## User Prompt

"We're thinking about caching the results of our nightly report generator to disk instead
of recomputing it every run. Should we do this? I want some pushback before we commit."

## Expected Behavior

1. Restate the decision precisely: "Should the nightly report generator's output be
   cached to disk instead of recomputed every run?"
2. Investigate the actual codebase to ground the debate: how long does generation
   currently take, how often does it run, what would invalidate a cache, is there
   already a caching pattern elsewhere in the repo to be consistent with.
3. Compose one identical brief containing those grounding facts.
4. Assign opposing roles: Advocate (for caching) and Skeptic (against/status quo) at
   minimum.
5. Spawn both reviewers in parallel (single message, 2+ tool calls).
6. Synthesize a verdict: proceed, do_not_proceed_for_now, or proceed_with_modification —
   not a restatement of both sides with no resolution.
7. If the verdict is do_not_proceed_for_now, state a concrete revisit trigger.
8. Offer to persist the verdict as a finding.

## Success Criteria

- Agent restates the decision as a precise question before investigating.
- Agent gathers concrete facts from the actual codebase before spawning reviewers.
- Agent assigns explicit opposing roles (Advocate/Skeptic), not neutral/balanced framing.
- Both reviewers are spawned in parallel, not sequentially.
- The final response includes an actual verdict (one of the three enum values in spirit,
  even if not stated in those exact words) with a named rationale referencing specific
  points from the reviews.
- Agent explicitly asks whether to persist the verdict as a finding, rather than
  silently deciding either way.

## Failure Conditions

- Agent skips investigation and lets reviewers debate from assumption.
- Agent asks both reviewers to "give balanced feedback" instead of assigning opposing
  stances.
- Agent spawns reviewers sequentially.
- Agent presents both opinions and asks the user to decide, without rendering its own
  verdict.
- Agent silently persists (or silently skips persisting) the verdict without asking.
