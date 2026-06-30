---
title: "ADR-009: Add D10 empirical validation dimension"
status: proposed
date: 2026-06-30
context:
  - path: .context/findings/skilllens-integration-2026-06-30.md
---

**Status:** Proposed
**Date:** 2026-06-30

## Context

The SkillLens research framework (Microsoft, MIT) defines an empirical methodology for validating skill quality: run a model on a benchmark without the skill (baseline), then with the skill injected, and measure the performance delta. The current 9-dimension framework scores skills analytically (rubric-based) but has no empirical validation component. SkillLens's map-reduce pipeline and failure mode taxonomy are a proven reference implementation.

## Decision

Add a D10 "Empirical Validation" dimension that runs a skill through SkillLens-style empirical validation:
1. Select a relevant benchmark matching the skill's domain
2. Run the target model on the benchmark without the skill (baseline)
3. Run the same model with the skill injected into the system prompt
4. Score = performance delta normalized by baseline

Score mapping: ≥+10% improvement → 20/20, +5-10% → 15/20, +1-5% → 10/20, ±1% → 5/20 (no effect), negative → 0/20 (harmful).

A lightweight benchmark harness in `testdata/benchmarks/` using existing fixtures is sufficient for relative scoring — full per-benchmark sandbox infrastructure is not needed.

## Consequences

- Skills gain an empirical quality signal alongside the analytic rubric
- New dependency: an LLM client and benchmark runner
- Benchmark selection is domain-specific — coverage depends on available benchmarks
- Low priority — implement after the native eval runner (ADR-001) is stable
