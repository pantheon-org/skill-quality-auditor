---
title: "Finding: olaservo/skilljack-evals is a Partial fit (behavioural harness; skill-lift + anti-trigger kernel worth keeping)"
type: FINDING
status: ACTIVE
date: 2026-07-07
value: MEDIUM
themes:
  - EVAL
  - SKILL-QUALITY
---

# Finding: olaservo/skilljack-evals fit assessment — 2026-07-07

Date: 2026-07-07
Status: DECISION-SUPPORT, not actioned

> [olaservo/skilljack-evals](https://github.com/olaservo/skilljack-evals) was linked with "does this fit?". It is a **behavioural (execution) eval harness** that runs skills through real agent SDKs/CLIs against SkillsBench-style task packages. Our project is a **static, rule-based content scorer** (D1-D9 over `SKILL.md`, no agent in the scoring path). Verdict: **Partial fit** — architecture does not fit, but it contributes the single most valuable transferable concept across the current wave of skill-eval tools.

## What was investigated

README (full), task-package schema, runner table, scoring/metrics section, security model, `evals/` example, and `src/` test layout. Licence MIT, TypeScript, ~5★ but conceptually the richest of the sources reviewed on this date.

## What it actually is

SkillsBench-style task packages run through real agent harnesses (Claude Agent SDK / Claude Code / Codex / OpenCode), scored by deterministic verifiers, with **paired no-skill baselines**. Its distinguishing measurements:

- **Skill Lift** — with-skill resolution rate minus a no-skill baseline on identical prompts. Isolates *what the skill actually adds* over the base model.
- **Skill Invocation Rate** — share of trials that actually discovered/loaded the expected skill; a discoverability signal decoupled from reward.
- **Anti-trigger (false-positive) tasks** — `expect_skill_invocation: false` tasks where loading the skill when it should not fire is a *failure*, catching over-broad triggers.
- Supporting machinery: TDD-for-skills authoring loop, content-addressed caching, `--judge` diagnostics that never gate, blind A/B of two skill versions.

- **Input:** task packages + skill(s). **Output:** resolution rate, pass@k, skill lift, invocation rate. **For:** skill authors (TDD loop) and CI gating.

## Mapping against this project

| Existing capability | Overlap |
| --- | --- |
| D1-D9 scorers (`scorer/`) | None — all its signals are runtime; D1-D9 read the text. |
| `validate` / `analyze` (`cmd/`) | None. |
| duplication engine (`duplication/`) | None. |
| native eval runner (`cmd/eval.go`, D9) | Partial — ours runs scenario tasks with an optional LLM judge against our own tile assets; it has no baseline/lift concept and does not run arbitrary skills through real agent CLIs. |

Vehicle if adopted: the native eval runner (D9). Adopting its ideas means running an agent, which our static scoring path deliberately does not do (model-agnostic, no API calls).

## Verdict

**Partial fit.** The overall architecture (TypeScript, real-agent harness, Docker verifier sandbox, five runners) does not fit our Go static scorer, but two measurement concepts are genuinely novel relative to what D1-D9 can see.

## The salvageable idea (built natively)

The kernel is **measuring what a skill adds, not just how it reads**:

1. **Skill lift (paired baseline).** Run scenarios twice — with the skill mounted and without — and report the delta. Our static D1-D9 cannot see this because it never executes the skill. Native design: an opt-in `--baseline` pass on `cmd/eval.go` that reruns each scenario with skills unmounted and reports lift, kept behind a flag and out of the deterministic static grade (it needs an LLM, is not offline).
2. **Anti-trigger scenarios.** Scenarios where the *correct* behaviour is *not* invoking the skill — the behavioural complement to what D4 (Specification Compliance) can only assess textually. Cheap to add to our eval-scenario schema as an `expect_invocation: false` scenario type.

NEVER port its literals (runner names, env-var conventions, issue numbers, task corpora). Extract the mechanism only. The same lift idea appears, in its cleanest generic form, in the `agent-skills-eval` finding of the same date.

## Recommendation

1. Record and hold. No import.
2. **Skill lift is the one idea to remember** if the project ever adds a behavioural dimension; build it natively into the eval runner rather than adopting this harness. No current gap forces it — this finding puts the concept on record so the source is not re-assessed.
3. **Anti-trigger scenarios** are the cheapest concrete follow-up for a small behavioural signal without a full lift harness; defer until eval-scenario work is next touched.

## Fit assessment (structured record)

<!-- fit-assessment -->
```yaml
schema_version: 1
source:
  name: olaservo/skilljack-evals
  url: https://github.com/olaservo/skilljack-evals
  license: MIT
  language: TypeScript
characterisation: >-
  A CLI that runs skills through real agent SDKs/CLIs against SkillsBench-style
  task packages, scored by deterministic verifiers with paired no-skill
  baselines, reporting resolution rate, pass@k, skill lift, and invocation
  rate. Input: task packages + skills. Output: those metrics. For: skill
  authors (TDD loop) and CI gating.
overlap:
  d1_d9_scorers:
    level: none
    note: All signals are runtime; D1-D9 read the text.
  validate_analyze:
    level: none
    note: Different axis.
  duplication:
    level: none
    note: No similarity detection.
  eval_runner:
    level: partial
    note: Ours judges our tile assets; no baseline/lift concept and no real agent CLIs.
  helper_skills:
    level: none
    note: No agent-workflow equivalent.
verdict: Partial fit
vehicle_if_adopted: go-cli
salvageable:
  present: true
  description: >-
    Skill lift via a paired no-skill baseline (rerun scenarios with the skill
    unmounted, report the delta) and anti-trigger scenarios
    (expect_invocation:false). Build natively as an opt-in --baseline pass on
    cmd/eval.go, behind a flag and outside the deterministic static grade.
recommendation:
  action: record-and-hold
  detail: >-
    No import. Skill lift is the idea to remember if a behavioural dimension is
    ever added; anti-trigger scenarios are the cheapest concrete follow-up.
value: MEDIUM
```
