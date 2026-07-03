# .context/ File Types Reference

This reference documents the three types of `.context/` files and their distinct purposes.

## Plan

Plans document multi-step implementation, migration, or remediation work.

**Location:** `.context/plans/<kebab-case-name>.md`

**Required frontmatter fields:** `title`, `type: plan`, `status`, `date`

| Field | Convention | Example |
|-------|-----------|---------|
| `title` | `"Plan: <concise title>"` | `"Plan: Migrate to Structured Logging"` |
| `status` | Start as draft | `draft` |
| `filename` | kebab-case with optional date | `migrate-structured-logging-2026-06-30.md` |

**Required sections:** Goal, Steps, Open Questions

## Finding

Findings document research output, code review results, audit findings, and prerequisite investigations.

**Location:** `.context/findings/<topic>-YYYY-MM-DD.md`

**Required frontmatter fields:** `title`, `type: finding`, `status`, `date`, optionally `related`

| Field | Convention | Example |
|-------|-----------|---------|
| `title` | `"Finding: <topic>"` | `"Finding: Database Migration Strategy"` |
| `status` | Start as active | `active` |
| `related` | Reference related plans/analyses | `../plans/improve-test-coverage.md` |
| `filename` | topic-YYYY-MM-DD.md | `database-migration-strategy-2026-06-30.md` |

**Required sections:** One-sentence summary, Summary, Detail, Recommended Action

## Analysis

Analyses document duplication reports, benchmark results, comparative reviews, and one-off audits.

**Location:** `.context/analysis/<topic>-YYYY-MM-DD.md`

**Required frontmatter fields:** `title`, `type: analysis`, `status`, `date`

| Field | Convention | Example |
|-------|-----------|---------|
| `title` | `"<Topic> Analysis — YYYY-MM-DD"` | `"CLI Flag Audit — 2026-06-30"` |
| `status` | Typically done on creation | `done` |

**Required sections:** Summary, Findings, Conclusion

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Using wrong type for the content | Match type to subdirectory: plan↔plans/, finding↔findings/, analysis↔analysis/ |
| Putting plans in `.context/audits/` | Use `.context/plans/` instead — audits/ is owned by skill-auditor |
| Missing frontmatter that blocks commits | Always start with `---\ntitle:\ntype:\nstatus:\ndate:\n---` |
| Not regenerating the index | Run `.context/plugins/pantheon-org/context-index/scripts/regenerate-context-index.sh` after creation |

