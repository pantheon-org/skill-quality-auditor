---
title: "Finding: benchflow-ai/skillsbench is a No fit (model-ranking benchmark + dataset, not a per-skill scorer)"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: LOW
themes:
  - EVAL
  - SKILL-QUALITY
---

# Finding: benchflow-ai/skillsbench fit assessment — 2026-07-07

Date: 2026-07-07
Status: DECISION-SUPPORT, not actioned

> [benchflow-ai/skillsbench](https://github.com/benchflow-ai/skillsbench) was linked with "does this fit?". It is a **research benchmark and dataset** for ranking frontier models on skill-use tasks, run via an external SDK. Our project is a **static, rule-based scorer** of individual `SKILL.md` files. Verdict: **No fit** — different purpose, different consumer, nothing to port.

## What was investigated

README, quick-start, task-package structure, and the bundled authoring/review skills under `.agents/skills/` (`task-creator`, `task-review`). Licence Apache-2.0, PDDL/Python, ~1454★, with a published HuggingFace dataset.

## What it actually is

A benchmark suite and dataset ("the first benchmark for evaluating how well AI agents use skills"), run via the external **BenchFlow** CLI (`bench eval run`). Gym-style tasks — many deliberately requiring composition of 2+ skills, designed so SOTA models score <50% — aimed at ranking frontier models (GPT-5.5, Opus 4.8, Gemini 3.1, GLM 5.1, …). It ships its own `task-creator` and `task-review` skills with a "good task" rubric for contributors.

- **Input:** the benchmark's own curated task corpus. **Output:** a model leaderboard. **For:** benchmark maintainers and model evaluators — not authors of an individual `SKILL.md`.

## Mapping against this project

| Existing capability | Overlap |
| --- | --- |
| D1-D9 scorers (`scorer/`) | None. |
| `validate` / `analyze` (`cmd/`) | None. |
| duplication engine (`duplication/`) | None. |
| native eval runner (`cmd/eval.go`, D9) | None — ours evaluates one skill's quality; this ranks models across a fixed corpus. |

There is no vehicle to adopt it into: it is external infrastructure (SDK + dataset + leaderboard), not a technique that slots into a per-skill scorer.

## Verdict

**No fit.** A model-ranking benchmark behind an external SDK is a different tool for a different question. Importing or depending on it would not serve our static per-skill scoring, and it does not overlap any existing capability.

## The salvageable idea (built natively)

At most a thin, secondary observation, not an action item: its `task-review` "good task" rubric (and the skill-composition framing) could sanity-check our own **eval-scenario authoring guidance** if we ever revise it — worth a glance then, nothing to port now. The core benchmark and dataset are not transferable to this project. There is no primary transferable kernel.

## Recommendation

Record the rejection so the source is not re-assessed. No import, no dependency. Glance at its `task-review` rubric only if eval-scenario authoring guidance is next revised.

## Fit assessment (structured record)

<!-- fit-assessment -->
```yaml
schema_version: 1
source:
  name: benchflow-ai/skillsbench
  url: https://github.com/benchflow-ai/skillsbench
  license: Apache-2.0
  language: PDDL/Python
characterisation: >-
  A research benchmark and HuggingFace dataset, run via the external BenchFlow
  SDK/CLI, of gym-style tasks (many requiring 2+ skills composed) designed so
  SOTA models score under 50%, used to rank frontier models. Input: its own
  curated task corpus. Output: a model leaderboard. For: benchmark maintainers
  and model evaluators, not individual SKILL.md authors.
overlap:
  d1_d9_scorers:
    level: none
    note: Ranks models across a corpus; does not score one skill's text.
  validate_analyze:
    level: none
    note: Different axis.
  duplication:
    level: none
    note: No similarity detection.
  eval_runner:
    level: none
    note: Ours evaluates one skill's quality; this ranks models.
  helper_skills:
    level: none
    note: Ships its own authoring/review skills for its corpus, not for us.
verdict: No fit
vehicle_if_adopted: none
salvageable:
  present: false
  description: >-
    No transferable kernel to build natively. At most a tangential glance at its
    task-review "good task" rubric if eval-scenario authoring guidance is revised;
    the benchmark and dataset themselves are not transferable.
recommendation:
  action: reject
  detail: >-
    Record the rejection so the source is not re-assessed. No import, no
    dependency on the external SDK.
value: LOW
```
