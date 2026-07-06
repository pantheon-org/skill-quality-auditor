# Scenario 01: Create an ADR from a Context File Decision

## User Prompt

"Here is a new finding about the database migration approach. It contains a decision — extract it and create an ADR."

## Input

```markdown
---
title: "Finding: Database Migration Strategy"
type: FINDING
status: ACTIVE
date: 2026-06-30
---

# Finding: Database Migration Strategy

Date: 2026-06-30
Status: DECISION-SUPPORT

> The team evaluated sqllite vs postgres for the skill audit storage layer.

## Summary

We need a storage backend for persisting audit results, trend data, and remediation plans. Two candidates were evaluated.

## Detail

**Option A: SQLite**
- Zero infrastructure, file-based
- Good for single-node use
- No concurrent write support needed for initial use case

**Option B: PostgreSQL**
- Full ACID compliance beyond what's needed
- Requires Docker or managed service for local dev
- Adds operational overhead

## Recommended Action

Adopt SQLite for now. The data model is simple (single writer, timestamped records), and SQLite eliminates DevOps dependencies during early development. Migration to PostgreSQL later is straightforward if concurrent writers or network access becomes necessary.
```

## Expected Behavior

1. Identify the binding decision: "Adopt SQLite for now" from the Recommended Action section.
2. Create a new ADR file at `docs/ADR/adr-NNN-database-migration-strategy.md` with proper frontmatter.
3. Frontmatter must include `title` (with `ADR-NNN:` prefix), `status: proposed`, `date: YYYY-MM-DD`, and `context:` linking to the source finding.
4. ADR body must have **Context**, **Decision**, and **Consequences** sections.
5. The `context:` field must reference the relative path to the source `.context/findings/` file.
6. Run `regenerate-adr-index.sh` after creating the ADR.
7. Do not edit or delete any existing ADRs.

## Success Criteria

- ADR file created with correct frontmatter schema (title, status, date, context).
- `context:` field links to the source `.context/findings/database-migration-strategy-2026-06-30.md`.
- ADR body contains Context, Decision, and Consequences sections.
- `regenerate-adr-index.sh` executed successfully after creation.
- No existing ADRs modified.

## Failure Conditions

- ADR created without a `context:` link to the source finding.
- ADR body missing any of the required sections (Context, Decision, Consequences).
- Frontmatter missing required fields or using wrong status value.
- Index not regenerated after ADR creation.
- Existing ADR body edited or deleted.
