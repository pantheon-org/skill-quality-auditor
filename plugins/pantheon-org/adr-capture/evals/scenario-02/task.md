# Scenario 02: Validate ADR Frontmatter and Regenerate Index

## User Prompt

"I've created some ADRs. Run validation on them and regenerate the index."

## Input

Three ADR files exist:

**`docs/ADR/adr-001-use-yaml-frontmatter.md`** (valid):
```yaml
---
title: "ADR-001: Use YAML Frontmatter for Context Files"
status: accepted
date: 2026-01-15
context:
  - path: .context/findings/frontmatter-evaluation-2026-01-15.md
---
```

**`docs/ADR/adr-002-nine-dimension-scoring.md`** (invalid — missing `context`):
```yaml
---
title: "Scoring framework"
status: accepted
date: 2026-01-20
---
```

**`docs/ADR/adr-003-split-reporter.md`** (invalid — wrong status value):
```yaml
---
title: "ADR-003: Split Reporter Into Sub-Packages"
status: reviewed
date: 2026-02-01
context:
  - path: .context/plans/refactor-reporter.md
---
```

## Expected Behavior

1. Run `validate-adr-frontmatter.sh` against all ADR files.
2. Identify that:
   - `adr-001` is valid and passes.
   - `adr-002` is invalid — title doesn't start with `ADR-NNN:` prefix, missing `context` field.
   - `adr-003` is invalid — `status: reviewed` is not in the allowed set (proposed, accepted, deprecated, superseded).
3. Report which ADRs pass and which fail, with specific reasons for each failure.
4. After reporting, run `regenerate-adr-index.sh` to refresh the index with valid entries only.

## Success Criteria

- `validate-adr-frontmatter.sh` executed against ADR files.
- Valid ADRs correctly identified as passing.
- Invalid ADRs identified with specific field-level failures.
- Index regenerated after validation.
- Clear output distinguishing pass/fail per file.

## Failure Conditions

- Only valid ADRs checked without identifying invalid ones.
- Invalid ADRs not flagged with field-level specifics.
- Index regenerated before or without validation.
- Validation script not used (validation done manually or not at all).
