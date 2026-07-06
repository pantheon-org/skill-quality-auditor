---
title: "Finding: EvoSkill Integration Opportunities"
type: FINDING
status: ACTIVE
date: 2026-07-02
value: LOW
themes:
  - EVAL
---

# Finding: EvoSkill Integration Opportunities

## What is EvoSkill

[EvoSkill](https://github.com/sentient-agi/EvoSkill) is a Python framework from Sentient AGI (arxiv:2603.02766)
that automatically discovers and refines agent skills through an evolutionary loop:

1. Run tasks against a benchmark → collect failures
2. Proposer analyzes failures → proposes skill/prompt changes
3. Generator materializes changes as structured skill folders
4. Evaluator scores the new candidate on held-out data
5. Pareto frontier retains only improvements

It supports Claude Code, OpenCode, Codex CLI, Goose, OpenHands, and Harbor runtimes.

## Current state of this project

**skill-quality-auditor** scores skills against a 9-dimension framework using static analysis
of SKILL.md content, duplication via Jaccard similarity, and an LLM-as-judge native eval
runner (`cmd/eval.go`). Remediation plans are **template-based static advice** — the same
generic suggestions every time (e.g., "Add expert-signal keywords: NEVER, ALWAYS").

The gap: we can tell you *what* is wrong with a skill and suggest *generic* fixes, but we
cannot *discover new skill content*, *test whether a specific fix actually helps*, or
*iterate toward improvement*.

## Integration Point 1: EvoSkill as the remediation engine (highest value)

**Replace `dimensionAdvice` with evolved fixes.**

`reporter/remediation.go:15-25` contains hardcoded advice strings per dimension. EvoSkill
could replace this by:

- Taking the diagnosed gaps (e.g., "D3 Anti-Pattern coverage is 5/15")
- Running the skill's evals (via our `cmd/eval.go`) as the fitness function
- Proposing and testing concrete skill edits across multiple iterations
- Returning the best-evolved candidate

This turns remediation from a static document into an automated improvement loop.

**What we'd need to build:**
- A bridge that reads our scorer diagnostics and translates them into EvoSkill "failure"
  signals (i.e., which D-gaps correspond to which task failures).
- A custom EvoSkill proposer that understands our D1-D9 vocabulary (so it proposes
  changes that close specific dimension gaps rather than generic accuracy).
- EvoSkill's evaluator calling our `cmd/eval.go` runner as the scoring function.

## Integration Point 2: Our 9-dim scorer as EvoSkill's fitness function

EvoSkill currently optimises for task accuracy (exact match, tolerance, or LLM-judge).
Our D1-D9 score provides a **richer fitness signal**: a skill that gets 90% task accuracy
but scores poorly on D3 (anti-patterns) and D6 (freedom calibration) is brittle.

We could:
1. Add a `--skill-quality` mode to EvoSkill that uses `scorer.Score()` as an additional
   fitness dimension alongside task accuracy.
2. Let the Pareto frontier optimise for both: high accuracy + high quality.
3. This prevents EvoSkill from evolving task-optimal but anti-pattern-ridden skills.

**What we'd need to build:**
- A Python wrapper around our Go scorer (subprocess call or cgo).
- A multi-objective Pareto front in EvoSkill (it currently uses scalar accuracy; multi-dim
  would require a small extension to its frontier logic).
- Done: our scorer exposes `scorer.Score()` as a single function call returning
  `Result{Total, Grade, Dimensions, ErrorDetails, WarningDetails}`.

## Integration Point 3: Cross-pollination — knowledge each side brings

| EvoSkill brings | This project brings |
|---|---|
| Automated skill discovery from failures | Structured D1-D9 quality rubric |
| Evolutionary optimisation loop | Static duplication detection (Jaccard) |
| Cross-agent/cross-model transferability | Schema-based eval validation (D9) |
| Pareto frontier selection | Native LLM-as-judge runner (`cmd/eval.go`) |
| Harbor + CSV benchmark integration | Remediation plan schema + CI gates |

The combination means: **discover skills automatically (EvoSkill), qualify them against
a proven framework (us), and iterate until they pass both task accuracy and quality gates
(combined).**

## Technical considerations

| Concern | Assessment |
|---|---|---|
| Language boundary | **Project is Go-only.** No Python runtime. The relevant EvoSkill algorithms must be ported to Go, not bridged. |
| What to port | EvoSkill's core is ~320 commits. The proposer+generator loop is the valuable part — LLM-driven prompt templates that analyze failures and materialise skill edits. Our existing Go code covers the rest: eval runner (replaces evaluator), git ops (replaces frontier), go-tokenizer (already have one). |
| What we already have | `cmd/eval.go` (LLM-as-judge), `scorer/` (D1-D9 quality fitness), git operations in `reporter/store.go`, scenario format under `cmd/assets/evals/`. |
| What we'd need to write | Proposer (LLM prompt templates + response parser for failure analysis), Generator (skill file writer), evolutionary loop controller, CSV/Harbor dataset loader (if we want benchmark-driven evolution). |
| Complexity | Medium-large. The proposer is mostly prompt engineering in Go structs. No deep algorithm — the intelligence is in the LLM calls, not in Python-specific code. |
| Prerequisite | Native eval runner (`cmd/eval.go`) must be stable — it's the evaluation callback for every candidate. The `--compare` (vanilla-vs-skill) mode is a stepping stone. |
| Relevance to existing plans | Supersedes static remediation in `reporter/remediation.go`. The `migrate-off-tessl-eval` draft plan's Phase 2 aligns closely — skills proving they change behaviour is the same signal EvoSkill optimises for. |

## Interface design

The evolutionary loop is gated behind a flag on the **existing `remediate` command**,
not a separate subcommand:

| Mode | Command | Behavior |
|---|---|---|
| Default | `skill-auditor remediate <skill>` | Generate static remediation plan as today. Deterministic, zero LLM cost. Leaves implementation to the current AI session. |
| Evolve | `skill-auditor remediate <skill> --evolve` | Run the full automated loop: propose → generate → evaluate → iterate. Requires API key, expensive, autonomous. Good for CI. |

Both paths produce a remediation — one is a plan for a human/AI to follow interactively,
the other is an automated fix loop. Same command, different execution strategy.

## Recommendation

**Port EvoSkill's core loop to Go as an `--evolve` flag on `cmd/remediate.go`.**

The value is in the proposer → generator → evaluate cycle, not in Python. Our existing Go
infrastructure already covers the evaluator and the quality scorer. The port scope is:

1. **Proposer** — LLM prompt templates that take (task, failure, current skill) and propose
   specific SKILL.md edits or new skill files. This is the key IP.
2. **Generator** — apply proposals as structured skill folders (file I/O, trivial).
3. **Loop controller** — iterate: run eval → collect failures → propose → generate → re-evaluate.
   Use git branches for the Pareto frontier (EvoSkill's approach), or simpler: keep the
   best N candidates in memory and serialise at the end.

Do not port: harness runners (we run the CLI directly), dataset loaders (use our existing
eval scenario format), or any Python-specific plumbing.

Integration Point 2 (scorer as fitness function) becomes part of the evolve mode natively —
no bridge needed since it's all Go.
