# Scenario 03: Handle ADR Supersession

## User Prompt

"We decided to switch from SQLite to PostgreSQL after all. The original decision was in ADR-004. Create a new ADR for the new decision and properly supersede the old one."

## Input

**`docs/ADR/adr-004-use-sqlite.md`**:
```yaml
---
title: "ADR-004: Use SQLite for Audit Storage"
status: accepted
date: 2026-06-15
context:
  - path: .context/findings/database-migration-strategy-2026-06-30.md
---
```

```markdown
**Status:** Accepted
**Date:** 2026-06-15

## Context

We need a storage backend that minimises operational overhead during early development.

## Decision

Use SQLite as the primary storage backend for audit results.

## Consequences

- Zero infrastructure for local development.
- Single-writer limitation may become a bottleneck.
- Migration path to PostgreSQL exists if needed.
```

Now the team has decided to migrate to PostgreSQL because concurrent writer access is now required and the deployment target (Kubernetes) makes SQLite impractical.

## Expected Behavior

1. Do NOT edit or delete `adr-004.md`'s body — only update `status` and `superseded_by` in its frontmatter.
2. Update `adr-004.md` frontmatter: set `status: superseded`, add `superseded_by: adr-005` (next available number).
3. Create `docs/ADR/adr-005-use-postgresql.md` with proper frontmatter and body.
4. The new ADR should reference `adr-004` in its `context:` field.
5. Include Context, Decision, and Consequences sections in the new ADR.
6. Run `regenerate-adr-index.sh` after all changes.

## Success Criteria

- Old ADR body unchanged (immutability respected) — only frontmatter updated.
- `adr-004` frontmatter has `status: superseded` and `superseded_by: adr-005`.
- New ADR `adr-005` created with next number, proper frontmatter, and body.
- New ADR `context:` references the old ADR and the original finding.
- Index regenerated after changes.

## Failure Conditions

- Old ADR body edited or deleted.
- Old ADR `superseded_by` omitted or pointing to wrong number.
- New ADR reuses a number or uses a non-sequential number.
- Index not regenerated.
- New ADR missing `context:` link to the superseded ADR.
