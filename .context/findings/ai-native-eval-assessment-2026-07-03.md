---
title: "Finding: ai-native-eval Assessment"
type: FINDING
status: ACTIVE
date: 2026-07-03
value: LOW
themes:
  - EVAL
related:
  - ../findings/evoskill-integration-2026-07-02.md
  - ../findings/arxiv-self-evolution-survey-2026-07-03.md
  - ../plans/evoskill-core-loop-port-2026-07-02.md
---

# Finding: ai-native-eval Assessment

Evaluated https://github.com/aiaccelerationism/ai-native-eval — an evidence-driven
evaluation and repair system for AI-native repository maturity. TypeScript + Codex
agent skills. Two projects have complementary strengths.

---

## Architecture

Recursive tree of evaluator plugins (each a SKILL.md with plugin manifest JSON fence).
Leaf evaluators output checklist-style deductions against a rubric. Parent nodes use
weighted average aggregation. A deterministic TypeScript tool validates, aggregates,
and renders reports — zero AI calls in the tool path.

**Scoring model:** Every leaf starts at 10/10. Deductions are applied from rubric groups
with per-group `budget` caps. Parent scores = `SUM(child.score * child.weight) / SUM(weights)`.
Confidence is separate from score (low/medium/high).

**Key novelty: "Why not 10/10"** — every deduction must cite evidence and carry a
recommendation. The tool strictly validates deduction IDs against the rubric JSON
in SKILL.md. Unknown deduction IDs are runtime errors.

---

## What to adopt (high priority)

### 1. Deduction rubrics with group budgets

Replace free-form diagnostic text with structured rubrics in each scorer's reference
doc: a JSON fence defining checklist items with `points`, `appliesWhen`,
`evidenceRequired`, and `recommendation`. The scorer must produce deductions matching
the rubric. Score = max_points - sum(capped deductions per group). This gives us:

- **Auditability**: every point lost traceable to a specific rubric item with evidence
- **Consistency**: rubric forces the scorer to evaluate the same checklist every time
- **Remediation linkage**: every deduction carries a recommendation, feeds `remediate`

### 2. Policy rules engine (ESLint-style)

Add per-evaluator gating rules with `off`/`warn`/`error` severity. A rule can mark
a result as `error` (blocked) when score falls below threshold, but numeric score
remains unchanged. Replaces hardcoded `--fail-below B` with configurable per-dimension
gates.

### 3. Run folder validation loop

Agent writes output JSON, tool validates against rubric schemas. We could adopt this
for `evaluate --store`: validate every dimension has a complete judgment with evidence
before accepting the audit.

---

## What to adopt (medium priority)

### 4. Separate confidence from score

We conflate evidence quality with score. An independent `confidence` field would
communicate "score is high but confidence is low" (limited evidence).

### 5. Append-only artifact bundles with manifest/snapshot

Our `.context/audits/<skill>/<date>/` structure already supports this. Adding
`manifest.json` and `snapshot.json` would enable incremental diffs between audits.

### 6. Context-aware evaluation routing

PRs, issues, periodic checks as evaluation targets alongside the existing single-skill
file evaluation.

---

## Contrast with our approach

| Dimension | ai-native-eval | skill-quality-auditor |
|-----------|---------------|----------------------|
| Scoring philosophy | Checklist deductions from 10/10 | Point accumulation toward max per dim |
| Plugin system | Tree of evaluator skills, dynamic | Flat registry of Go scorers |
| AI integration | Agent writes JSON, tool validates | Go functions score, LLM-judge advisory |
| Policy gating | ESLint-style rules, separate from score | Hardcoded `--fail-below` CLI flags |
| Deduction model | Group budgets cap double-penalizing | No budget concept |
| Config layering | Built-in/person/project/explicit | None (hardcoded defaults) |

---

## Recommended action

Adopt deduction rubrics with budgets (highest ROI — adds structural integrity to our
scoring without rewriting the Go architecture) and ESLint-style policy rules (replaces
brittle hardcoded gates with configurable per-dimension thresholds).
