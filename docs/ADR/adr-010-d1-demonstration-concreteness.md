---
title: "ADR-010: Add demonstration concreteness sub-criterion to D1 scorer"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d1-knowledge-delta-2026-04-29.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The D1 (Knowledge Delta) scorer only measured content volume and topical coverage, missing whether the skill provides concrete demonstrations. A skill can list facts without showing how to apply them, inflating its D1 score without delivering actionable knowledge.

## Decision

Add a "Demonstration Concreteness" sub-criterion (3 pts) to D1 that detects concrete demonstrations via three signals: code fences, arrow notation (`→`), and explicit output sections. Reduce `d1BaseScore` from 15 to 12 to accommodate the new sub-points without changing the D1 max of 20.

## Consequences

- D1 now rewards actionable demonstrations, not just content breadth
- Three-signal heuristic is simple and deterministic
- Existing high-scoring skills may regrade slightly lower on pure breadth without demos
- Implemented in the same refactor as the dimension registry pattern (ADR-006)
