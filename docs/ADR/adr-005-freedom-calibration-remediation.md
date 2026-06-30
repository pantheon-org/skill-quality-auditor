---
title: "ADR-005: Restructure D6 scorer for contextual action weighting"
status: proposed
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d6-freedom-calibration.md
---

**Status:** Proposed
**Date:** 2026-06-30

## Context

The D6 (Freedom Calibration) scorer currently assigns points based on the
presence of action types without weighting them by context, making it
possible to score highly by listing permitted actions generically. The
dimension improvement plan identified this as a structural gap.

## Decision

Restructure the D6 scorer to weight actions by their situational context.
Actions are assessed within the scenario or task they apply to, rather than
in a flat list. Contextual relevance and precision of freedom boundaries
become scoring factors alongside raw action count.

## Consequences

- D6 scores will better reflect real-world calibration quality
- Existing high-scoring skills may regrade lower — communicate this
- Scorer logic becomes more complex, requiring more test coverage
- Thresholds for grade boundaries may need recalibration
