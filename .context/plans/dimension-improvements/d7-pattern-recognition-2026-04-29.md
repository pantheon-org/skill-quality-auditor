---
title: "Improvement Plan: D7 — Pattern Recognition"
type: plan
status: active
date: 2026-04-29
---
# Improvement Plan: D7 — Pattern Recognition

**Date:** 2026-04-29  
**Current max score:** 10 pts  
**Priority:** Medium  
**Effort:** Medium

---

## Current Implementation

`scoreD7` in `scorer/d7_pattern_recognition.go` is a **pure description-length heuristic**. It calls `b.descriptionLen()` (which parses the char count from the `CheckFrontmatter` library result) and maps it to a fixed score band via a single `switch`:

| Description length | Score |
|---|---|
| > 200 chars | 10 |
| > 120 chars | 9 |
| > 60 chars | 8 |
| ≤ 60 chars | 6 (+ warning if ≤ 30 chars) |
| Bridge unavailable | 6 + warning |

There is **no trigger analysis**, **no keyword scanning**, **no workflow anchor detection**, and **no negative-anchor detection** in the current scorer. The "Problem Statement" in the prior version of this plan described a scorer that does not exist. Any plan that adjusts weights on non-existent components (e.g., "reduce keyword count weight from 4 to 3") is invalid and must be replaced.

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| AgentRouter: A Knowledge-Graph-Guided LLM Router for Collaborative Multi-Agent QA | Z Zhang, K Shi, Z Yuan, Z Wang, T Ma et al. | https://arxiv.org/abs/2510.05445 |
| Efficient and Interpretable Multi-Agent LLM Routing via Ant Colony Optimization | X Wang, C Zhang, J Zhang, C Li, Q Sun et al. | https://arxiv.org/abs/2603.12933 |
| AgentPoison: Red-Teaming LLM Agents via Poisoning Memory or Knowledge Bases | Z Chen, Z Xiang, C Xiao, D Song et al. | https://proceedings.neurips.cc/paper_files/paper/2024/hash/eb113910e9c3f6242541c1652e30dfd6-Abstract-Conference.html |

> **Citation notice:** arXiv:2510.05445, arXiv:2603.12933, and the NeurIPS 2024 AgentPoison link must be independently verified before being cited in `cmd/assets/references/framework-dimensions.md`. Do not update the framework-dimensions doc until verification is complete (Step 4 below is gated on this).

**Key finding:** AgentRouter (Zhang et al.) shows that knowledge-graph-guided routing outperforms keyword matching for skill selection — semantic context matters, not just surface trigger terms. Triggers that fire on surface vocabulary without semantic anchoring cause misfires on adjacent topics. The ACO routing paper (Wang et al.) establishes that routing quality is determined by *discriminativeness* (low false positives) as much as *recall* (low false negatives). AgentPoison (Chen et al., NeurIPS 2024) demonstrates that triggers are an active attack surface: adversarially crafted inputs can hijack skill routing via poisoned memory or knowledge bases. Together these papers establish that trigger quality has two distinct failure modes not currently scored: *over-triggering* (fires on wrong contexts) and *trigger hijacking* (fires on adversarially crafted inputs).

---

## Problem Statement

The current `scoreD7` awards up to 10 pts purely on description character length. While a richer description correlates with better pattern recognition, length alone does not measure whether triggers are *discriminative* — i.e., whether they avoid firing on plausibly-related but incorrect contexts.

A skill with a long description can still over-fire if its triggers appear naturally in adjacent domains. For example, a "database migration" skill that triggers on "migrate" will also fire on "migrate data between S3 buckets" — a different task. The current scorer gives no signal on this.

---

## Proposed Change

### Approach: additive discriminativeness signal on top of the existing length bands

Rather than rewriting the entire scorer (which would risk regression on existing calibration), this plan **preserves the existing length bands as the base score (6–10 pts)** and adds a `scoreDiscriminativeness()` helper that provides an **additive bonus signal** influencing diagnostic output and a future scoring revision. For this iteration, the discriminativeness check emits diagnostics only (warnings / infos) without changing the numeric score. This lets the team validate the signal quality before committing to a point-value change.

If the signal proves reliable after evaluation on at least 10 real skills, a follow-on plan will convert it to a 0–2 pt modifier (e.g., description length caps at 8, discriminativeness can add up to 2 pts to reach 10).

**Discriminativeness criteria checked:**
1. **Negative anchor** — description or a dedicated trigger section includes an explicit statement of a wrong context that should NOT activate the skill (e.g., `Does not apply`, `SKIP when`, `Not for`, `Exclude`, `DO NOT trigger`, `not intended for`).
2. **Workflow anchor** — description contains a trigger tied to a specific artifact or action rather than a generic topic (trigger phrase contains artifact nouns: `file`, `PR`, `commit`, `test`, `config`, `pipeline`, `migration`; or action-context phrases).

---

## Implementation Steps

1. **Verify citations** — before touching docs, open each of the three paper links and confirm title, authors, and year match the table above. Record the verification result. Do not proceed to Step 4 until done.

2. **Update `scorer/d7_pattern_recognition.go`** — add a `scoreDiscriminativeness(desc string) []Diagnostic` helper (unexported) that:
   - Accepts the raw description string (obtained from `b.rawDescription()` or equivalent — add that accessor to `validatorBridge` if it does not exist).
   - Scans for negative anchor markers (case-insensitive): `does not apply`, `skip when`, `not for`, `exclude`, `do not trigger`, `not intended for`.
   - Scans for workflow anchor tokens: `file`, `pr`, `commit`, `test`, `config`, `pipeline`, `migration` as whole words.
   - Returns an `INFO`-level diagnostic if both anchors are present (positive signal), a `WARN`-level diagnostic if neither is present, and no diagnostic if only one is present.
   - Call `scoreDiscriminativeness` from `scoreD7` and append its diagnostics to `diags` before returning.
   - Do **not** change the returned numeric score in this step.

3. **Update `scorer/d7_pattern_recognition_test.go`** — add test cases:
   - `TestD7_DiscriminativenessWarning`: description > 200 chars but no anchors → score 10, WARN diagnostic present.
   - `TestD7_DiscriminativenessInfo`: description > 200 chars with both anchors → score 10, INFO diagnostic present.
   - `TestD7_DiscriminativenessNeutral`: description > 200 chars with one anchor only → score 10, no discriminativeness diagnostic.
   - **Regression test** `TestD7_ExistingFixturesNoRegression`: run `scoreD7` against `testdata/fixtures/skill-full/SKILL.md` and `testdata/fixtures/skill-minimal/SKILL.md`; assert scores are **not lower** than the values returned by the unmodified scorer (capture baseline before editing, hard-code as expected minimums).

4. **Update `cmd/assets/references/framework-dimensions.md`** — document the discriminativeness criterion under the D7 section. **Gate: only do this after Step 1 citation verification passes.** Cite only papers whose links were confirmed. If a link is broken or title mismatches, omit that citation and note it as unverified.

5. **Update `testdata/` fixtures** — add two new SKILL.md files:
   - `testdata/fixtures/skill-d7-with-anchors/SKILL.md` — description > 200 chars including at least one negative anchor (`Does not apply to`) and one workflow anchor (`commit`).
   - `testdata/fixtures/skill-d7-no-anchors/SKILL.md` — description > 200 chars with no anchor language.
   These fixtures are used exclusively by the new tests in Step 3.

6. **Run `go test ./scorer/...`** — all tests must pass, including the regression test from Step 3.

---

## Acceptance Criteria

- `scoreD7` on `testdata/fixtures/skill-full/SKILL.md` returns a score **equal to or higher** than the pre-change baseline (regression guard).
- `scoreD7` on `testdata/fixtures/skill-minimal/SKILL.md` returns a score **equal to or higher** than the pre-change baseline (regression guard).
- A description > 200 chars with no anchor language produces a `WARN`-level D7 diagnostic mentioning discriminativeness.
- A description > 200 chars with both a negative anchor and a workflow anchor produces an `INFO`-level D7 diagnostic.
- The discriminativeness scanner does **not** produce false positives on generic negation language (e.g., "not recommended" within a content sentence rather than a trigger scope statement).
- `go test ./scorer/...` exits 0.
