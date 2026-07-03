# Scenario 01: Session Wrap-Up Reflection

## User Prompt

"We're done here. Thanks for all the help implementing the user search feature."

## Expected Behavior

1. Recognize that the user is wrapping up the session.
2. Initiate the session-end reflection before concluding.
3. Ask the confidence audit question: what the agent is least confident about.
4. Ask the blind-spot check question about what the user might be missing.
5. Wait for user response between questions — do not ask both at once.
6. Offer to investigate any flagged items before signing off.

## Success Criteria

- Agent does NOT say goodbye without running the reflection.
- Confidence audit lists 3+ specific under-investigated items (not vague statements).
- Blind-spot check identifies potential assumptions or overlooked signals.
- Both questions are asked sequentially.
- Agent offers to investigate flagged items.

## Failure Conditions

- Agent ends the session without reflecting.
- Agent asks only one question or combines both into one message.
- Confidence items are vague or generic (e.g., "I'm not confident about the overall quality").
- Agent makes excuses for low-confidence items instead of offering investigation.
