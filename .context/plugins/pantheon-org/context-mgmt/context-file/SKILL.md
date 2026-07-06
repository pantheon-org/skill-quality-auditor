---
name: context-file
description: "Create a new .context/ file (plan, finding, or analysis) with standard YAML frontmatter. Use when documenting a decision, writing an implementation plan, recording research findings, or capturing analysis output. DO NOT use for ephemeral notes, secrets storage, or skill remediation plans (use skill-auditor remediate instead). Triggers: 'create a plan', 'new finding', 'document this', 'write analysis', 'new context file', 'capture findings', 'draft a plan', 'record decision'."
---

# Context File

Create a new `.context/` file with standard YAML frontmatter and appropriate sections.

## Prerequisites

- A clear understanding of the file type needed: plan (multi-step work), finding (research), analysis (review), or known-issue (a real, verified gap that isn't being fixed right now)
- Familiarity with the `.context/` directory structure: `plans/`, `findings/`, `analysis/`, `known-issues/`
- The `context-index` skill available for index regeneration after creation

## When to Use

- **plan**: Multi-step implementation, migration, or remediation work with open tasks
- **finding**: Research output, code review results, audit findings, prerequisite investigations
- **analysis**: Duplication reports, benchmark results, comparative reviews, one-off audits
- **known-issue**: A concrete, verified gap or bug that's being consciously deferred rather than fixed now — see `session-reflection`'s workflow, which is the primary source of these

## When Not to Use

- For skill remediation plans, use `skill-auditor remediate` — it produces a richer schema
- Do not store secrets, credentials, or personal data in `.context/` files
- Do not create `.context/` files for ephemeral notes — use inline comments instead
- Do not create a `KNOWN_ISSUE` for something you're fixing in the same session — just fix it. `KNOWN_ISSUE` is for deferred, tracked work only.

## Frontmatter Schema

Every `.context/` file MUST start with this exact block:

```yaml
---
title: "Human-readable title"
type: PLAN | FINDING | ANALYSIS | INSTRUCTION | AUDIT | KNOWN_ISSUE
status: DRAFT | ACTIVE | DONE | SUPERSEDED
date: YYYY-MM-DD
related:
  - relative/path/to/related.md
---
```

Enum values are UPPER_CASE. Field rules:
- `title` — prose title matching the H1 heading; wrap in quotes
- `type` — matches the subdirectory (`plans/` → `PLAN`, `findings/` → `FINDING`, `analysis/` → `ANALYSIS`, `known-issues/` → `KNOWN_ISSUE`; note the underscore vs the hyphenated directory)
- `status` — `DRAFT` until reviewed, `ACTIVE` for in-progress work, `DONE` when complete, `SUPERSEDED` when replaced (for `KNOWN_ISSUE`: `ACTIVE` = still open, `DONE` = fixed)
- `date` — creation date in ISO format; do not update on edits
- `related` — relative paths from the file's location; omit the key entirely if there are no related files
- `severity` — required for `type: KNOWN_ISSUE` only: `CRITICAL | HIGH | MEDIUM | LOW`. Not applicable to other types.
- `value` — required for `type: PLAN`, `FINDING`, and `KNOWN_ISSUE` while `status` is `DRAFT` or `ACTIVE`: `HIGH | MEDIUM | LOW`. The benefit-of-action grade, distinct from `severity` (risk-of-inaction) and `effort` (cost-of-action). Grade against [`.context/instructions/value-rubric.md`](../../../../instructions/value-rubric.md); do not guess. Not applicable to `ANALYSIS` / `INSTRUCTION` / `AUDIT`. Exempt on `DONE` / `SUPERSEDED`.
- `themes` — required for `type: PLAN`, `FINDING`, and `KNOWN_ISSUE` while `status` is `DRAFT` or `ACTIVE`: a multi-valued **ordered** list from `EVAL | PR-TOOLING | DOCS | GOVERNANCE | SKILL-QUALITY | DISTRIBUTION`. The subject axis (what area the item is about), orthogonal to the magnitude axes. Write it **primary-first** — `themes[0]` is the primary theme and the only member used in the read-protocol tie-break. Draw members from [`.context/instructions/theme-vocabulary.md`](../../../../instructions/theme-vocabulary.md); do not invent themes. Block YAML style like `related`. Not applicable to `ANALYSIS` / `INSTRUCTION` / `AUDIT`. Exempt on `DONE` / `SUPERSEDED`.

## Workflow

1. Determine type: PLAN / FINDING / ANALYSIS / KNOWN_ISSUE
2. Choose a filename: kebab-case for plans (`migrate-off-tessl-eval.md`), `topic-YYYY-MM-DD.md` for timestamped reports and known-issues
3. Create the file using the template matching the type below
4. Set `status: DRAFT` until the content is reviewed (`KNOWN_ISSUE` starts at `ACTIVE` — it's already a confirmed, real gap by the time it's written down)
5. Run the context index regeneration script to update the index after creation

## Templates

**Plan** (`.context/plans/`):

```markdown
---
title: "Plan: <concise title>"
type: PLAN
status: DRAFT
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
type: FINDING
status: ACTIVE
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
type: ANALYSIS
status: DONE
date: YYYY-MM-DD
---
# <Topic> Analysis — YYYY-MM-DD
## Summary
## Findings
## Conclusion
```

**Known Issue** (`.context/known-issues/`):

```markdown
---
title: "Known Issue: <concrete, verified problem>"
type: KNOWN_ISSUE
status: ACTIVE
date: YYYY-MM-DD
severity: CRITICAL | HIGH | MEDIUM | LOW
value: HIGH | MEDIUM | LOW
themes:
  - GOVERNANCE   # ordered, primary-first; from theme-vocabulary.md
related:
  - ../plans/related-plan.md
---
# Known Issue: <concrete, verified problem>
> One-sentence statement of the verified impact, not a hypothesis.
## Why this exists
## Impact if unfixed
## Suggested fix (not yet applied — this is the tracked issue, not the fix)
```

## Mindset

- Write for the next agent or human who reads this cold — assume no prior context
- `status: DRAFT` is the safe default; promote to `ACTIVE` only when reviewed
- Date is creation date, not last-modified — do not update it on subsequent edits
- After creating or updating a `.context/` file, consider regenerating the index to keep it current
- Use production-grade terminology: pitfall, gotcha, ALWAYS, NEVER, anti-pattern

## Troubleshooting

- **Pre-commit hook blocks commit:** Run `check-context-frontmatter.sh` to find files missing YAML frontmatter — add the required block and re-run
- **Missing from index after creation:** Run `regenerate-context-index.sh` — the index is generated from frontmatter, not file existence alone
- **Wrong type directory:** Files must live under the matching subdirectory — plans in `plans/`, findings in `findings/`, analysis in `analysis/`, known-issues in `known-issues/`

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

