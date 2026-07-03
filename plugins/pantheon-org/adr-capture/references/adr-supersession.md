# ADR Supersession Workflow Reference

When a decision is reversed or replaced, do NOT edit or delete the existing ADR. Follow the supersession workflow.

## Process

1. Identify the existing ADR covering the old decision
2. Update its frontmatter only: change `status` to `superseded`, add `superseded_by: adr-NNN`
3. Do NOT modify the ADR body, title, date, or context fields
4. Create a new ADR with the next available number
5. In the new ADR's `context:` field, reference the superseded ADR via its relative path: `docs/ADR/adr-004-use-sqlite.md`
6. Run `scripts/regenerate-adr-index.sh`

## Example

### Before

`adr-004-use-sqlite.md`
```yaml
---
title: "ADR-004: Use SQLite for Audit Storage"
status: accepted
date: 2026-06-15
context:
  - path: .context/findings/database-migration-strategy-2026-06-30.md
---
```

### After (frontmatter only — body unchanged)

`adr-004-use-sqlite.md`
```yaml
---
title: "ADR-004: Use SQLite for Audit Storage"
status: superseded
date: 2026-06-15
superseded_by: "adr-005"
context:
  - path: .context/findings/database-migration-strategy-2026-06-30.md
---
```

### New ADR

`adr-005-use-postgresql.md`
```yaml
---
title: "ADR-005: Use PostgreSQL for Audit Storage"
status: proposed
date: 2026-06-30
context:
  - path: docs/ADR/adr-004-use-sqlite.md
---
```

## Key Rules

| Rule | Rationale |
|------|-----------|
| Never edit body of superseded ADR | Body content is a historical record; editing it distorts decision history |
| Always increment to next number | ADR numbers are permanent identifiers — never reuse |
| Link superseded ADR in new context: field | Maintains the decision chain so readers can trace evolution |
| Regenerate index after changes | The index is the discovery mechanism; stale index hides decisions |

