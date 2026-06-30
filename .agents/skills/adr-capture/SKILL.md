---
name: adr-capture
description: "Capture Architecture Decision Records (ADRs) from .context/ plans, findings, and analyses. Extracts binding decisions into numbered ADRs under docs/ADR/, maintains a machine-readable index, and validates coverage. Use when creating or reviewing context files that contain decisions. DO NOT use for observational findings without decisions, retroactive documentation of long-settled decisions, or inline comments. Triggers: 'capture decision', 'create ADR', 'document decision', 'architectural decision', 'record decision as ADR', 'extract decisions', 'ADR review', 'index ADRs', 'check ADR coverage'."
---

# ADR Capture

Capture Architecture Decision Records from `.context/` analysis, findings, plans, and reviews. Automatically extract decisions, create numbered ADRs, and maintain the index.

> **Immutability rule:** Once created, an ADR is never edited or deleted. Its title, body, and context fields are frozen. Only `status` and `superseded_by` may change. To replace a decision, supersede it — create a new ADR and mark the old one as superseded.

## Prerequisites

- A `.context/` file (plan, finding, or analysis) containing a **binding decision** — a choice that affects future development direction, architecture, conventions, or processes
- The `context-file` skill to create the source `.context/` file if it doesn't exist yet
- `assets/schemas/adr-frontmatter.schema.json` and `assets/templates/adr-template.yaml` for schema validation

## Quick Start

```bash
# Regenerate ADR index after adding or updating ADRs
scripts/regenerate-adr-index.sh

# Check for decisions in context files that lack ADRs
scripts/check-undocumented-decisions.sh
```

## ADR Location and Structure

ADRs live at `docs/ADR/adr-NNN-kebab-case-title.md`:

```text
docs/ADR/
  adr-001-use-yaml-frontmatter-for-context-files.md
  adr-002-nine-dimension-scoring-framework.md
  index.yaml
```

### ADR Frontmatter Schema

Every ADR MUST start with frontmatter. Use the template at `assets/templates/adr-template.yaml` to bootstrap new files.

```yaml
---
title: "ADR-NNN: Human-readable decision title"
status: proposed | accepted | deprecated | superseded
date: YYYY-MM-DD
superseded_by: "adr-NNN"     # only when status is "superseded"
context:
  - path: .context/findings/topic-YYYY-MM-DD.md
---
```

Field rules:
- `title` — Must start with `ADR-NNN: ` prefix; wrap in quotes
- `status` — `proposed` until reviewed, `accepted` for active decisions, `deprecated` or `superseded` when replaced
- `date` — Creation date in ISO format; do not update on edits
- `superseded_by` — Only when `status: superseded`; value is the replacement ADR name
- `context` — List of relative paths to `.context/` files that motivated the decision; omit entirely if none

### ADR Body Template

```markdown
**Status:** Proposed
**Date:** YYYY-MM-DD

## Context

What is the issue motivating this decision or change?

## Decision

What is the change being proposed or implemented?

## Consequences

What becomes easier or more difficult because of this change?
```

## When to Create an ADR

Create an ADR whenever a `.context/` file or a review makes a **binding decision** — a choice that affects future direction, architecture, conventions, or processes.

| Context file section | Decision example | ADR warranted? |
|---------------------|-----------------|----------------|
| Finding > Recommended Action | "Adopt Option A (native Go eval runner)" | Yes |
| Plan > Steps | "Phase 1: split reporter into sub-packages" | Yes |
| Plan > Open Questions | "Use sqllite vs postgres" — after resolved | Yes |
| Finding > Summary | "Observational research with no action" | No |

## Workflow

1. When creating a context file that contains a decision, create the ADR in the same session
2. Use the template — link the context file in the `context:` frontmatter field
3. Run the index regeneration script after creating the ADR
4. Set `status: proposed` initially; promote to `accepted` after implementation starts
5. When a decision is superseded: set `status: superseded` and `superseded_by` on the old ADR; create a new ADR referencing the old one via `context:`

## Scripts

```bash
scripts/validate-adr-frontmatter.sh        # Validate ADR frontmatter against JSON schema
scripts/regenerate-adr-index.sh            # Scan docs/ADR/ and regenerate index.yaml
scripts/check-undocumented-decisions.sh    # Find decisions without ADR coverage
```

## Integration with Context Workflow

ADR capture is the final step in the context file lifecycle:

```text
Create context file (context-file skill)
  → Regenerate context index (context-index skill)
  → Check for extractable decisions (adr-capture skill)
  → Create ADR if warranted (this skill)
  → Regenerate ADR index (this skill)
```

## Mindset

- Not every finding needs an ADR — only decisions that shape future work
- The `context:` frontmatter field links decisions to their evidence; always populate it
- `status: proposed` is the safe default; promote after implementation review
- ADRs are immutable — only `status` and `superseded_by` may be updated; superseding is the only way to replace a decision
- Consider marking `status: superseded` rather than deleting old ADRs; the historical record preserves context even for reversed decisions

## Troubleshooting

```text
Problem                                 | Solution
Index not updating after ADR creation   | Run scripts/regenerate-adr-index.sh
Pre-commit hook blocking                | Run scripts/validate-adr-frontmatter.sh
Decision not found in ADR index         | Check the ADR has context: and valid frontmatter
```

## Anti-Patterns

**NEVER** create an ADR without linking the source context file.
**WHY:** The `context:` field is the provenance chain — without it the decision is untethered from its evidence.
**BAD:** Creating an ADR with no `context:` references for a decision from a review.
**GOOD:** Always include `context:` with the relative path to the `.context/` file.

**NEVER** edit or delete an ADR after creation.
**WHY:** ADRs are immutable records. Editing distorts history; deleting erases the rationale trail.
**BAD:** Rewording the body of an ADR to reflect new understanding.
**GOOD:** Set `status: superseded` and `superseded_by`; create a new ADR.

**NEVER** reuse ADR numbers.
**WHY:** ADR numbers are permanent identifiers — reusing a number erases the mapping.
**GOOD:** Always increment to the next unused number.

**NEVER** skip creating an ADR when a `.context/` file contains a clear decision.
**WHY:** The decision will be invisible to agents reading the ADR index, causing re-debate.
**GOOD:** Create both the plan and ADR in the same session.

## References

| Topic | Reference | When to Use |
| --- | --- | --- |
| ADR frontmatter field rules, status lifecycle, and validation | [ADR Frontmatter Schema](references/adr-frontmatter-schema.md) | Validating or debugging ADR frontmatter errors |
| Step-by-step supersession workflow with examples | [ADR Supersession](references/adr-supersession.md) | Reversing or replacing an existing decision via supersession |

