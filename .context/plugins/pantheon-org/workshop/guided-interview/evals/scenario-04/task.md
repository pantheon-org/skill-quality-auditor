# Scenario 04: Override Signal — "Just Decide for Me"

## Setup

A guided interview is underway about choosing a logging format for a new service. The agent
asked its first question (structured vs. plain-text logs) and got an answer (structured/JSON).
The agent then asked a second question about log verbosity level.

## User Prompt (response to the second question)

"Honestly, just decide for me — I don't want to answer more questions about this."

## Expected Behavior

1. Recognize this as an explicit override signal to stop the interview.
2. Do NOT ask the planned question again or any further branching question.
3. State the reasonable default assumption(s) being used in place of the unanswered question(s).
4. Move directly to a recap/summary and final output — skip the confirmation-of-recap step only
   if the override itself already covers it, otherwise keep the recap brief.

## Success Criteria

- No further questions are asked after the override signal.
- The agent explicitly states what default it is assuming for the skipped question(s).
- The agent proceeds to a final synthesized answer rather than looping back into more questions.

## Failure Conditions

- The agent asks another question (e.g. "just to clarify, do you mean...") after the override.
- The agent silently picks a default without telling the user what it assumed.
- The agent apologizes repeatedly or argues against the override instead of respecting it.
