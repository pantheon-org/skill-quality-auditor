# Scenario 03: Skip the Interview and Skip the Unsolicited File

## User Prompt

"What port does Postgres run on by default?"

## Expected Behavior

1. Recognize this is a purely factual question with one right answer, not a topic with a real decision space.
2. Do NOT launch the guided-interview protocol — answer directly.
3. Do not offer 3-4 "options" for a fact that has a single correct answer.
4. Do not write any file — a one-line factual answer needs no brief or document.

## Success Criteria

- The agent answers directly (5432) without asking any interview-style question first.
- No multiple-choice options are presented for the factual answer.
- No file is created or proposed.

## Failure Conditions

- The agent asks "which environment: dev, staging, or prod?" or similar interview-style options for a question that has one universal answer.
- The agent proposes writing a summary/brief document for a one-line factual answer.
- The agent treats this as a topic requiring a multi-question interview.
