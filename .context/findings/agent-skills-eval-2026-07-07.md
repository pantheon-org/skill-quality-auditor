---
title: "Finding: darkrishabh/agent-skills-eval is a Partial fit (cleanest generic implementation of the skill-lift baseline)"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: MEDIUM
themes:
  - EVAL
  - SKILL-QUALITY
---

# Finding: darkrishabh/agent-skills-eval fit assessment — 2026-07-07

Date: 2026-07-07
Status: DECISION-SUPPORT, not actioned

> [darkrishabh/agent-skills-eval](https://github.com/darkrishabh/agent-skills-eval) was linked with "does this fit?". It is a **behavioural (execution) test runner** for agentskills.io skills. Our project is a **static, rule-based content scorer** (D1-D9 over `SKILL.md`, no agent in the scoring path). Verdict: **Partial fit** — wrong paradigm to import, but it is the cleanest, most generic implementation of the one idea worth keeping: the with-skill / without-skill baseline.

## What was investigated

README (full, including the ASCII flow diagram), file tree, and `src/` layout (`run-eval.ts`, `grade.ts`, `evaluate-skills.ts`, `openai-compatible-provider.ts`, artifacts/reporter modules). Licence MIT, TypeScript, ~614★.

## What it actually is

A CLI/SDK that runs each eval twice against the same prompt — once `with_skill` (SKILL.md in context), once `without_skill` (baseline, skill stripped) — has a judge model grade both sides against the eval's `expected_output` and `assertions`, and emits a side-by-side HTML report plus JSON/JSONL artifacts in the agentskills.io `iteration-N` layout. OpenAI-compatible by default (OpenAI/Together/Groq/Anthropic-compat/local Llama). Its entire premise is: prove the skill makes a measurable difference, or see that it does not.

- **Input:** a folder of skills with `evals/evals.json`, a target model, a judge model. **Output:** per-skill pass/fail with lift, as portable artifacts + static HTML. **For:** skill authors proving skill value; CI via `--strict`.

## Mapping against this project

| Existing capability | Overlap |
| --- | --- |
| D1-D9 scorers (`scorer/`) | None — its signal is runtime lift; D1-D9 read the text. |
| `validate` / `analyze` (`cmd/`) | None (though it also does spec-compliant `SKILL.md` validation as a side gate). |
| duplication engine (`duplication/`) | None. |
| native eval runner (`cmd/eval.go`, D9) | Partial — ours judges scenarios but has no with/without-skill baseline. |

Vehicle if adopted: the native eval runner (D9). It runs an LLM, which our static scoring path deliberately does not.

## Verdict

**Partial fit.** Same behavioural-vs-static axis mismatch as its siblings, so not something to import wholesale. It matters because it isolates the single transferable kernel — the paired baseline — more cleanly than the others: a small, runtime-agnostic, judge-graded with/without comparison with no Docker or five-runner apparatus.

## The salvageable idea (built natively)

**Skill lift via a paired with-skill / without-skill baseline**, judged against per-eval assertions. This is the same kernel flagged in the `skilljack-evals` finding, and this repo is the reference design to study for it: minimal, provider-agnostic, artifact-first. If the project ever adds a behavioural dimension, build a `--baseline` pass into `cmd/eval.go` that reruns each scenario with the skill unmounted and reports the delta — behind a flag, outside the deterministic static grade.

Secondary shape worth noting: its **portable JSON/JSONL artifacts diffable across runs** align with how our `--store`/`trend` outputs already work, so a lift metric would slot into existing trend tracking.

NEVER port its literals (provider names, workspace layout strings, flag names). Extract the mechanism only.

## Recommendation

Record and hold. No import. This is the design to reference if and when skill-lift is built natively; there is no current gap forcing it. Because it duplicates the kernel already captured under `skilljack-evals`, treat this finding as the "how to build it cleanly" companion rather than an independent action item.

## Fit assessment (structured record)

<!-- fit-assessment -->
```yaml
schema_version: 1
source:
  name: darkrishabh/agent-skills-eval
  url: https://github.com/darkrishabh/agent-skills-eval
  license: MIT
  language: TypeScript
characterisation: >-
  A CLI/SDK that runs each eval twice against the same prompt (with_skill and
  without_skill), has a judge model grade both sides against the eval's
  assertions, and emits a side-by-side HTML report plus JSON/JSONL artifacts.
  OpenAI-compatible. Input: skills + evals.json, target and judge models.
  Output: per-skill pass/fail with lift. For: skill authors proving skill value
  and CI.
overlap:
  d1_d9_scorers:
    level: none
    note: Its signal is runtime lift; D1-D9 read the text.
  validate_analyze:
    level: none
    note: Does spec-compliant SKILL.md validation as a side gate, not our axis.
  duplication:
    level: none
    note: No similarity detection.
  eval_runner:
    level: partial
    note: Ours judges scenarios but has no with/without-skill baseline.
  helper_skills:
    level: none
    note: No agent-workflow equivalent.
verdict: Partial fit
vehicle_if_adopted: go-cli
salvageable:
  present: true
  description: >-
    The cleanest reference design for skill lift via a paired
    with-skill/without-skill baseline, judged against per-eval assertions.
    Minimal, provider-agnostic, artifact-first. If a --baseline pass is built
    into cmd/eval.go, study this repo for the shape. Do not port its literals.
recommendation:
  action: record-and-hold
  detail: >-
    No import. Reference design for skill lift; same kernel as skilljack-evals,
    so a companion rather than an independent action item.
value: MEDIUM
```
