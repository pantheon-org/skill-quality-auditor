---
title: "ADR-012: Extend D3 anti-pattern format to include SYMPTOM and CONSEQUENCE"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/dimension-improvements/d3-anti-pattern-coverage.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The D3 (Anti-Pattern Coverage) scorer parsed skills for the standard NEVER/WHY/BAD/GOOD anti-pattern format. This format described the prohibition and the correct alternative but lacked explicit fields for how to recognise the anti-pattern (SYMPTOM) and what goes wrong (CONSEQUENCE), making detection and severity assessment harder.

## Decision

Extend the anti-pattern format from NEVER/WHY/BAD/GOOD to **NEVER/WHY/SYMPTOM/CONSEQUENCE/BAD/GOOD**. Switch from per-document parsing to per-block parsing to support multiple anti-patterns in a single skill. Rename the file from `d3_anti_pattern.go` to `d3_anti_pattern_coverage.go` to match the dimension name.

## Consequences

- Anti-patterns become more actionable with explicit symptom and consequence fields
- Per-block parsing supports multi-anti-pattern skills
- Existing skills with the old format continue to parse (backward compatible)
- File rename aligns source naming with dimension naming
