---
name: adr-capture
description: "Capture Architecture Decision Records (ADRs) from analysis, findings, plans, and reviews under .context/. Automatically extracts decisions into numbered ADRs under docs/ADR/, maintains a machine-readable index, and validates coverage. Use when creating or reviewing context files that contain decisions. Triggers: 'capture decision', 'create ADR', 'document decision', 'architectural decision', 'record decision as ADR', 'extract decisions', 'ADR review', 'index ADRs', 'check ADR coverage'."
---

# ADR Capture

Capture Architecture Decision Records from `.context/` analysis, findings, plans, and reviews. Automatically extract decisions, create numbered ADRs, and maintain the index.

> **Immutability rule:** Once created, an ADR is never edited or deleted. Its title, body, and context fields are frozen. Only `status` and `superseded_by` may change. To replace a decision, supersede it — create a new ADR and mark the old one as superseded.

## Quick Start

```bash
# Regenerate ADR index after adding or updating ADRs
.agents/skills/adr-capture/regenerate-adr-index.sh

# Check for decisions in context files that lack ADRs
.agents/skills/adr-capture/check-undocumented-decisions.sh
```

## ADR Location and Structure

ADRs live at `docs/ADR/adr-NNN-kebab-case-title.md`:

```text
docs/ADR/
  adr-001-use-yaml-frontmatter-for-context-files.md
  adr-002-nine-dimension-scoring-framework.md
  adr-003-split-reporter-into-sub-packages.md
  index.yaml                    # auto-generated; do not edit by hand
```

### ADR Frontmatter Schema

Every ADR MUST start with:

```yaml
---
title: "ADR-001: Human-readable decision title"
status: proposed | accepted | deprecated | superseded
date: YYYY-MM-DD
superseded_by: "adr-NNN"     # only when status is "superseded"
context:
  - path: .context/findings/topic-YYYY-MM-DD.md  # reference to source context file
  - path: .context/plans/related-plan.md
---
```

Field rules:
- `title` — Must start with `ADR-NNN: ` prefix; wrap in quotes. This is the document's single title — do not add a redundant `# ADR-NNN:` body heading
- `status` — `proposed` until reviewed, `accepted` for active decisions, `deprecated` or `superseded` when replaced
- `date` — Creation date in ISO format; do not update on edits
- `superseded_by` — Only when `status: superseded`; value is the replacement ADR name (e.g., `adr-002`)
- `context` — List of relative paths to `.context/` files that motivated the decision; omit entirely if none

### ADR Body Template

```markdown
**Status:** Proposed
**Date:** YYYY-MM-DD

## Context

What is the issue that we're seeing that is motivating this decision or change?

## Decision

What is the change that we're proposing and/or doing?

## Consequences

What becomes easier or more difficult to do because of this change?
```

## When to Create an ADR

Create an ADR whenever a `.context/` file (plan, finding, analysis) or a review makes a **binding decision** — a choice that affects future development direction, architecture, conventions, or processes.

### Examples of decision-triggering content

| Context file section | Decision example | ADR warranted? |
|---------------------|-----------------|----------------|
| Finding > Recommended Action | "Adopt Option A (native Go eval runner)" | Yes |
| Plan > Steps | "Phase 1: split reporter into sub-packages" | Yes |
| Analysis > Priority Remediation | "Replace global flag vars with option structs" | Yes |
| Plan > Open Questions | "Use sqllite vs postgres" — after resolved | Yes |
| Finding > Summary | "Observational research with no action" | No |
| Analysis > Findings | "Code has no interfaces" (observation only) | No |

### Decision capture workflow

1. When creating a `.context/` file that contains a decision, create the ADR in the same session
2. Use the template above — link the context file in the `context:` frontmatter field
3. Run `regenerate-adr-index.sh` after creating the ADR
4. Set `status: proposed` initially; promote to `accepted` after implementation starts
5. When a decision is superseded: set `status: superseded` and `superseded_by` on the old ADR (never edit its body); create a new ADR with higher number referencing the old one via `context:`

## Scripts

### `regenerate-adr-index.sh`

Scans `docs/ADR/adr-*.md` for frontmatter, generates `docs/ADR/index.yaml`.

```bash
.agents/skills/adr-capture/regenerate-adr-index.sh
```

### `check-undocumented-decisions.sh`

Scans `.context/**/*.md` for decision-related keywords and cross-references against ADR `context:` links. Reports context files that appear to contain decisions but are not referenced by any ADR.

```bash
.agents/skills/adr-capture/check-undocumented-decisions.sh
```

## Integration with Context Workflow

ADR capture is the final step in the context file lifecycle:

```
Create context file (context-file skill)
  → Regenerate context index (context-index skill)
  → Check for extractable decisions (adr-capture skill)
  → Create ADR if warranted (this skill)
  → Regenerate ADR index (this skill)
```

The hk pre-commit hook validates ADR frontmatter and index freshness automatically.

## Mindset

- Not every finding needs an ADR — only decisions that shape future work.
- The `context:` frontmatter field is the link between decisions and their evidence. Always populate it.
- `status: proposed` is the safe default; promote after implementation review.
- **ADRs are immutable.** Once created, the title, body, and context fields never change. Only `status` and `superseded_by` may be updated. Superseding is the only way to replace a decision.
- Superseding is normal — update `status: superseded` + `superseded_by`, create a new ADR. Never delete or overwrite old ADRs.

## Anti-Patterns

**NEVER** create an ADR without linking the source context file.
**WHY:** The `context:` field is the provenance chain — without it the decision is untethered from its evidence.
**BAD:** Creating an ADR with no `context:` references for a decision that came from a review.
**GOOD:** Always include `context:` with the relative path to the `.context/` file or review output.

**NEVER** edit or delete an ADR after creation.
**WHY:** ADRs are immutable records. Editing distorts the historical record; deleting erases the rationale trail entirely. Every ADR, even a superseded one, preserves the context and reasoning that led to a decision.
**BAD:** Rewording the body of `adr-002.md` to reflect a new understanding.
**BAD:** Deleting `adr-002.md` because the decision was reversed.
**GOOD:** Set `status: superseded` and `superseded_by: adr-007` on `adr-002.md`; create `adr-007.md` with the new decision.

**NEVER** reuse ADR numbers.
**WHY:** ADR numbers are permanent identifiers. Reusing a number erases the mapping between number and decision.
**BAD:** Deleting `adr-005.md` and creating a new `adr-005.md` for a different decision.
**GOOD:** Always increment to the next unused number, even if previous ADRs have been superseded.

**NEVER** create a `.context/` file with a clear decision section but skip the ADR.
**WHY:** The decision will be invisible to agents reading the ADR index, causing re-debate or re-research.
**BAD:** A plan with "Decision: Use Go 1.22" but no ADR referencing it.
**GOOD:** Create both the plan and the ADR in the same session.
