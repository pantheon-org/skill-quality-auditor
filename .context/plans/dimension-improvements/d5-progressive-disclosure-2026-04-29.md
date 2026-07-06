---
title: "Improvement Plan: D5 — Progressive Disclosure"
type: PLAN
status: DONE
date: 2026-04-29
value: MEDIUM
---
# Improvement Plan: D5 — Progressive Disclosure

**Date:** 2026-04-29  
**Current max score:** 15 pts  
**Priority:** Low  
**Effort:** Small

---

## Current Implementation

`scoreD5` in `scorer/d5_progressive_disclosure.go` is a **size/refs heuristic** — it does NOT parse skill content for load conditions, hub structure, or lazy-load instructions. The actual algorithm:

1. Counts `.md` files in `<skillDir>/references/` → sets `hasRefs bool` and `refCount int`.
2. Counts newlines in the SKILL.md content → `lines int`.
3. Prefers a token count from `validatorBridge.skillMDTokens()` when non-zero.
4. Dispatches to `scoreD5ByTokens` or `scoreD5ByLines`, both of which apply a two-axis lookup:
   - **With refs:** 15 / 13 / 11 / 10 pts depending on compactness thresholds.
   - **Without refs:** 12 / 10 / 7 / 5 pts depending on length thresholds.

There is also `isReferenceSectionCompliant(content string) bool` which checks whether `## References` is the last H2 and contains at least one bullet link. **This function is defined but never called anywhere in the codebase — it is dead code.**

The functions named in the old plan (`scoreHubStructure`, `scoreLoadConditions`, `scoreLazyLoadInstruction`) **do not exist**.

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| Progressive Disclosure: Designing for Effective Transparency | A Springer, S Whittaker | https://arxiv.org/abs/1811.02164 |
| The Role of Cognitive Load in Shaping Web Usability Requirements | A Timileyin | https://papers.ssrn.com/sol3/papers.cfm?abstract_id=5247018 |
| Designing Effective Training Dataset Explanations: The Impact of Information Depth and Progressive Disclosure | AI Anik, A Bunt | ACM CHI 2022 |
| AI-Enhanced Modular Information Architecture for Cognitive-Efficient User Experiences | F Pastrakis, M Konstantakis, G Caridakis | https://www.mdpi.com/2078-2489/17/1/92 |

**Key finding:** Springer & Whittaker (arXiv:1811.02164) is the canonical progressive disclosure paper for AI systems. It defines two failure modes: *over-disclosure* (too much information upfront, overwhelming working memory) and *passive availability* (deeper content exists but the agent is never told when to access it). Anik & Bunt's empirical study on dataset explanations shows that progressive disclosure only reduces cognitive load when *disclosure conditions are explicit* — agents and users must know not just that deeper content exists, but precisely when to access it. Passive availability produces the same outcome as no deeper content at all. Timileyin grounds this in Sweller's Cognitive Load Theory.

---

## Problem Statement

The current size/refs heuristic is a weak proxy for progressive disclosure quality. It rewards brevity and penalises length regardless of whether the skill actually teaches agents *when* to load references. A compact skill with no load conditions scores higher than a well-structured skill with explicit guidance. The heuristic also ignores `isReferenceSectionCompliant`, leaving useful signal completely unused.

The improvement adds `scoreNegativeConditions()` as a **purely additive sub-scorer** bolted onto the existing heuristic result, freeing 2 pts from the existing score ceiling to reward negative-trigger language in reference table rows.

---

## Proposed Change

### Option chosen: (b) — additive `scoreNegativeConditions()` integrated with the existing size-based score

The existing `scoreD5ByTokens` / `scoreD5ByLines` heuristics are retained unchanged. A new function `scoreNegativeConditions(content string) int` returns 0–2 pts by scanning only markdown table rows for negative-trigger keywords. `scoreD5WithMeta` is updated to call this function and add its result to the heuristic score, capped at 15.

**Why option (b) and not (a):** Rewriting the full scorer would invalidate all existing baselines and require updating every testdata fixture. The additive approach delivers the academic improvement incrementally without disrupting the scoring floor.

### Integration logic

```
finalScore = min(15, heuristicScore + scoreNegativeConditions(content))
```

`heuristicScore` is the existing return value of `scoreD5ByTokens` or `scoreD5ByLines`. Because the maximum heuristic score with refs is already 15, the +2 only lifts scores that are currently capped at 13 or lower. The effective ceiling remains 15.

### `scoreNegativeConditions` specification

```go
// scoreNegativeConditions scans markdown table rows in content for negative
// trigger language and returns 0–2 pts. Matching is restricted to table rows
// (lines beginning with '|') to avoid false positives from prose.
func scoreNegativeConditions(content string) int
```

**Keywords to match (case-insensitive, whole-word where practical):**
`skip if`, `only when`, `unless`, `not needed when`, `omit if`

**Scoping rule — table rows only:** A line qualifies for keyword matching only if it starts with `|` (after trimming leading whitespace). This prevents prose sentences such as "skip if you have time" from triggering the scorer.

**Scoring:**
- **2 pts:** At least one qualifying table row contains a negative-trigger keyword.
- **0 pts:** No qualifying table row contains a negative-trigger keyword.

*(A 1-pt band was considered but the matching is binary — either a table has "When to Skip" language or it does not. The 0/2 step avoids ambiguity about "most" vs "all" rows.)*

### Scoring adjustment (15 pts total, unchanged)

| Component | Before this change | After this change |
|---|---|---|
| Size / refs heuristic (existing) | 5–15 pts | 5–13 pts (ceiling lowered by 2) |
| **Negative conditions (`scoreNegativeConditions`)** | — | 0–2 pts |
| **Total** | 5–15 pts | 5–15 pts |

The heuristic ceiling is reduced from 15 to 13 by lowering the `d5TokenCompact` / `d5LinesCompact` band return value from 15 to 13. All other bands are unchanged. This preserves the score distribution for skills without a references directory while reserving the top 2 pts for explicit negative-condition language.

### Dead code: `isReferenceSectionCompliant`

`isReferenceSectionCompliant` is defined but never called. It must be **deleted** as part of this change. Its check (last H2 is `## References` with ≥1 bullet link) is not incorporated into `scoreNegativeConditions` — that concern belongs to D4 (Specification Compliance). Retaining dead code violates the no-deprecated-functionality project rule.

---

## Implementation Steps

1. **Delete `isReferenceSectionCompliant`** from `scorer/d5_progressive_disclosure.go` — it is dead code and must be removed.
2. **Lower the compact-band return values** in `scoreD5ByTokens` and `scoreD5ByLines`: change `return 15, ...` (the `tokens < d5TokenCompact` / `lines < d5LinesCompact` with-refs branch) to `return 13, ...`. All other return values are unchanged.
3. **Add `scoreNegativeConditions(content string) int`** — scans only `|`-prefixed lines for the five negative-trigger keywords; returns 2 if any match, 0 otherwise.
4. **Update `scoreD5WithMeta`** to call `scoreNegativeConditions(content)` and return `min(15, score+negScore)` in place of the bare `score` value.
5. **Update unit tests** in `scorer/d5_progressive_disclosure_test.go`:
   - Update `TestD5_WithRefsShort` expectation: compact with refs now scores 13 (heuristic) + 0 (no negative conditions in `makeLines` fixture) = 13. Add a parallel test `TestD5_WithRefsShort_NegativeConditions` that injects a table row with `skip if` language and expects 15.
   - Add `TestScoreNegativeConditions` covering: 0 pts (no table rows), 0 pts (table rows without keywords), 0 pts (keyword in prose paragraph only — false-positive guard), 2 pts (keyword in a `|`-prefixed row).
6. **Update `cmd/assets/references/framework-dimensions.md`** — document the negative condition requirement, cite Springer & Whittaker (arXiv:1811.02164) and Anik & Bunt.
7. **Update the references table template** in `cmd/assets/` to show a 4-column format (Topic / Reference / When to Load / When to Skip).
8. **Update `testdata/` fixtures** if any fixture relied on a score of 15 via the compact-with-refs path.
9. Run `go test ./scorer/...`.

---

## Acceptance Criteria

- `isReferenceSectionCompliant` is deleted; `go vet ./...` passes with no dead-code warnings.
- A skill with a compact SKILL.md, a references directory, and no negative-trigger table rows scores 13/15 (was 15/15 before this change).
- A skill with a compact SKILL.md, a references directory, and at least one `|`-prefixed row containing `skip if` scores 15/15.
- A skill with a negative-trigger keyword in **prose** (not a table row) scores 0 on `scoreNegativeConditions` — no false positive.
- `TestScoreNegativeConditions` covers all four paths: no table, table without keyword, keyword in prose, keyword in table row.
- The template update does not break existing skills that use the 3-column format (they score 0 on the new sub-criterion, not an error).
- `go test ./scorer/...` passes with no regressions.
