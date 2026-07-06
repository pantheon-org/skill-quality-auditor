---
title: "ADR-051: a `themes` frontmatter field adds a queryable subject axis to .context/"
status: accepted
date: 2026-07-06
context:
  - path: ".context/plans/context-theme-taxonomy-2026-07-06.md"
  - path: ".context/findings/context-taxonomy-gap-2026-07-06.md"
  - path: ".context/instructions/theme-vocabulary.md"
  - path: "docs/ADR/adr-049-context-value-frontmatter-field.md"
---

**Status:** Accepted
**Date:** 2026-07-06

## Context

With the `value` signal shipped (ADR-049), `.context/index.yaml` could rank the
backlog by magnitude but not by subject. Within a single `value`+`effort` bucket,
items were interchangeable, and questions like "show me all the eval-migration
work" or "which area carries the most open debt?" could not be answered from the
index. The classification was re-derived each session by reading titles and
walking `related` links. The finding `context-taxonomy-gap-2026-07-06.md`
documented the gap; the plan `context-theme-taxonomy-2026-07-06.md` closes it.
The three design forks were settled through a guided-interview on 2026-07-06.

## Decision

1. **Add a `themes` frontmatter field** to the three action-candidate types
   (`PLAN`, `FINDING`, `KNOWN_ISSUE`) — the subject axis, orthogonal to the
   magnitude axes (`value`/`effort`/`severity`). `ANALYSIS`, `INSTRUCTION`, and
   `AUDIT` are reference material and do not carry it.

2. **`themes` is a multi-valued, ordered list, not a single enum.** An entry can
   genuinely belong to several areas, so a single value would force a lossy
   choice. Every member is drawn from a controlled vocabulary, so the axis stays
   queryable; free-form tags are not permitted. Deriving the grouping from the
   `related` graph instead was rejected: it depends on link discipline the repo
   has already seen drift on, and cannot express "this is about X" independently
   of "this links to Y".

3. **The list is ordered and `themes[0]` is the primary theme.** The primary is
   the sole member that participates in the read protocol, as the final
   tie-breaker below `value` then `effort` — theme expresses preference-of-area,
   not priority. The remaining members are for filtering and cluster views only.
   Authors write the list primary-first.

4. **The vocabulary is six coarse themes, split only on evidence.** `EVAL`,
   `PR-TOOLING`, `DOCS`, `GOVERNANCE`, `SKILL-QUALITY`, `DISTRIBUTION`. A theme is
   subdivided only once it demonstrably dominates (rough guide: more than ~30% of
   active/draft action-candidates), and any split is recorded as an amendment to
   this ADR — not an ad-hoc edit. This mirrors ADR-049's deferral of a numeric
   value scale until ties proved blocking: ship the simple thing, refine on need.

5. **The vocabulary is canonical and lives at
   `.context/instructions/theme-vocabulary.md`** — the definitions, the
   primary-first rule, the split rule, and worked examples of primary selection.
   `ways-of-working.md` and `AGENTS.md` link to it rather than restating it.

6. **`themes` is required while `status` is `DRAFT` or `ACTIVE`; `DONE` and
   `SUPERSEDED` are exempt**, matching the `value` contract. Enforcement lives in
   `validate-context-frontmatter.sh`; values are UPPER_CASE per ADR-050.

## Consequences

- **Easier:** the backlog is sliceable by area ("all `EVAL` work") and
  interrelated clusters are visible for batch closure, straight from the index.
- **Easier:** same-`value`/`effort` ties now break deterministically on the
  primary theme instead of arbitrarily, and a future "what's next" skill has a
  stable grouping key.
- **Harder / ongoing cost:** `themes[0]` is load-bearing for the tie-break, so a
  carelessly ordered list sorts wrongly within its bucket (lower stakes than a
  `value` misgrade — it only affects same-tier ties). A multi-valued list also
  invites over-tagging; guidance is to tag only genuine areas.
- **Migration:** the field shipped optional-first (schema + validator accept it),
  all 45 remaining active/draft action-candidates were backfilled in one
  serialised pass, and only then did the validator flip to required — so the tree
  stayed green throughout.
- **Watch item:** at backfill, `GOVERNANCE` was the primary theme on 29% of
  active/draft entries — just under the split threshold. If it crosses ~30%,
  split it (e.g. into context-system vs ADR-governance) and amend Decision 4.
