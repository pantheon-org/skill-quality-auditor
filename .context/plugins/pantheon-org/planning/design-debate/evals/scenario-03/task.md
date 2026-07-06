# Scenario 03: Trivial Decision Boundary and Proceed Hand-off

## User Prompt (Part 1)

"Should we rename the local variable `tmp` to `tmpFile` in `scripts/install.sh` for
clarity? Run a design-debate on it."

## Expected Behavior (Part 1)

1. Recognize this is a trivial, fully-reversible decision with no real tradeoff on either
   side — exactly the case the skill's "When NOT to Use" section calls out.
2. Decline to run a full multi-agent debate for it, explaining why (no real tradeoff to
   debate), and either just make the call directly or ask if the user wants something
   more substantive debated instead.

## User Prompt (Part 2, after Part 1)

"Fair enough. Let's actually debate: should we replace our hand-rolled JSON parsing in
`internal/llmclient` with a well-known third-party library?"

## Expected Behavior (Part 2)

1. Recognize this now has a real tradeoff (dependency risk vs. maintenance burden) and
   run the full pattern: ground in facts, assign opposing roles, spawn in parallel,
   synthesize a verdict.
2. If the verdict is "proceed," explicitly hand off to `plan-create` next, using the
   debate's grounding facts and chosen approach as input — not re-deriving them from
   scratch in a separate step.

## Success Criteria

- Agent declines (or pushes back on) running the full debate pattern for the trivial
  rename, citing the lack of a real tradeoff.
- For the second, substantive question, the agent runs the full pattern: grounding,
  opposing roles, parallel spawn, synthesized verdict.
- If the verdict is "proceed," the agent explicitly names `plan-create` as the next step
  and references the debate's facts/design as its input.

## Failure Conditions

- Agent runs a full 2-3-agent debate for the trivial variable rename.
- Agent treats both prompts identically without distinguishing trivial from substantive.
- Agent reaches a "proceed" verdict on the second question but doesn't connect it to a
  concrete next step (`plan-create`), leaving the user to figure out what happens next.
