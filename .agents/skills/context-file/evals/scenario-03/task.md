# Scenario 03: Create an Analysis from Audit Review

## User Prompt

"Write up the analysis from the code review I just did. It's about the CLI command structure — there are overlapping flag definitions and inconsistent naming."

## Input

Review notes:
- Reviewed: all cobra commands under `cmd/` directory.
- Finding 1: Three commands define a `--repo-root` flag with identical logic (duplicate code).
- Finding 2: Flag naming inconsistency — some use `--dry-run`, others `--dryrun`, one uses `--noop`.
- Finding 3: Short flag collision — `-s` used for both `--store` and `--skip-baseline`.
- All findings are from today, 2026-06-30.

## Expected Behavior

1. Create a new `.context/analysis` file under `.context/analysis/` with type `analysis` in frontmatter.
2. Frontmatter must include `title`, `type: analysis`, `status: done`, and `date`.
3. The `title` should use format "Topic Analysis — YYYY-MM-DD".
4. Include a **Summary** section with a brief overview.
5. Include a **Findings** section with at least 3 findings (one for each issue identified).
6. Include a **Conclusion** section summarising the outcome and next steps.

## Success Criteria

- File created under `.context/analysis/`.
- Frontmatter has title, type, status, and date; status is `done`.
- Findings section covers all 3 issues (duplicate flag, naming inconsistency, short flag collision).
- Conclusion section present with recommended next steps.
- Title format follows "Topic Analysis — YYYY-MM-DD".

## Failure Conditions

- File created under wrong `.context/` subdirectory.
- Frontmatter missing required fields.
- Status not set to `done`.
- Findings missing one or more of the 3 issues.
- No conclusion section.
- Title not following the correct format.
