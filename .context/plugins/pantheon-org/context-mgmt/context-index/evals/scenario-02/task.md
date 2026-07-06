# Scenario 02: Detect and Fix Missing Frontmatter

## User Prompt

"The pre-commit hook is blocking me. It says some context files are missing YAML frontmatter. Find them and fix them."

## Input

**`.context/plans/migrate-tessl-evals.md`** (missing frontmatter):
```markdown
# Plan: Migrate Tessl Evals
```

**`.context/findings/go-version-evaluation-2026-06-30.md`** (valid):
```yaml
---
title: "Finding: Go Version Evaluation"
type: FINDING
status: ACTIVE
date: 2026-06-30
---
```

**`.context/analysis/api-audit-2026-06-30.md`** (missing frontmatter):
```markdown
# API Audit — 2026-06-30

## Summary
```

## Expected Behavior

1. Run `check-context-frontmatter.sh` against `.context/**/*.md` to identify files missing frontmatter.
2. Identify `migrate-tessl-evals.md` and `api-audit-2026-06-30.md` as missing frontmatter.
3. Fix both files by adding proper frontmatter with title, type, status, and date fields.
4. After fixing, run `regenerate-context-index.sh` and confirm zero warnings.
5. Do NOT add frontmatter to files that already have valid frontmatter (`go-version-evaluation`).

## Success Criteria

- `check-context-frontmatter.sh` used to identify missing frontmatter.
- Both missing-frontmatter files correctly identified.
- Proper frontmatter added to both files with all required fields.
- Index regenerated with zero warnings.
- Valid files not modified.

## Failure Conditions

- Check script not used (issues identified manually or by reading files only).
- Only some missing-frontmatter files fixed.
- Frontmatter added missing required fields.
- Index not regenerated after fixes.
- Valid file's frontmatter modified.
