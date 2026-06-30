---
title: "ADR-015: Add discriminativeness signal to D7 pattern recognition scorer"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d7-pattern-recognition.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The D7 (Pattern Recognition) scorer measured pattern coverage via length bands but had no way to assess whether the patterns it detected were genuinely discriminative — i.e., whether they actually helped an agent distinguish between similar situations. Length alone is not a quality signal.

## Decision

Add a discriminativeness check as a diagnostics-only signal (no score change yet). Two sub-criteria:
- **Negative anchors** — patterns that tell the agent what to look out for (distinguishing features)
- **Workflow anchors** — patterns that guide the agent through branching decisions

The existing length bands remain as the base score (6-10 pts). The discriminativeness signal is published in diagnostics for visibility but does not affect the numeric score until the rubric is validated.

## Consequences

- D7 diagnostics now surface pattern quality, not just coverage depth
- No score impact until the rubric is validated with real skills
- Negative anchors and workflow anchors provide a structural basis for future scoring
- Diagnostic-only deployment allows data collection before locking the scoring formula
