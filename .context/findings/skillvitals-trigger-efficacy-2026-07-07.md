---
title: "Finding: skillvitals Trigger and Efficacy Testing"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: MEDIUM
themes:
  - EVAL
  - SKILL-QUALITY
related:
  - ./cultivar-differential-eval-2026-07-07.md
---

# Finding: skillvitals Trigger and Efficacy Testing

## What is skillvitals

[skillvitals](https://github.com/ContextJet-ai/skillvitals) is a small (~9 files, ~15 KB)
MIT-licensed Python CLI plus bundled GitHub Action from the `ContextJet-ai` org, created
2026-07-06. It measures the "vital signs" of a `SKILL.md` cheaply enough to run in CI on a
small or local model (cents, not a research budget). Its thesis: agent skills ship on faith,
and only two things decide whether a skill works, so it measures exactly those two axes and
nothing else.

- **Axis A, Triggering (cheap):** given `cases.yaml` of `{prompt, should_fire}`, does the
  skill activate? Two backends: a zero-cost deterministic heuristic (extract quoted phrases
  from the `description` frontmatter plus name words >3 chars, fire if the prompt contains
  any), and an LLM classifier fed **only name + description** (not the body), yes/no at
  `temperature=0`.
- **Axis B, Efficacy (needs a model):** given `tasks.yaml` of `{prompt}` only (no gold
  answer), run the task model twice, once with the skill body injected and once without, then
  LLM-judge each output 0 to 10 and report the delta.

The scoring core (`metrics.py`) is pure and dependency-injected; the model layer is any
OpenAI-compatible endpoint.

## Current state of this project

**skill-quality-auditor** scores skills against a 9-dimension framework using static analysis
of SKILL.md content, Jaccard duplication detection, and an LLM-as-judge native eval runner
(`cmd/eval.go`) that runs authored scenarios **with the skill loaded**.

The gap: every one of D1 to D9 inspects the document text. Nothing empirically measures
whether the skill *fires* on the right requests, or whether it *helps* relative to not being
present. skillvitals tests exactly those two runtime behaviours.

## Concrete mechanisms worth adopting

### Metrics (`src/skillvitals/metrics.py`)

- **Trigger:** `recall` (fires when it should), `false_fire_rate` (fires when it should not,
  a first-class metric), `precision`, `f1`, `n`.
- **Efficacy:** `win_rate` (with > without), `tie_rate`, `avg_delta` (mean of with minus
  without, the headline number), `avg_with`, `avg_without`.

### Schemas

- `cases.yaml`: `{cases: [{prompt, should_fire}]}`, negative cases mandatory.
- `tasks.yaml`: `{tasks: [{prompt}]}`, no gold answer; grading is relative plus an absolute
  0 to 10 judge, so no reference output is needed.

### Gating and distribution

- `--min-recall` exits non-zero below threshold; the bundled Action defaults to `0.8`,
  heuristic by default.
- `badge_endpoint()` emits a shields.io JSON badge (`trigger f1 X` / `help +Y`), colour
  thresholded green >=0.8 / yellow >=0.6 / red.
- All model calls at `temperature=0` for reproducibility.

## Novel ideas a naive evaluator would miss

1. **Triggering is tested from the description alone, not the body.** This mirrors how real
   agent harnesses route: at activation time the model sees the description, not the full
   skill. A naive evaluator reads the whole skill to judge relevance, which does not reflect
   the actual routing decision.
2. **False-fire rate is a headline metric.** Over-triggering (stealing activation from
   neighbouring skills) is weighted equally with under-triggering. Negative cases are
   mandatory.
3. **Counterfactual efficacy isolates marginal value.** A skill can score well in isolation
   yet deliver `avg_delta ~ 0` because the base model already handled the task. Only the
   without-skill arm surfaces the true contribution, and reference-free judging makes it
   cheap to author.
4. **A deterministic zero-cost baseline** doubles as a lint on whether the description even
   contains concrete, matchable trigger vocabulary.

## Transferable ideas, mapped to this project

### Idea 1, Trigger Reliability scorer or eval mode (highest value)

No current dimension measures whether a skill fires. Ship a `cases.yaml`-style set and
measure recall / false-fire / F1 from the **description alone**. Backend either a
deterministic heuristic (pure Go, no API) or the existing Claude eval runner as a yes/no
router prompted with name + description only.

- Candidate: `scorer/d10_trigger_reliability.go` or an `evaluate --trigger` mode.
- **Pair with the duplication engine:** draw negative cases from the trigger phrases of
  near-neighbour skills the Jaccard engine surfaces, to detect **behavioural trigger
  collisions** (two skills that both grab the same prompt). Duplication finds textual
  overlap; this finds behavioural overlap. Feed collisions into the `remediate` generator.

### Idea 2, Counterfactual efficacy (see also cultivar finding)

Run each eval scenario with and without the skill body and report `win_rate` / `avg_delta`.
Our D9 and native runner run only the with-skill arm, so they cannot distinguish "additive"
from "base model already handled it". Cheapest path: add the without arm and the paired-delta
metric to the existing judge harness. This idea recurs independently in the cultivar finding,
which is the strongest signal it is worth building.

### Idea 3, Reference-free relative judging

`tasks.yaml` carries no gold answers; a useful signal comes from with-vs-without plus an
absolute judge. Adaptable as a lightweight "smoke efficacy" mode for skills that have no
authored evals yet, turning D9's binary "has evals?" into a graded "how much does it help?".
Implement as a `--no-reference` judge path in the eval runner.

### Idea 4, Deterministic zero-cost CI gate

A pure-Go function that parses frontmatter, extracts quoted phrases + name tokens, and
asserts the set is non-empty and at least one term appears in the skill's own example
prompts. Fast fail-fast in `hk run pre-push` ahead of any judge-token spend; catches vague
descriptions early.

### Idea 5, Vitals badge

Not a test, a distribution mechanism: a `--badge` output shape in `reporter/` that emits
shields.io JSON for an auditor grade or trigger-F1 as a live README badge.

## Technical considerations

| Concern | Assessment |
|---|---|
| Language boundary | Project is Go-only. skillvitals is Python; port the technique, do not bridge. The logic is trivial (confusion matrix, paired delta). |
| What we already have | Claude-API judge runner (`cmd/eval.go`), Jaccard duplication engine, `reporter/` JSON formatter, frontmatter parsing. |
| What we would write | A `cases.yaml` loader, a description-only router prompt, the confusion-matrix + paired-delta metrics, the deterministic heuristic, optional badge output. |
| Complexity | Small. Ideas 1 and 4 are the natural first slice; Idea 2 overlaps cultivar's differential mode. |

## Recommendation

Adopt **Idea 1 (trigger reliability)** and **Idea 4 (deterministic description lint)** as the
first slice, because they add an axis D1 to D9 do not touch and plug directly into the
duplication engine and pre-push gate. Treat **Idea 2 (counterfactual efficacy)** as shared
scope with the cultivar finding, not a separate build. skillvitals and this project are
complementary: this project is a deep static rubric auditor, skillvitals is a shallow
behavioural validator; the value is in importing the behavioural axis, not the tool.
