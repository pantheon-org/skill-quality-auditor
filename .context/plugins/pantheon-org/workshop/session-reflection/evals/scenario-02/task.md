# Scenario 02: Sub-Agent Spawn for Deep Session

## User Prompt

"Before I go, can you do that reflection thing where you get a second model to review the session? We did a lot of work today on the auth refactor."

## Expected Behavior

1. Recognize that the user is requesting the session-end reflection in sub-agent spawn mode.
2. Compose a detailed session summary covering: what work was done (auth refactor), files touched, assumptions made, what was skipped or deferred.
3. Spawn a sub-agent with the standard reflection prompt template.
4. The prompt must include both the confidence audit and blind-spot check questions.
5. Present the sub-agent's output as an independent review, clearly attributing it.
6. Offer to investigate any flagged items.

## Success Criteria

- Agent composes a session summary before spawning the sub-agent.
- Sub-agent is spawned with a prompt containing both reflection questions.
- Results are presented with clear attribution ("I asked a second agent to review...").
- Agent offers to investigate flagged items.

## Failure Conditions

- Agent does the reflection inline instead of spawning a sub-agent.
- Agent spawns sub-agent without a session summary (vague or empty prompt).
- Agent presents sub-agent output as its own analysis.
- Agent skips the investigation offer.
