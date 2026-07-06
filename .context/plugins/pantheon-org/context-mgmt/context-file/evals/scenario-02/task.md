# Scenario 02: Create a Finding from Research Output

## User Prompt

"I've evaluated three Go testing frameworks for the skill auditor project. Create a finding document capturing what I found."

## Input

Research notes:
- Evaluated: `testify`, `gotest`, and `is` for assertion-style testing in Go.
- **testify** — Most popular (16k+ stars), rich assertion set, suite support, mock package. Downside: heavy dependency tree, suite pattern encourages non-standard test structure.
- **gotest** — Minimal zero-dependency library, simple assertions, clean API. Downside: no mock support, smaller community.
- **is** — Minimal, zero-dependency, designed for table-driven tests. Downside: no suites, no mocks, very opinionated.
- **Recommendation:** Use `is` for new tests. It aligns with Go's philosophy of simple table-driven tests, adds zero dependency weight, and the project already uses table-driven patterns.
- Related to an existing plan: `.context/plans/improve-test-coverage-2026-06-29.md`

## Expected Behavior

1. Create a new `.context/finding` file under `.context/findings/` with type `FINDING` in frontmatter.
2. Frontmatter must include `title`, `type: FINDING`, `status`, `date`, and `related` linking to the existing plan.
3. The `title` should start with "Finding: ".
4. Include a **one-sentence summary** at the top (under the date/status line).
5. Include a **Summary** section with concise overview.
6. Include a **Detail** section covering all three frameworks evaluated.
7. Include a **Recommended Action** section recommending `is`.
8. Use the filename convention `topic-YYYY-MM-DD.md`.

## Success Criteria

- File created under `.context/findings/`.
- Frontmatter has title, type, status, date, and related fields.
- `related` field references `../plans/improve-test-coverage-2026-06-29.md`.
- One-sentence summary present at the top.
- Detail section covers all three frameworks.
- Recommended action clearly recommends `is` as the choice.

## Failure Conditions

- File created outside `.context/findings/`.
- Frontmatter missing `related` field (when related plan exists).
- No one-sentence summary.
- Detail section missing evaluation of one or more frameworks.
- Recommended action vague or missing.
