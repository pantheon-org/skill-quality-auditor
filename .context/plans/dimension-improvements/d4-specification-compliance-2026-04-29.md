---
title: "Improvement Plan: D4 — Specification Compliance"
type: plan
status: done
date: 2026-04-29
---
# Improvement Plan: D4 — Specification Compliance

**Date:** 2026-04-29  
**Current max score:** 15 pts  
**Priority:** High  
**Effort:** Medium

---

## Academic Basis

This plan is grounded in the following papers discovered via Google Scholar:

| Paper | Authors | Link |
|---|---|---|
| Test-Driven AI Agent Definition (TDAD): Compiling Tool-Using Agents from Behavioral Specifications | T Rehan | https://arxiv.org/abs/2603.08806 |
| Agentic AI for Behaviour-Driven Development Testing Using Large Language Models | C Paduraru, M Zavelca, A Stefanescu | https://www.researchgate.net/publication/390835646 |
| LLM-Skill Orchestration: Achieving 202/202 Subtask Completion via Rule-Augmented Multi-Model Collaboration | R Tao | https://www.researchsquare.com/article/rs-9323974/latest |
| Automated Structural Testing of LLM-Based Agents | J Kohl, O Kruse, Y Mostafa, A Luckow et al. | https://ieeexplore.ieee.org/abstract/document/11401679/ |

**Key finding:** TDAD (Rehan 2026) treats agent specifications as compiled artifacts and introduces three mechanisms relevant to D4: (1) *visible/hidden test splits* that prevent spec-to-eval gaming, (2) *semantic mutation testing* — generating plausibly-wrong prompt variants and verifying the test suite detects them, and (3) *spec evolution scenarios* measuring regression safety. TDAD achieves 97% regression safety on SpecSuite-Core. The BDD paper (Paduraru et al.) reinforces that natural-language specs like SKILL.md are only reliably validated via executable Given/When/Then scenarios, not structural inspection alone. Reference implementation: https://github.com/f-labs-io/tdad-paper-code

---

## Current Implementation

The scorer lives in `scorer/d4_specification.go` (tested by `scorer/d4_specification_test.go`). It operates as follows:

- **Base score** starts at 8.
- **`scoreD4Description`** — awards up to +3 pts for description length (>mid threshold: +2, >high threshold: +1) and subtracts up to −2 pts for overloaded descriptions with excessive `and`/`or` conjunctions.
- **`scoreD4HarnessRefs`** — awards +1 pt each for absence of a harness-specific path and absence of an agent-specific reference (e.g., Claude/Cursor/Windsurf/Copilot/Gemini/Goose mentions); emits `WARN` diagnostics when violations are found.
- **`scoreD4RelPaths`** — awards +1 pt if the skill uses relative asset paths (`scripts/`, `references/`, `assets/`).
- **`scoreD4ContentViolations`** — deducts pts for `../` outside code blocks (−2), absolute skill paths (−1), and `.context/`/`.agents/` references (−1); checks both SKILL.md prose and files under `scripts/` and `references/`.
- **`scoreD4Bonus`** — awards +1 pt if a `scripts/` directory contains a `.py`/`.ts`/`.js` file, and +1 pt if the last `##` section is `References` with at least one bullet link.
- Hard cap: max 15 pts (`d4Max`), with bonus taking it to `d4MaxWithBonus`.

**Gap:** The scorer has no behavioural constraint detection. A specification can score near-maximum while being entirely vague about what the agent must or must not do.

---

## Problem Statement

D4 currently scores structural compliance: does the skill have a description field, are triggers correctly formatted, does the metadata conform to tile.json schema. These are necessary but not sufficient. A specification that is structurally perfect can still be *behaviourally ambiguous* — two reasonable agents reading it could behave differently on the same input.

Structural compliance checks can be gamed: an author can write a specification that passes all format checks while leaving the actual behavioural constraints vague or contradictory. TDAD calls this the core risk: "small prompt changes cause silent regressions, tool misuse goes undetected, and policy violations emerge only after deployment."

---

## Proposed Change

### Add "Mutation Resistance" as a D4 behavioural criterion (4 pts)

A specification is compliant only if it is *tight enough* that a plausibly-wrong variant would behave differently. This is assessable at scoring time by checking whether the specification contains:

1. **At least one hard constraint** — a `MUST` or `NEVER` statement that is specific and testable (not "follow best practices" but "NEVER commit to main without a PR"). Specificity is required: the keyword must be followed by a verb phrase of at least 4 words, excluding generic filler phrases such as "follow best practices" or "be careful".
2. **At least one conditional branch** — an `if`/`when` clause that changes the required behaviour based on an input condition. Specificity is required: the clause must contain a subject noun phrase of at least 2 words adjacent to the keyword (e.g., "if the user provides" not a bare "if").
3. **At least one exclusion** — an explicit statement of what the skill does NOT do (scoping out-of-scope behaviour).

These three properties together constitute minimum mutation resistance: removing any one of them would produce a detectably different agent.

**Keyword detection specificity thresholds** — to prevent bare prose matches:

| Marker type | Keywords | Minimum context requirement |
|---|---|---|
| Hard constraint | `MUST`, `NEVER`, `always`, `ONLY` | Followed by ≥ 4-word verb phrase; not in a code block |
| Conditional | `if`, `when`, `unless`, `only if` | Adjacent noun phrase ≥ 2 words before or after the keyword; not in a code block |
| Exclusion | `does not`, `out of scope`, `SKIP`, `DO NOT` | Appears outside code blocks in a complete sentence (≥ 6 words) |

**Scoring adjustment (15 pts total, unchanged):**

| Component | Current | New |
|---|---|---|
| Description field quality | 5 pts | 4 pts |
| Trigger format + correctness | 4 pts | 3 pts |
| Metadata schema compliance | 3 pts | 2 pts |
| Allowed-tools correctness | 3 pts | 2 pts |
| **Mutation resistance** (new) | — | 4 pts |

**Mutation resistance breakdown:**
- Hard constraint present and specific: 1.5 pts
- Conditional branch present and specific: 1.5 pts
- Explicit exclusion / scope boundary: 1 pt

---

## Implementation Steps

1. Confirm the correct filename: the scorer is `scorer/d4_specification.go` and its test file is `scorer/d4_specification_test.go`. All references in this plan use these names.
2. Create `scorer/d4_specification_test.go` if it does not already exist (it currently does exist; verify with `ls scorer/d4_specification_test.go` before proceeding). The test file must be updated — not replaced — to add the new test cases in step 6.
3. Update `scorer/d4_specification.go` — add `scoreMutationResistance()` that applies the specificity thresholds above when scanning for hard constraint markers, conditional markers, and exclusion markers. Code blocks must be stripped before matching (reuse `removeCodeBlocks`).
4. Reweight existing sub-scorers to free 4 pts (see table above).
5. Update `cmd/assets/references/framework-dimensions.md` — document mutation resistance, cite Rehan 2026 and the TDAD benchmark.
6. Update `scorer/d4_specification_test.go` — add the acceptance-criteria test cases listed below.
7. Update `testdata/` fixtures — add a specification-compliant-but-vague fixture and a mutation-resistant fixture.
8. Run `go test ./scorer/...`.

---

## Acceptance Criteria

- A structurally valid SKILL.md with no hard constraints, no conditionals, and no exclusions scores ≤ 11/15.
- A SKILL.md with exactly one of the three mutation-resistance criteria present (e.g., hard constraint only, no conditional, no exclusion) scores between 9/15 and 12/15 (partial credit: 1.5 pts awarded).
- A SKILL.md with exactly two of the three criteria present (e.g., hard constraint + conditional, no exclusion) scores between 10/15 and 13/15 (partial credit: 3 pts awarded).
- A SKILL.md with all three criteria present and specific scores ≥ 13/15.
- A `MUST` or `if` that appears only inside a code block does not trigger the hard constraint or conditional scorer.
- A bare `MUST follow best practices` (non-specific, < 4-word verb phrase) does not award the hard-constraint point.
- The description field weight reduction does not change scores for skills that already have high-quality descriptions by more than 1 pt.
