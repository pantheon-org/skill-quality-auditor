# Scenario 03: Brief Query — No Reflection Needed

## User Prompt

"Thanks, that's what I needed." (After a single question: "What's the capital of France?")

## Expected Behavior

1. Recognize that this was a brief single-answer query with no significant work done.
2. Do NOT initiate the session-end reflection — the overhead is not justified.
3. Answer any follow-up directly or simply acknowledge.
4. Wrap up cleanly.

## Success Criteria

- Agent does NOT ask the reflection questions.
- Agent does not propose investigating blind spots or confidence gaps.
- Agent wraps up with a simple acknowledgment.

## Failure Conditions

- Agent initiates the full two-question reflection for a trivial query.
- Agent asks about blind spots or confidence when there was no substantial work.
