---
title: "Finding: Scoring Pattern Config Review — Naming and Dimension Mapping"
type: FINDING
status: ACTIVE
date: 2026-07-03
value: MEDIUM
themes:
  - SKILL-QUALITY
related:
  - ../plans/yaml-content-validation-config-2026-07-03.md
  - ../findings/yaml-content-validation-config-2026-07-03.md
  - ../../docs/ADR/adr-028-scoring-pattern-config.md
  - ../../cmd/assets/assets/config/scoring-patterns.yaml
  - ../../analysis/patterns.go
  - ../../scorer/d3_anti_pattern_coverage.go
  - ../../scorer/d6_freedom_calibration.go
---
# Finding: Scoring Pattern Config Review — Naming and Dimension Mapping

> Post-implementation review of `scoring-patterns.yaml` surfaced a naming inconsistency and an open question about whether the `analysis_quality` pattern group belongs under a scoring dimension.

## Summary

While implementing the `yaml-content-validation-config-2026-07-03` plan (Phase 1: externalise hardcoded scoring patterns to `cmd/assets/assets/config/scoring-patterns.yaml`), a review of the config's three top-level groups — `d1_knowledge_delta`, `analysis_quality`, `d6_freedom_calibration` — found that `analysis_quality` is structurally different from the other two, and confirmed the plan's YAML-only decision is documented independently in ADR-028.

## Detail

### 1. `analysis_quality` is the odd one out

`d1_knowledge_delta` and `d6_freedom_calibration` are named after scorer dimensions and map 1:1 to entries in `scorer/dimensions.go`'s `AllDimensions` (D1 "knowledgeDelta", D6 "freedomCalibration"). Both are consumed directly inside the corresponding `scoreD1`/`scoreD6` functions and therefore affect a skill's D1–D9 total.

`analysis_quality` (hedge words, vague words, passive-voice patterns) is not a dimension. Its sole consumer, `analysis.DetectAntiPatternSignals`, is called only from `cmd/analyze.go` — the standalone `skill-auditor analyze` command (TF-IDF keyword extraction + rule-based pattern report). It never feeds the `evaluate` D1–D9 scoring pipeline. Verified via `grep -rn "DetectAntiPatternSignals"`.

### 2. Why it can't be cleanly folded into D3 despite the "anti-pattern:" rule prefix

`DetectAntiPatternSignals`'s rule names (`anti-pattern:hedge-language`, `anti-pattern:vague-instructions`, `anti-pattern:passive-voice`) suggest a mapping to D3 (Anti-Pattern Coverage), but `scorer/d3_anti_pattern_coverage.go` already implements a much richer, structurally different check: it parses `**NEVER**`-anchored blocks and scores whether each has BAD/GOOD, WHY, SYMPTOM, and CONSEQUENCE components. D3 measures how well anti-patterns are *documented*; `analysis_quality`'s word lists measure whether the skill's own prose is hedgy/vague/passive. These are different concepts that happen to share the word "anti-pattern."

Folding the word lists into D3 would change D3's actual scoring behavior for every skill (regrading, updated tests) — a scoring-behavior change, not a config-externalization refactor, and out of scope for the current plan.

### 3. If folded into a dimension, D6 is the better conceptual fit

Hedge/vague/passive language signals weak instruction confidence, which is what D6 (Freedom Calibration) already measures via its hard/soft marker balance (`scoreConstraintTypology`, `scoreCalibrationBalance`). This is a plausible future direction but has not been decided or scoped.

### 4. ADR-028 confirms the YAML-only decision but is still `status: proposed`

`docs/ADR/adr-028-scoring-pattern-config.md` independently documents "Config format is YAML only. No TOML fallback" (decision point 2), matching the plan's Phase 1 scope as implemented. However its frontmatter status is still `proposed`, even though Phase 1 (decisions 1, 2, 4, 5) is now fully implemented and merged. Decision 3 (Phase 2 content-safety patterns) remains outstanding.

## Recommended Action

- Leave `analysis_quality` named and grouped as-is in `scoring-patterns.yaml` — renaming it to look dimension-shaped (e.g. `d3_...`) would misrepresent that it doesn't feed D1–D9 scoring.
- Treat "fold `analysis_quality` word lists into D6" as a separate, explicitly-scoped follow-up decision — not a config-file change.
- Flip ADR-028's frontmatter `status: proposed → accepted` to reflect that Phase 1 is implemented, keeping Decision 3 (Phase 2) tracked as the remaining open item.
