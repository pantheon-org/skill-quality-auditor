# Scenario 02: Duplicate Rule Prevention

## User Prompt

"New rule: always check the latest version before recommending a dependency."

## Expected Behavior

1. Read the existing `.agents/RULES.md` file first.
2. Identify that a rule with substantially the same directive already exists ("Never advise outdated dependencies" — checking latest version before recommending is the same concern).
3. Refuse to add a duplicate rule.
4. Inform the user that the rule already exists, referencing the existing rule by title.
5. If the user's instruction adds nuance, suggest amending the existing rule instead of creating a new one.
6. Do NOT create a duplicate rule entry.

## Success Criteria

- Existing `.agents/RULES.md` is read before any write operation.
- The overlap between the user's request and the existing rule is detected.
- No duplicate rule is created.
- User is informed the rule already exists, with a reference to the existing rule title.
- File content is unchanged (no new rule appended).

## Failure Conditions

- A duplicate rule is appended despite the existing one.
- Existing `.agents/RULES.md` is not read first.
- User is not informed of the existing rule.
- File content is modified unnecessarily.

**Context:**

- Repository root: current working directory
- Existing `.agents/RULES.md` contains:
  - Rule: Never advise outdated dependencies (Directive: NEVER recommend a dependency version without first verifying it is the latest stable release)
  - Rule: Always validate newly created SKILL files
