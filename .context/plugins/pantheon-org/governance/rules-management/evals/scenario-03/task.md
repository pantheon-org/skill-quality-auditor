# Scenario 03: Rule Format Enforcement

## User Prompt

"Make a rule that we should use structured logging everywhere."

## Expected Behavior

1. Recognise that the user's instruction is a valid trigger for rule creation.
2. Read the existing `.agents/RULES.md` to check for duplicates.
3. Recognise that the user's phrasing is vague — "use structured logging everywhere" is not a precise behavioural directive.
4. Reformulate the user's intent into a properly structured rule entry:
   - Title: "Always use structured logging in production services"
   - Directive: "ALWAYS use structured logging (JSON or equivalent) for production services. NEVER use fmt.Println or log.Printf for operational logging."
   - Rationale: explaining why — log aggregation, searchability, structured alerting
5. Optionally discuss the scope with the user before writing.
6. Append the reformulated rule to `.agents/RULES.md`.

## Success Criteria

- Existing rules are preserved.
- Title uses imperative phrasing appropriate for a rule heading.
- Directive uses ALWAYS and/or NEVER phrasing.
- Rationale is provided in 1–2 sentences.
- The rule is concrete and actionable — not just "use structured logging" but what to do and what to avoid.
- User is informed the rule was added.

## Failure Conditions

- Rule added verbatim from the user without reformulation ("we should use structured logging everywhere" as the directive).
- Missing NEVER/ALWAYS phrasing in the directive.
- No rationale provided.
- Rule too vague to be actionable ("Use good logging practices").
- Existing rules modified or deleted.
- User not notified.

**Context:**

- Repository root: current working directory
- Existing `.agents/RULES.md` contains:
  - Rule: Never advise outdated dependencies
  - Rule: Always validate newly created SKILL files
