---
title: "ADR-006: Register dimension scorers in a registry slice for OCP compliance"
status: accepted
date: 2026-06-30
context:
  - path: .context/analysis/code-review-se-principles.md
  - path: .context/plans/se-principles-remediation-plan.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The SE principles code review identified that `scorer.Score()` calls all nine dimension functions explicitly with a hard-coded dispatch chain. Adding a new dimension (D10+) requires editing the orchestrator function. This violates the Open/Closed Principle. The review recommended either a registry slice or an interface-based registration pattern.

## Decision

Replace the hard-coded dispatch chain in `scorer.Score()` with a dynamic registry approach:

```go
type dimensionFn func(content, skillDir string, b *validatorBridge) (int, []Diagnostic)

type dimensionEntry struct {
    Dimension
    fn dimensionFn
}
```

`ScoreFromContent` builds a local registry slice and iterates over it. The registry uses inline closure adapters (not a package-level `var registry`) because D1, D5, D9 require extra parameters captured from the enclosing scope (e.g., `evalsDir`, metadata side-effects). The OCP goal — no hard-coded `if d == "D1"` branches in `Score()` — is met.

## Consequences

- Adding D10 is now a one-file change: define the scorer function and add an entry to the registry
- No modification to `Score()` or `dimensionScores()` needed for new dimensions
- The closure-based approach (not package-level var) handles variable-arity scorers cleanly
- `go test ./scorer/... -race` passes
