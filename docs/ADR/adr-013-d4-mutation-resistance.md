---
title: "ADR-013: Add mutation resistance criterion to D4 specification compliance scorer"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d4-specification-compliance-2026-04-29.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The D4 (Specification Compliance) scorer checked that a skill's claims matched its specification but had no mechanism for testing whether the specification would survive modification. A skill could be perfectly compliant in one version yet degrade to non-compliant with trivial changes. There was no "stress test" for specification robustness.

## Decision

Add a "Mutation Resistance" behavioural criterion (4 pts) to D4 that tests specification robustness through hard constraints, conditional branches, and exclusion rules. The scorer checks for:
- Hard constraints that prevent invalid states
- Conditional branches that handle edge cases
- Explicit exclusions that define what is NOT covered

Specificity thresholds defined for keyword detection. Existing sub-scorers reweighted to free 4 pts.

## Consequences

- D4 now measures specification robustness, not just static compliance
- Skills with defensive specification design score higher
- Threshold constants ensure consistent scoring across skills
- Reweighting preserves D4 max of 15
