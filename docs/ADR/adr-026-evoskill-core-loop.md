---
title: "ADR-026: Port EvoSkill core loop to Go as --evolve mode"
status: proposed
date: 2026-07-02
context:
  - path: .context/findings/evoskill-integration-2026-07-02.md
  - path: .context/plans/evoskill-core-loop-port-2026-07-02.md
---
**Status:** Proposed
**Date:** 2026-07-02

## Context

The EvoSkill framework (Sentient AGI) automatically discovers and refines agent
skills through an evolutionary loop: propose changes → generate candidate files →
evaluate → retain improvements via Pareto frontier. Meanwhile, this project's
remediation engine (`reporter/remediation.go`) produces static, template-based
advice — the same generic suggestions regardless of the skill's specific gaps.

The gap is that we can tell a skill *what* is wrong and suggest *generic* fixes,
but we cannot discover new skill content, test whether a specific fix actually
helps, or iterate toward improvement.

A detailed integration analysis is in the findings document.

## Decision

Port EvoSkill's core evolutionary loop to Go as an `--evolve` flag on the
existing `skill-auditor remediate` command. The port scope is:

1. **Proposer** — LLM prompt templates that analyze (task, failure, current skill)
   and propose specific SKILL.md edits or new skill files.
2. **Generator** — apply proposals as structured skill folder edits (file I/O).
3. **Loop controller** — iterative: run eval → collect failures → propose →
   generate → re-evaluate, retaining top-K candidates per generation.

Do not port: harness runners (we run the CLI directly), dataset loaders (use
our existing eval scenario format), or Python-specific plumbing.

The native eval runner (`cmd/eval.go`) serves as the evaluation callback, and
the existing D1-D9 scorer provides a multi-objective fitness signal alongside
task accuracy.

## Consequences

**Positive:**
- Remediation evolves from static advice to automated improvement
- Reuses existing infrastructure: eval runner, scorer, git operations
- Same command (`remediate`), different execution strategy (`--evolve` flag)

**Negative:**
- LLM costs per generation (proposer calls + eval runs)
- Proposer prompt templates will need iteration to produce quality proposals
- No-regression guarantee requires careful frontier management

## Alternatives considered

- Keep static remediation only (no change) — loses the automation opportunity
- Bridge to Python EvoSkill via subprocess — rejected due to language boundary
  complexity and the project being Go-only
