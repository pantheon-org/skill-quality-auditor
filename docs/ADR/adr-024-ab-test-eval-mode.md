---
title: "ADR-024: Adopt A/B test evaluation mode from SkillEval"
status: proposed
date: 2026-06-30
context:
  - path: .context/findings/skilleval-analysis-2026-06-30.md
---

**Status:** Proposed
**Date:** 2026-06-30

## Context

Investigation of [justinwetch/SkillEval](https://github.com/justinwetch/SkillEval) — an A/B testing workbench for AI skills — identified three features worth adopting. SkillEval's core differentiator is running two skills against identical prompts with an LLM judge scoring outputs per-criterion and declaring a winner. Our project currently evaluates one skill at a time against fixed D1–D9 dimensions.

## Decision

Adopt an `ab-test` command as a new comparative evaluation mode alongside `evaluate` and `batch`. This gives skill authors a data-driven way to prove improvements (e.g., "my v2 skill scores 23% higher than v1"). The feature does not require dimensional changes — it's a cross-cutting command.

Secondary adoption:

- **AI-generated adaptive criteria** — add as an optional `--judge` flag on `evaluate` that produces LLM-scored criteria overlay alongside D1–D9. Slots into D9 (Eval Validation) or D8 (Practical Usability).
- **Visual evaluation via screenshots** — defer until the project expands into frontend/design skill evaluation. Would extend D8.

## Consequences

- New `ab-test` command provides comparative skill evaluation
- `--judge` flag adds dynamic LLM-scored criteria as an optional pass
- Visual evaluation remains on roadmap, not active work
