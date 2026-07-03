# Scenario 01: Regenerate the Index After Adding Files

## User Prompt

"I just created two new context files. Regenerate the index so they show up."

## Input

Two new files were created:

**`.context/findings/logging-evaluation-2026-06-30.md`**:
```yaml
---
title: "Finding: Logging Library Evaluation"
type: finding
status: active
date: 2026-06-30
---
```

**`.context/plans/add-structured-logging.md`**:
```yaml
---
title: "Plan: Add Structured Logging"
type: plan
status: draft
date: 2026-06-30
---
```

## Expected Behavior

1. Run `regenerate-context-index.sh` to update `.context/index.yaml`.
2. Verify both new files appear in the regenerated index.
3. Confirm no files are missing required frontmatter (no stderr warnings).
4. Explain that the index is a cache and the source of truth is the frontmatter in each `.md` file.

## Success Criteria

- `regenerate-context-index.sh` executed.
- Both new context files present in the regenerated index.
- No missing frontmatter warnings on stderr.
- Correctly states the index is regenerated and should not be edited by hand.

## Failure Conditions

- Index not regenerated.
- New files missing from the index output.
- Missing frontmatter warnings ignored without action.
- `.context/index.yaml` edited manually instead of regenerating.
