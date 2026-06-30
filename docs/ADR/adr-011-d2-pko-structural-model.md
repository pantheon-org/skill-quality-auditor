---
title: "ADR-011: Restructure D2 scorer around PKO structural model"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d2-mindset-procedures.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The D2 (Mindset & Procedures) scorer rewarded general procedure descriptions without assessing structural completeness. Skills could score highly by listing steps without specifying when to start, how to verify success, or how to handle decision points.

## Decision

Align D2 scoring against the PKO (Procedural Knowledge Ontology) structural model. Add three new sub-criteria:
- **Preconditions scoring** (2 pts) — does the skill describe what must be true before starting?
- **Postconditions / external checkpoints** (2 pts) — does it define how to verify success?
- **Decision points** (2 pts) — does it identify branching paths based on outcomes?

Reduce `scoreD2Structure` from 6 to 3 pts to accommodate the new sub-points without changing the D2 max of 15.

## Consequences

- D2 now measures structural completeness, not just procedural listing
- Skills must document pre/post conditions and decision branches for full marks
- PKO alignment makes the scoring rubric academically grounded
