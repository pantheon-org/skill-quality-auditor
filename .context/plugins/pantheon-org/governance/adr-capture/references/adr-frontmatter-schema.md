# ADR Frontmatter Schema Reference

This reference provides detailed field-level documentation for ADR frontmatter.

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Must start with `ADR-NNN: ` prefix. The NNN is a zero-padded 3-digit number. Wrap the full value in quotes. |
| `status` | enum | One of: `proposed`, `accepted`, `deprecated`, `SUPERSEDED` |
| `date` | string (date) | ISO 8601 date: `YYYY-MM-DD`. This is the creation date and must never be updated. |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `superseded_by` | string | ADR name like `adr-005`. Only valid when `status: superseded`. |
| `context` | array of objects | Each object has a `path` field with relative path to the source `.context/` file. Omit the key entirely if no context files exist. |

## Status Lifecycle

```text
proposed ──► accepted ──► deprecated
                  │
                  └──► superseded ──► superseded_by: adr-NNN
```

- `proposed`: Initial state for new ADRs before implementation review
- `accepted`: Decision is active and being implemented
- `deprecated`: Decision is no longer relevant but no direct replacement exists
- `SUPERSEDED`: Decision has been replaced by a newer ADR; requires `superseded_by`

## Common Validation Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `title` missing ADR-NNN prefix | Title doesn't start with `ADR-NNN:` | Prepend `ADR-NNN:` to the title value |
| `status` not in allowed set | Typo or wrong enum value | Use exactly: proposed, accepted, deprecated, superseded |
| `superseded_by` without superseded status | Field present but status is not superseded | Set `status: superseded` or remove `superseded_by` |
| Missing `context` field | ADR has no provenance link | Add `context:` with at least one source path |

## Validation Command

```bash
scripts/validate-adr-frontmatter.sh docs/ADR/adr-*.md
```

