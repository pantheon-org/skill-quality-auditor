---
title: "Improvement Plan: D6 — Freedom Calibration"
type: plan
status: active
date: 2026-04-29
---
# Improvement Plan: D6 — Freedom Calibration

**Date:** 2026-04-29  
**Current max score:** 15 pts  
**Priority:** Medium  
**Effort:** Medium (all sub-components built from scratch; scorer API extension required)

---

## Current Implementation

`scoreD6` in `scorer/d6_freedom_calibration.go` is a single-expression scorer with no sub-components:

```go
func scoreD6(b *validatorBridge) (int, []Diagnostic) {
    if b.Content == nil { return 0, [...] }
    if b.Content.StrongMarkers+b.Content.WeakMarkers == 0 { return 0, [...] }
    score := int(math.Round(b.Content.InstructionSpecificity * 15))
    // clamp [0,15]
    return score, nil
}
```

**What this means:**
- The entire 15-point score is a single linear scaling of `ContentReport.InstructionSpecificity`, a pre-computed float from the `skill-validator` library (`orchestrate.RunContentAnalysis`).
- `InstructionSpecificity` is defined as `StrongMarkers / (StrongMarkers + WeakMarkers)` (or `1.0` when both are zero, overridden to `0` by the zero-marker guard).
- There are **no separate sub-components** (no calibration-balance bucket, no when-not-to-use bucket, no permissive/imperative ratio bucket). The Proposed Change's scoring rebalance table describes a *desired future state*, not the current code.
- `scoreD6` receives only a `*validatorBridge` — it has no access to the raw skill text or `skillDir`. The registry closure discards both `content` and `dir` before calling it:
  ```go
  {AllDimensions[5], func(_, _ string, b *validatorBridge) (int, []Diagnostic) {
      return scoreD6(b)
  }}
  ```
- `ContentReport` exposes pre-aggregated counts (`StrongMarkers`, `WeakMarkers`, `InstructionSpecificity`, `ImperativeRatio`, `ListItemCount`, `CodeBlockCount`, `CodeLanguages`). It does **not** expose raw text or individual token positions — so `scoreConstraintTypology()` cannot be implemented against `b.Content` alone.

**Seven existing unit tests** all mock `ContentReport` fields directly and test the clamp/round behaviour of the single-expression formula. No test covers constraint-typology semantics.

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| Reasoning over Boundaries: Enhancing Specification Alignment via Test-time Deliberation | H Zhang, Y Li, X Hu, D Liu, Z Wang, B Li et al. | https://arxiv.org/abs/2509.14760 |
| Specification as the New Management | S Sorensen | https://www.researchgate.net/publication/401626622 |
| LLM-Skill Orchestration: Achieving 202/202 Subtask Completion via Rule-Augmented Multi-Model Collaboration | R Tao | https://www.researchsquare.com/article/rs-9323974/latest |

**Key finding:** "Reasoning over Boundaries" (Zhang et al. 2025) shows that specification alignment degrades when constraints are applied uniformly — agents need to distinguish *hard boundaries* (never cross, regardless of context) from *soft preferences* (default behaviour, overridable by context). Applying all constraints at the same rigidity level causes agents to over-refuse on edge cases where soft preferences should yield, and to under-enforce on cases where hard constraints should hold. "Specification as the New Management" (Sorensen 2026) frames this as *pre-delegation architecture*: the key enterprise skill is knowing which decisions to pre-specify vs. which to delegate to agent judgement. LLM-Skill Orchestration empirically demonstrates that rule-augmented agents with typed constraints (hard vs. soft) outperform uniformly-constrained agents on structured tasks, and that uniform over-specification degrades performance on open-ended tasks.

---

## Problem Statement

D6 currently scores whether a skill is appropriately calibrated — not too rigid, not too permissive — but treats this as a holistic judgment rather than a structural check. The scorer assesses the ratio of imperative vs. permissive language, and the presence of "when not to use" guidance.

The problem: a skill can pass D6 today by having some permissive language and some imperative language, without ever distinguishing *which* constraints are hard and *which* are soft. An agent reading such a skill has no basis for deciding which rules bend in unusual situations and which do not. This produces the two failure modes identified by Zhang et al.: over-refusal (treating soft preferences as hard limits) and under-enforcement (treating hard limits as negotiable).

---

## Proposed Change

### Add "Constraint Typology" criterion — explicit hard vs. soft distinction (4 pts)

A well-calibrated skill must visibly differentiate constraint strength. This can be done via explicit markers or via a dedicated section.

**Hard constraints** (MUST / NEVER / ALWAYS) — invariant, never yielded to context.  
**Soft defaults** (PREFER / AVOID / BY DEFAULT / UNLESS) — default behaviour that context may override.

**Scoring:**

- **4 pts:** Skill uses consistent markers distinguishing hard from soft constraints AND has at least 2 examples of each type.
- **3 pts:** Skill distinguishes types but inconsistently (some constraints unmarked or ambiguously typed).
- **2 pts:** Skill has hard constraints and soft guidance, but without explicit typology markers (ambiguous to the agent).
- **1 pt:** Skill is entirely hard constraints with no soft defaults, or entirely soft with no hard limits.
- **0 pts:** No constraint language at all, or the distinction is undetectable.

**Scoring rebalance (15 pts total, unchanged):**

| Component | Current | New |
|---|---|---|
| Calibration balance (not too rigid, not too free) | 6 pts | 5 pts |
| When-not-to-use guidance | 4 pts | 3 pts |
| Permissive/imperative ratio | 5 pts | 3 pts |
| **Constraint typology** (new) | — | 4 pts |

---

## Raw Text Access Strategy

`scoreConstraintTypology()` must scan the raw SKILL.md text for typed constraint markers. `ContentReport` does not expose this text — it only exposes pre-aggregated counts. The raw text is available in two ways:

**Option A — Change `scoreD6` signature to accept `content string` (recommended)**

Mirror the pattern used by `scoreD1`, `scoreD3`, and `scoreD8`, which accept a `content string` parameter. This requires:
1. Update `scoreD6(b *validatorBridge)` → `scoreD6(content string, b *validatorBridge)`.
2. Update the registry closure in `scorer/scorer.go` (`AllDimensions[5]`) to pass `c` instead of `_`:
   ```go
   {AllDimensions[5], func(c, _ string, b *validatorBridge) (int, []Diagnostic) {
       return scoreD6(c, b)
   }}
   ```
3. Update all existing unit tests to pass a `content` string — the 7 existing tests all call `scoreD6(b)` directly.

**Option B — Read SKILL.md inside scoreD6 via skillDir on validatorBridge**

`validatorBridge` does not carry `skillDir`. Adding it would be a larger API change and conflicts with the existing bridge design (bridge is constructed from `skillDir` but does not store it). Not recommended.

**Decision: implement Option A.** Call sites are limited to the registry closure in `scorer/scorer.go` (one line) and the 7 unit tests in `scorer/d6_freedom_calibration_test.go`. No other files reference `scoreD6`.

---

## Implementation Steps

All sub-components listed below must be **built from scratch** — there are currently zero sub-components in `scoreD6`. The existing scorer body is replaced, not extended.

### Step 1 — Extend `scoreD6` signature and update call site

- Change `func scoreD6(b *validatorBridge) (int, []Diagnostic)` to `func scoreD6(content string, b *validatorBridge) (int, []Diagnostic)`.
- In `scorer/scorer.go`, update `AllDimensions[5]` closure to pass `c` as content.
- Confirm no other call sites exist (grep confirms only `scorer.go` and the test file call `scoreD6`).

### Step 2 — Build `scoreCalibrationBalance()` (5 pts, built from scratch)

Replaces the current single `InstructionSpecificity * 15` expression as one sub-bucket. Scores whether the ratio of strong-to-weak markers falls in a balanced range (neither all-imperative nor all-permissive). Can continue to use `b.Content.InstructionSpecificity` — this is the correct signal for balance. Thresholds: 0.3–0.8 → 5 pts; 0.2–0.3 or 0.8–0.9 → 3 pts; outside that range → 1 pt; zero markers → 0 pts.

### Step 3 — Build `scoreWhenNotToUse()` (3 pts, built from scratch)

Scans `content` for a "when not to use" section or equivalent negative-scope language (`do not use`, `not intended for`, `outside the scope`, `avoid using`). Returns 3 pts if present, 0 pts otherwise.

### Step 4 — Build `scoreConstraintTypology()` (4 pts, built from scratch)

Scans `content` (raw markdown string) using `strings.ToUpper` + `strings.Contains` / `regexp.MustCompile`:
- Hard constraint markers: `MUST`, `NEVER`, `ALWAYS`, `REQUIRED`, `PROHIBITED`
- Soft default markers: `PREFER`, `AVOID`, `BY DEFAULT`, `UNLESS`, `TYPICALLY`, `RECOMMENDED`

Count occurrences of each class and apply the rubric:
- **4 pts:** ≥ 2 hard and ≥ 2 soft markers present.
- **3 pts:** ≥ 1 hard and ≥ 1 soft, but fewer than 2 of one type.
- **2 pts:** both types present but no explicit uppercase markers (detected via `b.Content.ImperativeRatio` > 0 and > 0 weak markers, but no ALLCAPS hits).
- **1 pt:** only hard or only soft markers, not both.
- **0 pts:** no constraint language detected at all.

### Step 5 — Compose `scoreD6`

```go
func scoreD6(content string, b *validatorBridge) (int, []Diagnostic) {
    if b.Content == nil { return 0, [...] }
    score := scoreCalibrationBalance(b)
      + scoreWhenNotToUse(content)
      + scoreConstraintTypology(content, b)
    // clamp [0,15]
}
```

Remove the zero-marker early-return guard — `scoreCalibrationBalance` handles zero markers with a 0-pt return.

### Step 6 — Update `cmd/assets/references/framework-dimensions.md`

Document the typology requirement; cite Zhang et al. (arXiv:2509.14760) and Sorensen 2026. Add the two-section guidance example:

```markdown
## Hard Constraints
- NEVER commit directly to main
- ALWAYS require a passing test suite before reporting completion

## Soft Defaults
- PREFER conventional commits; adapt if repo already uses a different convention
- AVOID inline comments unless the WHY is non-obvious
```

### Step 7 — Update testdata fixtures and tests

**Fixture delta scores (expected changes under the new scorer):**

| Fixture | Current D6 (approx) | Expected D6 (new) | Delta | Reason |
|---|---|---|---|---|
| `skill-minimal` | 0 (zero markers, no content) | 0 | 0 | Minimal fixture has no directive language — all three sub-buckets score 0. |
| `skill-full` | ~13 (InstructionSpecificity ≈ 1.0, 6 ALLCAPS hard markers, 0 soft markers counted by grep) | ~9 (5 balance + 3 when-not-to-use + 1 typology) | −4 | `skill-full` has only hard markers (MUST/NEVER/ALWAYS) and no soft defaults — typology scores 1 pt. If a "when not to use" section exists, adds 3 pts; review fixture to confirm. Balance bucket likely 5 pts given high `InstructionSpecificity`. Net is approximately 9 pts assuming no soft defaults are found. **Fixture must be updated to include ≥ 2 soft defaults to reach ≥ 13 pts for the "well-calibrated" test in Acceptance Criteria.** |

**Test updates required:**
- All 7 existing tests call `scoreD6(b)` — update to `scoreD6("", b)` or provide representative `content` strings. Tests covering clamping/rounding behaviour remain valid; mock content can be an empty string for bridge-only tests.
- Add new tests per sub-component: `TestD6_ConstraintTypology_BothTypes`, `TestD6_ConstraintTypology_HardOnly`, `TestD6_ConstraintTypology_SoftOnly`, `TestD6_ConstraintTypology_None`, `TestD6_WhenNotToUse_Present`, `TestD6_WhenNotToUse_Absent`.
- Add integration-level test using `skill-full` fixture to assert expected total score.

### Step 8 — Run tests

```bash
go test ./scorer/...
```

---

## Acceptance Criteria

- `go test ./scorer/...` passes with no regressions.
- A skill with only hard constraints (all MUST/NEVER, no soft defaults) scores ≤ 9/15.
- A skill with only soft guidance (all PREFER/AVOID, no hard limits) scores ≤ 9/15.
- A skill with ≥ 2 hard and ≥ 2 soft markers and a "when not to use" section scores ≥ 13/15.
- `skill-minimal` fixture continues to score 0 on D6.
- `skill-full` fixture score is re-baselined after adding soft defaults; expected post-update score ≥ 13 pts.
- `scoreD6` signature updated to `(content string, b *validatorBridge)` and registry closure updated in `scorer/scorer.go`.
