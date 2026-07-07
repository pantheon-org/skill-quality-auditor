# YAML Frontmatter Guide

This reference covers the YAML frontmatter schema used by all `.context/` files.

## Required Fields

```yaml
---
title: "Human-readable title"
type: PLAN | finding | analysis
status: DRAFT | active | deferred | done | superseded
date: YYYY-MM-DD
---
```

| Field | Type | Rules |
|-------|------|-------|
| `title` | string | Prose title matching the H1 heading; wrap in quotes |
| `type` | enum | One of: `PLAN`, `FINDING`, `ANALYSIS` — must match the subdirectory |
| `status` | enum | `DRAFT` until reviewed, `ACTIVE` for in-progress, `DEFERRED` for a real item intentionally parked (date-gated or blocked; ranked below ACTIVE by the read protocol), `DONE` when complete, `SUPERSEDED` when replaced |
| `date` | string (date) | ISO 8601 date `YYYY-MM-DD` — set once on creation, never updated |

## Optional Fields

### `related`

List of relative paths to related `.context/` files. The path is relative to the file's own location.

```yaml
related:
  - ../plans/related-plan.md
  - ../findings/related-finding-2026-06-30.md
```

**Rules:**
- Omit the key entirely when there are no related files
- Never use `related: []` (empty list) — either omit or populate
- Use relative paths with `../` to navigate between subdirectories

### Examples by Type

**Plan** (`.context/plans/`):
```yaml
---
title: "Plan: Add Structured Logging"
type: PLAN
status: DRAFT
date: 2026-06-30
---
```

**Finding** (`.context/findings/`):
```yaml
---
title: "Finding: Logging Library Evaluation"
type: FINDING
status: ACTIVE
date: 2026-06-30
related:
  - ../plans/add-structured-logging.md
---
```

**Analysis** (`.context/analysis/`):
```yaml
---
title: "CLI Flag Audit — 2026-06-30"
type: ANALYSIS
status: DONE
date: 2026-06-30
---
```

## Validation

Run `.context/plugins/pantheon-org/context-mgmt/context-index/scripts/check-context-frontmatter.sh .context/**/*.md` to validate all files.

