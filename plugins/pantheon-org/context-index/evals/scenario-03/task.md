# Scenario 03: Answer "What Context Files Exist?"

## User Prompt

"What context files exist right now? Give me a summary."

## Expected Behavior

1. Read `.context/index.yaml` to answer the question.
2. Summarise context files grouped by type (plans, findings, analysis).
3. Include a status breakdown (how many draft, active, done) for each type.
4. Do NOT edit `.context/index.yaml` — it is regenerated and should not be hand-edited.
5. State that the source of truth is the frontmatter in each `.md` file, not the index.

## Input (index.yaml content)

```yaml
- title: "Plan: Add Structured Logging"
  type: plan
  status: draft
  date: 2026-06-30
- title: "Plan: Improve Test Coverage"
  type: plan
  status: active
  date: 2026-06-28
- title: "Finding: Logging Library Evaluation"
  type: finding
  status: active
  date: 2026-06-30
- title: "Finding: Go Version Evaluation"
  type: finding
  status: done
  date: 2026-06-29
- title: "CLI Flag Audit — 2026-06-30"
  type: analysis
  status: done
  date: 2026-06-30
```

## Success Criteria

- Index read to produce the summary.
- Summary grouped by type (plans, findings, analysis).
- Status breakdown included (e.g., 1 draft, 1 active for plans).
- States that index is regenerated and should not be hand-edited.
- Source of truth correctly identified as frontmatter in `.md` files.

## Failure Conditions

- Files listed individually without type grouping.
- No status breakdown.
- Index edited or modified in any way.
- States that index is the source of truth.
- Index not consulted (files listed from memory or directory listing instead).
