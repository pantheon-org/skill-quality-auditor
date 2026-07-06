---
title: "Improvement Plan: D2 — Mindset and Procedures"
type: PLAN
status: DONE
date: 2026-04-29
value: MEDIUM
---
# Improvement Plan: D2 — Mindset + Procedures

**Date:** 2026-04-29  
**Current max score:** 15 pts  
**Priority:** Medium  
**Effort:** Small

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| Knowledge Activation: AI Skills as the Institutional Knowledge Primitive for Agentic Software Development | G Bakal | https://arxiv.org/abs/2603.14805 |
| Procedural Knowledge Ontology (PKO) | VA Carriero, M Scrocca et al. | https://link.springer.com/chapter/10.1007/978-3-031-94578-6_19 |
| Real-Time Procedural Learning From Experience for AI Agents | D Bi, Y Hu, MN Nasir | https://arxiv.org/abs/2511.22074 |
| Automating Skill Acquisition through Large-Scale Mining of Agentic Repositories | S Bi, M Wu, H Hao et al. | https://arxiv.org/abs/2603.11808 |

**Key finding:** The Procedural Knowledge Ontology (Carriero et al.) establishes a canonical structure for procedural knowledge: *preconditions* (when to start), *steps* (what to do), *decision points* (when to branch), and *postconditions* (how to verify completion). "Real-Time Procedural Learning" (Bi et al.) shows that agents without external checkpoint anchors in their procedures undergo recursive drift — self-feedback loops cause degradation rather than improvement. Bakal 2026 argues that skills are the institutional primitive: they must encode not just *what* to do but *when to stop and verify*.

---

## Current Implementation

`scoreD2` in `scorer/d2_mindset_procedures.go` awards up to 15 pts across four components:

| Component | Logic | Max pts |
|---|---|---|
| Mindset / philosophy heading | `reD2MindsetHeader` matches `## Mindset`, `## Philosophy`, or `## Principles` | 2 |
| Structural / procedural density | `scoreD2Structure()` — see detail below | 6 |
| When-to-use guidance | `countPattern` for `"when to use"` or `"when to apply"` | 4 |
| When-not-to-use guidance | `countPattern` for `"when not to"` | 3 |

**`scoreD2Structure` detail (up to 6 pts):**

When `validatorBridge.Content` is non-nil (library path):
- All three of `StrongMarkers`, `WeakMarkers`, `ImperativeRatio` are zero → return 0 + a warning diagnostic.
- Otherwise, `ImperativeRatio` drives a tiered score:
  - `≥ 0.40` → 4 pts
  - `≥ 0.25` → 3 pts
  - `≥ 0.10` → 2 pts
  - `< 0.10` → 0 pts
- `ListItemCount > 3` → +2 pts; `ListItemCount 1–3` → +1 pt; `0` → +0 pts.

When `validatorBridge.Content` is nil (fallback regex path):
- `reD2NumberedList` (`^\s*[0-9]+\.`) matches any numbered list → 2 pts, else 0.

The current implementation has no signal for preconditions, postconditions, or decision points. The `scoreD2Structure` component can award up to 6 pts (4 from `ImperativeRatio` + 2 from `ListItemCount`), but the plan table above lists it as capped at the values actually achievable in practice given the 15-pt ceiling.

---

## Problem Statement

Current D2 scoring checks for: a mindset/philosophy section, step-by-step workflows, and when/when-not guidance. It does not distinguish between procedures that are *structurally complete* (with decision points and postconditions) and those that are *linear lists* (steps with no branching or verification logic).

Linear procedure lists produce agents that drift when they encounter an unexpected state, because the procedure provides no signal for when to pause, branch, or verify with an external source.

---

## Proposed Change

### Align D2 scoring against the PKO structural model

**New scoring breakdown (15 pts total, unchanged):**

| Component | Current | New | Change |
|---|---|---|---|
| Mindset / philosophy framing | 4 pts | 3 pts | Slight reduction |
| Step-by-step workflow presence | 4 pts | 3 pts | Slight reduction |
| When / when-not guidance | 3 pts | 3 pts | Unchanged |
| **Preconditions** (explicit entry conditions) | — | 2 pts | New |
| **Postconditions / external checkpoints** | — | 2 pts | New |
| Decision points (branching logic) | — | 2 pts | New |

**Scoring guidance for new components:**

- **Preconditions (2 pts):** Does the procedure state what must be true before the agent begins? E.g., "Only invoke this when X is present" or "Requires Y to have already run." Partial credit (1 pt) if implied but not explicit.
- **Postconditions / external checkpoints (2 pts):** Does the procedure specify at least one step that requires external validation — a test result, a linting pass, a file artifact, a human confirmation — rather than agent self-assessment? Partial credit (1 pt) if a verification step exists but is self-assessed only.
- **Decision points (2 pts):** Does the procedure include at least one branch or conditional that tells the agent what to do when a step fails or produces an unexpected result? Partial credit (1 pt) if error handling is mentioned but not specified.

---

## Signal Vocabulary

Each new function uses regex patterns to detect signals. The tables below are the authoritative keyword lists.

### `scorePreconditions()` — signals for explicit entry conditions (2 pts)

| Signal type | Example patterns (case-insensitive) |
|---|---|
| Explicit guard clause | `only (invoke\|use\|run\|apply) (this\|when)`, `requires .* to (have\|be)` |
| Prerequisite statement | `prerequisite[s]?`, `before (starting\|running\|invoking\|applying)` |
| Entry condition header | `##\s*(prerequisites?\|preconditions?\|requirements?\|before you (start\|begin))` |
| Dependency declaration | `depends on`, `must (exist\|be present\|be configured\|have run)` |
| Conditional activation | `(do not\|don't) (use\|invoke\|run) (this\|if\|unless\|when)` |

**Scoring rule:** 2 pts if any explicit-guard, entry-condition-header, or dependency-declaration pattern matches. 1 pt (partial credit) if only a prerequisite-statement or conditional-activation pattern matches with no dedicated header.

### `scorePostconditions()` — signals for external checkpoints (2 pts)

| Signal type | Example patterns (case-insensitive) |
|---|---|
| Test/lint verification | `(run\|execute) .*(test[s]?\|spec[s]?\|lint\|check\|verify)`, `go test`, `npm test`, `pytest` |
| Artifact confirmation | `(confirm\|verify\|check) .*(file\|artifact\|output\|result) (exists?\|is present\|was created)` |
| Human confirmation gate | `(wait for\|requires?\|ask for) .*(approval\|confirmation\|sign.?off\|review)` |
| CI/CD gate | `(must pass\|pipeline\|ci\|cd).*(pass\|succeed\|green)` |
| External state assertion | `(assert\|ensure\|validate) .*(external\|remote\|database\|api\|service)` |

**Scoring rule:** 2 pts if a test/lint-verification, CI/CD-gate, or human-confirmation-gate pattern matches. 1 pt (partial credit) if only an artifact-confirmation or external-state-assertion pattern matches without a clear external agent (tool, human, or system) performing the check.

### `scoreDecisionPoints()` — signals for branching / error-handling logic (2 pts)

| Signal type | Example patterns (case-insensitive) |
|---|---|
| Explicit branch | `if .*(fail[s]?\|error[s]?\|unexpected\|not found\|missing)`, `otherwise`, `else` |
| Fallback instruction | `fall(back\| back) to`, `revert to`, `retry` |
| Conditional header | `##\s*(troubleshooting\|error handling\|if .* fails?\|when .* fails?)` |
| Stop condition | `(stop\|abort\|halt\|do not continue) (if\|when\|unless)` |
| Escalation signal | `(escalate\|raise\|report) .*(to\|with) .*(human\|user\|team\|engineer)` |

**Scoring rule:** 2 pts if an explicit-branch or conditional-header pattern matches. 1 pt (partial credit) if only a fallback-instruction, stop-condition, or escalation-signal matches without a named branch target.

---

## `scoreD2Structure` Refactor

The current `scoreD2Structure` awards up to 6 pts (4 from `ImperativeRatio` tiers + 2 from `ListItemCount`). Under the new model, the component budget for structural / procedural density is reduced to **3 pts** to make room for the three new 2-pt components.

**Revised `scoreD2Structure` (max 3 pts):**

| Condition | Points |
|---|---|
| `ImperativeRatio ≥ 0.40` | 3 |
| `ImperativeRatio ≥ 0.25` | 2 |
| `ImperativeRatio ≥ 0.10` OR fallback numbered list matches | 1 |
| All markers zero / below 0.10 and no numbered list | 0 (+ warning diagnostic) |

`ListItemCount` is **dropped** from `scoreD2Structure`; list structure is now captured indirectly by `scorePreconditions`, `scorePostconditions`, and `scoreDecisionPoints` through their own regex signals. This removes the double-counting that existed when a well-formatted numbered list could score both via `ListItemCount` and via the `when not to` pattern.

The all-zero-markers warning diagnostic is retained unchanged.

---

## Implementation Steps

1. Update `scorer/d2_mindset_procedures.go`:
   - Revise `scoreD2Structure` to drop `ListItemCount` branching and cap `ImperativeRatio` tiers at 3 pts (see table above).
   - Add `scorePreconditions(content string) (int, []Diagnostic)` using the signal vocabulary table above.
   - Add `scorePostconditions(content string) (int, []Diagnostic)` using the signal vocabulary table above.
   - Add `scoreDecisionPoints(content string) (int, []Diagnostic)` using the signal vocabulary table above.
   - Wire all three into `scoreD2`, replacing the lost pts from the reduced `scoreD2Structure` and `Mindset/philosophy` budgets.
2. Update `cmd/assets/references/framework-dimensions.md` — document the PKO-aligned structure, cite Carriero et al. and Bi et al. in the rationale.
3. Update `testdata/` fixtures — ensure at least one fixture has all three new components and at least one is missing them.
4. Run `go test ./scorer/...`.

---

## Test Cases

The following table specifies the required test coverage. Each row is a distinct `TestD2_*` function.

### `scoreD2Structure` (revised)

| Test name | Input (`ImperativeRatio`, `ListItemCount`) | Expected score | Notes |
|---|---|---|---|
| `TestD2Structure_AllZeroMarkers` | `(0, 0, 0)` strong/weak/imperative | 0, 1 diag | **Existing** — retained |
| `TestD2Structure_HighRatio` | `ImperativeRatio=0.45` | 3 | Full credit |
| `TestD2Structure_MidRatio` | `ImperativeRatio=0.30` | 2 | **Partial credit** |
| `TestD2Structure_LowRatio` | `ImperativeRatio=0.12` | 1 | **Partial credit** |
| `TestD2Structure_FallbackNumberedList` | nil bridge, content with `1. ... 2. ...` | 1 | Fallback path, partial credit |

### `scorePreconditions()`

| Test name | Content signal | Expected score | Notes |
|---|---|---|---|
| `TestD2Preconditions_ExplicitHeader` | `## Prerequisites\n- X must exist` | 2 | Full credit: dedicated header |
| `TestD2Preconditions_ImpliedOnly` | `Requires the config to have been initialised` | 1 | **Partial credit**: prerequisite statement, no header |
| `TestD2Preconditions_None` | No qualifying signals | 0 | Zero case |

### `scorePostconditions()`

| Test name | Content signal | Expected score | Notes |
|---|---|---|---|
| `TestD2Postconditions_TestGate` | `Run go test ./... before proceeding` | 2 | Full credit: test-verification pattern |
| `TestD2Postconditions_ArtifactOnly` | `Confirm the output file exists` | 1 | **Partial credit**: artifact confirmation, no external agent |
| `TestD2Postconditions_None` | No qualifying signals | 0 | Zero case |

### `scoreDecisionPoints()`

| Test name | Content signal | Expected score | Notes |
|---|---|---|---|
| `TestD2DecisionPoints_ExplicitBranch` | `If the command fails, revert to the previous state` | 2 | Full credit: explicit branch |
| `TestD2DecisionPoints_FallbackOnly` | `Retry up to three times` | 1 | **Partial credit**: fallback instruction, no named branch target |
| `TestD2DecisionPoints_None` | No qualifying signals | 0 | Zero case |

---

## Acceptance Criteria

- A procedure that is a flat numbered list with no branching or external checkpoints scores ≤ 9/15.
- A procedure with explicit preconditions, an external checkpoint, and a decision point scores ≥ 13/15.
- Existing well-structured skills (with if/else workflow language) do not lose more than 1 pt.
