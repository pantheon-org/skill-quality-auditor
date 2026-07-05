# Scenario 01: One Question, Concrete Options

## User Prompt

"I need to pick a caching strategy for our API. Can you interview me about it? Ask one question at a time and give me a few options to choose from."

## Expected Behavior

1. Recognize this is a request for a guided interview about a specific, nameable topic.
2. Ask a single question first — not a list of several questions in one message.
3. The question offers 3-4 concrete, mutually exclusive options, each with a short description of what picking it implies.
4. The free-text/"other" path remains open — the user is not forced into only the listed options.
5. The agent waits for the answer before asking the next question.

## Success Criteria

- Exactly one question is present in the first turn.
- The question has 3 or 4 distinct, real options (not vague labels like "option A/B").
- Each option includes a brief tradeoff or implication, not just a name.
- A free-text or "other" path is explicitly available.

## Failure Conditions

- The agent asks two or more questions in the same turn.
- The agent offers only 2 options (false binary) or 5+ options (overload).
- Options are bare labels with no description of what they mean.
- The agent proceeds to recommend a caching strategy without asking anything first.
