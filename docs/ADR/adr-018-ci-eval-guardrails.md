---
title: "ADR-018: CI gating guardrails for native eval runner"
status: proposed
date: 2026-06-30
context:
  - path: .context/findings/eval-gating-byok-2026-06-29.md
  - path: .context/plans/migrate-off-tessl-eval-2026-06-29.md
---

**Status:** Proposed
**Date:** 2026-06-30

## Context

The plan for a native Go eval runner (ADR-001) proposed replacing the Tessl CI step with `skill-auditor eval --fail-below 80`. Review found four unresolved CI-specific concerns documented in the Critical review (Section 11) of the migration plan: flaky single-sample LLM-judge gates, fork PR secret visibility, read-only vs write-summary semantics, and unquantified per-run costs.

## Decision

Adopt the following CI guardrails for the eval runner:

1. **Read-only CI:** The eval runner never writes `summary.json` in CI. `--write-summary` is reserved for local authoring. CI asserts and exits.
2. **Flakiness mitigation:** Use N-sample median per scenario with a margin band rather than a single-sample knife-edge threshold. The gate signal is the median of 3, with a ±5% advisory band.
3. **Fork PR handling:** The LLM-judge step does not run on fork PRs (no secret available). The structural D9 gate (`go test ./scorer/...`, `skill-auditor evaluate --fail-below B`) is the required gate for all PRs including forks.
4. **Cost awareness:** The runner logs estimated API cost per run. The CI cadence defaults to label-triggered (`run-eval` label) or nightly, not every PR, until costs are understood.
5. **Single-turn precondition:** The eval runner is valid only while scenarios remain single-turn reasoning tasks without tool-calling requirements. If future scenarios need tools, the runner approach is re-evaluated (switch to Option C — Agent SDK).

## Consequences

- CI stays green across re-runs (median + band absorbs non-determinism)
- Fork contributors are not blocked by secret availability
- `summary.json` stays as a committed marker, not a CI-generated drift source
- Cost is visible and gated, not unbounded
- The single-turn precondition constrains scenario design
