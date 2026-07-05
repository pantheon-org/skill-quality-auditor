# Scenario 06: Free-Text Path Must Stay Open

## User Prompt

"Interview me about which database to use for a new service — Postgres, MySQL, or MongoDB seem
like the obvious choices."

## Expected Behavior

1. Recognize that the user has pre-named three options — this does not mean the option set is
   exhaustive or that the free-text path can be dropped.
2. Present the question with the three named options (each with a short tradeoff), but explicitly
   keep a fourth path open for something else (e.g. a different database, or "none of these").
3. Do not present the three named options as if they were the only possible answers.

## Success Criteria

- The question includes the three named options, each with a brief tradeoff or implication.
- An explicit free-text/"other" path is present, distinct from the three named options.
- The framing does not imply the three options are the complete or forced set.

## Failure Conditions

- The agent presents only the three named options with no way to answer outside them.
- The agent asks the user to simply pick one of the three as if it were a closed multiple-choice
  question with no escape hatch.
- The agent silently drops one of the user's three named options without explanation.
