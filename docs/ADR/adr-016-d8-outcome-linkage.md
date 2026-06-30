---
title: "ADR-016: Add outcome linkage criterion to D8 practical usability scorer"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d8-practical-usability-2026-04-29.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The D8 (Practical Usability) scorer counted code blocks and examples but had no mechanism for assessing whether those examples were linked to specific outcomes. A skill could include many code snippets without connecting them to the results or side effects a user should expect.

## Decision

Add an "Outcome Linkage" criterion (3 pts, additive) that checks whether examples or code blocks are paired with outcome descriptions. Use Option B (additive within the flat model) rather than reweighting phantom components. Outcome indicators are only valid within or adjacent to code blocks to avoid false positives from general prose.

## Consequences

- D8 rewards example-outcome pairing, not just example volume
- The adjacent-to-code-block rule prevents prose-only skills from gaming the signal
- D8 max remains at 15 (additive, not reweighted)
- Implemented alongside the CLI standardization (ADR-022)
