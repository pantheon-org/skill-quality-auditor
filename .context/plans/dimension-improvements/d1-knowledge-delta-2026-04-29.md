---
title: "Improvement Plan: D1 — Knowledge Delta"
type: plan
status: done
date: 2026-04-29
value: medium
---
# Improvement Plan: D1 — Knowledge Delta

**Date:** 2026-04-29  
**Current max score:** 20 pts  
**Priority:** Medium  
**Effort:** Small

---

## Current Implementation

`scorer/d1_knowledge_delta.go` uses three independent adjustment layers applied to a base score:

| Constant | Value | Role |
|---|---|---|
| `d1BaseScore` | 15 | Starting score before any adjustments |
| `d1PenaltyPerPat` | 2 | Deducted per beginner-pattern hit (binary per pattern, not per occurrence) |
| `d1BonusPerPat` | 1 | Added per expert-signal pattern hit (binary per pattern) |
| `d1Max` | 20 | Hard ceiling; score floored at 0 |

**Layer 1 — Beginner-content penalty.** Seven patterns are checked (`npm install`, `yarn add`, `pip install`, `getting started`, `introduction`, `basic syntax`, `hello world`). Each pattern that appears subtracts 2 from the score. Maximum total deduction: 7 × 2 = **−14** (floored at 0).

**Layer 2 — Expert-signal bonus.** Six patterns are checked (`anti-pattern`, `NEVER`, `ALWAYS`, `production`, `gotcha`, `pitfall`). Each pattern that appears adds 1. Maximum total addition: 6 × **+6**.

**Layer 3 — Instructions ratio bonus/penalty.** Reads `evals/instructions.json`. If `(new_knowledge + preference) / total ≥ 70%` → **+2**. If ratio < 30% → **−2**. Otherwise **0**.

**Realistic score range without the new sub-criterion:**
- Best case (no beginner patterns, all 6 expert signals, ratio ≥ 70%): 15 − 0 + 6 + 2 = 23, capped → **20**
- Worst case (7 beginner patterns, no expert signals, ratio < 30%): 15 − 14 + 0 − 2 = −1, floored → **0**
- Typical expert skill (0 beginner patterns, 2–3 expert signals, no evals): 15 + 2–3 = **17–18**

There is **no separate "5-pt avoids-obvious-content bucket"** in the implementation. The negative-signal mechanic is purely `d1PenaltyPerPat × patterns_hit`, independent of any fixed allocation.

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| Instruction Agent: Enhancing Agent with Expert Demonstration | Y Li, H Hultquist, J Wagle, K Koishida | https://arxiv.org/abs/2509.07098 |
| From Novice to Expert: LLM Agent Policy Optimization via Step-wise Reinforcement Learning | Z Deng, Z Dou, Y Zhu et al. | https://arxiv.org/abs/2411.03817 |
| Grounding Open-Domain Knowledge from LLMs to Real-World RL Tasks: A Survey | H Yin, H Qian, Y Shi et al. | https://www.ijcai.org/proceedings/2025/1198.pdf |

**Key finding:** "Instruction Agent" (Li et al.) demonstrates that injecting expert demonstrations — not just declarative facts — into agent instructions measurably improves task performance. "From Novice to Expert" (Deng et al.) establishes a step-wise expertise model: skills should reflect progressive expert capability, not flat knowledge dumps. The grounding survey (Yin et al.) shows that knowledge which cannot be grounded to a real-world observable outcome degrades agent performance.

---

## Problem Statement

Current D1 scoring assesses whether skill content *sounds expert* (avoids basic syntax, avoids copying docs, avoids generic advice). It does not assess whether the knowledge is *actionable* — i.e., whether an agent equipped with this skill would demonstrably perform differently on a task than one without it.

A skill can score highly on D1 today by containing sophisticated prose that provides no concrete procedural handle for the agent. This is the "tacit knowledge gap": the content is expert in register but declarative in form.

---

## Proposed Change

### Add a "Demonstration Concreteness" sub-criterion (3 pts)

**Detection heuristic (implemented in `scoreDemonstrationConcreteness()`):**

The function checks for three signals in order of specificity:

1. **Fenced code block or `→`-arrow notation** — presence of a triple-backtick fenced block (` ```...``` `) or at least one `→` arrow on a line that is not inside a blockquote. This is the primary structural indicator of a worked example.
2. **Output section marker** — presence of a line or heading matching `(?i)output:`, `(?i)result:`, `(?i)expected:`, or `(?i)returns:` anywhere in the document (case-insensitive). This indicates an outcome is specified.
3. **Neither signal present** — the skill is entirely declarative prose.

**Scoring:**
- **3 pts:** Both signal 1 (code fence or `→` notation) and signal 2 (output/result/expected section) are present.
- **2 pts:** Signal 1 is present (concrete example structure exists) but signal 2 is absent (no verifiable outcome specified).
- **1 pt:** Neither signal is present but the content contains at least one of the Layer 2 expert-signal patterns (`anti-pattern`, `NEVER`, `ALWAYS`, `production`, `gotcha`, `pitfall`) — expert in register but no worked example.
- **0 pts:** No code fence, no `→` notation, no output section, and no expert-signal patterns.

### Scoring rebalance

The new sub-criterion adds up to 3 pts on top of the existing mechanics. To preserve the `d1Max = 20` ceiling without discarding the existing penalty/bonus layers, the base score (`d1BaseScore`) is reduced from 15 to 12. This keeps the arithmetic consistent with actual constants:

| Component | Current | Proposed |
|---|---|---|
| `d1BaseScore` | 15 | 12 |
| Beginner-pattern penalty | −2 per pattern (up to 7) | unchanged |
| Expert-signal bonus | +1 per pattern (up to 6) | unchanged |
| Instructions ratio adjustment | −2 / 0 / +2 | unchanged |
| Demonstration concreteness (new) | — | 0–3 pts |
| `d1Max` (hard ceiling) | 20 | 20 |

**Worked arithmetic examples after change:**

- Typical expert skill, 0 beginner patterns, 3 expert signals, no evals, concreteness = 2 pts: 12 + 3 + 2 = **17**
- Best-practice skill with code fence + output section, all 6 expert signals, ratio ≥ 70%: 12 + 6 + 2 + 3 = 23, capped → **20**
- Declarative prose skill, 0 expert signals, no evals, no concreteness: 12 + 0 + 0 + 0 = **12**
- Same declarative skill today (current scorer): 15 + 0 + 0 = **15** — this is the intentional regression that penalises content-free declarations.

---

## Implementation Steps

1. Update `scorer/d1_knowledge_delta.go`:
   - Change `d1BaseScore` constant from `15` to `12`.
   - Add `scoreDemonstrationConcreteness(content string) int` function implementing the three-signal heuristic above (fenced code block regex: `` /```[\s\S]+?```/ ``; `→` arrow: `/→/`; output section: `/(?i)^(output|result|expected|returns)\s*:/m`).
   - Call it in `scoreD1()` and add the returned value to `score` before the floor/ceiling clamp.
2. Update `cmd/assets/references/framework-dimensions.md` — add the new sub-criterion definition, detection rules, and worked examples under the D1 section.
3. Add test cases in `scorer/d1_knowledge_delta_test.go` (see Test Plan below).
4. Run `go test ./scorer/...` to verify all cases pass.

---

## Test Plan

Add the following explicit test cases in `scorer/d1_knowledge_delta_test.go`:

### TestD1_ConcretenessZero — 0-pt fixture

```
content := "---\ndescription: x\n---\n# Skill\nAlways validate inputs before processing."
```

No code fence, no `→`, no output section, no expert-signal patterns. Expected concreteness contribution: **0 pts**. Full score: `d1BaseScore(12) + 0 = 12`.

### TestD1_ConcretenessOne — 1-pt fixture (expert register, no example)

```
content := "---\ndescription: x\n---\n# Skill\nNEVER call this API without a timeout. This is a production gotcha."
```

No code fence, no `→`, no output section, but `NEVER`, `production`, and `gotcha` are present (Layer 2 = +3). Concreteness = 1 pt (expert signals present but no structural example). Score: `12 + 3 + 1 = 16`.

### TestD1_ConcretenessTwo — 2-pt fixture (code fence, no output section)

```
content := "---\ndescription: x\n---\n# Skill\n" +
    "Use the following pattern:\n" +
    "```go\nclient.SetTimeout(5 * time.Second)\n```"
```

Fenced block present, no output/result section. Concreteness = 2 pts. No beginner patterns, no expert signals, no evals. Score: `12 + 2 = 14`.

### TestD1_ConcretenessThree — full 3-pt fixture

```
content := "---\ndescription: x\n---\n# Skill\n" +
    "Run the migration:\n" +
    "```bash\ndb migrate --env production\n```\n" +
    "Expected output:\n" +
    "Migrations applied: 3"
```

Fenced block present, `Expected` section present. Concreteness = 3 pts. Score: `12 + 3 = 15`.

### Regression: TestD1_Penalties (existing — update expected value)

The existing test expects `11` (base 15 − 2 − 2). After lowering `d1BaseScore` to 12 the new expected value is `12 − 2 − 2 = 8`. Update the assertion accordingly.

### Regression: TestD1_Rewards (existing — verify cap still holds)

`12 + 6(expert signals) + 2(instructions) + 3(concreteness) = 23`, capped → 20. Assertion unchanged.

---

## Acceptance Criteria

- A skill with only declarative expert prose (no example, no output section) scores ≤ 15/20 on D1.
- A skill with a code fence and an explicit output/result section scores at least 2 pts higher than the same skill without them.
- The two new `TestD1_ConcretenessZero` and `TestD1_ConcretenessThree` tests pass.
- All existing `go test ./scorer/...` tests pass after updating the regression expected values noted above.
- Existing A-grade skills (those previously scoring 18+) are not retroactively pushed below grade B (14/20) by the base-score change alone.
