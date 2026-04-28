# Scenario 04: Anti-Pattern Documentation Enhancement

## User Prompt

"Rewrite this SKILL.md anti-patterns section to score D3: >=13/15."

## Expected Behavior

1. Recognise that the three existing bullets lack the required NEVER/WHY/BAD/GOOD structure and
   cannot score above D3: 5/15 in their current form.
2. Rewrite each of the three original issues as a full anti-pattern entry: `NEVER:` statement,
   `WHY:` explanation, fenced `BAD` code block, fenced `GOOD` code block.
3. Add at least one new anti-pattern entry (4+ total) relevant to CI/CD pipeline generation.
4. Cover all three original issues: secrets hardcoding, parallel dependency ordering, and floating action version tags.
5. Produce `anti-patterns-section.md` containing the complete rewritten section, ready to paste directly into SKILL.md.
6. Produce `before-after-diff.md` explaining which D3 scoring signals the new content satisfies
   and estimating the score delta.

## Success Criteria

- Each anti-pattern leads with `NEVER: [action]` phrasing.
- Each NEVER statement is immediately followed by a `WHY:` line explaining the consequence.
- Each anti-pattern includes a fenced `BAD` code block showing the incorrect pattern.
- Each anti-pattern includes a fenced `GOOD` code block showing the correct alternative.
- All three original issues covered: secrets hardcoding, parallel dependency ordering, floating action tags.
- Section contains 4 or more distinct NEVER entries (original 3 + at least 1 new).
- `before-after-diff.md` explains which D3 signals are satisfied and estimates the score delta.

## Failure Conditions

- Bullet-point "don't do X" format retained without NEVER/WHY/BAD/GOOD structure.
- Fewer than 4 anti-patterns produced.
- BAD or GOOD code examples missing for any entry.
- One or more of the three original issues not addressed.
- `before-after-diff.md` absent or does not explain which D3 signals are satisfied.

**Current content to rewrite:**

```markdown
## Common Mistakes

- Don't hardcode secrets in workflow files
- Avoid running all jobs in parallel when they have dependencies
- Don't use latest tags for actions
```
