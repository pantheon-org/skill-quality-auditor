---
title: "ADR-021: Split reporter package into single-mandate files for SRP compliance"
status: accepted
date: 2026-06-30
context:
  - path: .context/analysis/code-review-se-principles.md
  - path: .context/plans/se-principles-remediation-plan.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The SE principles code review identified the `reporter` package as carrying five distinct concerns in a single package: text formatting, JSON serialisation, aggregation plan generation, remediation plan generation, and trend analysis. Each had a separate reason to change, violating the Single Responsibility Principle.

## Decision

Split the `reporter` package into nine single-mandate files, each opening with a one-line mandate comment:

| File | Single mandate |
| ---- | -------------- |
| `reporter.go` | Result formatting: `scorer.Result` → human-readable text |
| `store.go` | Audit persistence: write to `.context/audits/` |
| `analysis.go` | Pattern analysis persistence: write to `.context/analysis/` |
| `aggregation.go` | Aggregation plan formatting |
| `combined_analysis.go` | `CombinedAnalysis` struct and serialisers |
| `duplication.go` | Duplication report formatting |
| `remediation.go` | Simple remediation: prioritised action plan from `scorer.Result` |
| `remediation_plan_generate.go` | Schema-compliant YAML-frontmatter plan generation |
| `remediation_plan_validate.go` | Plan validation against schema constraints |

The `remediation_plan.go` file (591 lines, two concerns) was split into separate generate and validate files.

## Consequences

- Each file has a single reason to change, improving maintainability
- New output features can be added as new files without modifying existing ones
- Tests can target individual mandates in isolation
- No behaviour change — the public API surface is preserved
