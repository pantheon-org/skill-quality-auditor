---
title: "ADR-014: Add negative-condition scoring to D5 progressive disclosure"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d5-progressive-disclosure.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The D5 (Progressive Disclosure) scorer measured how well a skill layered information from overview to detail but had no mechanism to reward negative conditions — the cases where the skill should NOT be applied. Skills could score highly on disclosure structure while omitting crucial "when not to use this" guidance.

## Decision

Add `scoreNegativeConditions()` as an additive sub-scorer (0-2 pts) that scans table rows for negative-condition indicators. Lower the compact-band ceiling by 2 pts (15→13) to accommodate the new sub-points. The approach is additive within the existing heuristic (Option B from the improvement plan) rather than a full rewrite (Option A).

Delete the dead `isReferenceSectionCompliant` code discovered during implementation.

## Consequences

- D5 now rewards skills that document off-label and contraindicated use cases
- Additive approach minimises refactoring risk
- Two-point ceiling keeps negative conditions as a bonus signal, not a dominant one
