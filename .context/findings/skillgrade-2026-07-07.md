---
title: "Finding: mgechev/skillgrade is a Partial fit (behavioural eval harness, not a static scorer)"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: LOW
themes:
  - EVAL
  - SKILL-QUALITY
---

# Finding: mgechev/skillgrade fit assessment — 2026-07-07

Date: 2026-07-07
Status: DECISION-SUPPORT, not actioned

> [mgechev/skillgrade](https://github.com/mgechev/skillgrade) was linked with "does this fit?". It is a **behavioural (execution) eval harness** that runs a skill through a real agent and grades the outputs. Our project is a **static, rule-based content scorer** (D1-D9 over `SKILL.md`, no agent in the scoring path). Verdict: **Partial fit** — wrong paradigm to import, one minor transferable grader shape.

## What was investigated

README, file tree, `eval.yaml` reference, grader contract, and the `src/` layout (agents, commands, core). Licence MIT, TypeScript, ~549★.

## What it actually is

"Unit tests for agent skills." A CLI that scaffolds an `eval.yaml`, runs the skill through a real agent (Gemini/Claude/Codex/OpenCode) inside Docker, executes author-defined tasks, and grades outputs with two grader types: `deterministic` (a script emitting `{score, details, checks}` JSON) and `llm_rubric` (weighted). Presets set trial counts (`--smoke` 5 / `--reliable` 15 / `--regression` 30); `--ci --threshold` gates a pass rate; `--validate` verifies graders against reference solutions.

- **Input:** a skill dir + `eval.yaml`. **Output:** per-trial pass rate vs threshold. **For:** skill authors doing behavioural regression testing.

## Mapping against this project

| Existing capability | Overlap |
| --- | --- |
| D1-D9 scorers (`scorer/`) | None — it measures runtime behaviour; D1-D9 read the text. |
| `validate` / `analyze` (`cmd/`) | None — different axis. |
| duplication engine (`duplication/`) | None. |
| native eval runner (`cmd/eval.go`, D9) | Partial — ours runs scenario tasks with an optional LLM judge against our own tile assets for CI; it does not run arbitrary skills through real agent CLIs. |

Vehicle if anything were adopted: the native eval runner (D9), not the static D1-D9 scorers. The axis mismatch is the same one recorded for SkillEval (`skilleval-analysis-2026-06-30.md`).

## Verdict

**Partial fit.** A mature behavioural harness in the wrong language and paradigm for our Go static scorer. The harness itself is not something to import; one grader shape is worth remembering.

## The salvageable idea (built natively)

The **weighted `deterministic` + `llm_rubric` grader schema** — `{score, details, checks}` with per-grader weights — is a clean, proven shape. If our eval scenarios ever grow richer per-scenario scoring, this structure is worth borrowing (redesigned in Go), not the code. NEVER port its literals (runner names, preset counts, env-var conventions).

The stronger baseline-lift idea shows up more clearly in its siblings — see the `skilljack-evals` and `agent-skills-eval` findings of the same date.

## Recommendation

Record and hold. No import. Remember the weighted grader schema for any future enrichment of eval-scenario scoring; there is no current gap forcing it.

## Fit assessment (structured record)

<!-- fit-assessment -->
```yaml
schema_version: 1
source:
  name: mgechev/skillgrade
  url: https://github.com/mgechev/skillgrade
  license: MIT
  language: TypeScript
characterisation: >-
  A CLI that runs a skill through a real agent (Gemini/Claude/Codex/OpenCode)
  inside Docker, executes author-defined tasks, and grades outputs with
  deterministic and llm_rubric graders, reporting a per-trial pass rate against
  a threshold. Input: a skill dir + eval.yaml. Output: pass rate. For: skill
  authors doing behavioural regression testing.
overlap:
  d1_d9_scorers:
    level: none
    note: Measures runtime behaviour; D1-D9 read the text.
  validate_analyze:
    level: none
    note: Different axis.
  duplication:
    level: none
    note: No similarity detection.
  eval_runner:
    level: partial
    note: Ours judges our own tile assets; it does not run arbitrary skills via real agent CLIs.
  helper_skills:
    level: none
    note: No agent-workflow equivalent.
verdict: Partial fit
vehicle_if_adopted: go-cli
salvageable:
  present: true
  description: >-
    The weighted deterministic + llm_rubric grader schema ({score, details,
    checks} with per-grader weights). Worth borrowing (redesigned in Go) if
    eval-scenario scoring is ever enriched. Do not port its literals.
recommendation:
  action: record-and-hold
  detail: >-
    No import. Remember the grader schema for future eval-scenario scoring
    enrichment; no current gap forces it.
value: LOW
```
