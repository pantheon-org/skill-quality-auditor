---
title: "Finding: research-proof/tools is a No fit (skill-specific eval harness + marker validator, kernel already covered by D4/validate and the eval runner)"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: LOW
themes:
  - EVAL
  - SKILL-QUALITY
---

# Finding: research-proof/tools fit assessment — 2026-07-07

Date: 2026-07-07
Status: DECISION-SUPPORT, not actioned

> The [`tools/`](https://github.com/tonyblu331/research-proof/tree/main/tools) directory of [tonyblu331/research-proof](https://github.com/tonyblu331/research-proof) was linked with "will it fit?". The parent project is a Claude Code skill that pressure-tests research claims (freeze the verifier, adversarial checks, proof ledger). The `tools/` are the JavaScript eval harness *for that skill itself*: a structural marker validator, a backtest rater, and a contract-based answer grader — all hardwired to research-proof's own vocabulary. Verdict: **No fit**. Same shape as the `degradation.py` case: project-specific gate, hardcoded literals, wrong language, generic kernels already covered.

## What was investigated

The `tools/` file tree plus the three files that embody the claim: `validate-research-skill.mjs`, `rate-research-skill.mjs`, and `lib/scorer.mjs` (with the `lib/*` helpers `answer-contracts`, `artifact-evidence`, `semantic-terms`, `text-matchers`). Licence MIT, JavaScript (.mjs), ~26★.

## What it actually is

Not a generic skill-quality toolkit. It is the eval/validation harness for the one research-proof skill:

- **`validate-research-skill.mjs`** — a structural conformance gate that asserts the research-proof `SKILL.md` and `evals/evals.json` contain a long list of **hardcoded literal marker strings** (required sections like `Core Discipline`, required pressure keys like `verifier_boundary`/`proof_ladder`, required eval-assertion IDs like `certainty-inflation`, `evaluator-hacking`). This is the same mechanism as `consensus-rnd`'s `degradation.py`: marker-presence linting bound to one skill's own decisions.
- **`rate-research-skill.mjs`** — reads `*.summary.json` and `benchmark.json` from a backtest workspace and rates the research-proof skill's own run outputs (scorer version 3).
- **`lib/scorer.mjs` + `lib/*`** — a domain-specific *answer grader*: it checks whether an agent answer contains research-proof's required structural moves (verifier boundary, ledger decision, refusal-near, conditional gate, verdict status, method concepts) via semantic-term and text-matcher heuristics.
- **`create-research-eval-pack.mjs` / `export-skill-creator-evals.mjs` / `run-research-backtest.mjs`** — eval-pack scaffolding and backtest plumbing for the skill.

Input: the research-proof skill dir and its backtest workspace. Output: pass/fail validation, ratings, and contract-graded answers. For: the research-proof maintainer's own CI and TDD loop.

## Mapping against this project

| Existing capability | Overlap |
| --- | --- |
| D1-D9 scorers (`scorer/`) | full for the validator's kernel — D4 already pattern-matches structure/markers generically. |
| `validate` / `analyze` (`cmd/`) | full — `validate` already checks artifact conventions; the marker gate duplicates it less generally. |
| duplication engine (`duplication/`) | none. |
| native eval runner (`cmd/eval.go`, D9) | partial — the rater/answer-grader is skill-specific eval machinery; we own a generic LLM-judge runner with scenario assertions. |
| helper skills (`.context/plugins/`) | none. |

Language and architecture mismatch (JavaScript vs the Go embedded-assets CLI) is a further cost, but it only matters if the idea were worth porting — and here it mostly is not.

## Verdict

**No fit.** The tools are a bespoke, hardwired eval harness for one external skill. The validator's generic kernel is already covered by D4 + `validate`; the rater and answer-grader are a project-specific instance of eval machinery we already own generically. Nothing here transfers as-is.

## The salvageable idea

Thin, and mostly already approximated. The one mildly interesting technique is **contract-based answer assertions**: grading an answer by asserting it contains required *structural moves* (a decision, a verifier boundary, a stated rejection) rather than matching gold text. Our eval scenarios already support `checks`/assertions that cover most of this, and D-scorers already do structural pattern-matching, so the residue is small. If ever wanted, it would be a new assertion type in our eval-scenario schema (built natively in Go), not a port of these `.mjs` files or their research-proof-specific term lists. Do not port the hardcoded marker strings or the answer-contract vocabulary.

The genuinely interesting part of the *parent project* is its methodology (freeze the verifier, adversarial counterexamples, proof ledger) — but that lives in the SKILL, not in `tools/`, and is out of scope for this question.

## Recommendation

Record the rejection so `tools/` is not re-assessed. No import, no port. The marker-validator kernel is already D4 + `validate`; the eval harness is already our native runner. If contract-based structural assertions are ever desired, add them as a native eval-scenario assertion type — do not resurrect these tools.

## Fit assessment (structured record)

<!-- fit-assessment -->
```yaml
schema_version: 1
source:
  name: tonyblu331/research-proof (tools/)
  url: https://github.com/tonyblu331/research-proof/tree/main/tools
  license: MIT
  language: JavaScript
characterisation: >-
  The JavaScript eval/validation harness for the research-proof skill itself: a
  structural marker validator (asserts SKILL.md/evals.json contain hardcoded
  literal strings), a backtest rater over the skill's own run outputs, and a
  contract-based answer grader checking for research-proof's required structural
  moves. Input: the research-proof skill dir + backtest workspace. Output:
  validation, ratings, graded answers. For: that skill's maintainer.
overlap:
  d1_d9_scorers:
    level: full
    note: The validator's marker-presence kernel is already covered generically by D4.
  validate_analyze:
    level: full
    note: validate already checks artifact conventions; the marker gate duplicates it less generally.
  duplication:
    level: none
    note: No similarity detection.
  eval_runner:
    level: partial
    note: The rater/answer-grader is skill-specific eval machinery; we own a generic LLM-judge runner.
  helper_skills:
    level: none
    note: No agent-workflow equivalent.
verdict: No fit
vehicle_if_adopted: go-cli
salvageable:
  present: true
  description: >-
    Contract-based answer assertions (grade an answer by required structural
    moves, not gold text) as a possible new native eval-scenario assertion type.
    Mostly already approximated by our scenario checks and D-scorers; do not port
    the .mjs files or the research-proof-specific marker/term vocabulary.
recommendation:
  action: record-and-hold
  detail: >-
    No import or port. Marker-validator kernel is already D4 + validate; the eval
    harness is already our native runner. Add contract assertions natively only
    if a need appears.
value: LOW
```
