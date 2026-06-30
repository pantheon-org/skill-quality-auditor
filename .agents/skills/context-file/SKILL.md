---
name: context-file
description: "Create a new .context/ file (plan, finding, or analysis) with standard YAML frontmatter. Use when documenting a decision, writing an implementation plan, recording research findings, or capturing analysis output. DO NOT use for ephemeral notes, secrets storage, or skill remediation plans (use skill-auditor remediate instead). Triggers: 'create a plan', 'new finding', 'document this', 'write analysis', 'new context file', 'capture findings', 'draft a plan', 'record decision'."
---

# Context File

Create a new `.context/` file with standard YAML frontmatter and appropriate sections.

## Prerequisites

- A clear understanding of the file type needed: plan (multi-step work), finding (research), or analysis (review)
- Familiarity with the `.context/` directory structure: `plans/`, `findings/`, `analysis/`
- The `context-index` skill available for index regeneration after creation

## Quick Start

```bash
# Determine type and create the file
.context/plans/<kebab-case-name>.md          # implementation or migration plan
.context/findings/<topic>-YYYY-MM-DD.md      # research or investigation output
.context/analysis/<topic>-YYYY-MM-DD.md      # duplication reports, audits, reviews

# After creating, regenerate the index
.agents/skills/context-index/regenerate-context-index.sh
```

## When to Use

- **plan**: Multi-step implementation, migration, or remediation work with open tasks
- **finding**: Research output, code review results, audit findings, prerequisite investigations
- **analysis**: Duplication reports, benchmark results, comparative reviews, one-off audits

## When Not to Use

- For skill remediation plans, use `skill-auditor remediate` — it produces a richer schema
- Do not store secrets, credentials, or personal data in `.context/` files
- Do not create `.context/` files for ephemeral notes — use inline comments instead

## Frontmatter Schema

Every `.context/` file MUST start with this exact block:

```yaml
---
title: "Human-readable title"
type: plan | finding | analysis
status: draft | active | done | superseded
date: YYYY-MM-DD
related:
  - relative/path/to/related.md
---
```

Field rules:
- `title` — prose title matching the H1 heading; wrap in quotes
- `type` — matches the subdirectory (`plans/` → `plan`, `findings/` → `finding`, `analysis/` → `analysis`)
- `status` — `draft` until reviewed, `active` for in-progress work, `done` when complete, `superseded` when replaced
- `date` — creation date in ISO format; do not update on edits
- `related` — relative paths from the file's location; omit the key entirely if there are no related files

## Workflow

1. Determine type: plan / finding / analysis
2. Choose a filename: kebab-case for plans (`migrate-off-tessl-eval.md`), `topic-YYYY-MM-DD.md` for timestamped reports
3. Create the file using the template matching the type below
4. Set `status: draft` until the content is reviewed
5. Run the context index regeneration script to update the index after creation

## Templates

**Plan** (`.context/plans/`):

```markdown
---
title: "Plan: <concise title>"
type: plan
status: draft
date: YYYY-MM-DD
---
# Plan: <title>
## Goal
One paragraph describing the desired end state.
## Steps
1. Step one
2. Step two
## Open Questions
- Question one
```

**Finding** (`.context/findings/`):

```markdown
---
title: "Finding: <topic>"
type: finding
status: active
date: YYYY-MM-DD
related:
  - ../plans/related-plan.md
---
# Finding: <topic>
> One-sentence summary.
## Summary
## Detail
## Recommended Action
```

**Analysis** (`.context/analysis/`):

```markdown
---
title: "<Topic> Analysis — YYYY-MM-DD"
type: analysis
status: done
date: YYYY-MM-DD
---
# <Topic> Analysis — YYYY-MM-DD
## Summary
## Findings
## Conclusion
```

## Mindset

- Write for the next agent or human who reads this cold — assume no prior context
- `status: draft` is the safe default; promote to `active` only when reviewed
- Date is creation date, not last-modified — do not update it on subsequent edits
- After creating or updating a `.context/` file, consider regenerating the index to keep it current

## Troubleshooting

- **Pre-commit hook blocks commit:** Run `check-context-frontmatter.sh` to find files missing YAML frontmatter — add the required block and re-run
- **Missing from index after creation:** Run `regenerate-context-index.sh` — the index is generated from frontmatter, not file existence alone
- **Wrong type directory:** Files must live under the matching subdirectory — plans in `plans/`, findings in `findings/`, analysis in `analysis/`

## Anti-Patterns

**NEVER** create a `.context/` file without the full frontmatter block.
**WHY:** The pre-commit hook will block the commit and the index cannot include the file.
**BAD:** Starting a file with `# Plan: ...` directly, with no frontmatter.
**GOOD:** Always open with `---\ntitle: ...\ntype: ...\nstatus: ...\ndate: ...\n---`.

**NEVER** include `related: []` when there are no related files.
**WHY:** An empty list is noise; the field should be absent.
**BAD:** `related: []`
**GOOD:** Omit the `related` key entirely.

**NEVER** put findings or plans under `.context/audits/`.
**WHY:** `.context/audits/` is owned by `skill-auditor --store`; writing there by hand conflicts with the tool's schema.
**GOOD:** Use `.context/plans/`, `.context/findings/`, or `.context/analysis/` only.

## References

| Topic | Reference | When to Use |
| --- | --- | --- |
| Three file types, field conventions, and common mistakes | [Context File Types](references/context-file-types.md) | Choosing the correct type or debugging file placement |
| Required and optional frontmatter fields with examples | [YAML Frontmatter Guide](references/yaml-frontmatter-guide.md) | Setting up frontmatter for a new file or fixing validation errors |

