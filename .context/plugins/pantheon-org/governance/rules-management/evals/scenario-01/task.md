# Scenario 01: Basic Rule Creation

## User Prompt

"Add a rule: never suggest unsolicited refactoring."

## Expected Behavior

1. Read the existing `.agents/RULES.md` file to check for duplicates.
2. Recognise the user's instruction as a rules-management request.
3. Append a new rule entry following the required format: `### Rule:` title, `**Directive:**` with NEVER/ALWAYS phrasing, and `**Rationale:**`.
4. Use NEVER phrasing for the directive (e.g., "NEVER suggest refactoring code that is not directly related to the task at hand").
5. Include a rationale that explains why — context-switching cost, scope creep, review burden.
6. Inform the user that the rule has been added.
7. Do NOT modify or remove any existing rules.

## Success Criteria

- Existing `.agents/RULES.md` content is preserved (no deletions or modifications to existing rules).
- A new rule entry is appended using the `### Rule:` / `**Directive:**` / `**Rationale:**` format.
- Directive uses NEVER or ALWAYS phrasing.
- Rationale explains the reasoning in 1–2 sentences.
- User is informed the rule was added.

## Failure Conditions

- Existing rules are modified or deleted.
- Rule added without the three-part format (title, directive, rationale).
- Directive lacks NEVER/ALWAYS phrasing.
- No rationale provided.
- User is not notified of the change.

**Context:**

- Repository root: current working directory
- Existing `.agents/RULES.md` already contains two rules (outdated dependencies, validate newly created SKILL files)
