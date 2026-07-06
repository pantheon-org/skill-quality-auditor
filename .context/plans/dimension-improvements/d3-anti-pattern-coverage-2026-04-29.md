---
title: "Improvement Plan: D3 — Anti-Pattern Coverage"
type: PLAN
status: DONE
date: 2026-04-29
value: MEDIUM
---
# Improvement Plan: D3 — Anti-Pattern Coverage

**Date:** 2026-04-29  
**Current max score:** 15 pts  
**Priority:** Medium  
**Effort:** Small

---

## Current Implementation

The scorer lives in `scorer/d3_anti_pattern.go`; its test file is `scorer/d3_anti_pattern_test.go`. Both exist today.

`scoreD3` is the top-level entry point (max 15 pts). It delegates to two helpers:

| Helper | What it checks | Points |
|---|---|---|
| `scoreD3DirectiveLanguage` | Counts strong-marker words (NEVER, ALWAYS, AVOID, DO NOT, anti-pattern) via `reD3AntiInstr`; uses `validatorBridge.Content.StrongMarkers` when available | 0–5 |
| Direct checks in `scoreD3` | `reD3BadGood` (`(?is)BAD.*GOOD` across full document); `countPattern(content, "WHY:")` | 0–4 |
| `scoreD3FromInstructions` | Reads `evals/instructions.json`; counts instructions whose `type == "anti-pattern"` or whose snippet matches `reD3AntiInstr` | 0–2 |

**Key gaps in current logic:**
- `reD3BadGood` is a single document-wide regex — it fires once regardless of how many anti-patterns are present. There is no per-block parsing.
- SYMPTOM and CONSEQUENCE components are not detected at all.
- The WHY check is document-wide (`countPattern`), not scoped to individual anti-pattern blocks.
- No NEVER count lower-bound enforces a minimum number of anti-patterns; point allocation is ad-hoc.

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| Software Process Anti-Patterns Catalogue | P Brada, P Picha | https://dl.acm.org/doi/abs/10.1145/3361149.3361178 |
| Software Process Anti-Pattern Detection in Project Data | P Picha, P Brada | https://dl.acm.org/doi/abs/10.1145/3361149.3361169 |
| Data Quality Anti-Patterns for Software Analytics | A Bhatia, D Lin, GK Rajbahadur, B Adams et al. | https://arxiv.org/abs/2408.12560 |
| Code Quality Alarms: Techniques, Datasets, and Emerging Trends in Detecting Smells and Anti-Patterns | YVA Amarasinghe, P Asanka et al. | https://jdrra.sljol.info/articles/10.4038/jdrra.v3i2.93 |

**Key finding:** Brada & Picha's canonical anti-pattern catalogue defines four required components for an actionable anti-pattern: *context* (when it emerges), *symptoms* (observable signals), *root cause* (WHY it is wrong), and *refactored solution* (what to do instead). The data quality anti-patterns paper (Bhatia et al.) empirically shows that anti-patterns without consequence statements are rarely acted on in practice — engineers acknowledge them but do not change behaviour without knowing *what goes wrong*. The code quality alarms survey confirms that symptom specification is the single strongest predictor of whether an anti-pattern is adopted by practitioners.

---

## Problem Statement

The current D3 format requires NEVER/WHY/BAD/GOOD structure. This is a strong baseline but is missing two components from the canonical anti-pattern structure:

1. **Symptom** — what the agent or reviewer will *observe* when the anti-pattern has occurred (the signal that something is wrong, before knowing the cause).
2. **Consequence** — what breaks downstream if the anti-pattern is not corrected.

Without these, an agent reading a NEVER/WHY/BAD/GOOD block knows what not to do and why in the abstract, but cannot self-diagnose when it has already done it, nor prioritise remediation.

---

## Proposed Change

### Extend to NEVER / WHY / SYMPTOM / CONSEQUENCE / BAD / GOOD

**New required format:**

```markdown
**NEVER** [concise statement of the forbidden action]

**WHY:** [root cause — the underlying constraint or invariant being violated]

**SYMPTOM:** [observable signal that this has occurred — what the agent or reviewer will see]

**CONSEQUENCE:** [what breaks downstream if uncorrected]

**BAD:**
[concrete bad example]

**GOOD:**
[concrete corrected example]
```

**Scoring adjustment (15 pts total, unchanged):**

| Component | Current | New |
|---|---|---|
| NEVER statement | 3 pts | 2 pts |
| WHY explanation | 3 pts | 2 pts |
| BAD example | 3 pts | 2 pts |
| GOOD example | 3 pts | 2 pts |
| **SYMPTOM** | — | 3 pts |
| **CONSEQUENCE** | — | 2 pts |
| Count (≥ 3 anti-patterns) | 3 pts | 2 pts |

> The existing 3-pt count bonus is reduced by 1 pt to fund the new components, while NEVER and WHY lose 1 pt each (they remain mandatory but are now shorter checks).

---

## Implementation Steps

1. Rename `scorer/d3_anti_pattern.go` → `scorer/d3_anti_pattern_coverage.go` and `scorer/d3_anti_pattern_test.go` → `scorer/d3_anti_pattern_coverage_test.go` to match the naming convention used by the other dimension scorers. The `package scorer` declaration and all exported/unexported symbols remain unchanged.

2. Create `scorer/d3_anti_pattern_coverage_test.go` (after the rename in step 1). The test file does not exist yet at the target path, so the rename doubles as its creation. Add table-driven tests covering:
   - Full 6-component block (all six keywords present) → expects score ≥ 13.
   - 4-component block (NEVER/WHY/BAD/GOOD only, no SYMPTOM/CONSEQUENCE) → expects score ≤ 10.
   - SYMPTOM keyword appearing only inside a `**WHY:**` paragraph → expects SYMPTOM score = 0 (disambiguation test).
   - CONSEQUENCE keyword appearing only inside a `**BAD:**` block → expects CONSEQUENCE score = 0 (disambiguation test).

3. Extend `scorer/d3_anti_pattern_coverage.go` — introduce `parseAntiPatternBlocks(content string) []antiPatternBlock` to split the skill document into individual NEVER-anchored blocks before scoring. Each block is delimited from one `**NEVER**` marker to the next (or end of document). Score SYMPTOM and CONSEQUENCE per block, not per document.

4. Add `scoreSymptom(block string) int` and `scoreConsequence(block string) int` with scoped keyword matching (see Disambiguation Strategy below). Update the rubric weights map to reflect the new 6-component point allocation.

5. Update `cmd/assets/references/framework-dimensions.md` and `cmd/assets/references/detailed-anti-patterns.md` — document the extended format, cite Brada & Picha and Bhatia et al.

6. Update `cmd/assets/references/anti-patterns.md` template to show the new 6-component structure.

7. Update `testdata/` fixtures — ensure at least one fixture exercises the full new format and one is missing SYMPTOM/CONSEQUENCE.

8. Run `go test ./scorer/...`.

### Disambiguation Strategy

SYMPTOM and CONSEQUENCE keywords appear naturally in other section headers (WHY explains root causes that can mention symptoms; BAD examples can describe consequences). To avoid false positives:

- **Parse per block first.** Call `parseAntiPatternBlocks` before any keyword search. Never run SYMPTOM/CONSEQUENCE regexes against the full document.
- **Match the section header, not prose.** A SYMPTOM is scored only when the block contains a line matching `^\*\*SYMPTOM[:\*]` (i.e., the keyword is a bold header, not embedded mid-sentence). Likewise for CONSEQUENCE.
- **Exclude subordinate sections.** Within a block, lines between `**WHY:**` and the next `**SYMPTOM:**` / `**CONSEQUENCE:**` / `**BAD:**` header are treated as WHY prose; SYMPTOM/CONSEQUENCE keywords in that span do not score.
- **Minimum content guard.** A matched SYMPTOM section must contain at least one non-header, non-blank line after the header; an empty `**SYMPTOM:**` header with no body does not score.

### Per-Block vs. Per-Document Clarification

All six components (NEVER, WHY, SYMPTOM, CONSEQUENCE, BAD, GOOD) are checked **per anti-pattern block**. A block is the text from one `**NEVER**` marker to the next (or end of document).

- The per-block score is then averaged (or summed with a cap) across all blocks to produce the document-level D3 score.
- The count bonus (2 pts) is awarded once at the document level when ≥ 3 valid blocks are present.
- A "valid block" requires at minimum a NEVER statement and one of WHY/BAD/GOOD — blocks with only a NEVER header and no body do not count toward the minimum.

---

## Acceptance Criteria

- An anti-pattern with NEVER/WHY/BAD/GOOD but no SYMPTOM or CONSEQUENCE scores ≤ 10/15.
- An anti-pattern with all six components scores ≥ 13/15.
- The scorer correctly identifies SYMPTOM keywords (e.g., "you will see", "the agent will", "observable when", "symptom:", "signs:") **only when they appear as a `**SYMPTOM:**` section header within a NEVER block**.
- The scorer correctly identifies CONSEQUENCE keywords (e.g., "this causes", "consequence:", "downstream", "results in", "breaks") **only when they appear as a `**CONSEQUENCE:**` section header within a NEVER block**.
- A SYMPTOM keyword that appears inside a `**WHY:**` paragraph does not score as a SYMPTOM.
- A CONSEQUENCE keyword that appears inside a `**BAD:**` code block does not score as a CONSEQUENCE.
- `go test ./scorer/...` passes with no regressions.
