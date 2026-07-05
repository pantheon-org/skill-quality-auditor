# Scenario 05: Downstream Work Requires a Written Output

## User Prompt

"Interview me about our caching strategy decision, then write it up as a short decision brief I
can share with the team."

## Setup

The interview has already run through 2-3 questions (cache location, invalidation trigger), the
agent has presented a bulleted recap, and the user has just replied "Yes, that's all correct."

## Expected Behavior

1. Recognize that the user asked upfront for a written artifact ("write it up as a ... brief"),
   which is exactly the case where a file is warranted, not the default chat-only case.
2. After the recap is confirmed, produce a written brief (e.g. via a file-writing tool) summarizing
   the decision and the reasoning behind it, matching what was confirmed in the recap.
3. Do not stop at just a chat summary when the user explicitly asked for a shareable document.

## Success Criteria

- A file/document is produced (or clearly proposed as the next concrete step) after confirmation.
- The file's content matches the confirmed recap — no answer is dropped or changed.
- The agent does not stop at a chat-only summary and treat the "write it up" request as satisfied
  by the recap alone.

## Failure Conditions

- The agent gives only a chat-based summary and never produces a document despite the explicit
  request to "write it up."
- The written output contradicts or omits something the user confirmed in the recap.
- The agent asks a redundant question about whether a file is wanted after already being told.
