---
title: "Plan: Port EvoSkill Core Loop as --evolve Mode"
type: plan
status: draft
date: 2026-07-02
related:
  - ../findings/evoskill-integration-2026-07-02.md
  - migrate-off-tessl-eval-2026-06-29.md
---
# Plan: Port EvoSkill Core Loop as --evolve Mode

Status: DRAFT for review
Date: 2026-07-02

## Goal

Replace the static `dimensionAdvice` in `reporter/remediation.go` with an automated
evolutionary improvement loop. When `skill-auditor remediate <skill> --evolve` is run,
the tool iteratively: runs the native eval runner → collects failures → proposes
targeted SKILL.md edits → generates candidate files → re-evaluates. The best-evolved
candidate across N generations is surfaced as the remediation output.

This turns remediation from a static document into an automated improvement loop,
and serves as the improvement engine that Phase 2 of the `migrate-off-tessl-eval` plan
describes ("skills proving they change behaviour").

## Steps

### 1. Proposer — LLM prompt templates for failure analysis

Build a Go package (`internal/evolve/proposer.go`) that:

- Takes as input: (a) the current SKILL.md content, (b) the scorer.Result with
  per-dimension diagnostics, (c) the native eval runner's per-scenario scores
  (structured output from `cmd/eval.go`).
- Constructs an LLM prompt that asks: *"Given these D-gaps and eval failures, what
  specific changes to SKILL.md would close the lowest-scoring dimensions?"*
- Parses the LLM response into structured proposals (target file, edit type:
  insert/delete/replace, content).
- Returns zero or more `Proposal` structs the Generator can apply.

Key design choice: the intelligence is in the prompt, not the code. The proposer
is ~200 lines of Go + prompt templates.

### 2. Generator — apply proposals as structured edits

Build `internal/evolve/generator.go` that:

- Takes a `Proposal` and applies it to the skill directory (file I/O).
- Handles inserts (new section), replacements (rewrite existing section), and
  deletes.
- Either edits in-place on a git branch (EvoSkill's approach) or copies the
  skill directory to a temp workspace.
- Returns the new skill directory path for evaluation.

This is the simplest component — standard file operations, no LLM calls.

### 3. Loop controller — evolutionary iteration

Build `internal/evolve/loop.go` that:

1. Runs the native eval runner (`cmd/eval.go`) on the current candidate → collects
   per-scenario scores + diagnostic summary.
2. Passes results to the Proposer → gets N candidate proposals.
3. For each proposal, runs the Generator → produces a candidate skill.
4. Re-evaluates each candidate → retains the top-K by score (Pareto frontier).
5. If the best candidate improved AND generations < max, loop back to step 2
   using the best candidate as the new base.
6. After max generations or no improvement, serialise the best candidate as
   the remediation output (same file format as today's static plan, but with
   the actual evolved skill files attached or referenced).

Configurable limits:
- `--max-generations` (default: 5)
- `--population-size` (default: 3 proposals per generation)
- `--retain-top` (default: 2 candidates per generation)

Use git branches for the frontier when git is available; fall back to temp
directories when not in a git repo.

### 4. Wire --evolve flag into cmd/remediate.go

`reporter/remediation.go:15-25` (`dimensionAdvice`) stays as-is for the default
(non-evolve) path. Add a flag to `cmd/remediate.go`:

```go
cmd.Flags().Bool("evolve", false, "Run automated evolutionary improvement loop")
```

When `--evolve` is set:
- Validate prerequisites: native eval runner compiled, API key present, skill
  directory is a git repo (or warn).
- Call `evolve.Run(skillDir, opts)` instead of `reporter.Remediation(result)`.
- Output the evolved candidate path and a delta report (before/after scores).

### 5. Verify with existing fixture skills

Run the evolve mode against `testdata/fixtures/skill-minimal` (low-hanging fruit
for improvement) and `testdata/fixtures/skill-full` (ceiling test — should improve
few or no dimensions). Verify:
- The loop terminates within `--max-generations`.
- Each generation produces valid skill files.
- Scores do not regress across generations (never-worse guarantee).

## Related

- [EvoSkill Integration Findings](../findings/evoskill-integration-2026-07-02.md) —
  full analysis of EvoSkill's architecture, integration points, and what to port.
- [migrate-off-tessl-eval](migrate-off-tessl-eval-2026-06-29.md) Phase 2 —
  the vanilla-vs-skill comparison that becomes the improvement signal for this
  loop.

## Open Questions

1. **LLM provider and cost.** Which model drives the Proposer? Claude (matching
   the eval runner) or a cheaper model for iteration? Cost per generation:
   1× eval run (N scenarios × [actor + judge]) + 1× proposer call. Estimate
   `$0.50-1.00` per generation at current Claude Sonnet pricing for 6 scenarios.
   CI cadence for `--evolve` should be manual-trigger or nightly, not every PR.
2. **Proposer prompt stability.** The prompt templates need rapid iteration.
   Should they live as embedded files (like `cmd/assets/`) for hot-reloading,
   or be hardcoded strings in Go? Embedded files are better for iteration but
   add a loading path.
3. **Pareto frontier strategy.** Git branches (EvoSkill's pattern) are clean
   but create clutter. In-memory candidate retention is simpler. Decision:
   start with in-memory, add git branches if we need persistence across
   interrupted runs.
4. **Skill file diversity.** If the Proposer always proposes edits to SKILL.md,
   we miss the chance to add `references/` files or `evals/` scenarios. Should
   proposals be scoped to SKILL.md only (simple first cut) or any file in the
   skill directory?
5. **No-regression guarantee.** How do we prevent the loop from converging on
   a high-score but useless skill (e.g., removing all negative diagnostic
   content)? The D9 scorer (eval validation) acts as a structural guard, but
   we may want a minimum eval scenario count check.
