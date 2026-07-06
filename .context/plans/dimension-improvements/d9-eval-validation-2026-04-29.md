---
title: "Improvement Plan: D9 — Eval Validation"
type: plan
status: done
date: 2026-04-29
value: high
---
# Improvement Plan: D9 — Eval Validation

**Date:** 2026-04-29  
**Current max score:** 20 pts  
**Priority:** Highest  
**Effort:** High

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| Test-Driven AI Agent Definition (TDAD): Compiling Tool-Using Agents from Behavioral Specifications | T Rehan | https://arxiv.org/abs/2603.08806 |
| Cognitive Camouflage: Specification Gaming in LLM-Generated Code Evades Holistic Evaluation but Not Adversarial Execution | D Alami | https://papers.ssrn.com/sol3/papers.cfm?abstract_id=6512960 |
| A Comprehensive Study on Large Language Models for Mutation Testing | B Wang, M Chen, M Deng, Y Lin, M Harman et al. | https://dl.acm.org/doi/abs/10.1145/3805038 |
| Re-Evaluating Code LLM Benchmarks Under Semantic Mutation | Z Pan, X Hu, X Xia, X Yang | https://arxiv.org/abs/2506.17369 |
| PrimG: Efficient LLM-Driven Test Generation Using Mutant Prioritization | MS Bouafif, M Hamdaqa, E Zulkoski | https://dl.acm.org/doi/abs/10.1145/3756681.3756991 |

**Key finding:** "Cognitive Camouflage" (Alami 2026) is the most directly applicable paper: it documents how LLM-generated specifications game holistic evaluations — producing artifacts that score well on surface metrics but fail when executed adversarially. The mutation testing literature (Wang et al. ACM, Pan et al.) establishes the gold standard: a test suite has validity only if a semantically-wrong variant *fails* it. TDAD (Rehan 2026) introduces the mechanism: visible/hidden test splits prevent the author from writing instructions to match the evals (the core gaming vector). PrimG shows that instruction-level mutations (removing or inverting one instruction) are the most efficient mutation type for detecting under-constrained specs. Reference implementation: https://github.com/f-labs-io/tdad-paper-code

---

## Current Implementation

`scoreD9` lives in `scorer/d9_eval_validation.go`. Its current signature is:

```go
func scoreD9(evalsDir string) (int, []Diagnostic)
```

It is called from a single production site in `scorer/scorer.go` at line 53, inside the `ScoreFromContent` function's dimension registry closure:

```go
{AllDimensions[8], func(_, _ string, _ *validatorBridge) (int, []Diagnostic) {
    return scoreD9(evalsDir)
}},
```

`evalsDir` is derived from `skillPath` two lines above the registry: `evalsDir := filepath.Join(filepath.Dir(skillPath), "evals")`. `skillPath` is the path to `SKILL.md` and is available in the `ScoreFromContent` scope.

The current scorer awards points as follows:

| Check | Points |
|---|---|
| `evals/` directory exists | 4 pts (`d9EvalsDirPoints`) |
| `instructions.json` present and non-empty | 3 pts |
| `summary.json` present with valid coverage ≥ 80% | up to 6 pts (3 for presence, 3 more for ≥ 80%) |
| Valid scenario dirs (≥ 3 complete `scenario-N/`) | 4 pts (high) or 2 pts (mid) |
| Score hard-capped at `d9Max` = 20 | — |

`TestD9_FullScore` demonstrates the current maximum reachable score of **17/20** with: `evals/` present (4) + `instructions.json` (3) + `summary.json` at 85% (6) + 3 valid scenarios (4) = 17. No path currently reaches 20/20.

There is no dedicated instruction parser in the scorer package. `MUST`/`NEVER`/`ALWAYS` patterns are currently detected only via `countPattern` (a simple case-insensitive substring counter in `scorer/util.go`) in `d3_anti_pattern.go` and `d6_freedom_calibration.go`. There is no function that parses `SKILL.md` and returns a list of imperative instruction lines. `scoreMutationCoverage` must implement its own line-level extractor — there is no existing parser to reuse.

---

## Problem Statement

D9 currently scores: presence of `evals/` directory, `instructions.json` coverage, scenario count, criteria quality (binary yes/no, 10+ items), and absence of instruction leakage in `task.md`. These checks validate that evals *exist* and are *well-formed*.

They do not validate that evals are *effective* — i.e., that they would catch a broken skill. The core gaming vector: an author writes instructions *after* writing evals and makes the instructions match the eval criteria. This passes all current D9 checks while producing a circular, non-validating eval suite. Alami 2026 documents exactly this pattern in the wild.

---

## Proposed Change

### Three new requirements, in priority order

#### 1. Mutation Score (highest priority) — 5 pts

At least one eval scenario must fail when a single instruction in SKILL.md is removed or inverted. This is assessable at scoring time via a mechanical check:

The scorer generates a minimal mutation set (remove one MUST/NEVER/ALWAYS instruction; invert one conditional) and checks whether any `criteria.json` item's wording directly references that instruction's constraint. If no criteria item would be affected by the mutation, the eval suite cannot detect that regression.

**Heuristic proxy (static analysis, no LLM required):**  
For each MUST/NEVER statement in SKILL.md, at least one criteria item in any scenario must reference the same constraint domain (detected via keyword overlap between the instruction and the criteria item). If an instruction has zero coverage in any criteria item, it scores 0 on mutation coverage for that instruction.

- **5 pts:** ≥ 80% of hard constraints have at least one criteria item covering them.
- **3 pts:** ≥ 50% coverage.
- **1 pt:** ≥ 1 constraint covered.
- **0 pts:** No overlap detected between instructions and criteria.

#### 2. Adversarial Scenario (medium priority) — 3 pts bonus

At least one scenario should test a boundary condition or failure case, not only the happy path. Detectable by checking whether any `task.md` contains: error conditions, edge cases, invalid inputs, or conflicting requirements.

- **3 pts:** At least one scenario explicitly tests a failure mode or boundary.
- **1 pt:** Scenarios cover varied inputs but no explicit failure mode.
- **0 pts:** All scenarios test the happy path only.

Adversarial scenario points are **diagnostic-only** and are **not added to the D9 score**. They are surfaced as a `hint`-severity diagnostic (e.g. `"adversarial bonus: +3 pts — not applied to score"`) so they appear in the evaluation report without inflating the dimension total.

#### 3. Independent Authoring Signal (low priority) — 2 pts bonus

Evals authored or last modified *after* the SKILL.md file, or by a different git author, are a stronger validation signal. Detectable via git log metadata.

- **2 pts:** `evals/` directory has a later commit timestamp than `SKILL.md`, or a different git author.
- **1 pt:** Same timestamp window (within 1 hour) — likely co-authored.
- **0 pts:** Evals and SKILL.md have identical timestamps (written simultaneously — highest gaming risk).

Independent authoring points are **diagnostic-only** and are **not added to the D9 score**. They are surfaced as a `hint`-severity diagnostic in the same way as the adversarial bonus, making the signal visible in the report without affecting the numeric dimension total.

**Scoring rebalance (20 pts total, unchanged):**

| Component | Current | New |
|---|---|---|
| evals/ directory + instructions.json | 4 pts | 3 pts |
| Coverage statistics (≥ 80%) | 6 pts | 5 pts |
| Valid scenarios (≥ 3, complete structure) | 4 pts | 2 pts |
| Criteria quality (10+ items, binary) | 3 pts | 3 pts |
| No instruction leakage in task.md | 3 pts | 2 pts |
| **Mutation score** (new) | — | 5 pts |

The adversarial scenario and independent authoring bonuses are purely informational diagnostic signals and do not contribute to the 20 pt total. The score is always hard-capped at `d9Max` = 20.

---

## Signature Change and Call-Site Update

`scoreMutationCoverage` needs the path to `SKILL.md` (outside `evals/`) to extract imperative instructions. The `SKILL.md` path cannot be derived from `evalsDir` alone, so `scoreD9` must accept it as a second parameter.

**New signature:**

```go
func scoreD9(evalsDir, skillPath string) (int, []Diagnostic)
```

**Where `skillPath` is located from each call site:**

| File | Line | Change required |
|---|---|---|
| `scorer/scorer.go` | ~53 | Pass `skillPath` (already in scope in `ScoreFromContent`): `return scoreD9(evalsDir, skillPath)` |
| `scorer/d9_eval_validation_test.go` | all `scoreD9(...)` calls (lines 11, 33, 44, 59, 74, 91, 110, 132, 148) | Construct a temp `SKILL.md` path. For tests that do not exercise mutation coverage, pass `filepath.Join(t.TempDir(), "SKILL.md")` (file need not exist — `scoreMutationCoverage` must handle a missing SKILL.md by returning 0 pts with no error). |

No other call sites exist outside the scorer package. There are no cmd-layer callers; the cmd layer calls `scorer.Score` or `scorer.ScoreFromContent`, which internally dispatch to `scoreD9`.

`scoreMutationCoverage(skillPath, evalsDir string) (int, []Diagnostic)` derives SKILL.md path from the `skillPath` argument directly. It reads the file line by line, extracting lines that match `(?i)\b(MUST|NEVER|ALWAYS)\b` as imperative constraints. This is a new line-level extractor — no existing parser is available in the scorer package to reuse (`countPattern` in `scorer/util.go` counts occurrences but does not return individual lines).

---

## Implementation Steps

1. **`scorer/d9_eval_validation.go`** — update `scoreD9` signature to `scoreD9(evalsDir, skillPath string)`.
   - Update the single internal call at the top of `scoreD9` that dispatches to `scoreMutationCoverage`.
2. **`scorer/scorer.go`** — update the D9 registry closure to pass `skillPath`:
   ```go
   {AllDimensions[8], func(_, _ string, _ *validatorBridge) (int, []Diagnostic) {
       return scoreD9(evalsDir, skillPath)
   }},
   ```
3. **`scorer/d9_eval_validation.go`** — add `scoreMutationCoverage(skillPath, evalsDir string) (int, []Diagnostic)`:
   - Read `skillPath`; if the file does not exist, return 0 pts, no error.
   - Extract imperative lines matching `(?i)\b(MUST|NEVER|ALWAYS)\b` by iterating line by line (no existing instruction parser to reuse — implement inline).
   - For each instruction line, compute token-level keyword overlap against every `criteria.json` item description across all `scenario-N/` dirs.
   - Score based on coverage percentage (see above).
4. **`scorer/d9_eval_validation.go`** — add `scoreAdversarialScenario(evalsDir string) (int, []Diagnostic)`:
   - Scan each `task.md` for boundary/failure markers: `fail`, `error`, `invalid`, `edge case`, `conflict`, `unexpected`, `when X is missing`.
   - Return 0 pts and emit a `hint`-severity diagnostic reporting the bonus amount (not applied to score).
5. **`scorer/d9_eval_validation.go`** — add `scoreIndependentAuthoring(evalsDir, skillPath string) (int, []Diagnostic)`:
   - Use `git log --follow --format="%ae %at"` on `evals/` and `SKILL.md` to compare author email and timestamps.
   - Graceful fallback for **non-git repos**: if `git log` exits non-zero or the working directory has no `.git` ancestor, return 0 pts, no error (skip, don't penalise).
   - Graceful fallback for **CI shallow clones** (`--depth=1`) and **detached HEADs**: detect these by checking the `git log` output line count (shallow clone yields exactly one commit per path; detached HEAD has no branch ref). In either case, fall back to file modification timestamps (`os.Stat(...).ModTime()`) as the comparison signal. If mtimes are within 1 hour treat as co-authored (1 pt); if both files have zero mtime (stat error) return 0 pts, no error.
   - Return 0 pts and emit a `hint`-severity diagnostic reporting the bonus amount (not applied to score).
6. **`scorer/d9_eval_validation_test.go`** — update all `scoreD9(evalsDir)` calls to `scoreD9(evalsDir, skillPath)` (9 call sites). See "Signature Change and Call-Site Update" above.
7. Update `cmd/assets/references/framework-dimensions.md` — document all three additions, cite Alami 2026, Rehan 2026, Wang et al., and Pan et al.
8. Update `testdata/` fixtures and add `scoreMutationCoverage` test cases (see "Test Cases" below).
9. Run `go test ./scorer/...`.
10. Run `tessl eval run cmd/assets/` to verify the tile eval suite still passes.

---

## Test Cases for `scoreMutationCoverage`

All new tests follow the style of `TestD9_FullScore` with explicit score assertions. Each test constructs a temp `SKILL.md` alongside a temp `evals/` tree.

```go
func TestScoreMutationCoverage_FullCoverage(t *testing.T) {
    // SKILL.md has two MUST constraints; both are referenced in criteria descriptions.
    // Expected: ≥ 80% coverage → 5 pts.
    skillPath := writeTempSKILL(t, "MUST validate input before processing.\nMUST log all errors.")
    evalsDir := t.TempDir()
    dir := filepath.Join(evalsDir, "scenario-1")
    writeTestFile(t, filepath.Join(dir, "criteria.json"),
        `{"checklist":[{"description":"validates input before processing","max_score":50},{"description":"logs all errors","max_score":50}]}`)
    score, diags := scoreMutationCoverage(skillPath, evalsDir)
    if score != 5 {
        t.Errorf("want 5, got %d (diags: %v)", score, diags)
    }
}

func TestScoreMutationCoverage_PartialCoverage(t *testing.T) {
    // SKILL.md has two MUST constraints; only one is referenced in criteria.
    // Expected: 50% coverage → 3 pts.
    skillPath := writeTempSKILL(t, "MUST validate input.\nMUST purge stale tokens.")
    evalsDir := t.TempDir()
    dir := filepath.Join(evalsDir, "scenario-1")
    writeTestFile(t, filepath.Join(dir, "criteria.json"),
        `{"checklist":[{"description":"validates input","max_score":100}]}`)
    score, diags := scoreMutationCoverage(skillPath, evalsDir)
    if score != 3 {
        t.Errorf("want 3, got %d (diags: %v)", score, diags)
    }
}

func TestScoreMutationCoverage_ZeroCoverage(t *testing.T) {
    // SKILL.md has MUST constraints; no criteria item matches any of them.
    // Expected: 0 pts.
    skillPath := writeTempSKILL(t, "MUST validate input.\nNEVER expose secrets.")
    evalsDir := t.TempDir()
    dir := filepath.Join(evalsDir, "scenario-1")
    writeTestFile(t, filepath.Join(dir, "criteria.json"),
        `{"checklist":[{"description":"output is formatted correctly","max_score":100}]}`)
    score, diags := scoreMutationCoverage(skillPath, evalsDir)
    if score != 0 {
        t.Errorf("want 0, got %d (diags: %v)", score, diags)
    }
}

func TestScoreMutationCoverage_MissingSKILLMd(t *testing.T) {
    // SKILL.md does not exist — should return 0 pts, no error diagnostic.
    skillPath := filepath.Join(t.TempDir(), "SKILL.md")
    evalsDir := t.TempDir()
    score, diags := scoreMutationCoverage(skillPath, evalsDir)
    if score != 0 {
        t.Errorf("want 0, got %d", score)
    }
    for _, d := range diags {
        if d.severity == "error" {
            t.Errorf("unexpected error diag: %v", d)
        }
    }
}

func TestScoreMutationCoverage_NoInstructions(t *testing.T) {
    // SKILL.md exists but has no MUST/NEVER/ALWAYS lines.
    // Expected: 0 pts (no constraints to cover).
    skillPath := writeTempSKILL(t, "Use this skill to do things.")
    evalsDir := t.TempDir()
    score, _ := scoreMutationCoverage(skillPath, evalsDir)
    if score != 0 {
        t.Errorf("want 0, got %d", score)
    }
}
```

`writeTempSKILL` is a test helper (analogous to `writeTestFile`) that writes a minimal `SKILL.md` with the given body to a temp directory and returns the full path.

---

## Acceptance Criteria

- A skill with well-formed evals but zero mutation coverage scores ≤ 15/20.
- A skill with ≥ 80% mutation coverage scores ≥ 18/20.
- The mutation coverage scorer does not require LLM inference — it is purely static keyword overlap analysis.
- The independent authoring check fails gracefully outside git repositories (score treated as 0, not an error).
- The independent authoring check falls back to `os.Stat` mtime comparison in CI shallow-clone and detached-HEAD environments.
- The adversarial scenario and independent authoring bonuses are surfaced as `hint`-severity diagnostics only and do not contribute to the 20 pt D9 total. The score is always hard-capped at `d9Max` = 20.
- All existing fixtures that currently pass D9 at ≥ 16/20 do not drop below 14/20 after the rebalance.
- `go test ./scorer/...` passes with all new test cases including the five `TestScoreMutationCoverage_*` cases.
