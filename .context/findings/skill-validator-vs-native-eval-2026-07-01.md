---
title: "Finding: Native eval runner vs skill-validator score evaluate — no overlap"
type: finding
status: done
date: 2026-07-01
value: low
related:
  - ../plans/native-eval-runner-2026-07-01.md
  - ../../docs/ADR/index.yaml
---
# Finding: Native eval runner vs skill-validator `score evaluate` — no overlap

> `skill-validator score evaluate` rates the skill *document* on prose quality (clarity, novelty, etc.). The native eval runner rates the skill's *behavioural effect* on LLM task performance. They are complementary, not duplicative.

## Summary

A source-code review of `github.com/agent-ecosystem/skill-validator` (`judge/judge.go`, `evaluate/evaluate.go`, `judge/client.go`) confirmed that its `score evaluate` subcommand performs a fundamentally different evaluation from this repo's native eval runner (`cmd/eval.go`). There is no meaningful overlap and no opportunity to consolidate.

## Detail

| | `skill-validator score evaluate` | `skill-auditor eval` (native runner) |
|---|---|---|
| **What's evaluated** | The SKILL.md *as prose* (quality of writing) | The LLM's *output* when following the skill (quality of task performance) |
| **LLM calls per scenario** | 1 — judge the document | 2 — actor (produces output) + judge (scores output) |
| **Input format** | Raw SKILL.md + reference files | `evals/scenario-N/` with `task.md`, `criteria.json`, `capability.txt` |
| **Tests behavioural effect?** | No — never asks the LLM to *do* anything | Yes — gives the LLM a concrete task, then judges the result |
| **Caching** | `.score_cache/` per skill dir | `summary.json` per evals/ dir |
| **Provider support** | Anthropic, OpenAI, Claude CLI | Anthropic, OpenAI, Gemini, OpenAI-compatible |
| **Samples/flakiness** | Single shot | N-sample median with margin band |

skill-validator's `score evaluate` sends SKILL.md content to an LLM with a rubric asking "How clear, actionable, and novel is this document?" The native runner instead runs a full actor/judge pipeline: (1) inject skill + task into an LLM, capture output, (2) send skill + task + output + weighted criteria to a judge LLM for scoring.

## Key architectural differences

1. **skill-validator has no scenario model.** It has no equivalent of `task.md`/`criteria.json`/`capability.txt` per scenario directory. It cannot test whether a skill improves task outcomes — only whether the skill text itself is well-written.

2. **The native runner has no document-quality scoring.** It does not evaluate clarity, novelty, or actionability of the SKILL.md itself. That gap is filled by D1–D8 of the static scorer (which uses skill-validator's structural/content analysis via `validatorBridge`).

3. **The only connection point** is `summary.json`: the native runner writes it (via `--write-summary`) and D9 reads it. This is a data handoff, not an overlap.

4. **The libraries are used differently:** skill-validator is imported as a Go library for static `structure.Validate()` and `RunContentAnalysis()` in the scoring pipeline. Its `score evaluate` CLI subcommand is never invoked — only the structural/content analysis functions are imported.

## Recommended Action

No action needed. The native eval runner and skill-validator's `score evaluate` serve different purposes and there is no duplication. The current architecture — using skill-validator for static analysis (via `validatorBridge`) and the native runner for behavioural evals — should be maintained.
