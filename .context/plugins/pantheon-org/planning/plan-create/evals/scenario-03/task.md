# Scenario 03: Oversized Phase Splitting

## User Prompt

"I need a plan for a big migration. Phase 1 is: update all configs, write migration scripts, test on staging, deploy to prod, update docs, train the team, and archive old data."

## Expected Behavior

1. Recognise that the user's "Phase 1" has 7 tasks — at the limit but not over.
2. Ask the user whether these tasks are truly sequential or if some can run in parallel.
3. Suggest splitting into 2 phases: "Phase 1: Preparation" (configs, scripts, staging test) and "Phase 2: Rollout" (deploy, docs, training, archive).
4. Annotate parallel tasks as waves within each phase.
5. Each phase gets an exit criterion.

## Success Criteria

- Agent identifies that the single-phase approach has too many tasks.
- Agent proposes splitting into at least 2 phases.
- Each split phase has 3-4 tasks (within the 2-5 range).
- Agent annotates which tasks can run in parallel (waves).
- Each phase has a concrete exit criterion.

## Failure Conditions

- Agent accepts the single-phase structure without question.
- Agent creates phases with more than 8 tasks.
- Agent does not annotate parallel tasks.
- Agent does not define exit criteria for phases.