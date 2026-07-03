---
title: "ADR-031: analysis_quality patterns stay outside the D1-D9 dimension model"
status: accepted
date: 2026-07-03
context:
  - path: ".context/findings/scoring-pattern-config-review-2026-07-03.md"
---

**Status:** Accepted
**Date:** 2026-07-03

## Context

`cmd/assets/assets/config/scoring-patterns.yaml` (ADR-028) groups pattern lists under three keys: `d1_knowledge_delta`, `analysis_quality`, and `d6_freedom_calibration`. The first and third map 1:1 to scorer dimensions (D1, D6) and feed the `evaluate` D1-D9 pipeline. `analysis_quality` (hedge/vague/passive-voice word lists) does not — its sole consumer, `analysis.DetectAntiPatternSignals`, is only called from the standalone `skill-auditor analyze` command and never affects a skill's scored total.

The rule names it produces (`anti-pattern:hedge-language`, `anti-pattern:vague-instructions`, `anti-pattern:passive-voice`) suggest a natural home in D3 (Anti-Pattern Coverage), prompting the question of whether it should be folded into a dimension.

## Decision

Keep `analysis_quality` outside the D1-D9 dimension model. Specifically:

1. Do not fold its word lists into D3. `scorer/d3_anti_pattern_coverage.go` already implements a structurally different check — it scores whether `**NEVER**`-anchored blocks have BAD/GOOD, WHY, SYMPTOM, and CONSEQUENCE components (documentation quality of anti-patterns), not whether the skill's own prose is hedgy or vague. Merging would conflate two distinct concepts and silently change D3's scored behaviour for every skill.
2. If this is revisited, D6 (Freedom Calibration) is the better conceptual fit — hedge/vague/passive language reflects instruction confidence, which D6 already measures via hard/soft marker balance. This is not decided or scoped; it is only recorded as the preferred direction if a future decision is made to fold it in.
3. Keep the `analysis_quality` key name and grouping as-is in `scoring-patterns.yaml`. Renaming it to look dimension-shaped (e.g. `d3_...`) would misrepresent that it doesn't feed D1-D9 scoring.

## Consequences

- **Easier:** No ambiguity for future readers of `scoring-patterns.yaml` about why one group doesn't follow the `d{N}_...` naming convention — this ADR is the answer.
- **Easier:** D3's scoring behaviour and tests remain untouched; no regrading risk introduced by this refactor.
- **Harder:** The `analyze` command's anti-pattern-style rule names remain slightly misleading (they say "anti-pattern" but don't feed D3) until/unless a future decision renames or relocates them.
- **Deferred:** Whether to eventually fold `analysis_quality` into D6 remains an open, unscoped question — not resolved by this ADR.
