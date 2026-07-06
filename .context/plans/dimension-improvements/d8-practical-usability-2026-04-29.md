---
title: "Improvement Plan: D8 — Practical Usability"
type: plan
status: done
date: 2026-04-29
value: medium
---
# Improvement Plan: D8 — Practical Usability

**Date:** 2026-04-29  
**Current max score:** 15 pts  
**Priority:** Medium  
**Effort:** Small

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| TheAgentCompany: Benchmarking LLM Agents on Consequential Real-World Tasks | FF Xu, Y Song, B Li, Y Tang, K Jain, M Bao et al. | https://arxiv.org/abs/2412.14161 |
| Evaluation and Benchmarking of LLM Agents: A Survey | M Mohammadi, Y Li, J Lo, W Yip | https://dl.acm.org/doi/abs/10.1145/3711896.3736570 |
| Evaluating LLM Metrics Through Real-World Capabilities | JK Miller, W Tang | https://arxiv.org/abs/2505.08253 |
| Survey on Evaluation of LLM-Based Agents | A Yehudai, L Eden, A Li, G Uziel, Y Zhao et al. | https://arxiv.org/abs/2503.16416 |

**Key finding:** TheAgentCompany (Xu et al. 2024) is the field's benchmark of record for real-world agent task evaluation — tasks span software engineering, data analysis, and administrative work, scored purely on *task completion* rather than output quality proxies. The agent benchmarking survey (Mohammadi et al.) and Miller & Tang's metrics analysis both find that common proxy metrics (response coherence, example realism, surface readability) correlate poorly with actual task completion rates. The Yehudai et al. survey identifies that usability benchmarks fail when they measure output form rather than outcome achievability: an agent can produce a beautiful response that doesn't actually complete the task.

---

## Current Implementation

`scoreD8` in `scorer/d8_practical_usability.go` is a **flat heuristic** with no component architecture:

```
score = 5 (baseline)
score += scoreD8CodeBlocks(content, b)   // 0–4 pts based on block count; +2 if any language tags present
score += 4 if hasRunCommand(content)     // detects ./  npm run  yarn  pnpm run  bun run  make  python  go run
score = min(score, 15)
```

`scoreD8CodeBlocks` delegates to `b.Content.CodeBlockCount` when the library bridge is available, and falls back to `codeBlockCount(content)` (regex count of fenced blocks) when it is nil. There are **no named components** (concrete examples, specificity, non-trivial use cases, copy-paste readiness) — those labels appear only in documentation; the scorer does not implement them as discrete scored buckets.

**Implication for this plan:** the scoring rebalance table in the Proposed Change section describes a *desired target architecture*, not the current code. The plan's implementation path must explicitly resolve this gap before `scoreOutcomeLinkage()` can be introduced with accurate point semantics.

---

## Problem Statement

D8 currently scores: presence of code blocks, language tag diversity, and run-command presence. These are proxy metrics — they measure whether examples *look* useful (structured, runnable), not whether they are *traceable to a verifiable outcome*.

The proxy metric problem: a skill can contain perfectly formatted, domain-specific examples that nevertheless fail to tell the agent (or reviewer) whether execution succeeded. Without a verifiable outcome, the example is a demonstration, not a test.

---

## Proposed Change

### Architecture decision: additive within flat model (Option B)

Because `scoreD8` is a flat heuristic and there is no existing component architecture to "reduce by 1 pt each", introducing `scoreOutcomeLinkage()` as a weight-neutral rebalance against phantom components would be misleading. This plan chooses **Option B: specify `scoreOutcomeLinkage()` as purely additive within the current flat architecture**, capped so the total cannot exceed 15.

The rebalance table below describes the *logical intent* of the current point budget — it is **not** a description of existing code. A follow-on refactor (tracked separately) may later decompose the flat heuristic into discrete component functions that align code structure with the rubric.

### Add "Outcome Linkage" criterion — each example specifies a verifiable artifact (up to 3 pts additive)

An example is outcome-linked when it specifies what the agent (or reviewer) should observe to confirm the example succeeded. This does not require a full test suite — it requires at least one falsifiable signal per example:
- A file path that should exist
- A command whose output should match a pattern
- A test that should pass
- A diff that should be reviewable
- A state that should be observable in a tool (PR opened, job succeeded, etc.)

**Scoring (additive, capped at 15 total):**

- **3 pts:** Every segmented example includes an explicit outcome indicator.
- **2 pts:** Most examples (≥ 50%) include outcome indicators.
- **1 pt:** At least one example includes an outcome indicator.
- **0 pts:** No examples include outcome indicators (demonstrations only).

**Logical point budget (15 pts total, unchanged — for rubric documentation only):**

| Component | Current heuristic proxy | Logical intent |
|---|---|---|
| Concrete examples present | code block count bonus (0–4 pts) | 3 pts |
| Copy-paste readiness | language tag bonus (+2 pts) | 3 pts |
| Run-command present | run-command bonus (+4 pts) | 3 pts |
| Example specificity / non-trivial use | (no current signal) | 3 pts |
| **Outcome linkage** (new) | — | 3 pts |

---

## Implementation Steps

1. **Add `scoreOutcomeLinkage(content string) int`** to `scorer/d8_practical_usability.go`.

   **Example segmentation:** split `content` into segments on each fenced code block boundary. A segment is defined as: the closing fence of a code block plus the immediately following prose paragraph (up to the next blank line or heading). Outcome indicators are only valid when they appear inside a code block or in the paragraph immediately following it — this prevents prose-section false positives.

   Each segment is checked for:
   - File path signals: `# output:`, `# result:`, `→`, `produces`, `creates`, `writes to`
   - Verification signals: `# verify:`, `# check:`, `should return`, `expected:`, `assert`
   - Observable state: `you should see`, `the PR will`, `the job will`, `confirms that`

   Count `linked` (segments with at least one indicator) and `total` (total segments). Return:
   - 3 if `total > 0 && linked == total`
   - 2 if `total > 0 && linked*2 >= total`
   - 1 if `linked > 0`
   - 0 otherwise

2. **Wire `scoreOutcomeLinkage` into `scoreD8`** as a purely additive term:
   ```go
   score += scoreOutcomeLinkage(content)
   ```
   The existing `min(score, d8Max)` cap already ensures the total cannot exceed 15.

3. **Add fixture sketches to `testdata/`** — at minimum one positive and two negative cases:

   **Positive fixture (outcome-linked example):**
   ```markdown
   Run the migration:
   ```bash
   go run ./cmd/migrate up
   ```
   # verify: migration table has 3 rows — `SELECT count(*) FROM schema_migrations` should return 3.
   ```
   This segment scores 1/1 linked — the `# verify:` comment appears immediately after the code block.

   **Negative fixture 1 — generic prose with outcome-like words (false-positive trap):**
   ```markdown
   This command creates a new project. The tool produces a scaffold that writes to disk.
   Run `init` to get started.
   ```
   No fenced code block is present. Segments: 0. `scoreOutcomeLinkage` returns 0. The words `creates`, `produces`, and `writes to` appear only in standalone prose, not adjacent to a code block, so they must NOT trigger a signal.

   **Negative fixture 2 — outcome language inside a prose section far from code (false-positive trap):**
   ```markdown
   ## Background
   The deploy pipeline confirms that all checks pass before merging.
   You should see green indicators in the dashboard.

   ```bash
   git push origin main
   ```
   ```
   The outcome-like phrases (`confirms that`, `you should see`) appear in the prose section above the code block, not in the paragraph immediately following it. `scoreOutcomeLinkage` must return 0 for this segment.

4. **Update tests** in `scorer/d8_practical_usability_test.go`:
   - `TestD8_OutcomeLinkage_AllLinked` — fixture with every code block followed by a `# verify:` comment → `scoreOutcomeLinkage` returns 3.
   - `TestD8_OutcomeLinkage_NoneLinked` — fixture with code blocks and no outcome indicators → returns 0.
   - `TestD8_OutcomeLinkage_FalsePositive_ProsePhrases` — negative fixture 1 above → returns 0.
   - `TestD8_OutcomeLinkage_FalsePositive_DistantProse` — negative fixture 2 above → returns 0.
   - `TestD8_OutcomeLinkage_Partial` — fixture where half the code blocks have outcome indicators → returns 2.

5. **Update `cmd/assets/references/framework-dimensions.md`** — document outcome linkage, cite Xu et al. (arXiv:2412.14161) and Miller & Tang (arXiv:2505.08253). Note the additive-within-flat-model architecture decision.

6. **Run `go test ./scorer/...`** — all existing tests must continue to pass; new tests must pass.

---

## Acceptance Criteria

- A skill with concrete but outcome-free examples scores ≤ 12/15.
- A skill where every example has an outcome indicator scores 14–15/15.
- The scorer correctly ignores outcome-like language inside prose sections (only scans within or adjacent to code blocks — negative fixtures 1 and 2 above both score 0 on `scoreOutcomeLinkage`).
- All four new test cases pass.
- No existing `TestD8_*` tests regress.
