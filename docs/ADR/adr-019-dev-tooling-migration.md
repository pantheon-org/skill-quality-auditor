---
title: "ADR-019: Migrate dev tooling to hk, markdownlint-cli2, and mise auto-bootstrap hooks"
status: accepted
date: 2026-06-30
context:
  - path: .context/plans/migrate-to-hk-and-markdownlint-cli2-2026-06-29.md
---

**Status:** Accepted
**Date:** 2026-06-30

## Context

The dev tooling stack used lefthook as the git hook runner and `github:swanysimon/mdlint` as the markdown linter. Both had maintenance gaps: lefthook was a separate binary install, and mdlint was less widely maintained than DavidAnson's markdownlint-cli2. The project already used mise (jdx) for tool version management, providing a natural integration path for hk (also by jdx).

## Decision

Replace the dev tooling stack with three changes:

1. **lefthook → hk** as the git hook runner, configured in `hk.pkl` (Pkl config language, evaluated by hk's built-in pklr evaluator — no pkl CLI needed)
2. **mdlint → markdownlint-cli2** (DavidAnson) as the markdown linter, configured in `.markdownlint-cli2.jsonc`
3. **mise auto-bootstrap hooks:** `enter` hook runs `mise install` in activated shells, `postinstall` hook runs `hk install` after every `mise install`. `HK_MISE=1` enables deep integration so hk wraps steps with `mise x` for local/CI parity.

Resolved decisions: versions pinned, same 5 rule disables initially (MD013, MD032, MD051, MD055, MD058), pre-push ordering via `depends` + no-glob on `go-build`, CI keeps explicit steps initially (defer `hk check` as single source of truth).

## Consequences

- Single-toolchain dev environment (mise manages everything, hk uses mise tools)
- `mise install` bootstraps hooks automatically — no manual `hk install`
- DavidAnson's markdownlint-cli2 is better maintained and more widely documented
- Pkl config is a new language for contributors to learn (but hk's built-in evaluator means no separate install)
- The `enter` hook is a convenience for activated shells only; `postinstall` remains the load-bearing bootstrap for CI and non-activated shells
