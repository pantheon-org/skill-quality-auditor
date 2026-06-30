---
title: "ADR-003: Replace global flag vars with per-command option structs"
status: accepted
date: 2026-06-30
context:
  - path: .context/analysis/code-review-se-principles.md
  - path: .context/plans/se-principles-remediation-plan.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The SE principles code review identified package-level mutable globals in
`cmd/` as the most critical issue in the codebase. Every command stores flags
in `var` declarations, preventing parallel tests and requiring fragile
`os.Stdout` piping in test helpers. 13 command files follow this pattern.

## Decision

Replace all global flag variables with per-command option structs that accept
an `io.Writer` for output. Each command gets a factory function:

```go
type evaluateOptions struct {
    asJSON   bool
    store    bool
    repoRoot string
    out      io.Writer
}

func newEvaluateCmd(out io.Writer) *cobra.Command { ... }
```

Root wiring moves to `NewRootCmd(out io.Writer)` as the composition root.

## Consequences

- Tests can run in parallel without race conditions
- Output capture uses `io.Writer` injection instead of `os.Pipe()`
- Standard Cobra pattern matching `kubectl`, `gh`, etc.
- Large refactor touching all 13 command files — coordinate branching
