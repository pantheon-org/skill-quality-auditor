# Scenario 01: Create a Plan from a Feature Request

## User Prompt

"Create a plan for adding a new `trend` command to the skill auditor CLI. It should show score history per skill with up/down arrows. I need to know the implementation steps, open questions, and what we're aiming for."

## Expected Behavior

1. Create a new `.context/plan` file under `.context/plans/` with type `plan` in frontmatter.
2. Frontmatter must include `title`, `type: plan`, `status: draft`, and `date`.
3. The `title` should start with "Plan: ".
4. Include a **Goal** section describing the desired end state (a trend command showing score history).
5. Include a **Steps** section with numbered implementation steps.
6. Include an **Open Questions** section for unresolved items (e.g., output format, date range, storage schema).
7. Use kebab-case filename with optional date, e.g., `add-trend-command-2026-06-30.md`.
8. Do NOT create the file under `.context/audits/` — use `.context/plans/`.

## Success Criteria

- File created under `.context/plans/` with `.md` extension.
- Frontmatter has all required fields: title, type, status, date.
- `status` is set to `draft`.
- Goal section describes the trend feature.
- Steps section has at least 3 numbered steps.
- Open Questions section includes at least 2 unresolved items.

## Failure Conditions

- File created under `.context/audits/` or `.context/findings/`.
- Frontmatter missing any required fields.
- Steps not numbered or missing.
- No Open Questions section.
- `status` set to `active` instead of `draft`.
- Filename not following kebab-case convention.
